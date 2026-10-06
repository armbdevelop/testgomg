package worker

import (
	"context"

	"github.com/armbdevelop/testgomg/internal/core/domain"
)

type ChannelQueue struct {
	tasks chan domain.CalcTask
}

func NewChannelQueue(buffer int) *ChannelQueue {
	return &ChannelQueue{tasks: make(chan domain.CalcTask, buffer)}
}

func (q *ChannelQueue) Enqueue(ctx context.Context, task domain.CalcTask) (err error) {
	if err = ctx.Err(); err != nil {
		return err
	}

	select {
	case q.tasks <- task:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (q *ChannelQueue) Dequeue(ctx context.Context) (task domain.CalcTask, err error) {
	if err = ctx.Err(); err != nil {
		return task, err
	}

	select {
	case task = <-q.tasks:
		return task, nil
	case <-ctx.Done():
		return task, ctx.Err()
	}
}
