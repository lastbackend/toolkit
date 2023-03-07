package plugin

import "context"

type Plugin interface {
	Start(ctx context.Context) error
	Stop() error
}
