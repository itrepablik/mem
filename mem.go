package mem

import (
	"fmt"
	"sync"
	"time"
)

// MemData is a struct that holds the data for the memory
type MemData struct {
	Key     string // key for the data
	Value   []byte // data to be stored in memory
	Expire  int64  // unix timestamp, 0 means never expire
	Created int64  // unix timestamp, the time the data stored in memory
}

// IsExpired returns true if the data is expired
func (m *MemData) IsExpired() bool {
	if m.Expire == 0 {
		return false
	}
	return m.Expire < time.Now().Unix()
}

// Cache is a struct that holds the data for the cache
type Cache struct {
	data map[string]*MemData // map of the data
	mu   *sync.RWMutex       // read-write mutex, multiple readers, single writer
}

// NewCache returns a new cache
func NewCache() *Cache {
	return &Cache{
		data: make(map[string]*MemData),
		mu:   &sync.RWMutex{}, // read-write mutex, multiple readers, single writer
	}
}

// Set sets the data in the cache
func (c *Cache) Set(m *MemData) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// If key already exists, return error
	if _, ok := c.data[m.Key]; ok {
		return fmt.Errorf("key already exists: %s", m.Key)
	}

	c.data[m.Key] = &MemData{
		Key:     m.Key,
		Value:   m.Value,
		Expire:  m.Expire,
		Created: time.Now().Unix(),
	}
	return nil
}

// Get gets the data from the cache
func (c *Cache) Get(key string) ([]byte, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if data, ok := c.data[key]; ok && !data.IsExpired() {
		return data.Value, true
	}
	return nil, false
}

// Replace replaces the data in the cache with the new data by the key
func (c *Cache) Replace(key string, m *MemData) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Get the data from the cache by the key
	if data, ok := c.data[key]; ok {
		// If the data is not expired, replace the data
		if !data.IsExpired() {
			data.Value = m.Value
			data.Expire = m.Expire
			return nil
		}
	}
	return fmt.Errorf("key not found: %s", key)
}

// Delete deletes the data from the cache
func (c *Cache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.data, key)
}

// ClearAll clears the cache
func (c *Cache) ClearAll() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data = make(map[string]*MemData)
}

// CleanExpired cleans the expired cached data
func CleanExpired(c *Cache) {
	c.mu.Lock()
	defer c.mu.Unlock()

	for k, v := range c.data {
		if v.IsExpired() {
			delete(c.data, k)
		}
	}
}
