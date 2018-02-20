package moreping

import (
	"time"
)

// IcmpCall models a single ICMP call
type IcmpCall struct {
	IpAddress string
	Success   bool
	Message   string
	Latency   time.Duration
}

// IcmpBatch models a batch of ICMP calls to a given IP address
type IcmpBatch struct {
	IpAddress    string
	Expertiments int
	PctPcktLoss  float32
	// this makes sense only if % of packet loss is close to 0.0
	// otherwise the data is clustered between timeouts and successful connections
	// with a multi-modal behaviour
	AvgLatency time.Duration
}

// TcpCall models a single TCP dial to an IP address and a TCP port
type TcpCall struct {
	IpAddress string
	TcpPort   int // TODO how about nil for numbers?
	Success   bool
	Latency   time.Duration
}

// TcpBatch models a batch of TCP dials to an IP address and a TCP port
type TcpBatch struct {
	IpAddress    string
	TcpPort      int // TODO how about nil for numbers?
	Expertiments int
	PctPcktLoss  float32
	// this makes sense only if % of packet loss is close to 0.0
	// otherwise the data is clustered between timeouts and successful connections
	// with a multi-modal behaviour
	AvgLatency time.Duration
}
