package ports

import (
	"context"

	"github.com/wallissonmarinho/GoSubs/internal/core/domain"
)

type SubtitleTranslator interface {
	TranslateBatch(ctx context.Context, sourceLang, targetLang string, items []domain.SubtitleTextChunk) ([]domain.SubtitleTextChunk, error)
}
