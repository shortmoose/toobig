package main

import (
	"fmt"
	"log"
	"os"

	"github.com/shortmoose/toobig/internal/base"
	"github.com/shortmoose/toobig/internal/cmd"
	"github.com/urfave/cli/v2"
)

type do func(ctx *base.Context) error

func wrap(c *cli.Context, fn do) error {
	// TODO: Do we want to allow multiple configs?
	args := c.Args().Slice()
	if len(args) != 1 {
		cli.ShowCommandHelpAndExit(c, c.Command.Name, 1)
	}

	for _, path := range args {
		var ctx base.Context
		ctx.Command = c.Command.Name
		ctx.ConfigPath = path

		ctx.DryRun = c.Bool("dry-run")
		ctx.Verbose = c.Bool("verbose")

		e := fn(&ctx)
		if e != nil {
			return e
		}
	}

	return nil
}

func main() {
	app := &cli.App{
		Flags: []cli.Flag{
			/*
				&cli.BoolFlag{
					Name:    "verbose",
					Aliases: []string{"v"},
					Value:   false,
					Usage:   "Usage ...",
				},
				&cli.BoolFlag{
					Name:  "dry-run",
					Value: false,
					Usage: "Usage ...",
				},
			*/
		},
		Commands: []*cli.Command{
			{
				Name:  "update",
				Usage: "update repos to match blobs",
				Action: func(c *cli.Context) error {
					return wrap(c, cmd.Update)
				},
			},
			{
				Name:  "restore",
				Usage: "restore blobs to match repos",
				Action: func(c *cli.Context) error {
					return wrap(c, cmd.Restore)
				},
			},
			{
				Name:  "fsck",
				Usage: "verify data integrity",
				Action: func(c *cli.Context) error {
					return wrap(c, cmd.Fsck)
				},
			},
			{
				Name:  "status",
				Usage: "info about current state",
				Action: func(c *cli.Context) error {
					return wrap(c, cmd.Status)
				},
			},
			{
				Name:  "version",
				Usage: "print version",
				Action: func(c *cli.Context) error {
					fmt.Printf("Version %s\n", VERSION)
					return nil
				},
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
