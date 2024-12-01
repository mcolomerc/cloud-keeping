package confluent

import (
	"encoding/json"
	"fmt"
	"mcolomer/cloud-keeping/pkg/client" // Import the package that defines the HTTPS type
	"strings"
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
	METRIC_TOPIC     = "metric.topic"
	METRIC_PRINCIPAL = "metric.principal_id"

	METRICS_RECEIVED_RECORDS   = "io.confluent.kafka.server/received_records"
	METRICS_ACTIVE_CONNECTIONS = "io.confluent.kafka.server/request_count"

	FIELD           = "resource.kafka.id"
	OPEREATION_EQ   = "eq"
	GRANULARITY_1_D = "P1D"
	LIMIT           = 1000

	SERVICE_ACCOUNT = "sa-"
)

func NewConfluentCloudMetricsClient(cluster, cluster_api_key, cluster_api_secret string) (*ConfluentCloudMetricsClient, error) {
	metricsClient := &ConfluentCloudMetricsClient{}
	metricsClient.HTTPS = *client.NewHTTPS(METRICS_ENDPOINT, cluster_api_key, cluster_api_secret)
	metricsClient.Cluster = cluster
	return metricsClient, nil
}

func (c *ConfluentCloudMetricsClient) QueryMetric(metric string, group string) (map[string]interface{}, error) {
	query := &ConfluentCloudMetricsQuery{
		Aggregations: []MetricDescriptor{
			{
				Metric: metric,
			},
		},
		Filter: MetricFilter{
			Field: FIELD,
			Op:    OPEREATION_EQ,
			Value: c.Cluster,
		},
		Granularity: GRANULARITY_1_D,
		GroupBy:     []string{group},
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
	return response.(map[string]interface{}), nil
}

/** io.confluent.kafka.server/request_count
 *  This metric is useful to identify the number of active clients connected to the cluster.
 *  Returns CC users as well (u-xxxxx)
 */
func (c *ConfluentCloudMetricsClient) GetConnectionsByServiceAccount() (map[string]float64, error) {
	// Extract the list of topics from the response
	principals := make(map[string]float64, 0)
	responseData, err := c.QueryMetric(METRICS_ACTIVE_CONNECTIONS, METRIC_PRINCIPAL)
	if err != nil {
		return principals, err
	}
	for _, row := range responseData["data"].([]interface{}) {
		principal := row.(map[string]interface{})[METRIC_PRINCIPAL].(string)
		value := row.(map[string]interface{})["value"].(float64)
		if strings.Contains(principal, SERVICE_ACCOUNT) {
			principals[principal] = principals[principal] + value
		}
	}
	return principals, nil
}

func (c *ConfluentCloudMetricsClient) GetActiveTopics() ([]string, error) {
	// Extract the list of topics from the response
	topics := make([]string, 0)
	// Build query payload
	responseData, err := c.QueryMetric(METRICS_RECEIVED_RECORDS, METRIC_TOPIC)
	if err != nil {
		return topics, err
	}
	for _, row := range responseData["data"].([]interface{}) {
		topic := row.(map[string]interface{})[METRIC_TOPIC].(string)
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
