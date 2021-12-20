package moreping_test

import (
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/onsi/gomega"
	"github.com/tappoz/moreping/src/moreping"
)

func TestICMPSyncPing(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	googleIPs, _ := net.LookupHost("google.com")
	// fmt.Println(googleIPs)
	googleIPv4 := googleIPs[1]

	icmpTimeout := 2 * time.Second
	icmpChan := make(chan moreping.IcmpCall)
	icmpPinger := moreping.NewIcmpPinger(icmpTimeout, icmpChan)

	// it should synchronously ping google.com"
	go icmpPinger.PingIP(googleIPv4)
	icmpCallMsg := <-icmpChan
	// fmt.Printf("%+v", icmpCallMsg)
	g.Expect(icmpCallMsg.IpAddress).To(gomega.Equal(googleIPv4))
	g.Expect(icmpCallMsg.Latency).Should(gomega.BeNumerically("<=", icmpTimeout))
	g.Expect(icmpCallMsg.Success).To(gomega.Equal(true))

	// it should synchronously ping the non-existing host 'foo' then timeout and return an unsuccessful ICMP message
	fooHost := "foo"
	go icmpPinger.PingIP(fooHost)
	icmpCallMsg = <-icmpChan
	g.Expect(icmpCallMsg.Message).To(gomega.HavePrefix(fmt.Sprintf("lookup %s", fooHost))) // on Windows: "lookup foo: no such host", on Linux: "lookup foo on 192.168.65.1:53: server misbehaving"
	g.Expect(icmpCallMsg.IpAddress).To(gomega.Equal(fooHost))
	g.Expect(icmpCallMsg.Latency).Should(gomega.Equal(moreping.InfiniteLatency))
	g.Expect(icmpCallMsg.Success).To(gomega.Equal(false))

	// it should synchronously and quickly timeout on an existing host 'google.com' and return an unsuccessful ICMP message
	veryLowIcmpTimeout := 100 * time.Nanosecond
	icmpPingerWithVeryLowTimeout := moreping.NewIcmpPinger(veryLowIcmpTimeout, icmpChan)
	go icmpPingerWithVeryLowTimeout.PingIP(googleIPv4)
	icmpCallMsg = <-icmpChan
	g.Expect(icmpCallMsg.Success).To(gomega.Equal(false))
	g.Expect(icmpCallMsg.IpAddress).To(gomega.Equal(googleIPv4))
	g.Expect(icmpCallMsg.Latency).Should(gomega.Equal(moreping.InfiniteLatency))
	g.Expect(icmpCallMsg.Message).To(gomega.Equal("This ping call is on timeout"))
}

func TestICMPAsyncPing(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	googleIPs, _ := net.LookupHost("google.com")
	// fmt.Println(googleIPs)
	googleIPv4 := googleIPs[1]

	icmpTimeout := 2 * time.Second
	icmpChan := make(chan moreping.IcmpCall)
	icmpPinger := moreping.NewIcmpPinger(icmpTimeout, icmpChan)

	// it should asynchronously ping google.com
	icmpPinger.SpawnPings([]string{googleIPv4})
	icmpCallMsg := <-icmpChan
	g.Expect(icmpCallMsg.IpAddress).To(gomega.Equal(googleIPv4))
	g.Expect(icmpCallMsg.Latency).Should(gomega.BeNumerically("<=", float32(icmpTimeout)))
	g.Expect(icmpCallMsg.Success).To(gomega.Equal(true))
}

