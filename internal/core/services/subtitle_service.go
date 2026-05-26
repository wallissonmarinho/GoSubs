package services

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"

	"github.com/wallissonmarinho/GoSubs/internal/core/domain"
	"github.com/wallissonmarinho/GoSubs/internal/core/ports"
)

const (
	targetLanguage    = "pt-BR"
	readyLabel        = "Português (Brasil)"
	generatableLabel  = "🟢 PT-BR (Traduzir)"
)

type SubtitleService struct {
	provider      ports.SubtitleProvider
	translator    ports.SubtitleTranslator
	cache         ports.SubtitleCache
	tokenCodec    ports.TokenCodec
	maxBatchChars int
}

func NewSubtitleService(provider ports.SubtitleProvider, translator ports.SubtitleTranslator, cache ports.SubtitleCache, tokenCodec ports.TokenCodec, maxBatchChars int) *SubtitleService {
	return &SubtitleService{
		provider:      provider,
		translator:    translator,
		cache:         cache,
		tokenCodec:    tokenCodec,
		maxBatchChars: maxBatchChars,
	}
}

func (s *SubtitleService) Manifest() map[string]any {
	return map[string]any{
		"id":          "org.wallissonmarinho.gosubs",
		"version":     "1.0.0",
		"name":        "GoSubs",
		"description": "Legendas PT-BR sob demanda com traducao via IA.",
		"resources":   []string{"subtitles"},
		"types":       []string{"movie", "series"},
		"idPrefixes":  []string{"tmdb", "tt"},
	}
}

func (s *SubtitleService) List(ctx context.Context, contentType, rawID, extra string, buildURL func(path string) string) (domain.SubtitleListResponse, error) {
	media, err := domain.ParseMediaRef(contentType, rawID, extra)
	if err != nil {
		return domain.SubtitleListResponse{Subtitles: []domain.ListedSubtitle{}}, nil
	}
	req := domain.SubtitleSearchRequest{
		Media:       media,
		ReleaseHint: parseReleaseHint(extra),
	}
	candidates, err := s.provider.Search(ctx, req)
	if err != nil {
		return domain.SubtitleListResponse{}, err
	}
	pt := pickReadyPortuguese(candidates)
	if pt != nil {
		token, err := s.tokenCodec.Encode(toTokenPayload(media, *pt))
		if err != nil {
			return domain.SubtitleListResponse{}, err
		}
		return domain.SubtitleListResponse{
			Subtitles: []domain.ListedSubtitle{{
				ID:   pt.ID,
				Lang: readyLabel,
				URL:  buildURL("/subtitles/source/" + token + ".srt"),
			}},
		}, nil
	}
	base := pickBaseCandidate(candidates)
	if base == nil {
		return domain.SubtitleListResponse{Subtitles: []domain.ListedSubtitle{}}, nil
	}
	payload := toTokenPayload(media, *base)
	key := cacheKey(payload)
	if _, ok, err := s.cache.Get(key); err == nil && ok {
		return domain.SubtitleListResponse{
			Subtitles: []domain.ListedSubtitle{{
				ID:   key,
				Lang: readyLabel,
				URL:  buildURL("/subtitles/cache/" + key + ".srt"),
			}},
		}, nil
	}
	token, err := s.tokenCodec.Encode(payload)
	if err != nil {
		return domain.SubtitleListResponse{}, err
	}
	return domain.SubtitleListResponse{
		Subtitles: []domain.ListedSubtitle{{
			ID:   base.ID,
			Lang: generatableLabel,
			URL:  buildURL("/subtitles/generate/" + token + ".srt"),
		}},
	}, nil
}

func (s *SubtitleService) ServeSource(ctx context.Context, token string, fetcher func(context.Context, domain.SubtitleTokenPayload) ([]byte, error)) ([]byte, error) {
	payload, err := s.tokenCodec.Decode(token)
	if err != nil {
		return nil, err
	}
	return fetcher(ctx, payload)
}

func (s *SubtitleService) Generate(ctx context.Context, token string, fetcher func(context.Context, domain.SubtitleTokenPayload) ([]byte, error)) ([]byte, error) {
	payload, err := s.tokenCodec.Decode(token)
	if err != nil {
		return nil, err
	}
	key := cacheKey(payload)
	if data, ok, err := s.cache.Get(key); err == nil && ok {
		return data, nil
	}
	sourceData, err := fetcher(ctx, payload)
	if err != nil {
		return nil, err
	}
	cues, err := domain.ParseSRT(sourceData)
	if err != nil {
		return nil, err
	}
	translated, err := s.translateCues(ctx, payload.SourceLang, cues)
	if err != nil {
		return nil, err
	}
	out := domain.BuildSRT(translated)
	if err := s.cache.Put(key, out); err != nil {
		return nil, err
	}
	return out, nil
}

func (s *SubtitleService) Cleanup() (domain.CleanupResult, error) {
	return s.cache.Cleanup()
}

