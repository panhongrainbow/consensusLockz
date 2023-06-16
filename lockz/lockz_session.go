package lockz

import (
	"github.com/hashicorp/consul/api"
	"strconv"
	"time"
)

func (locker *Locker) RenewSession() (err error) {
	_, _, err = locker.Client.Session().Renew(locker.SessionID, nil)
	return
}

func (locker *Locker) DestroySession() (err error) {
	if locker.SessionID != "" {
		_, err = locker.Client.Session().Destroy(locker.SessionID, nil)
		locker.SessionID = ""
		return err
	}
	return
}

func (locker *Locker) SwitchNewLockSession() (err error) {
	err = locker.DestroySession()
	if err != nil {
		return
	}

	// Define the session options
	sessionOpts := &api.SessionEntry{
		Name:      "consensusLockz",
		Behavior:  "delete",
		TTL:       strconv.Itoa(int(locker.Opts.SessionTTLOpt.Seconds())) + "s",
		LockDelay: 15 * time.Second,
	}

	//
	if locker.Opts.SessionTTLOpt == 0 {
		sessionOpts.TTL = DEFAULT_SESSION_TIMEOUT
	} else {
		sessionOpts.TTL = strconv.Itoa(int(locker.Opts.SessionTTLOpt.Seconds())) + "s"
	}

	// Create a new session
	locker.SessionID, _, err = locker.Client.Session().Create(sessionOpts, nil)
	if err != nil {
		return err
	}

	//
	return
}

func (locker *Locker) SwitchWaitLockSession() (err error) {

	err = locker.DestroySession()
	if err != nil {
		return
	}

	// Define the session options
	sessionOpts := &api.SessionEntry{
		Name:      "consensusLockz",
		Behavior:  "delete",
		TTL:       strconv.Itoa(int(locker.Opts.SessionTTLOpt.Seconds())) + "s",
		LockDelay: 15 * time.Second,
	}

	// Create a new session
	locker.SessionID, _, err = locker.Client.Session().Create(sessionOpts, nil)
	if err != nil {
		return err
	}

	//
	return
}
