package lockz

import (
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

// Test_Check_CheckOpts checks the locker options.
func Test_Check_CheckOpts(t *testing.T) {
	tests := []struct {
		description string
		opts        Options
		err         error
	}{
		{
			description: "Valid options",
			opts: Options{
				IpAddressPort: "127.0.0.1:8080",
				SessionTTL:    10,
				ExtendPeriod:  5,
				LockDelay:     1,
				ExtendLimit:   100,
			},
			err: nil,
		},
		{
			description: "Invalid IpAddressPort",
			opts: Options{
				IpAddressPort: "127.0.0.1",
				SessionTTL:    10,
				ExtendPeriod:  5,
				LockDelay:     1,
				ExtendLimit:   100,
			},
			err: ERROR_IPADDRESSPORT_FORMAT,
		},
		{
			description: "Negative SessionTTL",
			opts: Options{
				IpAddressPort: "127.0.0.1:8080",
				SessionTTL:    -10,
				ExtendPeriod:  5,
				LockDelay:     1,
				ExtendLimit:   100,
			},
			err: ERROR_SESSION_TTL_FORMAT,
		},
		{
			description: "Negative ExtendPeriod",
			opts: Options{
				IpAddressPort: "127.0.0.1:8080",
				SessionTTL:    10,
				ExtendPeriod:  -5,
				LockDelay:     1,
				ExtendLimit:   100,
			},
			err: ERROR_EXTENDED_PERIOD_FORMAT,
		},
		{
			description: "Negative LockDelay",
			opts: Options{
				IpAddressPort: "127.0.0.1:8080",
				SessionTTL:    10,
				ExtendPeriod:  5,
				LockDelay:     -1,
				ExtendLimit:   100,
			},
			err: ERROR_LOCK_DELAY_FORMAT,
		},
		{
			description: "ExtendLimit is ignored",
			opts: Options{
				IpAddressPort: "127.0.0.1:8080",
				SessionTTL:    10,
				ExtendPeriod:  5,
				LockDelay:     1,
				ExtendLimit:   -100,
			},
			err: nil,
		},
	}

	for _, test := range tests {
		err := CheckOpts(test.opts)
		require.Equal(t, test.err, err)
	}
}

// Test_Check_CheckIpAddressPort checks IP address and port validation.
func Test_Check_CheckIpAddressPort(t *testing.T) {
	// Test case 1: Valid IP address and port
	ipAddressPort := "192.168.0.1:8080"
	err := CheckIpAddressPort(ipAddressPort)
	require.NoError(t, err)

	// Test case 2: Invalid IP address format
	ipAddressPort = "256.168.0.1:8080"
	err = CheckIpAddressPort(ipAddressPort)
	require.Equal(t, ERROR_IPADDRESSPORT_FORMAT, err)

	// Test case 3: Invalid port number
	ipAddressPort = "192.168.0.1:99999"
	err = CheckIpAddressPort(ipAddressPort)
	require.Equal(t, ERROR_IPADDRESSPORT_FORMAT, err)

	// Test case 4: no ip address
	ipAddressPort = ":99999"
	err = CheckIpAddressPort(ipAddressPort)
	require.Equal(t, ERROR_IPADDRESSPORT_FORMAT, err)

	// Test case 5: no port
	ipAddressPort = "192.168.0.1"
	err = CheckIpAddressPort(ipAddressPort)
	require.Equal(t, ERROR_IPADDRESSPORT_FORMAT, err)

	// Test case 6: invalid delimiter
	ipAddressPort = "192.168.0.1,8080"
	err = CheckIpAddressPort(ipAddressPort)
	require.Equal(t, ERROR_IPADDRESSPORT_FORMAT, err)
}

// Test_Check_CheckDurationFormat checks time duration format.
func Test_Check_CheckDurationFormat(t *testing.T) {
	// Test case 1: Valid positive duration
	duration := time.Duration(10) * time.Second
	err := CheckDurationFormat(duration)
	require.NoError(t, err)

	// Test case 2: Valid zero duration
	duration = time.Duration(0)
	err = CheckDurationFormat(duration)
	require.NoError(t, err)

	// Test case 3: Invalid negative duration
	duration = time.Duration(-5) * time.Second
	err = CheckDurationFormat(duration)
	require.Equal(t, ERROR_NEGATIVE_TIME_DURATION, err)
}
