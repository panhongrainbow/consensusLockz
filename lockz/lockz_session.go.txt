package lockz

import (
	"github.com/hashicorp/consul/api"
	"strconv"
	"time"
)

func (locker *Locker) RenewSession() (err error) {
	_, _, err = locker.client.Session().Renew(locker.sessionID, nil)
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
		TTL:       strconv.Itoa(int(locker.Opts.SessionTTL.Seconds())) + "s",
		LockDelay: 15 * time.Second,
	}

	// Create a new session
	locker.sessionID, _, err = locker.client.Session().Create(sessionOpts, nil)
	if err != nil {
		return err
	}

	//
	return
}
