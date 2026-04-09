package terminal

import (
	"context"
	"fmt"
	"io"
	"time"
)

func RunSpinner(ctx context.Context, w io.Writer, message string) {
	frames := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	tick := time.NewTicker(90 * time.Millisecond)
	defer tick.Stop()
	for i := 0; ; i++ {
		_, _ = fmt.Fprintf(w, "\r\x1b[K%s %s", frames[i%len(frames)], message)
		select {
		case <-ctx.Done():
			_, _ = fmt.Fprint(w, "\r\x1b[K")
			return
		case <-tick.C:
		}
	}
}
