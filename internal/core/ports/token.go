package ports

import "github.com/wallissonmarinho/GoSubs/internal/core/domain"

type TokenCodec interface {
	Encode(payload domain.SubtitleTokenPayload) (string, error)
	Decode(token string) (domain.SubtitleTokenPayload, error)
}
