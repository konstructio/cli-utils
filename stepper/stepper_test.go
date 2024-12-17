package stepper

import (
	"bytes"
	"errors"
	"testing"
	"time"
)

// TestNewStep tests the creation of a new step.
func TestNewStep(t *testing.T) {
	var output bytes.Buffer
	step := New(&output, "Test Step")

	if step.GetName() != "Test Step" {
		t.Errorf("expected step name to be 'Test Step', got '%s'", step.GetName())
	}

	if step.done != 0 {
		t.Errorf("expected step done to be 0, got %d", step.done)
	}
}

// TestStepCompleteSuccess tests the successful completion of a step.
func TestStepCompleteSuccess(t *testing.T) {
	var output bytes.Buffer
	step := New(&output, "Test Step")

	err := step.Complete(nil)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if step.done != 1 {
		t.Errorf("expected step done to be 1, got %d", step.done)
	}

	expectedOutput := "\r✅ Test Step\n"
	if output.String() != expectedOutput {
		t.Errorf("expected output '%s', got '%s'", expectedOutput, output.String())
	}
}

// TestStepCompleteFailure tests the failure completion of a step.
func TestStepCompleteFailure(t *testing.T) {
	var output bytes.Buffer
	step := New(&output, "Test Step")

	err := step.Complete(errors.New("test error"))
	if err == nil {
		t.Errorf("expected error, got nil")
	}

	if step.done != 1 {
		t.Errorf("expected step done to be 1, got %d", step.done)
	}

	expectedOutput := "\r🔴 Test Step - error: test error\n"
	if output.String() != expectedOutput {
		t.Errorf("expected output '%s', got '%s'", expectedOutput, output.String())
	}
}

// TestStepCompleteAlreadyCompleted tests the completion of an already completed step.
func TestStepCompleteAlreadyCompleted(t *testing.T) {
	var output bytes.Buffer
	step := New(&output, "Test Step")

	_ = step.Complete(nil)
	err := step.Complete(nil)
	if err == nil {
		t.Errorf("expected AlreadyCompletedError, got nil")
	}

	if _, ok := err.(*AlreadyCompletedError); !ok {
		t.Errorf("expected AlreadyCompletedError, got %v", err)
	}
}

// TestStepSpinner tests the spinner functionality of a step.
func TestStepSpinner(t *testing.T) {
	var output bytes.Buffer
	step := New(&output, "Test Step")

	time.Sleep(10 * time.Millisecond)
	step.Complete(nil)

	if step.done != 1 {
		t.Errorf("expected step done to be 1, got %d", step.done)
	}

	if output.Len() == 0 {
		t.Errorf("expected spinner output, got empty output")
	}
}
