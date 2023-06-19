package lockz

import (
	"github.com/stretchr/testify/require"
	"testing"
)

// Test_Check_Locker is to confirm when the client function will change,
// such as AlterClient will not be modified until Lock is called to reload.
func Test_Check_Locker(t *testing.T) {
	// Declare variables
	var locker Locker
	var err error
	// Create new locker with empty options
	locker, err = NewLocker(Options{})
	require.NoError(t, err)

	// Save old client
	oldClient := locker.client

	// Alter the client to 127.0.0.1:8501
	err = locker.AlterClient("127.0.0.1:8501")
	require.NoError(t, err)

	// Check that locker client is not nil and reestablish is true
	require.NotNil(t, locker.client)
	require.Equal(t, true, locker.reestablish)
	// (Only when executing the Lock() function will it be re-established this client !)

	// Check that old client is equal to the locker client
	require.Equal(t, oldClient, locker.client)

	// Set locker client to nil and create the new client
	locker.client = nil
	err = locker.CreateClient()
	require.NoError(t, err)

	// Check that old client is not equal to the locker client
	require.NotEqual(t, oldClient, locker.client)

	//
	return
}
