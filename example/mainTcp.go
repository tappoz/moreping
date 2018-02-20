package main

import (
	"log"
	"net"
	"os"
	"time"

	"github.com/tappoz/moreping/src/moreping"
)

func init() {
	moreping.Logger = log.New(os.Stdout, "[TCP stuff] ", log.LstdFlags)
}

func main() {
	googleIps, _ := net.LookupHost("google.com")
	googleIp := googleIps[0]

	quitChan := moreping.Schedule(moreping.TCPBatchFunc([]string{googleIp}, []int{80, 443}, 10), 5*time.Second)

	time.Sleep(16 * time.Second)
	close(quitChan)
}
