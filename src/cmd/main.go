package main

import (
	"os"

	"github.com/urfave/cli"
)

func newApp() *cli.App {
	app := cli.NewApp()
	app.Name = "moreping"
	app.Author = "Alessio Gottardo"
	app.Version = "0.0.1"
	app.Usage = "ICMP ping and TCP/port dial"
	return app
}

func tcpCommand() cli.Command {
	return cli.Command{
		Name:   "tcp",
		Action: TcpCmd,
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "domain",
				Usage: "the domain to dial",
			},
			cli.Int64Flag{
				Name:  "port",
				Usage: "the port to dial",
			},
		},
	}
}

func icmpCommand() cli.Command {
	return cli.Command{
		Name:   "icmp",
		Usage:  "this *must* be run as root because of the internals of ICMP and raw sockets on Linux",
		Action: IcmpCmd,
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "domain",
				Usage: "the domain to dial",
			},
		},
	}
}

func main() {
	app := newApp()
	app.Commands = []cli.Command{tcpCommand(), icmpCommand()}
	app.Run(os.Args)
}
