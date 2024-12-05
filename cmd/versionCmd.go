package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of repr8ducer",
	Long:  `All software has versions. This is repr8ducer's`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Repr8ducer")
	},
}
