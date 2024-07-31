package confluent

import (
	"fmt"
	"mcolomer/cloud-keeping/pkg/client" // Import the package that defines the HTTPS type
)

type ConfluentCloudClient struct {
	client.HTTPS
	KafkaCluster ConfluentCloudCluster
	Environment  string
	ClusterID    string
}

const (
	// The Confluent Cloud Metrics API endpoint
	CONFLUENT_ENDPOINT = "https://api.confluent.cloud/cmk/v2/clusters/%s?environment=%s"
)

func NewConfluentCloudClient(environment, cluster, cluster_api_key, cluster_api_secret, cloud_api_key, cloud_api_secret string) (*ConfluentCloudClient, error) {

	confluentClient := &ConfluentCloudClient{}
	cloudUrl := fmt.Sprintf(CONFLUENT_ENDPOINT, cluster, environment)

	confluentClient.HTTPS = *client.NewHTTPS(cloudUrl, cloud_api_key, cloud_api_secret)
	confluentClient.Environment = environment
	confluentClient.ClusterID = cluster

	kafkaCluster, err := confluentClient.GetKafkaCluster(cluster_api_key, cluster_api_secret)
	if err != nil {
		fmt.Println("Error getting cluster")
		return nil, err
	}
	confluentClient.KafkaCluster = *kafkaCluster
	return confluentClient, nil
}

func (c *ConfluentCloudClient) GetKafkaCluster(cluster_api string, cluster_secret string) (*ConfluentCloudCluster, error) {
	response, err := c.HTTPS.Get()
	if err != nil {
		fmt.Printf("\n Error getting cluster : %v", err)
		return nil, err
	}
	responseData := response.(map[string]interface{})
	spec := responseData["spec"].(map[string]interface{})
	kafka_bootstrap_endpoint := spec["kafka_bootstrap_endpoint"].(string)
	rest_endpoint := spec["http_endpoint"].(string)
	kafkaCluster, err := NewKafkaCluster(c.ClusterID, kafka_bootstrap_endpoint, rest_endpoint, cluster_api, cluster_secret)
	if err != nil {
		fmt.Println("Error getting cluster")
		return nil, err
	}
	return kafkaCluster, nil

}

func (c *ConfluentCloudClient) GetTopics() ([]string, error) {
	return c.KafkaCluster.GetTopics()
}

func (c *ConfluentCloudClient) DeleteTopics(topics []string) {
	c.KafkaCluster.DeleteTopics(topics)
}

func (c *ConfluentCloudClient) GetInactiveTopics(activeTopics []string, topics []string) []string {
	return c.KafkaCluster.InactiveTopicsTable(activeTopics, topics)
}
