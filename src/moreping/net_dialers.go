// Package moreping contains the main components dealing with ICMP calls (ping),
// TCP dials, and scheduling.
package moreping

import (
	"fmt"
	"net"
	"time"

	"github.com/tatsushid/go-fastping"
)

// InfiniteLatency is the value used to simulate a timeout.
const InfiniteLatency = 9999 * time.Millisecond

// TCPPinger provides the functionality to dial IP addresses on a given TCP port.
type TCPPinger interface {
	DialBatchIP(targetIP string, targetPort int, batchSize int) TcpBatch
	AsyncTCPDialBatchesForIP(targetIP string, batchSize int)

	DialIP(targetIP string, targetPort int) TcpCall
	AsyncTCPDialsForIP(targetIP string)

	SpawnTCPDials(siteNetDetails []string)
	SpawnTCPDialBatches(siteNetDetails []string, batchSize int)
}

type tcpPinger struct {
	ports        []int
	timeout      time.Duration
	msgChan      chan TcpCall
	msgBatchChan chan TcpBatch
	batchSize    int
}

// NewTCPPinger creates a new instance of the TCP pinger
func NewTCPPinger(tcpPorts []int, tcpTimeout time.Duration, tcpChan chan TcpCall) TCPPinger {
	Logger.Printf("The TCP pinger is using this port list: %v\n", tcpPorts)
	return &tcpPinger{
		ports:   tcpPorts,
		timeout: tcpTimeout,
		msgChan: tcpChan,
	}
}

// NewTCPBatchPinger creates a new instance of the TCP *batch* pinger
func NewTCPBatchPinger(tcpPorts []int, tcpTimeout time.Duration, tcpBatchChan chan TcpBatch) TCPPinger {
	Logger.Printf("The TCP batch pinger is using this port list: %v\n", tcpPorts)
	return &tcpPinger{
		ports:   tcpPorts,
		timeout: tcpTimeout,
		// nil msgChan
		msgBatchChan: tcpBatchChan,
	}
}

// DialBatchIP performs a batch of TCP dials providing stats regarding the calls
func (t *tcpPinger) DialBatchIP(targetIP string, targetPort int, batchSize int) TcpBatch {
	avgLatency := float32(0)
	unSuccessCount := 0
	for i := 0; i < batchSize; i++ {
		outcome := t.DialIP(targetIP, targetPort)
		currLatencyFloat := float32(outcome.Latency)
		avgLatency = InPlaceAvg(avgLatency, currLatencyFloat, i)
		// Logger.Printf("Iteration %d curr latency %f curr avg %f\n", i, currLatencyFloat, avgLatency)
		if outcome.Latency >= t.timeout {
			unSuccessCount++
		}
	}
	durationAvgLatency := time.Duration(avgLatency)
	// Logger.Printf("Final avg: %s\n", durationAvgLatency)
	pctPacketLoss := float32(unSuccessCount) / float32(batchSize)
	// Logger.Printf("Percentage of packet loss: %f\n", pctPacketLoss)

	return TcpBatch{
		IpAddress: targetIP, TcpPort: targetPort,
		Expertiments: batchSize,
		PctPcktLoss:  pctPacketLoss,
		AvgLatency:   durationAvgLatency,
	}
}

// DialIP performs a TCP dial for a given IP address and TCP port
func (t *tcpPinger) DialIP(targetIP string, targetPort int) TcpCall {
	start := time.Now()
	tcpAddress := fmt.Sprintf("%s:%d", targetIP, targetPort)
	// Logger.Printf("The TCP address to dial is: %v with timeout (duration): %v\n", tcpAddress, t.timeout)
	conn, err := net.DialTimeout("tcp", tcpAddress, t.timeout)
	tcpProtoMsg := TcpCall{IpAddress: targetIP, TcpPort: targetPort}
	if err != nil {
		tcpProtoMsg.Latency = InfiniteLatency
		tcpProtoMsg.Success = false
		return tcpProtoMsg
	}
	// no errors, all good, calculate the latency
	conn.Close()
	elapsed := time.Now().Sub(start)
	tcpProtoMsg.Latency = elapsed
	tcpProtoMsg.Success = true
	return tcpProtoMsg
}

