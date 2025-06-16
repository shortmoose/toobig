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

func is_dir(dir string) bool {
	fileInfo, err := os.Stat(dir)
	if err != nil || !fileInfo.IsDir() {
		return false
	}

	return true
}

func wrap_cfg(ct context.Context, cd *cli.Command, fn do) error {
	args := cd.Args().Slice()
	if len(args) != 1 {
		_ = cli.ShowCommandHelp(ct, cd.Root(), cd.Name)
		os.Exit(3)
	}

	var ctx base.Context
	ctx.Command = cd.Name
	ctx.ConfigPath = args[0]

	ctx.DryRun = cd.Bool("dry-run")
	ctx.Verbose = cd.Bool("verbose")

	// TODO: Validate the configuration.
	cfg, err := config.ReadConfig(ctx.ConfigPath)
	if err != nil {
		return err
	}
	ctx.TooBig = cfg

	if !is_dir(cfg.FilePath) {
		fmt.Println("Error: file_path", cfg.FilePath, "is not a directory.")
		os.Exit(1)
	}
	if !is_dir(cfg.BlobPath) {
		fmt.Println("Error: blob_path", cfg.BlobPath, "is not a directory.")
		os.Exit(1)
	}
	if !is_dir(cfg.RefPath) {
		fmt.Println("Error: ref_path", cfg.RefPath, "is not a directory.")
		os.Exit(1)
	}
	if !is_dir(cfg.DupPath) {
		fmt.Println("Error: dup_path", cfg.DupPath, "is not a directory.")
		os.Exit(1)
	}

	return fn(&ctx)
}

func wrap0(ct context.Context, cd *cli.Command, fn do) error {
	args := cd.Args().Slice()
	if len(args) != 0 {
		_ = cli.ShowCommandHelp(ct, cd.Root(), cd.Name)
		os.Exit(1)
	}

	var ctx base.Context
	ctx.Command = cd.Name

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
					return wrap0(ctx, c, cmd.Config)
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
		fmt.Fprintf(os.Stderr, "Failed: %s\n", err)

		os.Exit(1)
	}
}
