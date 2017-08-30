package cli

import (
	"flag"
)

type CommandLineOptions struct {
	Config      string
	ShowVersion bool
}

func SetupFlagSet(name string, options *CommandLineOptions) *flag.FlagSet {
	flagSet := flag.NewFlagSet(name, flag.ContinueOnError)
	flagSet.StringVar(&options.Config, "config", "", "configuration file path")
	flagSet.BoolVar(&options.ShowVersion, "version", false, "show version")
	return flagSet
}
