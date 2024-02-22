package main

import (
	"context"
	"github.com/chelnak/ysmrr"
	"github.com/mattn/go-colorable"
	"io"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"syscall"
)

type consumerFunc[T any] func(ctx context.Context, queue chan T, s *ysmrr.Spinner)
type producerFunc[T any] func(ctx context.Context, queue chan T)

func getSpinnerWriter() io.Writer {
	if runtime.GOOS == "windows" {
		return colorable.NewColorableStderr()
	} else {
		return os.Stderr
	}
}

func withThreads[T any](producer producerFunc[T], consumer consumerFunc[T], threadCount int) {
	queue := make(chan T, threadCount)
	sm := ysmrr.NewSpinnerManager(ysmrr.WithWriter(getSpinnerWriter()))
	ctx, cancel := context.WithCancel(context.Background())

	var wg sync.WaitGroup

	for i := 1; i <= threadCount; i++ {
		s := sm.AddSpinner("Waiting...")

		go func() {
			defer wg.Done()
			consumer(ctx, queue, s)
		}()
	}

	wg.Add(threadCount)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		<-sig
		cancel()
	}()

	sm.Start()

	go func() {
		producer(ctx, queue)
		close(queue)
	}()

	wg.Wait()
	sm.Stop()
}
