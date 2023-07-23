package lockz

import (
	"encoding/json"
	"github.com/hashicorp/consul/api"
	"time"
)

// Extend continuously extends a distributed lock identified by a key by renewing the session.
func (locker *Locker) Extend(key string) (err error) {
	// Create a ticker for the extended period
	ticker := time.NewTicker(locker.Opts.Basic.ExtendPeriod)
	// Loop continuously
	for {
		// Select on either the ticker or release channel
		select {
		case <-ticker.C:
			// On ticker, renew the session and extend the lock
			_, _, err = locker.client.Session().Renew(locker.sessionID, nil)
			if err != nil {
				return
			}
			// Increment the value
			err = locker.Incr(key)
			if err != nil {
				return
			}
		case <-locker.release:
			// Complete the work and release the distributed lock
			err = locker.DestroySession()
			return
		}
	}
}

// Cancel is to send a signal in order to release the distributed lock.
func (locker *Locker) Cancel() (err error) {
	locker.release <- doneAndReleaseLock{}
	return
}

// Incr increments the value of a lock identified by a key.
func (locker *Locker) Incr(key string) (err error) {
	// Get the key-value pair from the client
	var keyPair *api.KVPair
	keyPair, _, err = locker.client.KV().Get(key, nil)
	if err != nil {
		return
	}

	// If the key-value pair is nil, return ERROR_LOCK_RELEASED
	if keyPair == nil {
		err = ERROR_LOCK_RELEASED
		return
	}

	// Unmarshal the value to a LockDetail struct
	var keyValue LockDetail
	err = json.Unmarshal(keyPair.Value, &keyValue)

	// If the ExtendLimit is reached, return ERROR_CANNOT_EXTEND
	if locker.Opts.Basic.ExtendLimit <= keyValue.Extend {
		err = ERROR_CANNOT_EXTEND
		return
	}

	// If the session ID does not match, return ERROR_OCCUPY_BY_OTHER
	if locker.sessionID != keyValue.SessionID {
		err = ERROR_OCCUPY_BY_OTHER
		return
	}

	// Increment the Extended field of the LockDetail struct
	value := LockDetail{
		SessionID:  locker.sessionID,
		Extend:     keyValue.Extend + 1,
		UpdateTime: time.Now(),
	}

	// Marshal the struct to JSON bytes
	b, err := json.Marshal(value)
	if err != nil {
		return
	}

	// Assemble the new key-value pair
	lockOpts := &api.KVPair{
		Key:     key,
		Value:   b,
		Session: locker.sessionID,
	}

	// Update the new key-value pair to the Consul
	_, err = locker.client.KV().Put(lockOpts, nil)
	if err != nil {
		return
	}

	// Return no error on success
	return
}
