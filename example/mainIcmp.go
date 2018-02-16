package main

import (
	"log"
	"net"
	"os"
	"time"

	"github.com/tappoz/tcp-pinger/src/service"
)

func init() {
	service.Logger = log.New(os.Stdout, "[ICMP stuff] ", log.LstdFlags)
}

func main() {
	googleIps, _ := net.LookupHost("google.com")
	googleIp := googleIps[0]

	quitChan := service.Schedule(service.IcmpBatchFunc([]string{googleIp}, 10), 5*time.Second)

	time.Sleep(16 * time.Second)
	close(quitChan)
}
