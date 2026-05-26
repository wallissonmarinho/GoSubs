package services

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/wallissonmarinho/GoSubs/internal/core/domain"
)

type stubProvider struct {
	items []domain.SubtitleCandidate
	err   error
}

func (s stubProvider) Search(ctx context.Context, req domain.SubtitleSearchRequest) ([]domain.SubtitleCandidate, error) {
	return s.items, s.err
}

type stubTranslator struct{}

func (stubTranslator) TranslateBatch(ctx context.Context, sourceLang, targetLang string, items []domain.SubtitleTextChunk) ([]domain.SubtitleTextChunk, error) {
	return items, nil
}

type stubCache struct{}

func (stubCache) Get(key string) ([]byte, bool, error) { return nil, false, nil }
func (stubCache) Put(key string, data []byte) error    { return nil }
func (stubCache) Cleanup() (domain.CleanupResult, error) {
	return domain.CleanupResult{}, nil
}

type stubToken struct{}

func (stubToken) Encode(payload domain.SubtitleTokenPayload) (string, error) {
	return "tok", nil
}
func (stubToken) Decode(token string) (domain.SubtitleTokenPayload, error) {
	return domain.SubtitleTokenPayload{}, nil
}

func TestListShowsGeneratablePTBRWhenOnlyEnglishExists(t *testing.T) {
	svc := NewSubtitleService(stubProvider{
		items: []domain.SubtitleCandidate{
			{ID: "1", URL: "http://x", Language: "en", Display: "English", Format: "srt"},
		},
	}, stubTranslator{}, stubCache{}, stubToken{}, 4000)

	out, err := svc.List(context.Background(), "series", "tmdb:86031:4:7", "", func(path string) string { return path })
	require.NoError(t, err)
	require.Len(t, out.Subtitles, 1)
	require.Equal(t, "🟢 PT-BR (Traduzir)", out.Subtitles[0].Lang)
}

func TestListShowsReadyPortugueseWhenAvailable(t *testing.T) {
	svc := NewSubtitleService(stubProvider{
		items: []domain.SubtitleCandidate{
			{ID: "1", URL: "http://x", Language: "pt", Display: "Português", Format: "srt"},
		},
	}, stubTranslator{}, stubCache{}, stubToken{}, 4000)

	out, err := svc.List(context.Background(), "series", "tmdb:86031:4:7", "", func(path string) string { return path })
	require.NoError(t, err)
	require.Len(t, out.Subtitles, 1)
	require.Equal(t, "Português (Brasil)", out.Subtitles[0].Lang)
}
