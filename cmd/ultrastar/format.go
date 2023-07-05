package main

import (
	"github.com/codello/ultrastar/txt"
	"github.com/spf13/cobra"
)

var formatCmd = &cobra.Command{
	Use:   "format",
	Short: "Format a Song",
	Long:  `Reads a song file and outputs a formatted version.`,
	Args:  cobra.NoArgs,
	RunE:  formatCommand,
}

func init() {
	rootCmd.AddCommand(formatCmd)
}

func formatCommand(cmd *cobra.Command, args []string) (err error) {
	input, err := inputFile()
	if err != nil {
		return err
	}
	defer input.Close()
	output, err := outputFile()
	if err != nil {
		return err
	}
	defer output.Close()

	song, err := txt.ReadSong(input)
	if err != nil {
		return err
	}
	err = txt.WriteSong(output, song)
	if err != nil {
		return err
	}
	return nil
}
