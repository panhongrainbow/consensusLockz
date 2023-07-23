package lockz

import (
	"context"
	"github.com/stretchr/testify/require"
	"sync"
	"testing"
	"time"
)

// Test_Check_Extend confirms whether the distributed lock can be renewed properly.
func Test_Check_Extend(t *testing.T) {
	// Create new locker with empty options
	var locker Locker
	var err error
	locker, err = NewLocker(BasicOptions{
		SessionTTL:   3 * time.Second,
		ExtendPeriod: 2 * time.Second,
		ExtendLimit:  3,
	})
	require.NoError(t, err)

	// Acquire the lock
	var acquired bool
	acquired, err = locker.Lock("extend_test")
	require.True(t, acquired)
	require.NoError(t, err)

	// Add 1 goroutine
	wg := sync.WaitGroup{}
	wg.Add(1)

	// Run the goroutine
	// This goroutine will automatically renew the distributed lock.
	go func() {
		err := locker.Extend("extend_test")
		require.NoError(t, err)
	}()

	// Set a timeout error to occur after 10 seconds.
	// This goroutine will check the status of the distributed lock.
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	// Add another goroutine
	go func() {
		for {
			// Just pause !
			time.Sleep(500 * time.Microsecond)
			// Select block
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
				// Keep paying attention to the latest status of the lock
				detail, err := locker.LockStatus("extend_test")
				// Here strictly confirm that the condition of the distributed lock is working properly, reaching ERROR_CANNOT_EXTEND and 3
				if err == ERROR_CANNOT_EXTEND && detail.Extend == 3 {
					wg.Done()
				}
				// Confirm that this lock belongs to this session
				require.Equal(t, locker.sessionID, detail.SessionID)
			}
		}
	}()

	// Wait for the goroutine to finish
	wg.Wait()
}

// Test_Check_Incr lets this test freely and unlimitedly increments the value of the distributed lock
func Test_Check_Incr(t *testing.T) {
	// Create new locker with empty options
	var locker Locker
	var err error
	locker, err = NewLocker(BasicOptions{
		SessionTTL:   10 * time.Second, // (Just set enough time to accumulate the Incr amount !)
		ExtendPeriod: 9 * time.Second,  // (Just set enough time to accumulate the Incr amount !)
		ExtendLimit:  20,               // (Just set enough time to accumulate the Incr amount !)
	})
	require.NoError(t, err)

	// Acquire the lock
	var acquired bool
	acquired, err = locker.Lock("incr_test")
	require.True(t, acquired)
	require.NoError(t, err)

	// Loop to increment the locker's value
	for i := 0; i < 10; i++ {
		err = locker.Incr("incr_test")
		require.Nil(t, err)
	}

	// Get lock status
	detail, err := locker.LockStatus("incr_test")
	require.NoError(t, err)

	// Confirm that this lock belongs to this session
	require.Equal(t, locker.sessionID, detail.SessionID)

	// Confirm the increment amount
	require.Equal(t, 10, detail.Extend)
}

// Test_Check_Cancel is to confirm whether the cancel function can release the distributed lock on time.
func Test_Check_Cancel(t *testing.T) {
	// Create new locker with empty options
	var locker Locker
	var err error
	locker, err = NewLocker(BasicOptions{
		SessionTTL:   10 * time.Second,
		ExtendPeriod: 9 * time.Second,
		ExtendLimit:  1,
	})
	require.NoError(t, err)

	// Acquire the lock
	var acquired bool
	acquired, err = locker.Lock("cancel_test")
	require.True(t, acquired)
	require.NoError(t, err)

	// Add 1 goroutine
	wg := sync.WaitGroup{}
	wg.Add(1)

	// Run the goroutine
	// This goroutine will automatically renew the distributed lock.
	go func() {
		err := locker.Extend("cancel_test")
		require.NoError(t, err)
		wg.Done()
	}()

	// Cancel the lock
	err = locker.Cancel()
	require.NoError(t, err)

	// Wait 1 second
	time.Sleep(1 * time.Second)

	// Get lock status
	_, err = locker.LockStatus("cancel_test")
	require.Equal(t, ERROR_LOCK_RELEASED, err)

	// Wait for the goroutine to finish
	wg.Wait()
}
