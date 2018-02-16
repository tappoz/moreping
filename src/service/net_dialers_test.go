package service_test

import (
	"fmt"
	"net"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/tappoz/tcp-pinger/src/model"
	"github.com/tappoz/tcp-pinger/src/service"
)

var _ = Describe("NetDialers", func() {

	googleIps, _ := net.LookupHost("google.com")
	googleIp := googleIps[0]

	Describe("ICMP pinger", func() {

		icmpTimeout := 2 * time.Second
		icmpChan := make(chan model.IcmpCall)
		icmpPinger := service.NewIcmpPinger(icmpTimeout, icmpChan)

		It("should synchronously ping google.com", func() {

			go icmpPinger.PingIp(googleIp)
			icmpCallMsg := <-icmpChan

			Expect(icmpCallMsg.IpAddress).To(Equal(googleIp))
			Expect(icmpCallMsg.Latency).Should(BeNumerically("<=", icmpTimeout))
			Expect(icmpCallMsg.Success).To(Equal(true))
		})

		It("should synchronously ping the non-existing host 'foo' then timeout and return an unsuccessful ICMP message", func() {

			fooHost := "foo"

			go icmpPinger.PingIp(fooHost)
			icmpCallMsg := <-icmpChan

			Expect(icmpCallMsg.Message).To(HavePrefix(fmt.Sprintf("lookup %s", fooHost))) // on Windows: "lookup foo: no such host", on Linux: "lookup foo on 192.168.65.1:53: server misbehaving"
			Expect(icmpCallMsg.IpAddress).To(Equal(fooHost))
			Expect(icmpCallMsg.Latency).Should(Equal(service.InfiniteLatency))
			Expect(icmpCallMsg.Success).To(Equal(false))
		})

		It("should synchronously and quickly timeout on an existing host 'google.com' and return an unsuccessful ICMP message", func() {

			veryLowIcmpTimeout := 100 * time.Nanosecond
			icmpPingerWithVeryLowTimeout := service.NewIcmpPinger(veryLowIcmpTimeout, icmpChan)

			go icmpPingerWithVeryLowTimeout.PingIp(googleIp)
			icmpCallMsg := <-icmpChan

			Expect(icmpCallMsg.Success).To(Equal(false))
			Expect(icmpCallMsg.IpAddress).To(Equal(googleIp))
			Expect(icmpCallMsg.Latency).Should(Equal(service.InfiniteLatency))
			Expect(icmpCallMsg.Message).To(Equal("This ping call is on timeout"))
		})

		It("should asynchronously ping google.com", func() {
			icmpPinger.SpawnPings([]string{googleIp})

			icmpCallMsg := <-icmpChan

			Expect(icmpCallMsg.IpAddress).To(Equal(googleIp))
			Expect(icmpCallMsg.Latency).Should(BeNumerically("<=", float32(icmpTimeout)))
			Expect(icmpCallMsg.Success).To(Equal(true))
		})
	})

	Describe("ICMP batch pinger", func() {

		icmpTimeout := 2 * time.Second
		icmpBatchChan := make(chan model.IcmpBatch)
		icmpBatchPinger := service.NewIcmpBatchPinger(icmpTimeout, icmpBatchChan)

		It("should synchronously perform a batch of ping calls to google.com", func() {
			batchSize := 5
			icmpBatch := icmpBatchPinger.PingBatchIp(googleIp, batchSize)

			Expect(icmpBatch.IpAddress).To(Equal(googleIp))
			Expect(icmpBatch.Expertiments).To(Equal(batchSize))
			Expect(icmpBatch.PctPcktLoss).To(Equal(float32(0.0))) // assuming Google always responds promptly
			Expect(icmpBatch.AvgLatency).Should(BeNumerically("<=", icmpTimeout))
		})

		It("should asynchronously perform a batch of ping calls to google.com", func() {
			batchSize := 5

			icmpBatchPinger.AsyncPingBatchIp(googleIp, batchSize)
			icmpBatch := <-icmpBatchChan

			Expect(icmpBatch.IpAddress).To(Equal(googleIp))
			Expect(icmpBatch.Expertiments).To(Equal(batchSize))
			Expect(icmpBatch.PctPcktLoss).To(Equal(float32(0.0))) // assuming Google always responds promptly
			Expect(icmpBatch.AvgLatency).Should(BeNumerically("<=", icmpTimeout))
		})
	})

	Describe("TCP dialer", func() {

		tcpTimeout := 2 * time.Second
		tcpChan := make(chan model.TcpCall)
		tcpPorts := []int{80, 443}
		tcpPinger := service.NewTcpPinger(tcpPorts, tcpTimeout, tcpChan)

		It("should synchronously dial google.com on port 80", func() {
			tcpPort := 80
			tcpCallMsg := tcpPinger.DialIp(googleIp, tcpPort)

			Expect(tcpCallMsg.IpAddress).To(Equal(googleIp))
			Expect(tcpCallMsg.TcpPort).To(Equal(tcpPort))
			Expect(tcpCallMsg.Latency).Should(BeNumerically("<=", tcpTimeout))
			Expect(tcpCallMsg.Success).To(Equal(true))
		})

		It("should synchronously dial google.com on port 443", func() {
			tcpPort := 443
			tcpCallMsg := tcpPinger.DialIp(googleIp, tcpPort)

			Expect(tcpCallMsg.IpAddress).To(Equal(googleIp))
			Expect(tcpCallMsg.TcpPort).To(Equal(tcpPort))
			Expect(tcpCallMsg.Latency).Should(BeNumerically("<=", tcpTimeout))
			Expect(tcpCallMsg.Success).To(Equal(true))
		})

		It("should synchronously dial 'foo' on port 666 then timeout and give back an unsuccessful TCP call message", func() {
			fooHost := "foo"
			tcpPort := 666
			tcpCallMsg := tcpPinger.DialIp(fooHost, tcpPort)

			Expect(tcpCallMsg.IpAddress).To(Equal(fooHost))
			Expect(tcpCallMsg.TcpPort).To(Equal(tcpPort))
			Expect(tcpCallMsg.Latency).Should(Equal(service.InfiniteLatency))
			Expect(tcpCallMsg.Success).To(Equal(false))
		})

		It("should asynchronously dial google.com on all the available TCP ports", func() {
			tcpPinger.AsyncTcpDialsForIp(googleIp)

			for range tcpPorts {
				tcpCallMsg := <-tcpChan
				//fmt.Printf("the msg: %v\n", tcpCallMsg)
				Expect(tcpCallMsg.IpAddress).To(Equal(googleIp))
				Expect(tcpPorts).To(ContainElement(tcpCallMsg.TcpPort))
				Expect(tcpCallMsg.Latency).Should(BeNumerically("<=", tcpTimeout))
				Expect(tcpCallMsg.Success).To(Equal(true))
			}
		})

		It("should asynchronously dial google.com (from an array of hosts) on all the available TCP ports", func() {
			tcpPinger.SpawnTcpDials([]string{googleIp})

			for range tcpPorts {
				tcpCallMsg := <-tcpChan
				//fmt.Printf("the msg: %v\n", tcpCallMsg)
				Expect(tcpCallMsg.IpAddress).To(Equal(googleIp))
				Expect(tcpPorts).To(ContainElement(tcpCallMsg.TcpPort))
				Expect(tcpCallMsg.Latency).Should(BeNumerically("<=", tcpTimeout))
				Expect(tcpCallMsg.Success).To(Equal(true))
			}
		})
	})

	Describe("TCP batch dialer", func() {

		tcpTimeout := 500 * time.Millisecond // trying to be strict on the timeout on Google
		tcpBatchChan := make(chan model.TcpBatch)
		tcpPorts := []int{80, 443}
		tcpBatchPinger := service.NewTcpBatchPinger(tcpPorts, tcpTimeout, tcpBatchChan)
		batchSize := 5

		It("should synchronously perform a batch of dials to google.com on port 80", func() {
			tcpPort := 80 // tcpPorts[0]
			tcpBatch := tcpBatchPinger.DialBatchIp(googleIp, tcpPort, batchSize)

			Expect(tcpBatch.IpAddress).To(Equal(googleIp))
			Expect(tcpBatch.TcpPort).To(Equal(tcpPort))
			Expect(tcpBatch.Expertiments).To(Equal(batchSize))
			Expect(tcpBatch.PctPcktLoss).To(Equal(float32(0.0))) // assuming Google always responds promptly
			Expect(tcpBatch.AvgLatency).Should(BeNumerically("<=", tcpTimeout))
		})

		It("should asynchronously dial google.com on all the available TCP ports", func() {
			tcpBatchPinger.AsyncTcpDialBatchesForIp(googleIp, batchSize)

			for range tcpPorts {
				tcpBatch := <-tcpBatchChan

				Expect(tcpBatch.IpAddress).To(Equal(googleIp))
				Expect(tcpPorts).To(ContainElement(tcpBatch.TcpPort))
				Expect(tcpBatch.Expertiments).To(Equal(batchSize))
				Expect(tcpBatch.PctPcktLoss).To(Equal(float32(0.0))) // assuming Google always responds promptly
				Expect(tcpBatch.AvgLatency).Should(BeNumerically("<=", tcpTimeout))
			}
		})

		It("should asynchronously dial (a batch of calls) google.com (from an array of hosts) on all the available TCP ports", func() {
			tcpBatchPinger.SpawnTcpDialBatches([]string{googleIp}, batchSize)

			for range tcpPorts {
				tcpBatch := <-tcpBatchChan

				Expect(tcpBatch.IpAddress).To(Equal(googleIp))
				Expect(tcpPorts).To(ContainElement(tcpBatch.TcpPort))
				Expect(tcpBatch.Expertiments).To(Equal(batchSize))
				Expect(tcpBatch.PctPcktLoss).To(Equal(float32(0.0))) // assuming Google always responds promptly
				Expect(tcpBatch.AvgLatency).Should(BeNumerically("<=", tcpTimeout))
			}
		})
	})
})