// AsyncTCPDialsForIP is a non blocking attempt at TCP dialing
// publishing the outcomes to a channel
func (t *tcpPinger) AsyncTCPDialsForIP(targetIP string) {
	for _, targetPort := range t.ports {
		go func(targetIP string, targetPort int) {
			Logger.Printf("Dialing IP %s and TCP port %d\n", targetIP, targetPort)
			t.msgChan <- t.DialIP(targetIP, targetPort)
		}(targetIP, targetPort)
	}
	Logger.Printf("Done spawning TCP dials for host %s \n", targetIP)
}

// AsyncTCPDialBatchesForIP is a non blocking attempt at TCP dialing
// in batch, publishing the outcomes to a channel
func (t *tcpPinger) AsyncTCPDialBatchesForIP(targetIP string, batchSize int) {
	for _, targetPort := range t.ports {
		go func(t *tcpPinger, targetIP string, targetPort int, batchSize int) {
			// Logger.Printf("Batch dialing IP %s and TCP port %d\n", targetIP, targetPort)
			t.msgBatchChan <- t.DialBatchIP(targetIP, targetPort, batchSize)
		}(t, targetIP, targetPort, batchSize)
		// Logger.Printf("Done async spawn of TCP calls for IP %s and port %d", targetIP, targetPort)
	}
	// Logger.Printf("Done spawning TCP dials for host %s \n", targetIP)
}

// SpawnTCPDials performs the TCP calls for all the TCP ports of the given IP addresses.
func (t *tcpPinger) SpawnTCPDials(siteNetDetails []string) {
	for _, targetIP := range siteNetDetails {
		t.AsyncTCPDialsForIP(targetIP)
		Logger.Printf("Spawned TCP dials for IP %s", targetIP)
	}
	Logger.Printf("Done spawning TCP dials for: %v Flex Entities", len(siteNetDetails))
}

// SpawnTCPDials performs the TCP calls for all the TCP ports of the given IP addresses.
// For each IP address a batch of calls is performed in order to return stats on those executions.
func (t *tcpPinger) SpawnTCPDialBatches(siteNetDetails []string, batchSize int) {
	for _, targetIP := range siteNetDetails {
		t.AsyncTCPDialBatchesForIP(targetIP, batchSize)
		// Logger.Printf("Spawned TCP dials for IP %s", targetIP)
	}
}

// ---------------------------------------------------------------------------------------

// IcmpPinger provides the functionality to ping IP addresses.
// To use this on a Linux machine make sure that you are running as root user.
// This is due to how raw sockets work and the internals of the ICMP protocol.
type IcmpPinger interface {
	PingIP(targetIP string)
	PingBatchIP(targetIP string, batchSize int) IcmpBatch
	AsyncPingBatchIP(targetIP string, batchSize int)
	SpawnPings(ips []string)
	SpawnBatchPings(ips []string, batchSize int)
}

type icmpPinger struct {
	timoutForIcmpCall time.Duration
	msgChan           chan IcmpCall
	batchMsgChan      chan IcmpBatch
}

// NewIcmpPinger is intended to be used when running `PingIP(...)`
// just unique ICMP calls with their latency
// no batch calls to get the pct of loss packets
func NewIcmpPinger(timeout time.Duration, icmpChan chan IcmpCall) IcmpPinger {
	return &icmpPinger{
		timoutForIcmpCall: timeout,
		msgChan:           icmpChan,
		// nil batchMsgChan
	}
}

// NewIcmpBatchPinger is intended to be used when running `PingBatchIP(...)`
// batch calls to get the pct of loss packets based on multiple ICMP calls (with their latency)
func NewIcmpBatchPinger(timeout time.Duration, icmpBatchChan chan IcmpBatch) IcmpPinger {
	return &icmpPinger{
		timoutForIcmpCall: timeout,
		batchMsgChan:      icmpBatchChan,
		msgChan:           make(chan IcmpCall), // this is not exposed!
	}
}

