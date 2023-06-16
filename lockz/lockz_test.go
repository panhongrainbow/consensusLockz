package lockz

import (
	"fmt"
	"testing"
	"time"
)

/*
	Address       string
	SessionTTLOpt time.Duration
	RenewPeriod   time.Duration
	ExtendLimit   int
	LockDelay     time.Duration
*/

func Test_Lockz(t *testing.T) {
	lockOpts := Options{
		SessionTTLOpt: 10 * time.Second,
		RenewPeriod:   8 * time.Second,
		ExtendLimit:   5,
		LockDelay:     5 * time.Second,
	}

	distributedLocker, _ := NewLocker(lockOpts)
	acquired, err := distributedLocker.Lock("key1")
	fmt.Println(acquired)
	fmt.Println(1, err)

	go func() {
		err := distributedLocker.Extend("key1")
		fmt.Println(2, err)
	}()

	time.Sleep(30 * time.Second)

	err = distributedLocker.Cancel()
	fmt.Println(3, err)
}
