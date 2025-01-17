// Copyright (c) 2024 BlueRock Security, Inc.

package src

// cSpell:ignore controlplane modfix

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
)

func Run() {

	var workPathArg string
	var goPathArg string
	var debugArg bool
	var recursiveArg bool
	var enableHiddenArg bool
	var verboseArg bool
	var dryRunArg bool
	var addAllArg bool

	command := &cobra.Command{}

	command.PersistentFlags().StringVarP(&workPathArg, "work-path", "w", ".", "The directory to process go.mod files")
	command.PersistentFlags().StringVarP(&goPathArg, "go-path", "g", "", "The directory to read go.mod files (default is $GOPATH)")
	command.PersistentFlags().BoolVarP(&recursiveArg, "recursive", "r", false, "Enables recursive search")
	command.PersistentFlags().BoolVar(&enableHiddenArg, "enable-hidden", false, "Enables reading of hidden directories in the go-path")
	command.PersistentFlags().BoolVarP(&debugArg, "debug", "d", false, "Enables debug output")
	command.PersistentFlags().BoolVarP(&verboseArg, "verbose", "v", false, "Enables verbose report")
	command.PersistentFlags().BoolVar(&dryRunArg, "dry-run", false, "Enables dry run (no changes will be made)")
	command.PersistentFlags().BoolVar(&addAllArg, "add-all", false, "Adds all existing replaces regardless of requirement")

	command.SilenceUsage = true

	rc := 0

	command.RunE = func(cmd *cobra.Command, args []string) error {

		config := &Config{}

		err := config.LoadFromEnv()
		if err != nil {
			return err
		}

		if workPathArg != "" {
			config.WorkPath = workPathArg
		}

		if goPathArg != "" {
			config.GoPath = goPathArg
		}

		if command.Flags().Lookup("debug").Changed {
			config.Debug = debugArg
		}

		if command.Flags().Lookup("recursive").Changed {
			config.Recursive = recursiveArg
		}

		if command.Flags().Lookup("enable-hidden").Changed {
			config.EnableHidden = enableHiddenArg
		}

		if command.Flags().Lookup("verbose").Changed {
			config.Verbose = verboseArg
		}

		if command.Flags().Lookup("dry-run").Changed {
			config.DryRun = dryRunArg
		}

		if command.Flags().Lookup("add-all").Changed {
			config.AddAll = addAllArg
		}

		report, err := Execute(config)
		if err != nil {
			rc = 1
			return err
		}

		if report.State == ReportStateFail {
			rc = 1
		}

		fmt.Fprint(os.Stderr, report.Text())

		return nil
	}

	ctx := context.Background()

	ctx, cancel := context.WithCancel(ctx)

	signals := make(chan os.Signal, 1)

	signal.Notify(signals, os.Interrupt, syscall.SIGHUP, syscall.SIGTERM)

	defer func() {
		signal.Stop(signals)
		cancel()
	}()

	go func() {
		select {

		case <-signals:
			cancel()

		case <-ctx.Done():

		}

	}()

	command.ExecuteContext(ctx)

	// tmp := recover()
	// if tmp != nil {
	// 	fmt.Fprintf(os.Stderr, "%v\n", tmp)
	// 	return 1
	// }

	if rc != 0 {
		os.Exit(rc)
	}
}
