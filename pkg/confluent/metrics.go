package confluent

import (
	"encoding/json"
	"fmt"
	"mcolomer/cloud-keeping/pkg/client" // Import the package that defines the HTTPS type
	"time"
)

type ConfluentCloudMetricsClient struct {
	client.HTTPS
	Cluster string
}

type MetricDescriptor struct {
	Metric string `json:"metric"`
}

type MetricFilter struct {
	Field string `json:"field"`
	Op    string `json:"op"`
	Value string `json:"value"`
}

type ConfluentCloudMetricsQuery struct {
	Aggregations []MetricDescriptor `json:"aggregations"`
	Filter       MetricFilter       `json:"filter"`
	Granularity  string             `json:"granularity"`
	GroupBy      []string           `json:"group_by"`
	Intervals    []string           `json:"intervals"`
	Limit        int                `json:"limit"`
}

const (
	// The Confluent Cloud Metrics API endpoint
	METRICS_ENDPOINT = "https://api.telemetry.confluent.cloud/v2/metrics/cloud/query"
	TOPIC            = "metric.topic"
	FIELD            = "resource.kafka.id"
	RECEIVED_RECORDS = "io.confluent.kafka.server/received_records"
	OPEREATION_EQ    = "eq"
	GRANULARITY_1_D  = "P1D"
	LIMIT            = 1000
)

func NewConfluentCloudMetricsClient(cluster, cluster_api_key, cluster_api_secret string) (*ConfluentCloudMetricsClient, error) {
	metricsClient := &ConfluentCloudMetricsClient{}
	metricsClient.HTTPS = *client.NewHTTPS(METRICS_ENDPOINT, cluster_api_key, cluster_api_secret)
	metricsClient.Cluster = cluster
	return metricsClient, nil
}

func (c *ConfluentCloudMetricsClient) GetActiveTopics() ([]string, error) {
	// Build query payload
	query := &ConfluentCloudMetricsQuery{
		Aggregations: []MetricDescriptor{
			{
				Metric: RECEIVED_RECORDS,
			},
		},
		Filter: MetricFilter{
			Field: FIELD,
			Op:    OPEREATION_EQ,
			Value: c.Cluster,
		},
		Granularity: GRANULARITY_1_D,
		GroupBy:     []string{TOPIC},
		Intervals:   []string{getLastWeekRange()},
		Limit:       LIMIT,
	}

	queryBytes, err := json.Marshal(query)
	if err != nil {
		return nil, err
	}

	response, err := c.HTTPS.Post(queryBytes)
	if err != nil {
		fmt.Printf("\nError getting active topics: %v", err)
		return nil, err
	}
	responseData := response.(map[string]interface{})
	// Extract the list of topics from the response
	topics := make([]string, 0)
	for _, row := range responseData["data"].([]interface{}) {
		topic := row.(map[string]interface{})[TOPIC].(string)
		topics = append(topics, topic)
	}

	return topics, nil

}

func getLastWeekRange() string {
	// Get the current time
	now := time.Now().UTC()

	// Calculate the start and end of the last week
	// Subtract 7 days from the current time for the start of the last week
	startOfLastWeek := now.AddDate(0, 0, -7)
	// The end of the last week is the current time
	endOfLastWeek := now

	// Format the time in the desired format
	startString := startOfLastWeek.Format(time.RFC3339)
	endString := endOfLastWeek.Format(time.RFC3339)

	// Return the result in the desired format
	return fmt.Sprintf("%s/%s", startString, endString)
}
