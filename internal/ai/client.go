package ai

import "context"

// Completer returns assistant text (raw model output).
type Completer interface {
	Complete(ctx context.Context, system, user string) (string, error)
}
