package service

import (
	"time"

	"github.com/tappoz/moreping/src/model"
)

var TcpBatchChan = make(chan model.TcpBatch)
var IcmpBatchChan = make(chan model.IcmpBatch)

// TcpBatchFunc is a "func" type that can be used to schedule TCP dials
func TcpBatchFunc(websites []string, tcpPorts []int, batchSize int) func() {
	tcpPinger := NewTcpBatchPinger(tcpPorts, 1*time.Second, TcpBatchChan)
	return func() {
		tcpPinger.SpawnTcpDialBatches(websites, batchSize)
	}
}

// IcmpBatchFunc is a "func" type that can be used to schedule ICMP calls
func IcmpBatchFunc(websites []string, batchSize int) func() {
	icmpPinger := NewIcmpBatchPinger(1*time.Second, IcmpBatchChan)
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
			case tcpMsg := <-TcpBatchChan:
				Logger.Printf("Stats: %#v", tcpMsg)
			case icmpMsg := <-IcmpBatchChan:
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
