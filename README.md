![mem package](https://user-images.githubusercontent.com/58651329/155666747-72a17ca0-57a9-4021-8cc9-1d37019634fc.png)

# Installation
```
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
	// Create a new cache instance
	c := mem.NewCache()

	// *********************************************
	// Set a new data in the cache
	// *********************************************
	m := &mem.MemData{
		Key:    "key",
		Value:  []byte("value"),
		Expire: time.Now().Add(time.Second * 3).Unix(),
	}
	err := c.Set(m)
	if err != nil {
		fmt.Printf("Error setting data: %s", err)
		return
	}

	// *********************************************
	// Get the data from the cache
	// *********************************************
	v, ok := c.Get("key")
	if !ok {
		fmt.Printf("Key not found")
		return
	}
	fmt.Println("Get:", string(v), ok)

	// *********************************************
	// Replace the data in the cache
	// *********************************************
	m = &mem.MemData{
		Key:    "key",
		Value:  []byte("new value"),
		Expire: time.Now().Add(time.Second * 3).Unix(),
	}
	err = c.Replace("key", m)
	if err != nil {
		fmt.Printf("Error replacing data: %s", err)
		return
	}

	v, ok = c.Get("key")
	if !ok {
		fmt.Printf("Key not found")
		return
	}
	fmt.Println("Get cached data after replaced event:", string(v), ok)

	// *********************************************
	// Delete the data in the cache
	// *********************************************
	c.Delete("key")

	v, ok = c.Get("key")
	if !ok {
		fmt.Printf("Key not found, already deleted!")
		return
	}
	fmt.Println("Get cached data after the delete event:", string(v), ok)

	// *********************************************
	// ClearAll the data in the cache manually
	// *********************************************
	c.ClearAll()

	// *********************************************
	// CleanExpired clear all expired data in the cache manually
	// *********************************************
	mem.CleanExpired(c)

	// *********************************************
	// Auto clean expired cached data
	// *********************************************
	sched := mem.NewScheduledCleanCached(0, 3)
	mem.RunAutoCleanExpiredCached(c, sched)

	// Stop ending the program
	fmt.Fscanln(os.Stdin)
}
```

# Subscribe to Maharlikans Code Youtube Channel:
Please consider subscribing to my Youtube Channel to recognize my work on any of my tutorial series. Thank you so much for your support!
https://www.youtube.com/c/MaharlikansCode?sub_confirmation=1

# License
Code is distributed under MIT license, feel free to use it in your proprietary projects as well.
