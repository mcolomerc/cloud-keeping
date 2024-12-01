package confluent

import (
	"context"
	"fmt"
	"mcolomer/cloud-keeping/pkg/client"

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
	CrnPatern         string
}

func NewKafkaCluster(cluster, bootstrap, rest_endpoint, cluster_api_key, cluster_api_secret, crn_pattern string) (*ConfluentCloudCluster, error) {

	fmt.Println("\n Validating cluster configuration. ")
	fmt.Println("  - Cluster: ", cluster)
	fmt.Println("  - Bootstrap: ", bootstrap)
	fmt.Println("  - Cluster API KEY: ", cluster_api_key)

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
		CrnPatern:         crn_pattern,
	}, nil
}

// TOPICS
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

func (c *ConfluentCloudCluster) DeleteTopics(topics []string) []string {
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
		return nil
	}
	return deletedTopics
}

// ACLs
func (c *ConfluentCloudCluster) GetACLs() ([]kafka.ACLBinding, error) {
	c.ClusterAPI.Endpoint = fmt.Sprintf(ACL_ENDPOINT, c.RestEndpoint, c.ClusterID)
	fmt.Println("\n Getting ACLs")
	fmt.Println("  - Endpoint: ", c.ClusterAPI.Endpoint)
	response, err := c.ClusterAPI.Get()
	if err != nil {
		fmt.Printf("\n Error getting ACLs: %v", err)
		return nil, err
	}

	responseData := response.(map[string]interface{})
	spec := responseData[DATA].([]interface{})

	var acls []kafka.ACLBinding
	for _, row := range spec {
		var resource_type kafka.ResourceType
		var resource_pattern_type kafka.ResourcePatternType
		var operation kafka.ACLOperation
		var permission kafka.ACLPermissionType

		if row.(map[string]interface{})["resource_type"] != nil {
			rType := row.(map[string]interface{})["resource_type"].(string)
			if rType == "CLUSTER" {
				rType = "BROKER"
			}
			res_type, err := kafka.ResourceTypeFromString(rType)
			if err != nil {
				fmt.Printf("Invalid resource type: %s: %v\n", row.(map[string]interface{})["resource_type"].(string), err)
				return nil, err
			}
			resource_type = res_type
		}
		if row.(map[string]interface{})["pattern_type"] != nil {
			pattern_type, err := kafka.ResourcePatternTypeFromString(row.(map[string]interface{})["pattern_type"].(string))
			if err != nil {
				fmt.Printf("Invalid resource pattern type: %s: %v\n", row.(map[string]interface{})["pattern_type"].(string), err)
				return nil, err
			}
			resource_pattern_type = pattern_type
		}
		if row.(map[string]interface{})["operation"] != nil {
			oper, err := kafka.ACLOperationFromString(row.(map[string]interface{})["operation"].(string))
			if err != nil {
				fmt.Printf("Invalid operation: %s: %v\n", row.(map[string]interface{})["operation"].(string), err)
				return nil, err
			}
			operation = oper
		}
		if row.(map[string]interface{})["permission"] != nil {
			perm, err := kafka.ACLPermissionTypeFromString(row.(map[string]interface{})["permission"].(string))
			if err != nil {
				fmt.Printf("Invalid permission: %s: %v\n", row.(map[string]interface{})["permission"].(string), err)
				return nil, err
			}
			permission = perm
		}
		var name string
		if row.(map[string]interface{})["resource_name"] != nil {
			name = row.(map[string]interface{})["resource_name"].(string)
		}
		var principal string
		if row.(map[string]interface{})["principal"] != nil {
			principal = row.(map[string]interface{})["principal"].(string)
		}
		var host string
		if row.(map[string]interface{})["host"] != nil {
			host = row.(map[string]interface{})["host"].(string)
		}
		if name == "" || principal == "" || host == "" {
			fmt.Println("Invalid ACL entry")
			continue
		}
		acl := kafka.ACLBinding{
			Type:                resource_type,
			Name:                name,
			ResourcePatternType: resource_pattern_type,
			Principal:           principal,
			Host:                host,
			Operation:           operation,
			PermissionType:      permission,
		}
		acls = append(acls, acl)
	}
	return acls, nil
}

func (c *ConfluentCloudCluster) DeleteACLs(acls []kafka.ACLBinding) ([]kafka.DescribeACLsResult, error) {
	// Contexts are used to abort or limit the amount of time
	// the Admin call blocks waiting for a result.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Delete ACLs on cluster
	// Set Admin options to wait for the operation to finish (or at most 60s)
	maxDur, err := time.ParseDuration("60s")
	if err != nil {
		fmt.Println("Error:: ParseDuration(60s):: ", err)
	}
	results, err := c.AdminClient.DeleteACLs(ctx, acls, kafka.SetAdminRequestTimeout(maxDur))
	if err != nil {
		fmt.Printf("Failed to delete ACLs: %v\n", err)
	}
	return results, nil
}
