package main

import (
	"fmt"
	"os"

	"github.com/shortmoose/toobig/internal/base"
	"github.com/shortmoose/toobig/internal/cmd"
	"github.com/shortmoose/toobig/internal/config"
	"github.com/urfave/cli/v2"
)

type do func(ctx *base.Context) error

func wrap_cfg(c *cli.Context, fn do) error {
	args := c.Args().Slice()
	if len(args) != 1 {
		cli.ShowCommandHelpAndExit(c, c.Command.Name, 1)
	}

	var ctx base.Context
	ctx.Command = c.Command.Name
	ctx.ConfigPath = args[0]

	ctx.DryRun = c.Bool("dry-run")
	ctx.Verbose = c.Bool("verbose")

	// TODO: Validate the configuration.
	cfg, err := config.ReadConfig(ctx.ConfigPath)
	if err != nil {
		return err
	}
	ctx.TooBig = cfg

	return fn(&ctx)
}

func wrap0(c *cli.Context, fn do) error {
	args := c.Args().Slice()
	if len(args) != 0 {
		cli.ShowCommandHelpAndExit(c, c.Command.Name, 1)
	}

	var ctx base.Context
	ctx.Command = c.Command.Name

	return fn(&ctx)
}

func main() {
	app := &cli.App{
		Usage: "manage large binary files (photos, videos, etc)",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "verbose",
				Aliases: []string{"v"},
				Value:   false,
				Usage:   "Usage ...",
			},
			/*
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
				Usage: "update blobs and metadata files to match files",
				Action: func(c *cli.Context) error {
					return wrap_cfg(c, cmd.Update)
				},
			},
			{
				Name:  "restore",
				Usage: "restore files to match blobs and metadata files",
				Action: func(c *cli.Context) error {
					return wrap_cfg(c, cmd.Restore)
				},
			},
			{
				Name:  "fsck",
				Usage: "verify data integrity",
				Action: func(c *cli.Context) error {
					return wrap_cfg(c, cmd.Fsck)
				},
			},
			{
				Name:  "status",
				Usage: "info about current state",
				Action: func(c *cli.Context) error {
					return wrap_cfg(c, cmd.Status)
				},
			},
			{
				Name:  "config",
				Usage: "print an example config",
				Action: func(c *cli.Context) error {
					return wrap0(c, cmd.Config)
				},
			},
			{
				Name:  "version",
				Usage: "print version",
				Action: func(c *cli.Context) error {
					fmt.Printf("Version %s\n", version)
					return nil
				},
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed: %s\n", err)

		os.Exit(1)
	}
}
