package lockz

import (
	"github.com/hashicorp/consul/api"
	"time"
)

const (
	DEFAULT_SESSION_TIMEOUT = "10s" // Seconds
)

const (
	STATUS_LOCK_CHECKED_OPTIONS uint32 = iota + 1
	STATUS_LOCK_INITED
	STATUS_BLOCK_ON_RELEASE
	STATUS_LOCK_COMPETITION
	STATUS_LOCK_LOCKING
	STATUS_LOCK_EXTENDED_LIMIT
)

type Error string

func (e Error) Error() string {
	return string(e)
}

// When trying to lock or modify, it may not be possible to proceed due to the following errors, so these errors will be responded.
const (
	ERROR_CANNOT_EXTEND    = Error("Distributed lock error because the lease could not be extended")
	ERROR_OCCUPY_BY_OTHER  = Error("Distributed lock error because the lock was already acquired by another client")
	ERROR_LOCK_RELEASED    = Error("Distributed lock error because the lock was released")
	ERROR_CLIENT_NO_DRIVER = Error("Distributed lock error because the client has no driver configured")

	// It is impossible to have this error, the lock will time out if over time, this does not need to be considered.
	// ERROR_LOCK_NO_CHANGE  = Error("distributed lock error because no changes in the TTL duration")
)

const (
	ERROR_NO_AUTH_DEL = Error("Distributed lock error because there is no permission to delete the key")
)

// Locker is the distributed lock entity.
type Locker struct {
	client      *api.Client             // Client for the lock service (single Goroutine Lock protect)
	reEstablish bool                    // Re-establish the Consul client
	sessionID   string                  // ID of the session
	sessionTTL  string                  // Time-to-live for the session
	release     chan doneAndReleaseLock // Channel for releasing the lock (single Goroutine Lock protect)
	status      uint32                  // The Locker's status
	Opts        LockerOptions           // BasicOptions for the lock
}

// doneAndReleaseLock is the signal to send when the work is done to release the lock.
type doneAndReleaseLock struct{}

// LockDetail needs to be written into the lock key of Consul.
type LockDetail struct {
	SessionID  string    `json:"session_id"`
	Extend     int       `json:"extend"`
	UpdateTime time.Time `json:"update_time"`
}

// NewLocker creates a locker entity.
func NewLocker(opts BasicOptions) (locker Locker, err error) {
	// Reload Session TTL
	err = locker.ReloadSessionTTL()
	if err != nil {
		return
	}

	// Create a consul client
	locker.Opts.Basic = opts
	err = locker.CreateClient()

	// SessionID is only available when the lock is acquired.
	// I want to ensure that when the lock is not acquired, the SessionID immediately becomes empty.
	// (没抢到锁，立刻为空)

	// Set the channel for releasing the lock
	locker.release = make(chan doneAndReleaseLock)

	// Change the status to initialization.
	locker.status = STATUS_LOCK_INITED

	return
}

// CreateClient initializes a locker client.
func (locker *Locker) CreateClient() (err error) {
	// If a client is nil, proceed to create one.
	// The main reason is to maintain client stability and avoid arbitrarily reconstructing.
	// (为了稳定，不随意重建)
	if locker.client == nil {
		//
		switch locker.Opts.Basic.Driver {
		case "consul":
			// Use default config
			config := api.DefaultConfig()
			if locker.Opts.Basic.IpAddressPort != "" {
				config.Address = locker.Opts.Basic.IpAddressPort
			}
			// Create a client based on config
			locker.client, err = api.NewClient(config)
			if err != nil {
				return
			}
		case "mock":
			/*consulMock := mockconsul.NewConsul(t)
			cfg.Address = consulMock.URL()*/
			SetupMock()
		default:
			err = ERROR_CLIENT_NO_DRIVER
			return
		}

	}

	// Return no error if a client created or already exists
	return
}

// AlterClient updates options, status.
func (locker *Locker) AlterClient(IpAddressPort string) (err error) {
	// Check if input IpAddressPort is valid.
	err = CheckIpAddressPort(IpAddressPort)
	if err != nil {
		return
	}

	// Update IpAddressPort option
	locker.Opts.Basic.IpAddressPort = IpAddressPort

	// Make a mark here, at the right time, switch the client.
	locker.reEstablish = true

	// Return no error
	return
}
