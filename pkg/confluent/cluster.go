package confluent

import (
	"context"
	"fmt"
	"mcolomer/cloud-keeping/pkg/client"
	"mcolomer/cloud-keeping/pkg/table"
	"slices"
	"strings"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

type ConfluentCloudCluster struct {
	ClusterID         string
	BootstrapEndpoint string
	ApiEndpoint       string
	RestEndpoint      string
	ClusterAPI        client.HTTPS
	AdminClient       kafka.AdminClient
}

const (
	KAFKA_ENDPOINT  = "%s/kafka/v3/clusters/%s/topics"
	INTERNAL_PREFIX = "__"
	DATA            = "data"
	TOPIC_NAME      = "topic_name"
	SASL_SSL        = "SASL_SSL"
	PLAIN           = "PLAIN"
)

func NewKafkaCluster(cluster, bootstrap, rest_endpoint, cluster_api_key, cluster_api_secret string) (*ConfluentCloudCluster, error) {

	fmt.Println("Validating configuration")
	fmt.Println("Cluster: ", cluster)
	fmt.Println("Cluster API KEY: ", cluster_api_key)
	fmt.Println("Cloud API SECRET: ", cluster_api_secret)

	// Create a new AdminClient.
	config := &kafka.ConfigMap{
		"bootstrap.servers": bootstrap,
		"security.protocol": SASL_SSL,
		"sasl.mechanisms":   PLAIN,
		"sasl.username":     cluster_api_key,
		"sasl.password":     cluster_api_secret,
	}

	admin, err := kafka.NewAdminClient(config)
	if err != nil {
		fmt.Println("Failed to create Admin client:", err)
		return nil, err
	}

	client := client.NewHTTPS(rest_endpoint, cluster_api_key, cluster_api_secret)

	return &ConfluentCloudCluster{
		ClusterID:         cluster,
		BootstrapEndpoint: bootstrap,
		RestEndpoint:      rest_endpoint,
		ClusterAPI:        *client,
		AdminClient:       *admin,
	}, nil
}

func (c *ConfluentCloudCluster) GetTopics() ([]string, error) {
	c.ClusterAPI.Endpoint = fmt.Sprintf(KAFKA_ENDPOINT, c.RestEndpoint, c.ClusterID)
	response, err := c.ClusterAPI.Get()
	if err != nil {
		fmt.Printf("\n Error getting topics: %v", err)
		return nil, err
	}
	responseData := response.(map[string]interface{})
	spec := responseData[DATA].([]interface{})
	var topics []string
	for _, row := range spec {
		topic := row.(map[string]interface{})[TOPIC_NAME].(string)
		if !strings.HasPrefix(topic, INTERNAL_PREFIX) {
			topics = append(topics, topic)
		}
	}
	return topics, nil
}

func (c *ConfluentCloudCluster) DeleteInactiveTopics(activeTopics, topics []string) {
	var inactiveTopics []string
	for _, topic := range topics {
		if !slices.Contains(activeTopics, topic) {
			inactiveTopics = append(inactiveTopics, topic)
		}
	}
	c.DeleteTopics(inactiveTopics)
}

func (c *ConfluentCloudCluster) DeleteTopics(topics []string) {
	// Contexts are used to abort or limit the amount of time
	// the Admin call blocks waiting for a result.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Delete topics on cluster
	// Set Admin options to wait for the operation to finish (or at most 60s)
	maxDur, err := time.ParseDuration("60s")
	if err != nil {
		fmt.Println("Error:: ParseDuration(60s):: ", err)
	}
	results, err := c.AdminClient.DeleteTopics(ctx, topics, kafka.SetAdminOperationTimeout(maxDur))
	if err != nil {
		fmt.Printf("Failed to delete topics: %v\n", err)
	}

	// Print results
	var deletedTopics []string
	for _, result := range results {
		deletedTopics = append(deletedTopics, result.Topic)
	}
	if len(deletedTopics) == 0 {
		fmt.Println("No topics deleted")
	} else {
		c.TopicsTable(deletedTopics)
	}
}

func (c *ConfluentCloudCluster) TopicsTable(topics []string) {
	y := make([][]interface{}, len(topics))
	for i, topic := range topics {
		row := make([]interface{}, 2)
		row[0] = topic
		row[1] = "DELETED"
		y[i] = row
	}
	table.NewTable([]string{"Topic", ""}, y)
}

func (c *ConfluentCloudCluster) InactiveTopicsTable(activeTopics []string, topics []string) []string {
	allTopics := make([][]interface{}, len(topics))
	var inactiveTopics []string

	for i, topic := range topics {
		row := make([]interface{}, 2)
		row[0] = topic
		if slices.Contains(activeTopics, topic) {
			row[1] = "YES"
		} else {
			row[1] = "EMPTY"
			inactiveTopics = append(inactiveTopics, topic)
		}
		allTopics[i] = row
	}
	table.NewTable([]string{"Topic", "Active (Last 7 Days)"}, allTopics)
	return inactiveTopics
}
