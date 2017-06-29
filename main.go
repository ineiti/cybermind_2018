package main

import (
	"os"

	"time"

	"github.com/ineiti/cybermind/broker"
	"github.com/ineiti/cybermind/modules/data"
	"github.com/ineiti/cybermind/modules/input"
	"github.com/ineiti/cybermind/modules/user"
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
	if err := data.RegisterConfig(broker); err != nil {
		return err
	}
	if err := input.RegisterEmail(broker); err != nil {
		return err
	}
	if err := input.RegisterFiles(broker); err != nil {
		return err
	}
	if err := user.RegisterWeb(broker); err != nil {
		return err
	}
	if err := user.RegisterCLI(broker); err != nil {
		return err
	}
	if err := broker.Start(); err != nil {
		return err
	}
	for {
		time.Sleep(time.Hour)
	}
	return nil
}
