package main

import (
	"github.com/spf13/cobra"
	"os"
)

var (
	file    string
	inPlace bool

	rootCmd = &cobra.Command{
		Use:           "ultrastar",
		Short:         "UltraStar is a command line application for interacting with UltraStar Songs.",
		Long:          `A versatile command line application for interacting with UltraStar songs Unix-style.`,
		Args:          cobra.NoArgs,
		SilenceUsage:  true,
		SilenceErrors: true,
	}
)

func init() {
	rootCmd.PersistentFlags().StringVarP(&file, "file", "f", "", "input file.")
	rootCmd.PersistentFlags().BoolVarP(&inPlace, "inplace", "i", false, "write output to --file instead of stdout.")
	_ = rootCmd.MarkPersistentFlagFilename("file", "txt")
}

func inputFile() (*os.File, error) {
	if file != "" && file != "-" {
		return os.Open(file)
	}
	return os.Stdin, nil
}

func outputFile() (*os.File, error) {
	return os.Stdout, nil
}
