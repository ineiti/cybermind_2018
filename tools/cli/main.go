package main

import (
	"os"

	"github.com/ineiti/cybermind/broker"
	"github.com/ineiti/cybermind/modules/user"
	"gopkg.in/dedis/onet.v1/log"
	"gopkg.in/urfave/cli.v1"
)

var configPath string
var fifo *os.File

func main() {
	app := cli.NewApp()

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "config, c",
			Usage: "path to cybermind-config",
		},
		cli.IntFlag{
			Name:  "debug, d",
			Usage: "debug-verbosity: 0(silent) - 5(spam)",
			Value: 0,
		},
	}

	app.Commands = []cli.Command{
		{
			Name:    "module",
			Aliases: []string{"m"},
			Usage:   "module handling",
			Subcommands: []cli.Command{
				{
					Name:      "spawn",
					Aliases:   []string{"s"},
					Usage:     "starts a new module",
					ArgsUsage: "module_name [key1=value1[,key2=value2[,...]]]",
					Action:    moduleSpawn,
				},
				{
					Name:    "list",
					Aliases: []string{"ls"},
					Usage:   "lists available or spawned modules",
					Action:  moduleList,
					Flags: []cli.Flag{
						cli.BoolFlag{
							Name: "spawned, s",
						},
					},
				},
			},
		},
	}

	app.Before = func(c *cli.Context) error {
		log.SetDebugVisible(c.Int("debug"))
		configPath = c.String("config")
		if configPath == "" {
			configPath = broker.ConfigPath
		}
		return nil
	}

	app.Run(os.Args)
}

func moduleSpawn(c *cli.Context) error {
	return nil
}

func moduleList(c *cli.Context) error {
	log.Print("Going to ask for list")
	rep, err := user.CLICmd(configPath, "list")
	if err != nil {
		return err
	}
	log.Print(rep)
	return nil
}
