package main

import (
	"context"
	"fmt"
	"os"

	"github.com/shortmoose/toobig/internal/base"
	"github.com/shortmoose/toobig/internal/cmd"
	"github.com/shortmoose/toobig/internal/config"
	"github.com/urfave/cli/v3"
)

// This is ugly, not sure how to get the parent.
type do func(ctx *base.Context) error

func wrap_cfg(ct context.Context, cd *cli.Command, fn do) error {
	args := cd.Args().Slice()
	if len(args) != 1 {
		_ = cli.ShowCommandHelp(ct, cd.Root(), cd.Name)
		os.Exit(3)
	}

	var ctx base.Context
	ctx.Command = cd.Name
	ctx.ConfigPath = args[0]

	ctx.Verbose = cd.Bool("verbose")
	ctx.UpdateIsError = cd.Bool("update-is-error")
	ctx.FilePathOverride = cd.String("file-path")

	cfg, err := config.ReadConfig(ctx.ConfigPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(12)
	}
	ctx.TooBig = cfg

	return fn(&ctx)
}

func main() {
	app := &cli.Command{
		EnableShellCompletion: true,
		Name:                  "toobig",
		Usage:                 "manage large binary files (photos, videos, etc)",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "verbose",
				Aliases: []string{"v"},
				Value:   false,
			},
			&cli.BoolFlag{
				Name:  "update-is-error",
				Value: false,
			},
		},
		Commands: []*cli.Command{
			// TODO: Is there a way to specify these commands take a config file?
			{
				Name:  "update",
				Usage: "update blobs and metadata files to match files",
				Action: func(ctx context.Context, c *cli.Command) error {
					return wrap_cfg(ctx, c, cmd.Update)
				},
			},
			{
				Name:  "restore",
				Usage: "restore files to match blobs and metadata files",
				Action: func(ctx context.Context, c *cli.Command) error {
					return wrap_cfg(ctx, c, cmd.Restore)
				},
				Flags: []cli.Flag{
					&cli.StringFlag{
						Usage:    "path to restore TooBig files",
						Name:     "file-path",
						Required: true,
					},
				},
			},
			{
				Name:  "status",
				Usage: "current state of repository - are there current file changes",
				Action: func(ctx context.Context, c *cli.Command) error {
					return wrap_cfg(ctx, c, cmd.Status)
				},
			},
			{
				Name:  "fsck",
				Usage: "verify data integrity - are the refs and blobs valid",
				Action: func(ctx context.Context, c *cli.Command) error {
					return wrap_cfg(ctx, c, cmd.Fsck)
				},
			},
			{
				Name:  "config",
				Usage: "print an example config",
				Action: func(ctx context.Context, c *cli.Command) error {
					fmt.Println(`Here is an example config.
Add a path to a directory for each of these settings.
Remember all paths are relative to wherever the config file is located.`)
					fmt.Println("")
					config.WriteExampleConfig()
					return nil
				},
			},
			{
				Name:  "version",
				Usage: "print version",
				Action: func(ctx context.Context, c *cli.Command) error {
					fmt.Printf("Version %s\n", version)
					return nil
				},
			},
		},
	}

	err := app.Run(context.Background(), os.Args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)

		os.Exit(1)
	}
}
