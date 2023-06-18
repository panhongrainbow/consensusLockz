package lockz

import "github.com/hashicorp/consul/api"

const (
	STATUS_LOCK_CHECKED_OPTIONS uint32 = iota + 1
	STATUS_IP_ADDRESS_PORT_ALTERED
	STATUS_LOCK_INITED
	STATUS_BLOCK_ON_RELEASE
	STATUS_LOCK_COMPETITION
	STATUS_LOCK_LOCKING
	STATUS_LOCK_EXTENDED_LIMIT
)

// Locker is the distributed lock entity.
type Locker struct {
	client     *api.Client             // Client for the lock service (single Goroutine Lock protect)
	sessionID  string                  // ID of the session
	sessionTTL string                  // Time-to-live for the session
	release    chan doneAndReleaseLock // Channel for releasing the lock (single Goroutine Lock protect)
	status     uint32                  // The Locker's status
	Opts       Options                 // Options for the lock
}

// doneAndReleaseLock is the signal to send when the work is done to release the lock.
type doneAndReleaseLock struct{}

// CreateClient initializes a locker client.
func (locker *Locker) CreateClient() (err error) {
	// If client is nil, proceed to create one.
	if locker.client == nil {
		// Use default config
		config := api.DefaultConfig()
		if locker.Opts.IpAddressPort != "" {
			config.Address = locker.Opts.IpAddressPort
		}
		// Create client based on config
		locker.client, err = api.NewClient(api.DefaultConfig())
		if err != nil {
			return
		}
	}

	// Return no error if client created or already exists
	return
}

// AlterClient updates options, status.
func (locker *Locker) AlterClient(IpAddressPort string) (err error) {
	// Check if input IpAddressPort is valid.
	err = CheckIpAddressPort(IpAddressPort)
	if err != nil {
		return
	}

	// Update IpAddressPort option
	locker.Opts.IpAddressPort = IpAddressPort

	// Make a mark here, at the right time, switch the client.
	locker.status = STATUS_IP_ADDRESS_PORT_ALTERED

	// Return no error
	return
}
