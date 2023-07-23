package lockz

import (
	"encoding/json"
	"github.com/hashicorp/consul/api"
	"time"
)

// Lock retries until lock obtained or unknown errors return failure.
func (locker *Locker) Lock(key string) (acquired bool, err error) {
	// Destroy any existing session
	_ = locker.DestroySession()

	// If client connection status changed, recreate a client
	if locker.reEstablish == true {
		locker.client = nil
		err = locker.CreateClient()
		if err != nil {
			return
		}
	}

	// Check the lock status
	_, err = locker.LockStatus(key)
	switch err {
	case ERROR_OCCUPY_BY_OTHER, ERROR_CANNOT_EXTEND:
		err = locker.BlockOnReleased(key)
		if err != ERROR_LOCK_RELEASED {
			// If there are unknown errors, just directly return the error!
			return
		}
	case ERROR_LOCK_RELEASED:
		// do not thing!
	default:
		// If there are unknown errors, just directly return the error!
		return
	}

	// Create a new session
	err = locker.NewSession()
	if err != nil {
		return
	}

	// Try to lock
	return locker.TryLock(key)
}

// UnLock releases the distributed locks.
func (locker *Locker) UnLock(key string) (acquired bool, err error) {
	// Get the key-value pair for the lock
	var keyPair *api.KVPair
	keyPair, _, err = locker.client.KV().Get(key, nil)
	if err != nil {
		return
	}

	// Unmarshal the lock value JSON to a LockDetail struct
	var keyValue LockDetail
	err = json.Unmarshal(keyPair.Value, &keyValue)

	// Check the permission to delete the lcok key
	if locker.sessionID != keyValue.SessionID {
		err = ERROR_NO_AUTH_DEL
		return
	}

	// Delete the key-value pair to release the lock
	_, err = locker.client.KV().Delete(key, nil)

	// Return released status and no error on success
	return
}

// LockStatus queries lock status, validating ownership and limits not exceeded util the lock
func (locker *Locker) LockStatus(key string) (lockDetail LockDetail, err error) {
	// Get the key-value pair for the key
	var keyPair *api.KVPair
	keyPair, _, err = locker.client.KV().Get(key, nil)
	if err != nil {
		return
	}

	// If no key-value pair is returned, the lock has been released. Return ERROR_LOCK_RELEASED.
	// (Know it for the first time, hurry up and lock it !)
	if keyPair == nil {
		err = ERROR_LOCK_RELEASED
		return
	}

	// Unmarshal the value into a LockDetail struct
	err = json.Unmarshal(keyPair.Value, &lockDetail)
	if err != nil {
		return
	}

	// If the lock has been extended beyond the limit, return ERROR_CANNOT_EXTEND.
	// (Unlock soon, ready to grab the lock !)
	if lockDetail.Extend >= locker.Opts.Basic.ExtendLimit {
		err = ERROR_CANNOT_EXTEND
	}

	// If the lock session ID does not match, return error
	// (The lock is still occupied, just wait !)
	if lockDetail.SessionID != locker.sessionID {
		err = ERROR_OCCUPY_BY_OTHER
	}

	// Return the LockDetail and no error if validation succeeds
	return
}

// BlockOnReleased queries key repeatedly, blocking until release the distributed lock
func (locker *Locker) BlockOnReleased(key string) (err error) {
	// Declare variables to hold the key-value pair and query metadata
	var keyPair *api.KVPair
	var queryMeta *api.QueryMeta

	// Start blocking
	q := &api.QueryOptions{WaitIndex: 0}

	// Set the status to STATUS_BLOCK_ON_RELEASE
	locker.status = STATUS_BLOCK_ON_RELEASE

	for {
		// Get the key-value pair and query metadata from the key
		keyPair, queryMeta, err = locker.client.KV().Get(key, q)

		// If no key-value pair is returned, the lock has been released. Return ERROR_LOCK_RELEASED.
		if keyPair == nil {
			err = ERROR_LOCK_RELEASED
			return
		}
		// Return any other error
		if err != nil {
			return
		}

		// Update the wait index to the latest for the next query
		q.WaitIndex = queryMeta.LastIndex
	}
}

// TryLock attempts to acquire a lock using a session ID and a key
func (locker *Locker) TryLock(key string) (acquired bool, err error) {
	// Define the LockDetails struct
	value := LockDetail{
		SessionID:  locker.sessionID,
		Extend:     0,
		UpdateTime: time.Now(),
	}

	// Marshal the struct to JSON bytes
	b, err := json.Marshal(value)
	if err != nil {
		return
	}

	// Use the session to acquire a locker
	lockOpts := &api.KVPair{
		Key:     key,
		Value:   b,
		Session: locker.sessionID,
	}

	// Try acquiring the lock using the session
	acquired, _, err = locker.client.KV().Acquire(lockOpts, nil)
	if err != nil {
		return
	}

	// If the lock acquisition fails, delete the session immediately.
	if acquired == false {
		_ = locker.DestroySession()
	}

	// Return acquired status and no error on success
	return
}
