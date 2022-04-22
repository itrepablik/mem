package mem

import (
	"testing"
	"time"
)

func TestGetTime(t *testing.T) {
	inputTime := "1s:00"

	startTimeHour, startTimeMinute := GetTime(inputTime)
	if startTimeHour != 0 || startTimeMinute != 0 {
		t.Errorf("GetTime failed")
	}
	t.Logf("GetTime: %d, %d", startTimeHour, startTimeMinute)
}

func TestNewCleaner(t *testing.T) {
	// To create a new cache instance, use this method
	c := NewCache()
	Client(c)

	// Set a new data in the cache with a key and a value and a expiration time
	m := &MemData{
		Key:    "key",
		Value:  []byte("value"),
		Expire: time.Now().Add(time.Second * 3).Unix(),
	}
	err := Set(m)
	if err != nil {
		t.Errorf("Error setting data: %s", err)
	}

	cleaner, err := NewCleaner(FREQUENTLY, WithIntervalValue(EVERY_SECOND, 3))
	if err != nil {
		t.Errorf("Error creating a new cleaner: %s", err)
	}
	cleaner.Run(c)
}
