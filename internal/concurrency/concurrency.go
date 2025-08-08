package concurrency

import (
	"context"
	"errors"

	"github.com/chelnak/ysmrr"
	"github.com/mattn/go-colorable"
	"golang.org/x/sync/errgroup"
)

type (
	consumerFunc[T any] func(ctx context.Context, queue chan T, s *ysmrr.Spinner) error
	producerFunc[T any] func(ctx context.Context, queue chan T) error
)

func WithThreads[T any](ctx context.Context, producer producerFunc[T], consumer consumerFunc[T], workerCount int) error {
	queue := make(chan T, workerCount-1)
	sm := ysmrr.NewSpinnerManager(ysmrr.WithWriter(colorable.NewColorableStderr()))
	wg, ctx := errgroup.WithContext(ctx)

	for i := 1; i < workerCount; i++ {
		s := sm.AddSpinner("Waiting...")

		wg.Go(func() error {
			err := consumer(ctx, queue, s)

			if err == nil {
				s.CompleteWithMessage("Done.")
			} else if errors.Is(err, context.Canceled) {
				s.ErrorWithMessage("Canceled.")
			} else {
				s.ErrorWithMessage(err.Error())
			}

			return err
		})
	}

	sm.Start()

	defer sm.Stop()

	wg.Go(func() error {
		defer close(queue)

		return producer(ctx, queue)
	})

	err := wg.Wait()

	return err
}
