package task

import (
	"context"

	"dev.azure.com/msazure/One/_git/symphony/gitops/pkg/logger"
	"golang.org/x/sync/errgroup"
)

// Group collects a set of tasks and runs them until they all succeed or one of
// them fails, at which time the rest are shut down by their logger's embedded
// context being cancelled.
type Group struct {
	eg     *errgroup.Group
	cancel func()
}

type Task interface {
	// Run begins processing. This function is expected to block the thread
	// rather than spawning its own goroutine. The goroutine lifecycle will be
	// managed by the caller of Start().
	//
	// If the processing encounters an issue and cannot continue, or is shut
	// down normally, it should return with the error that caused it to quit, if
	// any. This function MUST NOT return until all internal processing has
	// fully completed.
	//
	// The context is guaranteed to live for the lifetime of the parent
	// (typically the lifetime of the process)
	// and is cancelled to signal to the inner function that it should shut
	// down. Implementors of this interface MUST listen for the Done() channel
	// closing on this context and shut down if it triggers. Once all shutdown
	// is complete, the function should return.
	Run(ctx context.Context, log logger.Logger) error
}

// NewGroup creates a group with the provided base logger and the series of
// tasks provided.
func NewGroup(ctx context.Context, log logger.Logger, tasks ...Task) *Group {
	log.Infof("task", "NewGroup", "Starting TaskGroup of %d tasks", len(tasks))

	ctxCancelable, cancel := context.WithCancel(ctx)
	eg, egCtx := errgroup.WithContext(ctxCancelable)
	for _, task := range tasks {
		// reassign so we don't fall into only having the last task
		task := task

		eg.Go(func() error {
			return task.Run(egCtx, log)
		})
	}

	// Wait for one of the tasks in the group to finish and then signal the rest
	// to complete. The errgroup context is cancelled (i.e. Done() is closed) as
	// soon as one of the following things happens:
	//
	//  - One of the tasks returns with a non-nil error
	//  - All of the tasks return with nil errors
	//
	// Cancelling the context we created above allows us to propagate the signal
	// for the error case back into the other tasks so that they know to shut
	// themselves down. It has no effect for the successful case.
	go func() {
		<-egCtx.Done()
		cancel()
	}()

	return &Group{
		eg:     eg,
		cancel: cancel,
	}
}

// Wait waits until all of the tasks have completed. If any of the tasks failed,
// the first task to fail will have its error reported in the return value.
func (g *Group) Wait() error {
	return g.eg.Wait()
}

// Cancel cancels the inner context provided to each of the tasks, initiating a
// shutdown sequence. Once this has been called, consumers can call Wait() to
// wait until all of the tasks have fully completed.
func (g *Group) Cancel() {
	g.cancel()
}
