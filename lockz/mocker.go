package lockz

// UsingMock is a flag indicating whether mock Consul is being used.
const (
	UsingMock      = true
	MockSessionID1 = "00000000-0000-0000-0000-000000000000"
	MockSessionID2 = "00000000-0000-0000-0000-000000000001"
	MockSessionID3 = "00000000-0000-0000-0000-000000000002"
)

// TestConsulIPPort is used for unit testing.
const (
	TestConsulIPPort string = "127.0.0.1:8500"
)

func SetupMock() {}