func TestICMPBatchPing(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	googleIPs, _ := net.LookupHost("google.com")
	// fmt.Println(googleIPs)
	googleIPv4 := googleIPs[1]

	icmpTimeout := 2 * time.Second
	icmpBatchChan := make(chan moreping.IcmpBatch)
	icmpBatchPinger := moreping.NewIcmpBatchPinger(icmpTimeout, icmpBatchChan)
	batchSize := 5

	// it should synchronously perform a batch of ping calls to google.com
	icmpBatch := icmpBatchPinger.PingBatchIP(googleIPv4, batchSize)
	g.Expect(icmpBatch.IpAddress).To(gomega.Equal(googleIPv4))
	g.Expect(icmpBatch.Expertiments).To(gomega.Equal(batchSize))
	g.Expect(icmpBatch.PctPcktLoss).To(gomega.Equal(float32(0.0))) // assuming Google always responds promptly
	g.Expect(icmpBatch.AvgLatency).Should(gomega.BeNumerically("<=", icmpTimeout))

	// it should asynchronously perform a batch of ping calls to google.com
	icmpBatchPinger.AsyncPingBatchIP(googleIPv4, batchSize)
	icmpBatch = <-icmpBatchChan
	g.Expect(icmpBatch.IpAddress).To(gomega.Equal(googleIPv4))
	g.Expect(icmpBatch.Expertiments).To(gomega.Equal(batchSize))
	g.Expect(icmpBatch.PctPcktLoss).To(gomega.Equal(float32(0.0))) // assuming Google always responds promptly
	g.Expect(icmpBatch.AvgLatency).Should(gomega.BeNumerically("<=", icmpTimeout))
}

func TestTCPSyncDial(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	googleIPs, _ := net.LookupHost("google.com")
	// fmt.Println(googleIPs)
	googleIPv4 := googleIPs[1]

	tcpTimeout := 2 * time.Second
	tcpChan := make(chan moreping.TcpCall)
	tcpPorts := []int{80, 443}
	TCPPinger := moreping.NewTCPPinger(tcpPorts, tcpTimeout, tcpChan)

	// it should synchronously dial google.com on port 80
	tcpPort := 80
	tcpCallMsg := TCPPinger.DialIP(googleIPv4, tcpPort)
	g.Expect(tcpCallMsg.IpAddress).To(gomega.Equal(googleIPv4))
	g.Expect(tcpCallMsg.TcpPort).To(gomega.Equal(tcpPort))
	g.Expect(tcpCallMsg.Latency).Should(gomega.BeNumerically("<=", tcpTimeout))
	g.Expect(tcpCallMsg.Success).To(gomega.Equal(true))

	// it should synchronously dial google.com on port 443
	tcpPort = 443
	tcpCallMsg = TCPPinger.DialIP(googleIPv4, tcpPort)
	g.Expect(tcpCallMsg.IpAddress).To(gomega.Equal(googleIPv4))
	g.Expect(tcpCallMsg.TcpPort).To(gomega.Equal(tcpPort))
	g.Expect(tcpCallMsg.Latency).Should(gomega.BeNumerically("<=", tcpTimeout))
	g.Expect(tcpCallMsg.Success).To(gomega.Equal(true))

	// it should synchronously dial 'foo' on port 666 then timeout and give back an unsuccessful TCP call message
	fooHost := "foo"
	tcpPort = 666
	tcpCallMsg = TCPPinger.DialIP(fooHost, tcpPort)
	g.Expect(tcpCallMsg.IpAddress).To(gomega.Equal(fooHost))
	g.Expect(tcpCallMsg.TcpPort).To(gomega.Equal(tcpPort))
	g.Expect(tcpCallMsg.Latency).Should(gomega.Equal(moreping.InfiniteLatency))
	g.Expect(tcpCallMsg.Success).To(gomega.Equal(false))
}

