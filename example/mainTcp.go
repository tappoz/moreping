package main

import (
	"log"
	"net"
	"os"
	"time"

	"github.com/tappoz/moreping/src/service"
)

func init() {
	service.Logger = log.New(os.Stdout, "[TCP stuff] ", log.LstdFlags)
}

func main() {
	googleIps, _ := net.LookupHost("google.com")
	googleIp := googleIps[0]

	quitChan := service.Schedule(service.TCPBatchFunc([]string{googleIp}, []int{80, 443}, 10), 5*time.Second)

	time.Sleep(16 * time.Second)
	close(quitChan)
}
