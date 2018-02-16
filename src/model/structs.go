package model

import (
	"time"
)

type IcmpCall struct {
	IpAddress string
	Success   bool
	Message   string
	Latency   time.Duration
}

type IcmpBatch struct {
	IpAddress    string
	Expertiments int
	PctPcktLoss  float32
	// this makes sense only if % of packet loss is close to 0.0
	// otherwise the data is clustered between timeouts and successful connections
	// with a multi-modal behaviour
	AvgLatency time.Duration
}

type TcpCall struct {
	IpAddress string
	TcpPort   int // TODO how about nil for numbers?
	Success   bool
	Latency   time.Duration
}

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
