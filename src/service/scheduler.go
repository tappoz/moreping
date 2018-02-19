package service

import (
	"time"

	"github.com/tappoz/moreping/src/model"
)

var tcpBatchChan = make(chan model.TcpBatch)
var icmpBatchChan = make(chan model.IcmpBatch)

// TCPBatchFunc is a "func" type that can be used to schedule TCP dials
func TCPBatchFunc(websites []string, tcpPorts []int, batchSize int) func() {
	tcpPinger := NewTCPBatchPinger(tcpPorts, 1*time.Second, tcpBatchChan)
	return func() {
		tcpPinger.SpawnTCPDialBatches(websites, batchSize)
	}
}

// IcmpBatchFunc is a "func" type that can be used to schedule ICMP calls
func IcmpBatchFunc(websites []string, batchSize int) func() {
	icmpPinger := NewIcmpBatchPinger(1*time.Second, icmpBatchChan)
	return func() {
		icmpPinger.SpawnBatchPings(websites, batchSize)
	}
}

// Schedule can be used to schedule any of the "func" types.
func Schedule(f func(), recurring time.Duration) chan struct{} {
	ticker := time.NewTicker(recurring)
	quit := make(chan struct{})
	go func(inFunc func()) {
		go inFunc()
		for {
			select {
			case <-ticker.C:
				// Logger.Printf("! Ticked")
				go inFunc()
			case tcpMsg := <-tcpBatchChan:
				Logger.Printf("Stats: %#v", tcpMsg)
			case icmpMsg := <-icmpBatchChan:
				Logger.Printf("Stats: %#v", icmpMsg)
			case <-quit:
				// Logger.Printf("! Stopping the scheduler")
				ticker.Stop()
				return
			}
		}
	}(f)

	return quit
}
