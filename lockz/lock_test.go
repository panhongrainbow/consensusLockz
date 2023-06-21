package lockz

import (
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

// Test_Check_Lock tests two lockers, lock and unlock them alternatively.
func Test_Check_Lock(t *testing.T) {
	// Create a new locker with a session TTL of 3 seconds
	var locker0, locker1 Locker
	var err error
	locker0, err = NewLocker(Options{
		SessionTTL: 3 * time.Second,
	})
	require.NoError(t, err)

	// Create a new locker with a session TTL of 3 seconds
	locker1, err = NewLocker(Options{
		SessionTTL: 3 * time.Second,
	})
	require.NoError(t, err)

	// Loop 10 times
	for i := 0; i < 10; i++ {
		// Acquire a lock on locker0
		var acquired bool
		acquired, err = locker0.Lock("lock_test")
		require.NoError(t, err)
		require.True(t, acquired)
		// Unlock locker0
		_, _ = locker0.UnLock("lock_test")

		// Acquire a lock on locker1
		acquired, err = locker1.Lock("lock_test")
		require.NoError(t, err)
		require.True(t, acquired)
		// Unlock locker1
		_, _ = locker1.UnLock("lock_test")
	}
}
