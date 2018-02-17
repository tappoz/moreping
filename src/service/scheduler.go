package service

import (
	"time"

	"github.com/tappoz/moreping/src/model"
)

func TcpBatchFunc(websites []string, tcpPorts []int, batchSize int) func() {
	tcpBatchChan := make(chan model.TcpBatch)
	tcpPinger := NewTcpBatchPinger(tcpPorts, 1*time.Second, tcpBatchChan)
	return func() {
		tcpPinger.SpawnTcpDialBatches(websites, batchSize)
	}
}

func IcmpBatchFunc(websites []string, batchSize int) func() {
	icmpBatchChan := make(chan model.IcmpBatch)
	icmpPinger := NewIcmpBatchPinger(1*time.Second, icmpBatchChan)
	return func() {
		icmpPinger.SpawnBatchPings(websites, batchSize)
	}
}

func Schedule(f func(), recurring time.Duration) chan struct{} {
	ticker := time.NewTicker(recurring)
	quit := make(chan struct{})
	go func(inFunc func()) {
		go inFunc()
		for {
			select {
			case <-ticker.C:
				Logger.Printf("! Ticked")
				go inFunc()
			case <-quit:
				Logger.Printf("! Stopping the scheduler")
				ticker.Stop()
				return
			}
		}
	}(f)

	return quit
}
