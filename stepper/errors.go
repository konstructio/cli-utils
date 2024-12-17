package stepper

import "fmt"

// AlreadyCompletedError is returned when a step is completed more than once.
type AlreadyCompletedError struct {
	Name string
}

// Error returns the error message. Implements the error interface.
func (e *AlreadyCompletedError) Error() string {
	return fmt.Sprintf("step %q already completed", e.Name)
}

// Is returns true if the error is an AlreadyCompletedError.
func (e *AlreadyCompletedError) Is(err error) bool {
	_, ok := err.(*AlreadyCompletedError)
	return ok
}
