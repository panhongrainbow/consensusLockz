package lockz

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"testing"
)

func Test_Check_SessionTTL(t *testing.T) {
	locker, err := NewLocker(Options{})
	fmt.Println(locker.GetSessionTTL())
	require.NoError(t, err)
	t.Run("", func(t *testing.T) {
		// err = locker.SetReloadSessionTTL(20 * time.Second)
		require.NoError(t, err)
	})
	fmt.Println(locker.GetSessionTTL())
}
