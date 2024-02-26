package secretService

import "context"

type SecretService interface {
	GetSecretValue(ctx context.Context, secretKey string) (<-chan string, <-chan error)
}
