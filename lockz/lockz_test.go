package lockz

import (
	"fmt"
	"testing"
	"time"
)

/*
	IpAddressPort       string
	SessionTTL time.Duration
	ExtendPeriod   time.Duration
	ExtendLimit   int
	LockDelay     time.Duration
*/

var lockKey = "key1"

func Test_Lockz(t *testing.T) {
	lockOpts := Options{
		SessionTTL:   10 * time.Second,
		ExtendPeriod: 8 * time.Second,
		ExtendLimit:  5,
		LockDelay:    5 * time.Second,
	}

	distributedLocker, _ := NewLocker(lockOpts)
	acquired, err := distributedLocker.Lock(lockKey)
	fmt.Println(acquired)
	fmt.Println(1, err)

	go func() {
		err := distributedLocker.Extend(lockKey)
		fmt.Println(2, err)
	}()

	time.Sleep(30 * time.Second)

	err = distributedLocker.Cancel()
	fmt.Println(3, err)
}
