package lockz

import (
	"github.com/hashicorp/consul/api"
	"strconv"
)

// NewSession creates a new session of locker.
func (locker *Locker) NewSession() (err error) {
	// Destroy any existing session
	_ = locker.DestroySession()

	// Define the session options
	sessionOpts := &api.SessionEntry{
		Name:      "consensusLockz",
		Behavior:  "delete",
		TTL:       locker.sessionTTL,
		LockDelay: locker.Opts.LockDelay,
	}

	// Create a new session
	locker.sessionID, _, err = locker.client.Session().Create(sessionOpts, nil)
	if err != nil {
		return err
	}

	// Return no error if session created successfully
	return
}

// DestroySession deletes the associated session and resources.
func (locker *Locker) DestroySession() (err error) {
	if locker.sessionID != "" {
		_, err = locker.client.Session().Destroy(locker.sessionID, nil)
		locker.sessionID = ""
	}
	return
}

// ReloadSessionTTL reloads the time-to-live (TTL) value for a session in a locker.
func (locker *Locker) ReloadSessionTTL() (err error) {
	if locker.Opts.SessionTTL == 0 {
		locker.sessionTTL = DEFAULT_SESSION_TIMEOUT
	} else {
		locker.sessionTTL = strconv.Itoa(int(locker.Opts.SessionTTL.Seconds())) + "s"
	}
	return
}
