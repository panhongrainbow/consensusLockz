package lockz

import (
	"github.com/hashicorp/consul/api"
	"sync"
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
	ERROR_CANNOT_EXTEND   = Error("distributed lock error because the lease cannot be extended")
	ERROR_OCCUPY_BY_OTHER = Error("distributed lock error because the lock was occupied by others")
	ERROR_LOCK_RELEASED   = Error("distributed lock error because the lock was released")

	// It is impossible to have this error, the lock will time out if over time, this does not need to be considered.
	// ERROR_LOCK_NO_CHANGE  = Error("distributed lock error because no changes in the TTL duration")
)

const (
	ERROR_NO_AUTH_DEL = Error("distributed lock error because there is no permission to delete the key")
)

// Locker is the distributed lock entity.
type Locker struct {
	client      *api.Client             // Client for the lock service (single Goroutine Lock protect)
	reestablish bool                    // Re-establish the Consul client
	sessionID   string                  // ID of the session
	sessionTTL  string                  // Time-to-live for the session
	release     chan doneAndReleaseLock // Channel for releasing the lock (single Goroutine Lock protect)
	status      uint32                  // The Locker's status
	Opts        Options                 // Options for the lock
}

// doneAndReleaseLock is the signal to send when the work is done to release the lock.
type doneAndReleaseLock struct{}

// LockDetail needs to be written into the lock key of Consul.
type LockDetail struct {
	SessionID  string    `json:"session_id"`
	Extend     int       `json:"extend"`
	UpdateTime time.Time `json:"update_time"`
}

// The distributed lock needs to ensure that it is not accessed by multiple goroutines.
// The singleGoroutineLock is used for protection.
// Since the distributed lock follows a two-phase acquisition approach, there should be no need to specifically use a mutex lock.
var singleGoroutineLock sync.Mutex

// NewLocker creates a locker entity.
func NewLocker(opts Options) (locker Locker, err error) {
	// Reload Session TTL
	err = locker.ReloadSessionTTL()
	if err != nil {
		return
	}

	// The following code block needs to ensure that it is not accessed by multiple goroutines.
	// The singleGoroutineLock is used for protection.
	// Since the distributed lock follows a two-phase acquisition approach, there should be no need to specifically use a mutex lock.
	singleGoroutineLock.Lock() // <----- single goroutine lock

	// Create a consul client
	locker.Opts = opts
	err = locker.CreateClient()

	// SessionID is only available when the lock is acquired.
	// I want to ensure that when the lock is not acquired, the SessionID immediately becomes empty.
	// (没抢到锁，立刻为空)

	// Set the channel for releasing the lock
	locker.release = make(chan doneAndReleaseLock)

	// Change the status to initialization.
	locker.status = STATUS_LOCK_INITED

	// Unlock single goroutine Lock
	singleGoroutineLock.Unlock() // <----- single goroutine unlock

	return
}

// CreateClient initializes a locker client.
func (locker *Locker) CreateClient() (err error) {
	// If a client is nil, proceed to create one.
	if locker.client == nil {
		// Use default config
		config := api.DefaultConfig()
		if locker.Opts.IpAddressPort != "" {
			config.Address = locker.Opts.IpAddressPort
		}
		// Create a client based on config
		locker.client, err = api.NewClient(api.DefaultConfig())
		if err != nil {
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
	locker.Opts.IpAddressPort = IpAddressPort

	// Make a mark here, at the right time, switch the client.
	locker.reestablish = true

	// Return no error
	return
}
