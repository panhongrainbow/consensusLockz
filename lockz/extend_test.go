package lockz

import (
	"context"
	"github.com/stretchr/testify/require"
	"sync"
	"testing"
	"time"
)

func Test_Check_Extend(t *testing.T) {
	// Create new locker with empty options
	var locker Locker
	var err error
	locker, err = NewLocker(Options{
		SessionTTL:   3 * time.Second,
		ExtendPeriod: 2 * time.Second,
		ExtendLimit:  3,
	})
	require.NoError(t, err)

	//
	var acquired bool
	acquired, err = locker.Lock("extend_test")
	require.True(t, acquired)
	require.NoError(t, err)

	//
	wg := sync.WaitGroup{}
	wg.Add(1)

	//
	go func() {
		err := locker.Extend("extend_test")
		require.NoError(t, err)
	}()

	// Set a timeout error to occur after 10 seconds.
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	//
	go func() {
		for {
			//
			select {
			case <-ctx.Done():
				// Here is an explanation:
				// The TTL is set to 3 seconds each time, delayed every 2 seconds, and delayed up to 3 times,
				// so it can be delayed for a total of 2 * 3 = 6 seconds to 2 * 2 + 3 = 7 seconds.
				// Therefore, even with some time deviation, it will not time out for more than 10 seconds.
				// If it times out for 10 seconds, an error must be occurred.
				// (TTL为3秒，每2秒延时一次，最多延时3次，所以一共可以延时 2*3 = 6秒至2*2+3 = 7秒，超时10秒是错误)
				panic("timeout happens!")
			default:
				detail, err := locker.LockStatus("extend_test")
				if err == ERROR_CANNOT_EXTEND && detail.Extend == 3 {
					wg.Done()
				}
				require.Equal(t, locker.sessionID, detail.SessionID)
			}

		}
	}()

	wg.Wait()
}
