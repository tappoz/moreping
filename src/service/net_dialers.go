package service

import (
	"fmt"
	"net"
	"time"

	"github.com/tappoz/moreping/src/model"
	"github.com/tatsushid/go-fastping"
)

const InfiniteLatency = 9999 * time.Millisecond

type TcpPinger interface {
	DialBatchIp(targetIp string, targetPort int, batchSize int) model.TcpBatch
	AsyncTcpDialBatchesForIp(targetIp string, batchSize int)

	DialIp(targetIp string, targetPort int) model.TcpCall
	AsyncTcpDialsForIp(targetIp string)

	SpawnTcpDials(siteNetDetails []string)
	SpawnTcpDialBatches(siteNetDetails []string, batchSize int)
}

type tcpPinger struct {
	ports        []int
	timeout      time.Duration
	msgChan      chan model.TcpCall
	msgBatchChan chan model.TcpBatch
	batchSize    int
}

func NewTcpPinger(tcpPorts []int, tcpTimeout time.Duration, tcpChan chan model.TcpCall) TcpPinger {
	Logger.Printf("The TCP pinger is using this port list: %v\n", tcpPorts)
	return &tcpPinger{
		ports:   tcpPorts,
		timeout: tcpTimeout,
		msgChan: tcpChan,
	}
}

func NewTcpBatchPinger(tcpPorts []int, tcpTimeout time.Duration, tcpBatchChan chan model.TcpBatch) TcpPinger {
	Logger.Printf("The TCP batch pinger is using this port list: %v\n", tcpPorts)
	return &tcpPinger{
		ports:   tcpPorts,
		timeout: tcpTimeout,
		// nil msgChan
		msgBatchChan: tcpBatchChan,
	}
}

func (t *tcpPinger) DialBatchIp(targetIp string, targetPort int, batchSize int) model.TcpBatch {
	avgLatency := float32(0)
	unSuccessCount := 0
	for i := 0; i < batchSize; i++ {
		outcome := t.DialIp(targetIp, targetPort)
		currLatencyFloat := float32(outcome.Latency)
		avgLatency = InPlaceAvg(avgLatency, currLatencyFloat, i)
		Logger.Printf("Iteration %d curr latency %f curr avg %f\n", i, currLatencyFloat, avgLatency)
		if outcome.Latency >= t.timeout {
			unSuccessCount++
		}
	}
	durationAvgLatency := time.Duration(avgLatency)
	Logger.Printf("Final avg: %s\n", durationAvgLatency)
	pctPacketLoss := float32(unSuccessCount) / float32(batchSize)
	Logger.Printf("Percentage of packet loss: %f\n", pctPacketLoss)

	return model.TcpBatch{
		IpAddress: targetIp, TcpPort: targetPort,
		Expertiments: batchSize,
		PctPcktLoss:  pctPacketLoss,
		AvgLatency:   durationAvgLatency,
	}
}

