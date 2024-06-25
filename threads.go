package main

import (
	"context"
	"io"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"github.com/chelnak/ysmrr"
	"github.com/mattn/go-colorable"
	"golang.org/x/sync/errgroup"
)

type consumerFunc[T any] func(ctx context.Context, queue chan T, s *ysmrr.Spinner) error
type producerFunc[T any] func(ctx context.Context, queue chan T) error

func getSpinnerWriter() io.Writer {
	if runtime.GOOS == "windows" {
		return colorable.NewColorableStderr()
	}

	return os.Stderr
}

func withThreads[T any](producer producerFunc[T], consumer consumerFunc[T], threadCount int) error {
	queue := make(chan T, threadCount)
	sm := ysmrr.NewSpinnerManager(ysmrr.WithWriter(getSpinnerWriter()))
	ctx, cancel := context.WithCancel(context.Background())

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		<-sig
		cancel()
	}()

	wg, ctx := errgroup.WithContext(ctx)

	for i := 1; i <= threadCount; i++ {
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
	wg.Go(func() error {
		defer close(queue)

		return producer(ctx, queue)
	})

	err := wg.Wait()

	sm.Stop()

	return err
}
