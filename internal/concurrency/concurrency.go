package concurrency

import (
	"context"
	"github.com/chelnak/ysmrr"
	"github.com/mattn/go-colorable"
	"golang.org/x/sync/errgroup"
	"io"
	"os"
	"runtime"
)

type consumerFunc[T any] func(ctx context.Context, queue chan T, s *ysmrr.Spinner) error
type producerFunc[T any] func(ctx context.Context, queue chan T) error

func getSpinnerWriter() io.Writer {
	if runtime.GOOS == "windows" {
		return colorable.NewColorableStderr()
	}

	return os.Stderr
}

func WithThreads[T any](ctx context.Context, producer producerFunc[T], consumer consumerFunc[T], threadCount int) error {
	queue := make(chan T, threadCount-1)
	sm := ysmrr.NewSpinnerManager(ysmrr.WithWriter(getSpinnerWriter()))
	wg, ctx := errgroup.WithContext(ctx)

	for i := 1; i < threadCount; i++ {
		s := sm.AddSpinner("Waiting...")

		wg.Go(func() error {
			err := consumer(ctx, queue, s)

			if err == nil {
				s.CompleteWithMessage("Done.")
			} else {
				s.ErrorWithMessagef("%s", err)
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
