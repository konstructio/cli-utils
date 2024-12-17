package stepper

import "testing"

func Test_ErrorInterface(t *testing.T) {
	// Check that AlreadyCompletedError implements the error interface
	var e error = &AlreadyCompletedError{}

	// Learned from the Go standard library that you can create inline
	// interface assertions to bring methods from a struct that lost it
	// when converted to a weak type (interface). See:
	// https://cs.opensource.google/go/go/+/refs/tags/go1.23.4:src/errors/wrap.go;l=17-25
	if _, ok := e.(interface{ Is(err error) bool }); !ok {
		t.Errorf("error does not implement \"Is\" method")
	}
}
