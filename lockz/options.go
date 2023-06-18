package lockz

import (
	"net"
	"strconv"
	"strings"
	"time"
)

const (
	ERROR_IPADDRESSPORT_FORMAT   = Error("lock options error because ip and port string format are not correct")
	ERROR_NEGATIVE_TIME_DURATION = Error("lock options error because ip and the time duration is negative")
	ERROR_SESSION_TTL_FORMAT     = Error("lock options error because ip and the session ttl format is not correct")
	ERROR_EXTENDED_PERIOD_FORMAT = Error("lock options error because ip and the extended period format is not correct")
	ERROR_LOCK_DELAY_FORMAT      = Error("lock options error because ip and the lock delay format is not correct")
)

type Options struct {
	IpAddressPort string        // The address of the lock service, such as a Consul address
	SessionTTL    time.Duration // The lifetime of a session in the lock service
	ExtendPeriod  time.Duration // The period to extend a session before it expires
	LockDelay     time.Duration // Allow temporary interruption time when locking on consul.
	ExtendLimit   int           // The maximum number of times a lock may be extended
}

// >>>>> >>>>> >>>>> >>>>> >>>>>> for IpAddressPort Option

// CheckOpts validates locker options.
func CheckOpts(opts Options) (err error) {
	// Check if IpAddressPort option is valid
	err = CheckIpAddressPort(opts.IpAddressPort)
	if err != nil {
		return
	}
	// Check if SessionTTL option is valid and not negative
	err = CheckDurationFormat(opts.SessionTTL)
	if err == ERROR_NEGATIVE_TIME_DURATION {
		err = ERROR_SESSION_TTL_FORMAT
		return
	}
	if err != nil {
		return
	}
	// Check if ExtendPeriod option is valid and not negative
	err = CheckDurationFormat(opts.ExtendPeriod)
	if err == ERROR_NEGATIVE_TIME_DURATION {
		err = ERROR_EXTENDED_PERIOD_FORMAT
		return
	}
	if err != nil {
		return
	}
	// Check if LockDelay option is valid and not negative
	err = CheckDurationFormat(opts.LockDelay)
	if err == ERROR_NEGATIVE_TIME_DURATION {
		err = ERROR_LOCK_DELAY_FORMAT
		return
	}
	if err != nil {
		return
	}

	// ignore the ExtendLimit option

	return
}

// CheckIpAddressPort validates IP address and port format.
func CheckIpAddressPort(ipAddressPort string) (err error) {
	// The ip address can be empty.
	if ipAddressPort == "" {
		return
	}

	// Split the ipAddressPort string into parts using ":" as the delimiter
	parts := strings.Split(ipAddressPort, ":")
	if len(parts) != 2 {
		err = ERROR_IPADDRESSPORT_FORMAT
		return
	}

	// If the split result does not contain exactly two parts, set an error for invalid IP address port format
	ip := net.ParseIP(parts[0])
	if ip == nil {
		err = ERROR_IPADDRESSPORT_FORMAT
		return
	}

	// If the IP address parsing fails, set an error for invalid IP address format
	port, err := strconv.Atoi(parts[1])
	if err != nil || port < 1 || port > 65535 {
		err = ERROR_IPADDRESSPORT_FORMAT
		return
	}

	// Return nil to indicate no error
	return
}

// CheckDurationFormat validates duration format.
// This checks SessionTTL, ExtendPeriod, and ExtendPeriod.
func CheckDurationFormat(d time.Duration) (err error) {
	if d < 0 {
		err = ERROR_NEGATIVE_TIME_DURATION
	}

	return
}

// No need to check ExtendLimit because when ExtendLimit is 0 or a negative value, the distributed lock cannot be renewed.
