package lockz

import (
	"encoding/json"
	"fmt"
	"github.com/hashicorp/consul/api"
)

func (locker *Locker) Incr(key string) (err error) {
	var keyPair *api.KVPair
	keyPair, _, err = locker.client.KV().Get(key, nil)
	if err != nil {
		return
	}

	if keyPair == nil {
		err = ERROR_LOCK_RELEASED
		return
	}

	var keyValue LockDetail
	err = json.Unmarshal(keyPair.Value, &keyValue)

	if locker.sessionID != keyValue.SessionID {
		err = ERROR_OCCPUY_BY_OTHER
		return
	}

	if locker.Opts.ExtendLimit <= keyValue.Extend {
		err = ERROR_CANNOT_EXTEND
		return
	}

	value := LockDetail{
		SessionID: locker.sessionID,
		Extend:    keyValue.Extend + 1,
	}

	b, err := json.Marshal(value)
	if err != nil {
		return
	}

	lockOpts := &api.KVPair{
		Key:     key,
		Value:   b,
		Session: locker.sessionID,
	}

	_, err = locker.client.KV().Put(lockOpts, nil)
	if err != nil {
		return
	}

	return
}

func (locker *Locker) TryLock(key string) (err error) {
	if locker.sessionID != "" {
		_, err = locker.client.Session().Destroy(locker.sessionID, nil)
		locker.sessionID = ""
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
		SessionID: locker.sessionID,
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
		Session: locker.sessionID,
	}

	acquired, _, err = locker.client.KV().Acquire(lockOpts, nil)
	if err != nil {
		panic(err)
	}

	return
}

func (locker *Locker) UnLock(key string) (acquired bool, err error) {
	var keyPair *api.KVPair
	keyPair, _, err = locker.client.KV().Get(key, nil)
	if err != nil {
		return
	}

	var keyValue LockDetail
	err = json.Unmarshal(keyPair.Value, &keyValue)

	if locker.sessionID != keyValue.SessionID {
		fmt.Println("delete err")
		return
	}

	_, err = locker.client.KV().Delete(key, nil)
	if err != nil {
		panic(err)
	}

	return
}
