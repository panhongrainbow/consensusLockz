package lockz

import (
	"github.com/hashicorp/consul/api"
	"strconv"
	"time"
)

func (lock *Locker) RenewSession(key string) (err error) {
	ticker := time.NewTicker(lock.Opts.RenewPeriod)
	go func() {
		for {
			select {
			case <-ticker.C:
				_, _, err = lock.Client.Session().Renew(lock.SessionID, nil)
				if err != nil {
					return
				}
				err = lock.Incr(key)
				if err != nil {
					return
				}
			case <-lock.cancel:
				err = lock.DestroySession()
				return
			}
		}
	}()

	return
}

func (lock *Locker) CancelSession() (err error) {
	lock.cancel <- done{}
	return
}

func (lock *Locker) DestroySession() (err error) {
	if lock.SessionID != "" {
		_, err = lock.Client.Session().Destroy(lock.SessionID, nil)
		lock.SessionID = ""
		return err
	}
	return
}

func (lock *Locker) SwitchNewLockSession() (err error) {
	err = lock.DestroySession()
	if err != nil {
		return
	}

	// Define the session options
	sessionOpts := &api.SessionEntry{
		Name:      "consensusLockz",
		Behavior:  "delete",
		TTL:       strconv.Itoa(int(lock.Opts.SessionTTLOpt.Seconds())) + "s",
		LockDelay: 15 * time.Second,
	}

	//
	if lock.Opts.SessionTTLOpt == 0 {
		sessionOpts.TTL = DEFAULT_SESSION_TIMEOUT
	} else {
		sessionOpts.TTL = strconv.Itoa(int(lock.Opts.SessionTTLOpt.Seconds())) + "s"
	}

	// Create a new session
	lock.SessionID, _, err = lock.Client.Session().Create(sessionOpts, nil)
	if err != nil {
		return err
	}

	//
	return
}

func (lock *Locker) SwitchWaitLockSession() (err error) {

	err = lock.DestroySession()
	if err != nil {
		return
	}

	// Define the session options
	sessionOpts := &api.SessionEntry{
		Name:      "consensusLockz",
		Behavior:  "delete",
		TTL:       strconv.Itoa(int(lock.Opts.SessionTTLOpt.Seconds())) + "s",
		LockDelay: 15 * time.Second,
	}

	// Create a new session
	lock.SessionID, _, err = lock.Client.Session().Create(sessionOpts, nil)
	if err != nil {
		return err
	}

	//
	return
}
