package cleanup

import (
	"fmt"
	"mcolomer/cloud-keeping/pkg/confluent"
	"os"

	"github.com/spf13/cobra"
)

var topicsCmd = &cobra.Command{
	Use:     "topics",
	Aliases: []string{"tpcs"},
	Short:   "Clean Topics ",
	Long:    ` Command to Clean Confluent Cloud Topics.`,
	Run: func(cmd *cobra.Command, args []string) {
		if !Validate() {
			fmt.Println("Error validating configuration.")
			cmd.Help()
			os.Exit(1)
		}
		cflt := confluent.NewConfluentClean(environment, cluster, cluster_api_key, cluster_api_secret, cloud_api_key, cloud_api_secret)
		cflt.HandleInactiveTopics(confirm)
	},
}
