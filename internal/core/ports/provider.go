package ports

import (
	"context"

	"github.com/wallissonmarinho/GoSubs/internal/core/domain"
)

type SubtitleProvider interface {
	Search(ctx context.Context, req domain.SubtitleSearchRequest) ([]domain.SubtitleCandidate, error)
}
