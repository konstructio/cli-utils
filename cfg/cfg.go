package cfg

import (
	"encoding/gob"
	"errors"
	"fmt"
	"io"
	"os"
	"sync"
)

// Config holds the configuration data and manages concurrent access.
type Config struct {
	mu   sync.RWMutex
	data map[string]interface{}
	file *os.File
	path string
}

func New(path string) (*Config, error) {
	c := &Config{
		path: path,
	}

	if err := c.readFromFile(); err != nil {
		return nil, err
	}

	return c, nil
}

func (c *Config) readFromFile() error {
	file, err := os.OpenFile(c.path, os.O_RDWR|os.O_CREATE, 0o644)
	if err != nil {
		return fmt.Errorf("unable to open config file: %w", err)
	}

	stat, err := file.Stat()
	if err != nil {
		file.Close()
		return fmt.Errorf("unable to stat config file: %w", err)
	}

	var data map[string]interface{}
	if stat.Size() > 0 {
		if err := gob.NewDecoder(file).Decode(&data); err != nil && err != io.EOF {
			file.Close()
			return fmt.Errorf("unable to decode config file as JSON: %w", err)
		}
	} else {
		data = make(map[string]interface{})
	}

	// Move file pointer back to start
	if _, err := file.Seek(0, 0); err != nil {
		file.Close()
		return fmt.Errorf("unable to seek config file: %w", err)
	}

	c.mu.Lock()
	c.data = data
	c.file = file
	c.mu.Unlock()

	return nil
}

func (c *Config) GetString(key string) (string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	val, ok := c.data[key]
	if !ok {
		return "", false
	}
	s, ok := val.(string)
	return s, ok
}

func (c *Config) GetInt(key string) (int, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	val, ok := c.data[key]
	if !ok {
		return 0, false
	}

	var i int
	switch v := val.(type) {
	case float64:
		i = int(v)
	case int:
		i = v
	default:
		return 0, false
	}

	return i, true
}

func (c *Config) GetFloat64(key string) (float64, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	val, ok := c.data[key]
	if !ok {
		return 0.0, false
	}

	v, ok := val.(float64)
	return v, ok
}

func (c *Config) GetBool(key string) (bool, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	val, ok := c.data[key]
	if !ok {
		return false, false
	}

	v, ok := val.(bool)
	return v, ok
}

// Set stores a value and flushes changes immediately to disk.
func (c *Config) Set(key string, value interface{}) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Normalize ints/floats to float64 for consistency
	switch v := value.(type) {
	case int:
		value = float64(v)
	case int32:
		value = float64(v)
	case int64:
		value = float64(v)
	case float32:
		value = float64(v)
	case string, bool, float64:
		// All supported types
	default:
		return fmt.Errorf("unsupported type: %T", value)
	}

	c.data[key] = value
	return c.flushUnsafeLocked()
}

// Finish flushes (if needed) and closes the file.
func (c *Config) Finish() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.file == nil {
		return errors.New("config file already closed")
	}

	if err := c.flushUnsafeLocked(); err != nil {
		return fmt.Errorf("failed to flush config: %w", err)
	}

	if err := c.file.Close(); err != nil {
		return fmt.Errorf("failed to close config file: %w", err)
	}
	c.file = nil

	return nil
}

// flushUnsafeLocked writes the current data to disk.
// Call with c.mu.Lock() held.
func (c *Config) flushUnsafeLocked() error {
	if c.file == nil {
		return errors.New("config file not available")
	}
	if err := c.file.Truncate(0); err != nil {
		return err
	}
	if _, err := c.file.Seek(0, 0); err != nil {
		return err
	}

	if err := gob.NewEncoder(c.file).Encode(c.data); err != nil {
		return err
	}
	return c.file.Sync()
}

func (c *Config) GetAll() (map[string]interface{}, error) {
	if err := c.readFromFile(); err != nil {
		return nil, err
	}

	returned := make(map[string]interface{}, len(c.data))
	c.mu.RLock()
	for k, v := range c.data {
		returned[k] = v
	}
	c.mu.RUnlock()

	return returned, nil
}
