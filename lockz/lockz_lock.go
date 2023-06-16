package lockz

import (
	"encoding/json"
	"fmt"
	"github.com/hashicorp/consul/api"
)

func (locker *Locker) Lock(key string) (acquired bool, err error) {
	//
	err = locker.SwitchWaitLockSession()
	if err != nil {
		return
	}
	_, err = locker.GetLockDetailNonblocking(key)
	switch err {
	case ERROR_OCCPUY_BY_OTHER, ERROR_CANNOT_EXTEND:
		err = locker.BlockOnReleased(key)
		if err == ERROR_LOCK_RELEASED {
			err = locker.SwitchNewLockSession()
			if err != nil {
				return
			}
			return locker.ObtainLock(key)
		}
	case ERROR_LOCK_RELEASED:
		err = locker.SwitchNewLockSession()
		if err != nil {
			return
		}
		return locker.ObtainLock(key)
	default:
		return
	}
	return
}

func (locker *Locker) Incr(key string) (err error) {
	var keyPair *api.KVPair
	keyPair, _, err = locker.Client.KV().Get(key, nil)
	if err != nil {
		return
	}

	if keyPair == nil {
		err = ERROR_LOCK_RELEASED
		return
	}

	var keyValue LockDetail
	err = json.Unmarshal(keyPair.Value, &keyValue)

	if locker.SessionID != keyValue.SessionID {
		err = ERROR_OCCPUY_BY_OTHER
		return
	}

	if locker.Opts.ExtendLimit <= keyValue.Extend {
		err = ERROR_CANNOT_EXTEND
		return
	}

	value := LockDetail{
		SessionID: locker.SessionID,
		Extend:    keyValue.Extend + 1,
	}

	b, err := json.Marshal(value)
	if err != nil {
		return
	}

	lockOpts := &api.KVPair{
		Key:     key,
		Value:   b,
		Session: locker.SessionID,
	}

	_, err = locker.Client.KV().Put(lockOpts, nil)
	if err != nil {
		return
	}

	return
}

func (locker *Locker) GetLockDetailNonblocking(key string) (lockDetail LockDetail, err error) {
	var keyPair *api.KVPair
	keyPair, _, err = locker.Client.KV().Get(key, nil)
	if err != nil {
		return
	}

	if keyPair == nil {
		err = ERROR_LOCK_RELEASED
		return
	}

	err = json.Unmarshal(keyPair.Value, &lockDetail)
	if err != nil {
		return
	}

	if lockDetail.Extend >= locker.Opts.ExtendLimit {
		err = ERROR_CANNOT_EXTEND
	}

	if lockDetail.SessionID != locker.SessionID {
		err = ERROR_OCCPUY_BY_OTHER
	}

	return
}

func (locker *Locker) BlockOnReleased(key string) (err error) {
	var keyPair *api.KVPair
	var queryMeta *api.QueryMeta

	// start blocking
	q := &api.QueryOptions{WaitIndex: 0}
	for {
		keyPair, queryMeta, err = locker.Client.KV().Get(key, q)

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

func (locker *Locker) TryLock(key string) (err error) {
	if locker.SessionID != "" {
		_, err = locker.Client.Session().Destroy(locker.SessionID, nil)
		locker.SessionID = ""
		return err
	}
	return
}

func (locker *Locker) ObtainLock(key string) (acquired bool, err error) {
	// err = locker.CheckSession()
	if err != nil {
		return
	}

	value := LockDetail{
		SessionID: locker.SessionID,
		Extend:    0,
	}

	b, err := json.Marshal(value)
	if err != nil {
		return false, err
	}

	// Use the session to acquire a locker
	lockOpts := &api.KVPair{
		Key:     key,
		Value:   b,
		Session: locker.SessionID,
	}

	acquired, _, err = locker.Client.KV().Acquire(lockOpts, nil)
	if err != nil {
		panic(err)
	}

	return
}

func (locker *Locker) UnLock(key string) (acquired bool, err error) {
	var keyPair *api.KVPair
	keyPair, _, err = locker.Client.KV().Get(key, nil)
	if err != nil {
		return
	}

	var keyValue LockDetail
	err = json.Unmarshal(keyPair.Value, &keyValue)

	if locker.SessionID != keyValue.SessionID {
		fmt.Println("delete err")
		return
	}

	_, err = locker.Client.KV().Delete(key, nil)
	if err != nil {
		panic(err)
	}

	return
}