// PingIP dials via ICMP a target IP address
func (i *icmpPinger) PingIP(targetIP string) {
	// TODO either make this async or return the struct?
	responseReceived := false

	p := fastping.NewPinger()
	p.MaxRTT = i.timoutForIcmpCall
	// Logger.Printf("Setting up the ICMP call for IP %v and timeout (duration): %v\n", targetIP, i.timoutForIcmpCall)

	ra, err := net.ResolveIPAddr("ip4:icmp", targetIP)
	if err != nil {
		// fmt.Println("Resolve error on IP:", targetIP)
		i.msgChan <- IcmpCall{IpAddress: targetIP, Message: err.Error(), Success: false, Latency: InfiniteLatency}
		return
	}
	p.AddIPAddr(ra)

	p.OnRecv = func(addr *net.IPAddr, rtt time.Duration) {
		// Logger.Printf("Received ICMP response for IP %v with latency %v", addr, rtt)
		i.msgChan <- IcmpCall{IpAddress: addr.String(), Latency: time.Duration(rtt.Nanoseconds()), Success: true}
		responseReceived = true
		return
	}
	p.OnIdle = func() {
		if !responseReceived {
			// Logger.Printf("Idling on ICMP call to IP %v with timeout (duration) %v\n", targetIP, i.timoutForIcmpCall)
			i.msgChan <- IcmpCall{IpAddress: targetIP, Message: "This ping call is on timeout", Latency: InfiniteLatency, Success: false}
		}
		return
	}
	err = p.Run()
	if err != nil {
		// fmt.Println("Run error on IP:", targetIP) // TODO make something about running this as sudo! (e.g. error channel or panic here?)
		i.msgChan <- IcmpCall{IpAddress: targetIP, Message: err.Error(), Success: false, Latency: InfiniteLatency}
		return
	}
}

// PingBatchIP dials via ICMP an IP address for a given amount of times.
// The returned struct contains stats on the percentage of packet loss and
// the average amount of time required to perform the calls.
func (i *icmpPinger) PingBatchIP(targetIP string, batchSize int) IcmpBatch {
	// Logger.Printf("Running ICMP batch \n")
	avgLatency := float32(0)
	unSuccessCount := 0
	for idx := 0; idx < batchSize; idx++ {
		// Logger.Printf("ICMP iteration %d \n", idx)
		go i.PingIP(targetIP)
		outcome := <-i.msgChan
		currLatencyFloat := float32(outcome.Latency)
		avgLatency = InPlaceAvg(avgLatency, currLatencyFloat, idx)
		// Logger.Printf("Iteration %d curr latency %f curr avg %f\n", idx, currLatencyFloat, avgLatency)
		if outcome.Latency >= i.timoutForIcmpCall {
			unSuccessCount++
		}
	}
	durationAvgLatency := time.Duration(avgLatency)
	// Logger.Printf("Final avg: %s\n", durationAvgLatency)
	pctPacketLoss := float32(unSuccessCount) / float32(batchSize)
	// Logger.Printf("Percentage of packet loss: %f\n", pctPacketLoss)

	return IcmpBatch{
		IpAddress:    targetIP,
		Expertiments: batchSize,
		PctPcktLoss:  pctPacketLoss,
		AvgLatency:   durationAvgLatency,
	}
}

// AsyncPingBatchIP performs a batch of ICMP calls in an asynchronous way.
// A channel to read these outcomes needs to be consumed.
func (i *icmpPinger) AsyncPingBatchIP(targetIP string, batchSize int) {
	// publish the result to the channel
	go func(i *icmpPinger, targetIP string, batchSize int) {
		icmpBatchMsg := i.PingBatchIP(targetIP, batchSize)
		i.batchMsgChan <- icmpBatchMsg
	}(i, targetIP, batchSize)
}

// SpawnPings performs ICMP calls for a list of input IP addresses.
// This is an asynchronous process.
func (i *icmpPinger) SpawnPings(ips []string) {
	for idx, targetIP := range ips {
		go i.PingIP(targetIP)
		Logger.Printf("ICMP iteration %d: sent request for IP %s\n", idx, targetIP)
	}
	Logger.Printf("Done spawning %v ICMP pings\n", len(ips))
}

// SpawnBatchPings performs ICMP calls for a list of input IP addresses and a batch size for each of them.
// For each IP address a number of ICMP calls proportional to the batch size is performed.
// This is an asynchronous process.
func (i *icmpPinger) SpawnBatchPings(ips []string, batchSize int) {
	for _, targetIP := range ips {
		go i.AsyncPingBatchIP(targetIP, batchSize)
		// Logger.Printf("ICMP iteration %d: sent requests for IP %s with batch size: %v\n", idx, targetIP, batchSize)
	}
	// Logger.Printf("Done spawning %v ICMP pings\n", len(ips))
}
