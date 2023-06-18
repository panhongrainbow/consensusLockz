package lockz

import (
	"encoding/json"
	"github.com/hashicorp/consul/api"
)

// FollowLockStatus queries lock status, validating ownership and limits not exceeded
func (locker *Locker) FollowLockStatus(key string) (lockDetail LockDetail, err error) {
	// Get the key-value pair for the key
	var keyPair *api.KVPair
	keyPair, _, err = locker.client.KV().Get(key, nil)
	if err != nil {
		return
	}

	// If no key-value pair is returned, the lock has been released. Return error.
	if keyPair == nil {
		err = ERROR_LOCK_RELEASED
		return
	}

	// Unmarshal the value into a LockDetail struct
	err = json.Unmarshal(keyPair.Value, &lockDetail)
	if err != nil {
		return
	}

	// If the lock has been extended beyond the limit, return error
	if lockDetail.Extend >= locker.Opts.ExtendLimit {
		err = ERROR_CANNOT_EXTEND
	}

	// If the lock session ID does not match, return error
	if lockDetail.SessionID != locker.sessionID {
		err = ERROR_OCCPUY_BY_OTHER
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

func (locker *Locker) Lock(key string) (acquired bool, err error) {
	//
	/*_ = locker.DestroySession()
	if err != nil {
		return
	}*/

	_, err = locker.FollowLockStatus(key)
	switch err {
	case ERROR_OCCPUY_BY_OTHER, ERROR_CANNOT_EXTEND:
		err = locker.BlockOnReleased(key)
		if err == ERROR_LOCK_RELEASED {
			err = locker.NewLockSession()
			if err != nil {
				return
			}
			return locker.ObtainLock(key)
		}
	case ERROR_LOCK_RELEASED:
		err = locker.NewLockSession()
		if err != nil {
			return
		}
		return locker.ObtainLock(key)
	default:
		return
	}

	return
}