func (t *tcpPinger) DialIp(targetIp string, targetPort int) model.TcpCall {
	start := time.Now()
	tcpAddress := fmt.Sprintf("%s:%d", targetIp, targetPort)
	Logger.Printf("The TCP address to dial is: %v with timeout (duration): %v\n", tcpAddress, t.timeout)
	conn, err := net.DialTimeout("tcp", tcpAddress, t.timeout)
	tcpProtoMsg := model.TcpCall{IpAddress: targetIp, TcpPort: targetPort}
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

func (t *tcpPinger) AsyncTcpDialsForIp(targetIp string) {
	for _, targetPort := range t.ports {
		go func(targetIp string, targetPort int) {
			Logger.Printf("Dialing IP %s and TCP port %d\n", targetIp, targetPort)
			t.msgChan <- t.DialIp(targetIp, targetPort)
		}(targetIp, targetPort)
	}
	Logger.Printf("Done spawning TCP dials for host %s \n", targetIp)
}

func (t *tcpPinger) AsyncTcpDialBatchesForIp(targetIp string, batchSize int) {
	for _, targetPort := range t.ports {
		go func(t *tcpPinger, targetIp string, targetPort int, batchSize int) {
			Logger.Printf("Batch dialing IP %s and TCP port %d\n", targetIp, targetPort)
			t.msgBatchChan <- t.DialBatchIp(targetIp, targetPort, batchSize)
		}(t, targetIp, targetPort, batchSize)
		Logger.Printf("Done async spawn of TCP calls for IP %s and port %d", targetIp, targetPort)
	}
	Logger.Printf("Done spawning TCP dials for host %s \n", targetIp)
}

func (t *tcpPinger) SpawnTcpDials(siteNetDetails []string) {
	for _, targetIp := range siteNetDetails {
		t.AsyncTcpDialsForIp(targetIp)
		Logger.Printf("Spawned TCP dials for IP %s", targetIp)
	}
	Logger.Printf("Done spawning TCP dials for: %v Flex Entities", len(siteNetDetails))
}

func (t *tcpPinger) SpawnTcpDialBatches(siteNetDetails []string, batchSize int) {
	for _, targetIp := range siteNetDetails {
		t.AsyncTcpDialBatchesForIp(targetIp, batchSize)
		Logger.Printf("Spawned TCP dials for IP %s", targetIp)
	}
}

// ---------------------------------------------------------------------------------------

type IcmpPinger interface {
	PingIp(targetIp string)
	PingBatchIp(targetIp string, batchSize int) model.IcmpBatch
	AsyncPingBatchIp(targetIp string, batchSize int)
	SpawnPings(ips []string)
	SpawnBatchPings(ips []string, batchSize int)
}

type icmpPinger struct {
	timoutForIcmpCall time.Duration
	msgChan           chan model.IcmpCall
	batchMsgChan      chan model.IcmpBatch
}

// NewIcmpPinger is intended to be used when running `PingIp(...)`
// just unique ICMP calls with their latency
// no batch calls to get the pct of loss packets
func NewIcmpPinger(timeout time.Duration, icmpChan chan model.IcmpCall) IcmpPinger {
	return &icmpPinger{
		timoutForIcmpCall: timeout,
		msgChan:           icmpChan,
		// nil batchMsgChan
	}
}

// NewIcmpPinger is intended to be used when running `PingBatchIp(...)`
// batch calls to get the pct of loss packets based on multiple ICMP calls (with their latency)
func NewIcmpBatchPinger(timeout time.Duration, icmpBatchChan chan model.IcmpBatch) IcmpPinger {
	return &icmpPinger{
		timoutForIcmpCall: timeout,
		batchMsgChan:      icmpBatchChan,
		msgChan:           make(chan model.IcmpCall), // this is not exposed!
	}
}

// TODO either make this async or return the struct?
func (i *icmpPinger) PingIp(targetIp string) {
	responseReceived := false

	p := fastping.NewPinger()
	p.MaxRTT = i.timoutForIcmpCall
	Logger.Printf("Setting up the ICMP call for IP %v and timeout (duration): %v\n", targetIp, i.timoutForIcmpCall)

	ra, err := net.ResolveIPAddr("ip4:icmp", targetIp)
	if err != nil {
		fmt.Println("Resolve error on IP:", targetIp)
		i.msgChan <- model.IcmpCall{IpAddress: targetIp, Message: err.Error(), Success: false, Latency: InfiniteLatency}
		return
	}
	p.AddIPAddr(ra)

	p.OnRecv = func(addr *net.IPAddr, rtt time.Duration) {
		Logger.Printf("Received ICMP response for IP %v with latency %v", addr, rtt)
		i.msgChan <- model.IcmpCall{IpAddress: addr.String(), Latency: time.Duration(rtt.Nanoseconds()), Success: true}
		responseReceived = true
		return
	}
	p.OnIdle = func() {
		if !responseReceived {
			Logger.Printf("Idling on ICMP call to IP %v with timeout (duration) %v\n", targetIp, i.timoutForIcmpCall)
			i.msgChan <- model.IcmpCall{IpAddress: targetIp, Message: "This ping call is on timeout", Latency: InfiniteLatency, Success: false}
		}
		return
	}
	err = p.Run()
	if err != nil {
		fmt.Println("Run error on IP:", targetIp) // TODO make something about running this as sudo! (e.g. error channel or panic here?)
		i.msgChan <- model.IcmpCall{IpAddress: targetIp, Message: err.Error(), Success: false, Latency: InfiniteLatency}
		return
	}
}

func (i *icmpPinger) PingBatchIp(targetIp string, batchSize int) model.IcmpBatch {
	Logger.Printf("Running ICMP batch \n")
	avgLatency := float32(0)
	unSuccessCount := 0
	for idx := 0; idx < batchSize; idx++ {
		Logger.Printf("ICMP iteration %d \n", idx)
		go i.PingIp(targetIp)
		outcome := <-i.msgChan
		currLatencyFloat := float32(outcome.Latency)
		avgLatency = InPlaceAvg(avgLatency, currLatencyFloat, idx)
		Logger.Printf("Iteration %d curr latency %f curr avg %f\n", idx, currLatencyFloat, avgLatency)
		if outcome.Latency >= i.timoutForIcmpCall {
			unSuccessCount++
		}
	}
	durationAvgLatency := time.Duration(avgLatency)
	Logger.Printf("Final avg: %s\n", durationAvgLatency)
	pctPacketLoss := float32(unSuccessCount) / float32(batchSize)
	Logger.Printf("Percentage of packet loss: %f\n", pctPacketLoss)

	return model.IcmpBatch{
		IpAddress:    targetIp,
		Expertiments: batchSize,
		PctPcktLoss:  pctPacketLoss,
		AvgLatency:   durationAvgLatency,
	}
}

func (i *icmpPinger) AsyncPingBatchIp(targetIp string, batchSize int) {
	// publish the result to the channel
	go func(i *icmpPinger, targetIp string, batchSize int) {
		icmpBatchMsg := i.PingBatchIp(targetIp, batchSize)
		i.batchMsgChan <- icmpBatchMsg
	}(i, targetIp, batchSize)
}

func (i *icmpPinger) SpawnPings(ips []string) {
	for idx, targetIp := range ips {
		go i.PingIp(targetIp)
		Logger.Printf("ICMP iteration %d: sent request for IP %s\n", idx, targetIp)
	}
	Logger.Printf("Done spawning %v ICMP pings\n", len(ips))
}

func (i *icmpPinger) SpawnBatchPings(ips []string, batchSize int) {
	for idx, targetIp := range ips {
		go i.AsyncPingBatchIp(targetIp, batchSize)
		Logger.Printf("ICMP iteration %d: sent requests for IP %s with batch size: %v\n", idx, targetIp, batchSize)
	}
	Logger.Printf("Done spawning %v ICMP pings\n", len(ips))
}
