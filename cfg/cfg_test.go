package cfg

import (
	"io/fs"
	"os"
	"path/filepath"
	"sync"
	"testing"
)

func TestNewConfig(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.gob")

	// Test creation of new config
	c, err := New(path)
	if err != nil {
		t.Fatalf("failed to create new config: %v", err)
	}
	defer c.Finish()

	// The file should now exist.
	if _, err := os.Stat(path); err != nil {
		t.Errorf("expected config file to exist, got error: %v", err)
	}
}

func TestGetAndSetString(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.gob")

	c, err := New(path)
	if err != nil {
		t.Fatalf("failed to create new config: %v", err)
	}
	defer c.Finish()

	// Initially key shouldn't exist
	if val, ok := c.GetString("key"); ok || val != "" {
		t.Errorf("expected no value, got %v, %v", val, ok)
	}

	if err := c.Set("key", "value"); err != nil {
		t.Fatalf("failed to set string: %v", err)
	}

	val, ok := c.GetString("key")
	if !ok || val != "value" {
		t.Errorf("expected value='value', got %v, ok=%v", val, ok)
	}
}

func TestGetAndSetInt(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.gob")

	c, err := New(path)
	if err != nil {
		t.Fatalf("failed to create config: %v", err)
	}
	defer c.Finish()

	if err := c.Set("int_key", 42); err != nil {
		t.Fatalf("failed to set int: %v", err)
	}

	val, ok := c.GetInt("int_key")
	if !ok || val != 42 {
		t.Errorf("expected int=42, got %v, ok=%v", val, ok)
	}
}

func TestGetAndSetFloat64(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.gob")

	c, err := New(path)
	if err != nil {
		t.Fatalf("failed to create config: %v", err)
	}
	defer c.Finish()

	if err := c.Set("float_key", 3.14159); err != nil {
		t.Fatalf("failed to set float: %v", err)
	}

	val, ok := c.GetFloat64("float_key")
	if !ok || val != 3.14159 {
		t.Errorf("expected float=3.14159, got %v, ok=%v", val, ok)
	}
}

func TestGetAndSetBool(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.gob")

	c, err := New(path)
	if err != nil {
		t.Fatalf("failed to create config: %v", err)
	}
	defer c.Finish()

	if err := c.Set("bool_key", true); err != nil {
		t.Fatalf("failed to set bool: %v", err)
	}

	val, ok := c.GetBool("bool_key")
	if !ok || val != true {
		t.Errorf("expected bool=true, got %v, ok=%v", val, ok)
	}
}

func TestUnsupportedType(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.gob")

	c, err := New(path)
	if err != nil {
		t.Fatalf("failed to create config: %v", err)
	}
	defer c.Finish()

	type custom struct{}
	err = c.Set("custom_key", custom{})
	if err == nil {
		t.Errorf("expected error for unsupported type, got nil")
	}
}

func TestDataPersistsAcrossReopen(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.gob")

	// First create and set some values
	c, err := New(path)
	if err != nil {
		t.Fatalf("failed to create config: %v", err)
	}
	if err := c.Set("foo", "bar"); err != nil {
		t.Fatalf("failed to set string: %v", err)
	}
	if err := c.Finish(); err != nil {
		t.Fatalf("failed to finish: %v", err)
	}

	// Reopen and check if "foo" persists
	c2, err := New(path)
	if err != nil {
		t.Fatalf("failed to reopen config: %v", err)
	}
	defer c2.Finish()

	val, ok := c2.GetString("foo")
	if !ok || val != "bar" {
		t.Errorf("expected foo='bar', got %v, ok=%v", val, ok)
	}
}

func TestFinishTwice(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.gob")

	c, err := New(path)
	if err != nil {
		t.Fatalf("failed to create config: %v", err)
	}

	if err := c.Finish(); err != nil {
		t.Fatalf("failed to finish once: %v", err)
	}

	if err := c.Finish(); err == nil {
		t.Error("expected error on second finish, got nil")
	}
}

func TestCorruptedFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.gob")

	// Create a corrupted file
	if err := os.WriteFile(path, []byte("not a valid gob"), fs.ModePerm); err != nil {
		t.Fatalf("failed to write corrupted file: %v", err)
	}

	_, err := New(path)
	if err == nil {
		t.Errorf("expected error for corrupted file, got nil")
	}
}

func TestConcurrentAccess(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.gob")

	c, err := New(path)
	if err != nil {
		t.Fatalf("failed to create config: %v", err)
	}
	defer c.Finish()

	// Set initial value
	if err := c.Set("counter", 0); err != nil {
		t.Fatalf("failed to set initial counter: %v", err)
	}

	var wg sync.WaitGroup
	n := 50
	wg.Add(n)
	for i := 0; i < n; i++ {
		go func() {
			defer wg.Done()
			c.mu.Lock()
			val, ok := c.data["counter"].(float64)
			if !ok {
				// If we fail to read it for some reason, just return
				return
			}
			val++
			c.data["counter"] = val
			c.mu.Unlock()
		}()
	}

	wg.Wait()

	v, ok := c.GetFloat64("counter")
	if !ok {
		t.Fatalf("counter key does not exist")
	}

	if int(v) != n {
		t.Errorf("expected counter=%d after concurrent increments, got %d", n, int(v))
	}
}

func TestFlushAfterSet(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.gob")

	c, err := New(path)
	if err != nil {
		t.Fatalf("failed to create config: %v", err)
	}

	// Set a value and immediately re-open file in another instance to confirm flush
	if err := c.Set("test_key", "test_val"); err != nil {
		t.Fatalf("failed to set value: %v", err)
	}

	c2, err := New(path)
	if err != nil {
		t.Fatalf("failed to create second config: %v", err)
	}
	defer c2.Finish()

	val, ok := c2.GetString("test_key")
	if !ok || val != "test_val" {
		t.Errorf("expected test_key='test_val', got %v, ok=%v", val, ok)
	}

	// Close first config
	if err := c.Finish(); err != nil {
		t.Fatalf("failed to finish first config: %v", err)
	}
}
