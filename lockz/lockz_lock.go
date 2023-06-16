package lockz

import (
	"encoding/json"
	"fmt"
	"github.com/hashicorp/consul/api"
)

func (lock *Locker) Lock(key string) (acquired bool, err error) {
	//
	err = lock.SwitchWaitLockSession()
	if err != nil {
		return
	}
	_, err = lock.GetLockDetailNonblocking(key)
	switch err {
	case ERROR_OCCPUY_BY_OTHER, ERROR_CANNOT_EXTEND:
		err = lock.BlockOnReleased(key)
		if err == ERROR_LOCK_RELEASED {
			err = lock.SwitchNewLockSession()
			if err != nil {
				return
			}
			return lock.ObtainLock(key)
		}
	case ERROR_LOCK_RELEASED:
		err = lock.SwitchNewLockSession()
		if err != nil {
			return
		}
		return lock.ObtainLock(key)
	default:
		fmt.Println("teen or older")
	}
	return
}

func (lock *Locker) Incr(key string) (err error) {
	var keyPair *api.KVPair
	keyPair, _, err = lock.Client.KV().Get(key, nil)
	if err != nil {
		return
	}

	if keyPair.Value == nil {
		err = ERROR_LOCK_RELEASED
		return
	}

	var keyValue LockDetail
	err = json.Unmarshal(keyPair.Value, &keyValue)

	if lock.SessionID != keyValue.SessionID {
		err = ERROR_OCCPUY_BY_OTHER
		return
	}

	if lock.Opts.ExtendLimit <= keyValue.Extend {
		err = ERROR_CANNOT_EXTEND
		return
	}

	value := LockDetail{
		SessionID: lock.SessionID,
		Extend:    keyValue.Extend + 1,
	}

	b, err := json.Marshal(value)
	if err != nil {
		return
	}

	lockOpts := &api.KVPair{
		Key:     key,
		Value:   b,
		Session: lock.SessionID,
	}

	_, err = lock.Client.KV().Put(lockOpts, nil)
	if err != nil {
		return
	}

	return
}

func (lock *Locker) GetLockDetailNonblocking(key string) (lockDetail LockDetail, err error) {
	var keyPair *api.KVPair
	keyPair, _, err = lock.Client.KV().Get(key, nil)
	if err != nil {
		return
	}

	err = json.Unmarshal(keyPair.Value, &lockDetail)
	if err != nil {
		return
	}

	if keyPair == nil {
		err = ERROR_LOCK_RELEASED
	}

	if lockDetail.Extend >= lock.Opts.ExtendLimit {
		err = ERROR_CANNOT_EXTEND
	}

	if lockDetail.SessionID != lock.SessionID {
		err = ERROR_OCCPUY_BY_OTHER
	}

	return
}

func (lock *Locker) BlockOnReleased(key string) (err error) {
	var keyPair *api.KVPair
	var queryMeta *api.QueryMeta

	// start blocking
	q := &api.QueryOptions{WaitIndex: 0}
	for {
		keyPair, queryMeta, err = lock.Client.KV().Get(key, q)

		if keyPair == nil {
			err = ERROR_LOCK_RELEASED
			return
		}

		if err != nil {
			return
		}

		q.WaitIndex = queryMeta.LastIndex
	}
}

func (lock *Locker) TryLock(key string) (err error) {
	if lock.SessionID != "" {
		_, err = lock.Client.Session().Destroy(lock.SessionID, nil)
		lock.SessionID = ""
		return err
	}
	return
}

func (lock *Locker) ObtainLock(key string) (acquired bool, err error) {
	// err = lock.CheckSession()
	if err != nil {
		return
	}

	value := LockDetail{
		SessionID: lock.SessionID,
		Extend:    0,
	}

	b, err := json.Marshal(value)
	if err != nil {
		return false, err
	}

	// Use the session to acquire a lock
	lockOpts := &api.KVPair{
		Key:     key,
		Value:   b,
		Session: lock.SessionID,
	}

	acquired, _, err = lock.Client.KV().Acquire(lockOpts, nil)
	if err != nil {
		panic(err)
	}

	return
}

func (lock *Locker) UnLock(key string) (acquired bool, err error) {
	var keyPair *api.KVPair
	keyPair, _, err = lock.Client.KV().Get(key, nil)
	if err != nil {
		return
	}

	var keyValue LockDetail
	err = json.Unmarshal(keyPair.Value, &keyValue)

	if lock.SessionID != keyValue.SessionID {
		fmt.Println("delete err")
		return
	}

	_, err = lock.Client.KV().Delete(key, nil)
	if err != nil {
		panic(err)
	}

	return
}
