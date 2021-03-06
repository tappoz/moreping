package main

import (
	"log"
	"os"
	"time"

	"github.com/tappoz/moreping/src/moreping"
	"github.com/urfave/cli"
)

func tcpCmd(c *cli.Context) {
	moreping.Logger = log.New(os.Stdout, "[TCP stuff] ", log.LstdFlags)

	domain := c.String("domain")
	port := c.Int64("port")

	moreping.Schedule(moreping.TCPBatchFunc([]string{domain}, []int{int(port)}, 10), 5*time.Second)
	for {
	}
}

func sudoCheck() {
	// TODO check sudo usage parsing this log file: /var/log/auth.log
	// sudoUid := os.Getenv("SUDO_UID")
	// sudoGid := os.Getenv("SUDO_GID")
	// sudoUser := os.Getenv("SUDO_USER")
	// moreping.Logger.Printf("Stuff: %s, %s, %s", sudoUid, sudoGid, sudoUser)
}

func icmpCmd(c *cli.Context) {
	moreping.Logger = log.New(os.Stdout, "[ICMP stuff] ", log.LstdFlags)

	sudoCheck()
	domain := c.String("domain")

	moreping.Schedule(moreping.IcmpBatchFunc([]string{domain}, 10), 5*time.Second)

	for {
	}
}
