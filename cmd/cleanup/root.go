package cleanup

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var version = "0.0.1"

var rootCmd = &cobra.Command{
	Use:     "cleanup",
	Aliases: []string{"cleanup, clean, cln"},
	Version: version,
	Short:   "cleanup - a simple CLI to clean unused respurces",
	Long:    `a simple CLI to clean unused respurces`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Run cleanup without command.")
		cmd.Help()
	},
}

func Execute() {

	if err := rootCmd.Execute(); err != nil {
		fmt.Errorf("Error executing command: %v", err)
		os.Exit(1)
	}
}
