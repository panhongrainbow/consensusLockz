package lockz

import (
	"runtime"
	"strconv"
	"sync"
	"time"
)

type Error string

func (e Error) Error() string {
	return string(e)
}

var singleGoroutineLock sync.Mutex

const (
	DEFAULT_SESSION_TIMEOUT = "10s" // Seconds
)

const (
	ERROR_CANNOT_EXTEND        = Error("distributed lock error because the lease cannot be extended")
	ERROR_OCCPUY_BY_OTHER      = Error("distributed lock error because the lock was occupied by others")
	ERROR_LOCK_RELEASED        = Error("distributed lock error because the lock was released")
	ERROR_LOCK_NO_CHANGE       = Error("distributed lock error because no changes in the TTL duration")
	ERROR_STATUS_EXCHANGE_FAIL = Error("")
)

type LockDetail struct {
	SessionID  string    `json:"session_id"`
	Extend     int       `json:"extend"`
	UpdateTime time.Time `json:"update_time"`
}

func NewLocker(opts Options) (locker Locker, err error) {

	// Set Session TTL
	if locker.Opts.SessionTTL == 0 {
		locker.sessionTTL = DEFAULT_SESSION_TIMEOUT
	} else {
		locker.sessionTTL = strconv.Itoa(int(locker.Opts.SessionTTL.Seconds())) + "s"
	}

	// The following code block needs to ensure that it is not accessed by multiple goroutines.
	// The singleGoroutineLock is used for protection.
	// Since the distributed lock follows a two-phase acquisition approach, there should be no need to specifically use a mutex lock.
	singleGoroutineLock.Lock()

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
	singleGoroutineLock.Unlock()

	return
}

func (locker *Locker) SetIpAddressPort(ipAddressPort string) (err error) {

	err = locker.Cancel()

	/*if locker.client == nil {
		config := api.DefaultConfig()
		if locker.Opts.IpAddressPort != "" {
			config.Address = locker.Opts.IpAddressPort
		}
		locker.client, err = api.NewClient(api.DefaultConfig())
		if err != nil {
			return
		}
	}*/
	return
}

func (locker *Locker) GetSessionTTL() (ttl string) {
	ttl = locker.sessionTTL
	return
}

func (locker *Locker) DestroyClient() (err error) {
	if locker.client != nil {
		locker.client = nil
		runtime.GC()
	}
	return
}
