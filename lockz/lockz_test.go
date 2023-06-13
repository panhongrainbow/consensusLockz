package lockz

import (
	"fmt"
	"testing"
	"time"
)

func Test_Lockz(t *testing.T) {
	lockOpts := Options{
		TTL:         10 * time.Second,
		ExtendLimit: 5,
		LockDelay:   5 * time.Second,
	}

	distributedLock := NewLock(lockOpts)
	acquired, err := distributedLock.Lock("key1")
	fmt.Println(acquired)
	fmt.Println(err)

	for i := 0; i <= 20; i++ {
		acquired, err = distributedLock.ExtendLease("key1")
		fmt.Println(acquired)
		fmt.Println(err)
		time.Sleep(2 * time.Second)
	}

	// distributedLock.UnLock("key1")
}
