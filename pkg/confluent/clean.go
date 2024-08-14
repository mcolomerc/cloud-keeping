package confluent

import (
	"fmt"
	"os"

	"github.com/manifoldco/promptui"
)

type ConfluentClean struct {
	Environment, Cluster, Cluster_api_key, Cluster_api_secret, Cloud_api_key, Cloud_api_secret string
}

func NewConfluentClean(environment, cluster, cluster_api_key, cluster_api_secret, cloud_api_key, cloud_api_secret string) *ConfluentClean {
	return &ConfluentClean{
		Environment:        environment,
		Cluster:            cluster,
		Cluster_api_key:    cluster_api_key,
		Cluster_api_secret: cluster_api_secret,
		Cloud_api_key:      cloud_api_key,
		Cloud_api_secret:   cloud_api_secret,
	}
}

func (c *ConfluentClean) HandleInactiveTopics(confirm bool) {
	fmt.Println("\n Detecting inactive Topics...")
	cfltMetrics, err := NewConfluentCloudMetricsClient(c.Cluster, c.Cloud_api_key, c.Cloud_api_secret)
	if err != nil {
		fmt.Println("Error creating Confluent Cloud Metrics Client")
		os.Exit(1)
	}
	activeTopics, err := cfltMetrics.GetActiveTopics()
	if err != nil {
		fmt.Println("Error getting active topics")
		os.Exit(1)
	}
	confluentApi, err := NewConfluentCloudClient(c.Environment, c.Cluster, c.Cluster_api_key, c.Cluster_api_secret, c.Cloud_api_key, c.Cloud_api_secret)
	if err != nil {
		fmt.Println("Error getting active topics")
		os.Exit(1)
	}
	topics, err := confluentApi.GetTopics()
	if err != nil {
		fmt.Println("Error getting topics")
		os.Exit(1)
	}

	inactiveTopics := confluentApi.GetInactiveTopics(activeTopics, topics)
	if len(inactiveTopics) > 0 {
		if !confirm {
			prompt := promptui.Prompt{
				Label:     "Delete all inactive topics?",
				IsConfirm: true,
			}

			result, err := prompt.Run()
			if err != nil {
				return
			}
			if result == "y" {
				confirm = true
			}
		} else {
			confluentApi.DeleteTopics(inactiveTopics)
		}
	} else {
		fmt.Println("No inactive topics found.")
	}
}

func (c *ConfluentClean) InactiveConnectors() {
	fmt.Println("InactiveConnectors")
}
