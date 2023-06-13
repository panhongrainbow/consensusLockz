package lockz

import (
	"encoding/json"
	"fmt"
	"github.com/hashicorp/consul/api"
	"strconv"
	"sync"
	"time"
)

type Options struct {
	Address     string
	TTL         time.Duration
	ExtendLimit int
	LockDelay   time.Duration
}

type Lock struct {
	Client    *api.Client
	SessionID string
	CreateAt  time.Time
	MuLock    sync.Mutex
	Opts      Options
}

type Value struct {
	SessionID string `json:"session_id"`
	Extend    int    `json:"extend"`
}

func NewLock(opts Options) (lock Lock) {
	lock.Opts = opts

	return
}

func (c *Lock) CheckSession() error {
	// Create a new Consul Client
	var err error
	if c.Client == nil {
		config := api.DefaultConfig()
		if c.Opts.Address != "" {
			config.Address = c.Opts.Address
		}
		c.Client, err = api.NewClient(api.DefaultConfig())
		if err != nil {
			return err
		}
	}
	
	c.MuLock.Lock()
	if (time.Now().Unix()-c.CreateAt.Unix()) > int64(c.CreateAt.Second()) ||
		c.SessionID == "" ||
		c.Opts.TTL.Seconds() == 0 {

		// Define the session options
		sessionOpts := &api.SessionEntry{
			Name:      "consensusLockz",
			Behavior:  "delete",
			TTL:       strconv.Itoa(int(c.Opts.TTL.Seconds())) + "s",
			LockDelay: 15 * time.Second,
		}

		// Create a new session
		c.SessionID, _, err = c.Client.Session().Create(sessionOpts, nil)
		if err != nil {
			return err
		}
	}
	c.MuLock.Unlock()

	return nil
}

func (c *Lock) Lock(key string) (acquired bool, err error) {
	err = c.CheckSession()
	if err != nil {
		return
	}

	value := Value{
		SessionID: c.SessionID,
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
		Session: c.SessionID,
	}

	acquired, _, err = c.Client.KV().Acquire(lockOpts, nil)
	if err != nil {
		panic(err)
	}

	return
}

// https://developer.hashicorp.com/consul/docs/dynamic-app-config/sessions
/*
If the lock is already held by the given session during an acquire,
then the LockIndex is not incremented but the key contents are updated.
This lets the current lock holder update the key contents without having to give up the lock and reacquire it.
*/
func (c *Lock) ExtendLease(key string) (acquired bool, err error) {
	var keyPair *api.KVPair
	keyPair, _, err = c.Client.KV().Get(key, nil)
	if err != nil {
		return
	}

	var keyValue Value
	err = json.Unmarshal(keyPair.Value, &keyValue)

	if c.SessionID != keyValue.SessionID {
		fmt.Println("correct err")
		return
	}

	value := Value{
		SessionID: c.SessionID,
		Extend:    keyValue.Extend + 1,
	}

	b, err := json.Marshal(value)
	if err != nil {
		return false, err
	}

	lockOpts := &api.KVPair{
		Key:     key,
		Value:   b,
		Session: c.SessionID,
	}

	_, err = c.Client.KV().Put(lockOpts, nil)
	if err != nil {
		panic(err)
	}

	return
}

func (c *Lock) UnLock(key string) (acquired bool, err error) {
	var keyPair *api.KVPair
	keyPair, _, err = c.Client.KV().Get(key, nil)
	if err != nil {
		return
	}

	var keyValue Value
	err = json.Unmarshal(keyPair.Value, &keyValue)

	if c.SessionID != keyValue.SessionID {
		fmt.Println("delete err")
		return
	}

	_, err = c.Client.KV().Delete(key, nil)
	if err != nil {
		panic(err)
	}

	return
}
