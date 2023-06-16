package lockz

import (
	"github.com/hashicorp/consul/api"
	"runtime"
	"sync"
	"time"
)

type Error string

func (e Error) Error() string {
	return string(e)
}

const (
	DEFAULT_SESSION_TIMEOUT = "10s" // Seconds
)

const (
	ERROR_CANNOT_EXTEND   = Error("distributed lock error because the lease cannot be extended")
	ERROR_OCCPUY_BY_OTHER = Error("distributed lock error because the lock was occupied by others")
	ERROR_LOCK_RELEASED   = Error("distributed lock error because the lock was released")
)

type Options struct {
	Address       string
	SessionTTLOpt time.Duration
	RenewPeriod   time.Duration
	ExtendLimit   int
	LockDelay     time.Duration
}

type Locker struct {
	Client     *api.Client
	SessionID  string
	SessionTTL string
	CreateAt   time.Time
	MuLock     sync.Mutex
	Opts       Options
	cancel     chan done
}

type done struct{}

type LockDetail struct {
	SessionID string `json:"session_id"`
	Extend    int    `json:"extend"`
}

func NewLocker(opts Options) (lock Locker, err error) {
	lock.Opts = opts
	err = lock.CreateClient()
	lock.cancel = make(chan done)
	return
}

func (lock *Locker) CreateClient() (err error) {
	if lock.Client == nil {
		config := api.DefaultConfig()
		if lock.Opts.Address != "" {
			config.Address = lock.Opts.Address
		}
		lock.Client, err = api.NewClient(api.DefaultConfig())
		if err != nil {
			return
		}
	}
	return
}

func (lock *Locker) DestroyClient() (err error) {
	if lock.Client != nil {
		lock.Client = nil
		runtime.GC()
	}
	return
}
