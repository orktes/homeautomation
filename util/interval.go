package util

import (
	"fmt"
	"time"
)

// Interval starts an interval and return a function that can be used to stop the interval
func Interval(cb func(), interval time.Duration) func() {
	fmt.Printf("Creating interval %f\n", interval.Minutes())
	closeChannel := make(chan struct{})
	go func() {
		for {
			select {
			case <-time.After(interval):
				cb()
			case <-closeChannel:
				return
			}
		}
	}()

	return func() {
		close(closeChannel)
	}
}
