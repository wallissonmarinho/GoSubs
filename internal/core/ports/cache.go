package ports

import "github.com/wallissonmarinho/GoSubs/internal/core/domain"

type SubtitleCache interface {
	Get(key string) ([]byte, bool, error)
	Put(key string, data []byte) error
	Cleanup() (domain.CleanupResult, error)
}
