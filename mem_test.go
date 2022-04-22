package mem

import (
	"fmt"
	"testing"
	"time"
)

func initCache() (*Cache, *MemData) {
	c := NewCache()
	Client(c)
	expireOn := time.Now().Add(time.Second * 3).Unix()

	// Use the MemData struct to set the data
	m := &MemData{
		Key:    "key",
		Value:  []byte("value"),
		Expire: expireOn,
	}
	return c, m
}

func setCache(c *Cache, m *MemData) error {
	// Set the new data in the cache
	err := Set(m)
	if err != nil {
		return fmt.Errorf("Error setting data: %s", err)
	}

	if v, ok := Get("key"); !ok || string(v) != "value" {
		return fmt.Errorf("Set failed")
	}
	return nil
}

func TestSet(t *testing.T) {
	c, m := initCache()
	err := setCache(c, m)

	if err != nil {
		t.Error(err)
	}
}

func TestGet(t *testing.T) {
	c, m := initCache()

	// Set the data in the cache
	err := setCache(c, m)
	if err != nil {
		t.Error(err)
	}

	// Get the data from the cache
	v, ok := Get("key")
	if !ok || string(v) != "value" {
		t.Error("Get failed")
	}
	t.Logf("Get: %s, %v", string(v), ok)
}

func TestDelete(t *testing.T) {
	c, m := initCache()

	// Set the data in the cache
	err := setCache(c, m)
	if err != nil {
		t.Error(err)
	}

	// Delete the data from the cache
	Delete("key")

	// Get the data from the cache
	v, ok := Get("key")
	if ok || v != nil {
		t.Error("Delete failed")
	}
	t.Logf("Get: %s, %v", string(v), ok)
}

func TestReplace(t *testing.T) {
	c, m := initCache()

	// Set the data in the cache
	err := setCache(c, m)
	if err != nil {
		t.Error(err)
	}

	// Replace the data in the cache
	m.Value = []byte("new value")
	err = Replace("key", m)
	if err != nil {
		t.Error(err)
	}

	// Get the data from the cache
	v, ok := Get("key")
	if !ok || string(v) != "new value" {
		t.Error("Replace failed")
	}
	t.Logf("Get: %s, %v", string(v), ok)
}

func TestClear(t *testing.T) {
	c, m := initCache()

	// Set the data in the cache
	err := setCache(c, m)
	if err != nil {
		t.Error(err)
	}

	// Clear all the cache
	ClearAll()

	// Get the data from the cache
	v, ok := Get("key")
	if ok || v != nil {
		t.Error("Clear failed")
	}
	t.Logf("Get: %s, %v", string(v), ok)
}
