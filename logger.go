package tower

import "context"

type Logger interface {
	Log(ctx context.Context, entry Entry)
	LogError(ctx context.Context, err Error)
}
