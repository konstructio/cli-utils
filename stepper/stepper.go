package stepper

import (
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/tj/go-spin"
)

// Step represents a single step in a process. It can be marked as completed
// by calling the Complete method. The step will display a spinner until it is
// marked as completed.
type Step struct {
	name         string
	output       io.Writer
	spinner      *spin.Spinner
	chResult     chan error
	chDone       chan struct{}
	done         uint32
	onceComplete sync.Once
}

// New creates a new step with the given name and output writer.
// The progress is written to the output writer.
func New(output io.Writer, name string) *Step {
	sp := spin.New()
	sp.Set("🕐🕑🕒🕓🕔🕕🕖🕗🕘🕙🕚🕛")

	chDone := make(chan struct{})
	chResult := make(chan error)

	go func() {
		defer close(chDone)
		for {
			select {
			case err := <-chResult:
				if err == nil {
					fmt.Fprint(output, "\r", "✅ ", name, "\n")
				} else {
					fmt.Fprint(output, "\r", "🔴 ", name, " - error: ", err.Error(), "\n")
				}
				return
			default:
				fmt.Fprint(output, "\r", sp.Next(), " ", name)
				time.Sleep(100 * time.Millisecond)
			}
		}
	}()

	return &Step{
		name:     name,
		output:   output,
		spinner:  sp,
		chResult: chResult,
		chDone:   chDone,
	}
}

// GetName returns the name of the step.
func (s *Step) GetName() string {
	return s.name
}

// Complete marks the step as completed. If a non-nil error is passed, the
// step is marked as failed.
func (s *Step) Complete(err error) error {
	var returnErr error

	// Ensure the step is only completed once, subsequent calls to Complete
	// will return an AlreadyCompletedError.
	if s.done++; s.done > 1 {
		returnErr = &AlreadyCompletedError{Name: s.name}
	}

	s.onceComplete.Do(func() {
		s.chResult <- err
		close(s.chResult)
		<-s.chDone
		returnErr = err
	})

	return returnErr
}
