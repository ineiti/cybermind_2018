package main

import (
	"os"

	"time"

	"github.com/ineiti/cybermind/broker"
	"github.com/ineiti/cybermind/modules"
	"gopkg.in/dedis/onet.v1/log"
	"gopkg.in/urfave/cli.v1"
)

func main() {
	app := cli.NewApp()

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "config, c",
			Usage: "config-directory of cybermind",
		},
		cli.IntFlag{
			Name:  "debug, d",
			Usage: "debug-level: 0 (silent) - 5 (spam)",
			Value: 0,
		},
	}
	app.Action = start
	app.Run(os.Args)
}

func start(c *cli.Context) error {
	configPath := c.GlobalString("config")
	log.Print(configPath)
	if configPath != "" {
		broker.ConfigPath = configPath
	}

	log.Info("Starting some services")
	log.SetDebugVisible(c.Int("debug"))
	broker := broker.NewBroker()
	modules.RegisterAll(broker)
	if err := broker.Start(); err != nil {
		return err
	}
	for {
		time.Sleep(time.Hour)
	}
	return nil
}
