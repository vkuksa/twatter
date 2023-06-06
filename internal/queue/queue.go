package msgqueue

import "context"

type Queue interface {
	Enqueue(ctx context.Context, content string)
}
