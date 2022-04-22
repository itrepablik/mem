![mem package](https://user-images.githubusercontent.com/58651329/155666747-72a17ca0-57a9-4021-8cc9-1d37019634fc.png)

# Installation
```go
go get -u github.com/itrepablik/mem
```

# mem
The `mem` stands for memory. It is a simple tool to manage the storage of data in memory. It is useful to store data in memory for a short period of time or to store data in memory for a long period of time until the service is restarted. It's not persistent storage, rather it's simple memory storage suitable for single-machine applications.

# Usage
This is how you can use the mem package in your next Go project.
```go
package main

import (
	"fmt"
	"github.com/itrepablik/mem"
	"os"
	"time"
)

func main() {
	// To create a new cache instance, use this method
	c := mem.NewCache()
	mem.Client(c)

	// Set a new data in the cache with a key and a value and a expiration time
	m := &mem.MemData{
		Key:    "key",
		Value:  []byte("value"),
		Expire: time.Now().Add(time.Second * 3).Unix(),
	}
	err := mem.Set(m)
	if err != nil {
		fmt.Printf("Error setting data: %s", err)
		return
	}

	// Use this method to get the data from the cache by key
	v, ok := mem.Get("key")
	if !ok {
		fmt.Printf("Key not found")
		return
	}
	fmt.Println("Get:", string(v), ok)

	// Replace the data in the cache with a new one using the same key, but with a new value
	m = &mem.MemData{
		Key:    "key",
		Value:  []byte("new value"),
		Expire: time.Now().Add(time.Second * 3).Unix(),
	}

	err = mem.Replace("key", m)
	if err != nil {
		fmt.Printf("Error replacing data: %s", err)
		return
	}

	v, ok = mem.Get("key")
	if !ok {
		fmt.Printf("Key not found")
		return
	}
	fmt.Println("Get cached data after replaced event:", string(v), ok)

	// This is how to delete the data from the cache using the key
	mem.Delete("key")

	v, ok = mem.Get("key")
	if !ok {
		fmt.Printf("Key not found, already deleted!")
		return
	}
	fmt.Println("Get cached data after the delete event:", string(v), ok)

	// Use this method to clear all the data in the cache manually regardless of the expiration time
	mem.ClearAll()

	// To clean all the expired cached data in the cache manually you can use this method
	mem.CleanExpired(c)

	// Use goroutine to start the mem cleaner routine
	// Initialize a new cleaner, e.g mem.WithIntervalValue(mem.EVERY_SECOND, 3)
	// the interval options are: EVERY_SECOND, EVERY_MINUTE, EVERY_HOUR
	// WithStartTime option is not allowed for this mem.FREQUENTLY cleaner type
	go func() {
		cleaner, err := mem.NewCleaner(mem.FREQUENTLY, mem.WithIntervalValue(mem.EVERY_SECOND, 3))
		if err != nil {
			fmt.Printf("Error creating a new cleaner: %s", err)
			return
		}
		cleaner.Run(c)
	}()
	
	// Stop ending the program
	fmt.Fscanln(os.Stdin)
}
```
# Examples to run the cleaner preferrably inside your main.go file.
For the Frequently cleaner example, the following options are available:
The interval options are: EVERY_SECOND, EVERY_MINUTE, EVERY_HOUR
Use the WithStartTime option to set the start time of the cleaner.
```go
	cleaner, err := mem.NewCleaner(mem.FREQUENTLY, mem.WithIntervalValue(mem.EVERY_SECOND, 3))
	if err != nil {
		fmt.Printf("Error creating a new cleaner: %s", err)
		return
	}
	cleaner.Run(c)
```

For the Daily cleaner example, the following day will be the first time the cleaner will run. It requires the WithStartTime option.
```go
	cleaner, err := mem.NewCleaner(mem.DAILY, mem.WithStartTime("10:30"))
	if err != nil {
		fmt.Printf("Error creating a new cleaner: %s", err)
		return
	}
	cleaner.Run(c)
```

Weekly cleaner example, the following week will be the first time the cleaner will run. It requires the WithWeekDay and WithStartTime options.
```go
	cleaner, err := mem.NewCleaner(mem.WEEKLY, mem.WithWeekDay(mem.FRIDAY), mem.WithStartTime("10:30"))
	if err != nil {
		fmt.Printf("Error creating a new cleaner: %s", err)
		return
	}
	cleaner.Run(c)
```

Monthly cleaner example, the following month will be the first time the cleaner will run. It requires the WithDayOfMonth and WithStartTime options.
```go
	cleaner, err := mem.NewCleaner(mem.MONTHLY, mem.WithDayOfMonth(15), mem.WithStartTime("10:30"))
	if err != nil {
		fmt.Printf("Error creating a new cleaner: %s", err)
		return
	}
	cleaner.Run(c)
```

# Subscribe to Maharlikans Code Youtube Channel:
Please consider subscribing to my Youtube Channel to recognize my work on any of my tutorial series. Thank you so much for your support!
https://www.youtube.com/c/MaharlikansCode?sub_confirmation=1

# License
Code is distributed under MIT license, feel free to use it in your proprietary projects as well.
