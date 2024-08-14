package cleanup

import (
	"fmt"
	"mcolomer/cloud-keeping/pkg/confluent"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	environment        string
	cluster            string
	cluster_api_key    string
	cluster_api_secret string
	cloud_api_key      string
	cloud_api_secret   string
	confirm            bool
)

var confluentCmd = &cobra.Command{
	Use:     "confluent",
	Aliases: []string{"confluent, confl, cfl"},
	Short:   "Clean Confluent resources",
	Long:    ` Command to clean unused Confluent resources.`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

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

		fmt.Println(confirm) // Print the value of the confirm variable

		cflt.HandleInactiveTopics(confirm)
	},
}

func init() {
	viper.AutomaticEnv()
	// Flags
	topicsCmd.Flags().StringVarP(&environment, "environment", "", viper.GetString("ENVIRONMENT"), "Confluent Cloud environment Id (env-xxxxx) or set ENVIRONMENT environment variable")
	viper.BindPFlag("environment", topicsCmd.Flags().Lookup("environment"))

	topicsCmd.Flags().StringVarP(&cluster, "cluster", "", viper.GetString("CLUSTER"), "A Confluent Cloud cluster Id (lkc-xxxxx) or set CLUSTER environment variable")
	viper.BindPFlag("cluster", topicsCmd.Flags().Lookup("cluster"))

	topicsCmd.Flags().StringVarP(&cluster_api_key, "cluster_api_key", "", viper.GetString("CLUSTER_API_KEY"), "Cluster API KEY or set CLUSTER_API_KEY environment variable")
	viper.BindPFlag("cluster_api_key", topicsCmd.Flags().Lookup("cluster_api_key"))

	topicsCmd.Flags().StringVarP(&cluster_api_secret, "cluster_api_secret", "", viper.GetString("CLUSTER_API_SECRET"), "Cluster API SECRET or set CLOUD_API_KEY environment variable")
	viper.BindPFlag("cluster_api_secret", topicsCmd.Flags().Lookup("cluster_api_secret"))

	topicsCmd.Flags().StringVarP(&cloud_api_key, "cloud_api_key", "", viper.GetString("CLOUD_API_KEY"), "Cloud API KEY with Metrics API access or set CLOUD_API_KEY environment variable")
	viper.BindPFlag("cloud_api_key", topicsCmd.Flags().Lookup("cloud_api_key"))

	topicsCmd.Flags().StringVarP(&cloud_api_secret, "cloud_api_secret", "", viper.GetString("CLOUD_API_SECRET"), "Cloud API SECRET or set CLOUD_API_SECRET environment variable")
	viper.BindPFlag("cloud_api_secret", topicsCmd.Flags().Lookup("cloud_api_secret"))

	topicsCmd.Flags().BoolVarP(&confirm, "yes", "y", false, "Confirm delete - no prompt")
	viper.BindPFlag("yes", topicsCmd.Flags().Lookup("yes"))

	confluentCmd.AddCommand(topicsCmd)
	rootCmd.AddCommand(confluentCmd)
}

func Validate() bool {
	if environment == "" {
		fmt.Println("Environment required. Please provide the environment id (env-xxxxx), using the --environment flag or the ENVIRONMENT environment variable")
		return false
	}
	if cluster == "" {
		fmt.Println("Cluster required. Please provide the cluster id (lkc-xxxxx), using the --cluster flag or the CLUSTER environment variable")
		return false
	}
	if cluster_api_key == "" {
		fmt.Println("Cluster API KEY required. Please provide the cluster api key, using the --cluster_api_key flag or the CLUSTER_API_KEY environment variable")
		return false
	}
	if cluster_api_secret == "" {
		fmt.Println("Cluster API SECRET required. Please provide the cluster api secret, using the --cluster_api_secret flag or the CLUSTER_API_SECRET environment variable")
		return false
	}
	if cloud_api_key == "" {
		fmt.Println("Cloud API KEY required. Please provide the cloud api key, using the --cloud_api_key flag or the CLOUD_API_KEY environment variable")
		return false
	}
	if cloud_api_secret == "" {
		fmt.Println("Cloud API SECRET required. Please provide the cloud api secret, using the --cloud_api_secret flag or the CLOUD_API_SECRET environment variable")
		return false
	}
	return true
}