func (s *SubtitleService) translateCues(ctx context.Context, sourceLang string, cues []domain.SubtitleCue) ([]domain.SubtitleCue, error) {
	items := make([]domain.SubtitleTextChunk, 0, len(cues))
	for _, cue := range cues {
		items = append(items, domain.SubtitleTextChunk{
			ID:   fmt.Sprintf("%d", cue.Index),
			Text: cue.Text,
		})
	}
	translated := make(map[string]string, len(items))
	for _, batch := range splitIntoBatches(items, s.maxBatchChars) {
		out, err := s.translator.TranslateBatch(ctx, sourceLang, targetLanguage, batch)
		if err != nil {
			return nil, err
		}
		for _, item := range out {
			translated[item.ID] = item.Text
		}
	}
	outCues := make([]domain.SubtitleCue, 0, len(cues))
	for _, cue := range cues {
		text, ok := translated[fmt.Sprintf("%d", cue.Index)]
		if !ok {
			return nil, fmt.Errorf("missing translated cue %d", cue.Index)
		}
		cue.Text = text
		outCues = append(outCues, cue)
	}
	return outCues, nil
}

func splitIntoBatches(items []domain.SubtitleTextChunk, maxChars int) [][]domain.SubtitleTextChunk {
	if maxChars <= 0 {
		maxChars = 4000
	}
	var batches [][]domain.SubtitleTextChunk
	var cur []domain.SubtitleTextChunk
	curChars := 0
	for _, item := range items {
		n := len(item.Text)
		if len(cur) > 0 && curChars+n > maxChars {
			batches = append(batches, cur)
			cur = nil
			curChars = 0
		}
		cur = append(cur, item)
		curChars += n
	}
	if len(cur) > 0 {
		batches = append(batches, cur)
	}
	return batches
}

func parseReleaseHint(extra string) string {
	extra = strings.TrimPrefix(strings.TrimSpace(extra), "/")
	if extra == "" {
		return ""
	}
	extra = strings.TrimSuffix(extra, ".json")
	for _, part := range strings.Split(extra, "&") {
		if strings.HasPrefix(part, "videoName=") {
			return strings.TrimPrefix(part, "videoName=")
		}
		if strings.HasPrefix(part, "filename=") {
			return strings.TrimPrefix(part, "filename=")
		}
	}
	return ""
}

func toTokenPayload(media domain.MediaRef, c domain.SubtitleCandidate) domain.SubtitleTokenPayload {
	return domain.SubtitleTokenPayload{
		Media:      media,
		SourceID:   c.ID,
		SourceURL:  c.URL,
		SourceLang: c.Language,
		Format:     c.Format,
		Encoding:   c.Encoding,
		Release:    c.Release,
		TargetLang: targetLanguage,
	}
}

func cacheKey(p domain.SubtitleTokenPayload) string {
	h := sha256.Sum256([]byte(fmt.Sprintf("%s|%s|%s|%s|%d|%d|%d", p.SourceID, p.SourceURL, p.SourceLang, p.TargetLang, p.Media.TMDBID, p.Media.Season, p.Media.Episode)))
	return hex.EncodeToString(h[:])
}

func pickReadyPortuguese(candidates []domain.SubtitleCandidate) *domain.SubtitleCandidate {
	for _, c := range rankCandidates(candidates) {
		if isPortuguese(c.Language, c.Display) {
			cp := c
			return &cp
		}
	}
	return nil
}

func pickBaseCandidate(candidates []domain.SubtitleCandidate) *domain.SubtitleCandidate {
	for _, c := range rankCandidates(candidates) {
		if c.AI {
			continue
		}
		if isEnglish(c.Language, c.Display) || isSpanish(c.Language, c.Display) {
			cp := c
			return &cp
		}
	}
	for _, c := range rankCandidates(candidates) {
		if c.AI || isPortuguese(c.Language, c.Display) {
			continue
		}
		cp := c
		return &cp
	}
	return nil
}

func rankCandidates(in []domain.SubtitleCandidate) []domain.SubtitleCandidate {
	out := append([]domain.SubtitleCandidate(nil), in...)
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].AI != out[j].AI {
			return !out[i].AI
		}
		if out[i].IsHearingImpaired != out[j].IsHearingImpaired {
			return !out[i].IsHearingImpaired
		}
		return strings.Compare(out[i].Source+out[i].Release, out[j].Source+out[j].Release) < 0
	})
	return out
}

func isPortuguese(lang, display string) bool {
	v := strings.ToLower(strings.TrimSpace(lang + " " + display))
	return strings.Contains(v, "pt") || strings.Contains(v, "portuguese") || strings.Contains(v, "português")
}

func isEnglish(lang, display string) bool {
	v := strings.ToLower(strings.TrimSpace(lang + " " + display))
	return strings.Contains(v, "en") || strings.Contains(v, "english")
}

func isSpanish(lang, display string) bool {
	v := strings.ToLower(strings.TrimSpace(lang + " " + display))
	return strings.Contains(v, "es") || strings.Contains(v, "spanish") || strings.Contains(v, "español")
}
