package lockz

import (
	"encoding/json"
	"github.com/hashicorp/consul/api"
	mockconsul "github.com/panhongrainbow/consul-mock-api"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

// Test_Basic_Mock is just to test Mock, testing if the basic locking operations work normally.
// Whether Lock works normally means being able to write lock detail data to the fake Consul
func Test_Basic_Mock(t *testing.T) {
	// Create a mock Consul client to test against
	consulMock := mockconsul.NewConsul(t)

	// Set filtered headers for cleaner mock requests
	consulMock.SetFilteredHeaders([]string{
		"Accept-Encoding",
		"Content-Length",
		"Content-Type",
		"User-Agent",
	})

	// Marshal sample lock detail data
	lockDetail, err := json.Marshal(struct {
		SessionID  string    `json:"session_id"`
		Extend     int       `json:"extend"`
		UpdateTime time.Time `json:"update_time"`
	}{
		SessionID:  MockSessionID1,
		Extend:     0,
		UpdateTime: time.Now(),
	})
	require.NoError(t, err)

	// Mock a KV put to store the lock detail
	consulMock.KVPut("mock_basic_test", nil, lockDetail, 200).Once()

	// Create a client with the mock Consul
	cfg := api.DefaultConfig()
	cfg.Address = consulMock.URL()
	client, err := api.NewClient(cfg)
	require.NoError(t, err)

	// Try writing lock to mock Consul
	_, err = client.KV().Put(&api.KVPair{Key: "mock_basic_test", Value: lockDetail}, nil)
	require.NoError(t, err)
}