func TestTCPAsyncDial(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	googleIPs, _ := net.LookupHost("google.com")
	// fmt.Println(googleIPs)
	googleIPv4 := googleIPs[1]

	tcpTimeout := 2 * time.Second
	tcpChan := make(chan moreping.TcpCall)
	tcpPorts := []int{80, 443}
	TCPPinger := moreping.NewTCPPinger(tcpPorts, tcpTimeout, tcpChan)

	// it should asynchronously dial google.com on all the available TCP ports
	TCPPinger.AsyncTCPDialsForIP(googleIPv4)
	for range tcpPorts {
		tcpCallMsg := <-tcpChan
		//fmt.Printf("the msg: %v\n", tcpCallMsg)
		g.Expect(tcpCallMsg.IpAddress).To(gomega.Equal(googleIPv4))
		g.Expect(tcpPorts).To(gomega.ContainElement(tcpCallMsg.TcpPort))
		g.Expect(tcpCallMsg.Latency).Should(gomega.BeNumerically("<=", tcpTimeout))
		g.Expect(tcpCallMsg.Success).To(gomega.Equal(true))
	}

	// it should asynchronously dial google.com (from an array of hosts) on all the available TCP ports
	TCPPinger.SpawnTCPDials([]string{googleIPv4})
	for range tcpPorts {
		tcpCallMsg := <-tcpChan
		//fmt.Printf("the msg: %v\n", tcpCallMsg)
		g.Expect(tcpCallMsg.IpAddress).To(gomega.Equal(googleIPv4))
		g.Expect(tcpPorts).To(gomega.ContainElement(tcpCallMsg.TcpPort))
		g.Expect(tcpCallMsg.Latency).Should(gomega.BeNumerically("<=", tcpTimeout))
		g.Expect(tcpCallMsg.Success).To(gomega.Equal(true))
	}
}

func TestTCPBatchDial(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	googleIPs, _ := net.LookupHost("google.com")
	// fmt.Println(googleIPs)
	googleIPv4 := googleIPs[1]

	tcpTimeout := 500 * time.Millisecond // trying to be strict on the timeout on Google
	tcpBatchChan := make(chan moreping.TcpBatch)
	tcpPorts := []int{80, 443}
	tcpBatchPinger := moreping.NewTCPBatchPinger(tcpPorts, tcpTimeout, tcpBatchChan)
	batchSize := 5

	// it should synchronously perform a batch of dials to google.com on port 80
	tcpPort := 80 // tcpPorts[0]
	tcpBatch := tcpBatchPinger.DialBatchIP(googleIPv4, tcpPort, batchSize)
	g.Expect(tcpBatch.IpAddress).To(gomega.Equal(googleIPv4))
	g.Expect(tcpBatch.TcpPort).To(gomega.Equal(tcpPort))
	g.Expect(tcpBatch.Expertiments).To(gomega.Equal(batchSize))
	g.Expect(tcpBatch.PctPcktLoss).To(gomega.Equal(float32(0.0))) // assuming Google always responds promptly
	g.Expect(tcpBatch.AvgLatency).Should(gomega.BeNumerically("<=", tcpTimeout))

	// it should asynchronously dial google.com on all the available TCP ports
	tcpBatchPinger.AsyncTCPDialBatchesForIP(googleIPv4, batchSize)
	for range tcpPorts {
		tcpBatch := <-tcpBatchChan

		g.Expect(tcpBatch.IpAddress).To(gomega.Equal(googleIPv4))
		g.Expect(tcpPorts).To(gomega.ContainElement(tcpBatch.TcpPort))
		g.Expect(tcpBatch.Expertiments).To(gomega.Equal(batchSize))
		g.Expect(tcpBatch.PctPcktLoss).To(gomega.Equal(float32(0.0))) // assuming Google always responds promptly
		g.Expect(tcpBatch.AvgLatency).Should(gomega.BeNumerically("<=", tcpTimeout))
	}

	// it should asynchronously dial (a batch of calls) google.com (from an array of hosts) on all the available TCP ports
	tcpBatchPinger.SpawnTCPDialBatches([]string{googleIPv4}, batchSize)
	for range tcpPorts {
		tcpBatch := <-tcpBatchChan

		g.Expect(tcpBatch.IpAddress).To(gomega.Equal(googleIPv4))
		g.Expect(tcpPorts).To(gomega.ContainElement(tcpBatch.TcpPort))
		g.Expect(tcpBatch.Expertiments).To(gomega.Equal(batchSize))
		g.Expect(tcpBatch.PctPcktLoss).To(gomega.Equal(float32(0.0))) // assuming Google always responds promptly
		g.Expect(tcpBatch.AvgLatency).Should(gomega.BeNumerically("<=", tcpTimeout))
	}

}
