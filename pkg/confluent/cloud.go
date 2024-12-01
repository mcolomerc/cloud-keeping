package confluent

import (
	"fmt"
	"mcolomer/cloud-keeping/pkg/client" // Import the package that defines the HTTPS type
	"strings"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

type ConfluentCloudClient struct {
	client.HTTPS
	KafkaCluster ConfluentCloudCluster
	Environment  string
	ClusterID    string
}

type ConfluentCloudRoleBinding struct {
	Role      string
	Resource  string
	Principal string
	Id        string
}

func NewConfluentCloudClient(environment, cluster, cluster_api_key, cluster_api_secret, cloud_api_key, cloud_api_secret string) (*ConfluentCloudClient, error) {

	confluentClient := &ConfluentCloudClient{}
	cloudUrl := fmt.Sprintf(CONFLUENT_ENDPOINT+CLUSTER, cluster, environment)

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

// Kafka Cluster
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
	metadata := responseData["metadata"].(map[string]interface{})
	resource_name := metadata["resource_name"].(string)
	// Find the index of "/kafka="
	kafkaIndex := strings.Index(resource_name, "/kafka=")
	// If "/kafka=" is found, slice the string up to that index
	if kafkaIndex != -1 {
		resource_name = resource_name[:kafkaIndex]
	}
	kafkaCluster, err := NewKafkaCluster(c.ClusterID, kafka_bootstrap_endpoint, rest_endpoint, cluster_api, cluster_secret, resource_name)
	if err != nil {
		fmt.Println("Error getting cluster")
		return nil, err
	}
	return kafkaCluster, nil
}

// TOPICS
func (c *ConfluentCloudClient) GetTopics() ([]string, error) {
	return c.KafkaCluster.GetTopics()
}
func (c *ConfluentCloudClient) DeleteTopics(topics []string) []string {
	return c.KafkaCluster.DeleteTopics(topics)
}

// ACLS
func (c *ConfluentCloudClient) GetACLs() ([]kafka.ACLBinding, error) {
	return c.KafkaCluster.GetACLs()
}

func (c *ConfluentCloudClient) DeleteACLs(acls []kafka.ACLBinding) ([]kafka.DescribeACLsResult, error) {
	return c.KafkaCluster.DeleteACLs(acls)
}

// API_KEYS
func (c *ConfluentCloudClient) GetClusterApiKeys() (map[string][]string, error) {
	apiKeys := make(map[string][]string)
	c.HTTPS.Endpoint = fmt.Sprintf(CONFLUENT_ENDPOINT+CLUSTER_API_KEYS, c.ClusterID)
	response, err := c.HTTPS.Get()
	if err != nil {
		fmt.Printf("\n Error getting cluster API keys: %v", err)
		return nil, err
	}
	responseData := response.(map[string]interface{})
	for _, apiKey := range responseData["data"].([]interface{}) {
		id := apiKey.(map[string]interface{})["id"].(string)
		spec := apiKey.(map[string]interface{})["spec"].(map[string]interface{})
		owner := spec["owner"].(map[string]interface{})
		ownerid := owner["id"].(string)
		apiKeys[ownerid] = append(apiKeys[ownerid], id)
	}
	return apiKeys, nil
}

func (c *ConfluentCloudClient) DeleteApiKeys(apiKeys []string) error {
	fmt.Println("\n Deleting API keys: ", apiKeys)
	c.HTTPS.Endpoint = fmt.Sprintf(CONFLUENT_ENDPOINT + API_KEYS)
	for _, key := range apiKeys {
		c.HTTPS.Endpoint = fmt.Sprintf("%s/%s", c.HTTPS.Endpoint, key)
		_, err := c.HTTPS.Delete()
		if err != nil {
			fmt.Printf("\n Error deleting API keys: %v", err)
			return err
		}
	}
	return nil
}

// RBAC
func (c *ConfluentCloudClient) GetRoleBindings(principal string) ([]ConfluentCloudRoleBinding, error) {
	c.HTTPS.Endpoint = fmt.Sprintf(CONFLUENT_ENDPOINT+RBAC_ENDPOINT, principal, c.KafkaCluster.CrnPatern)
	response, err := c.HTTPS.Get()
	if err != nil {
		fmt.Printf("\n Error getting role bindings: %v", err)
		return nil, err
	}
	responseData := response.(map[string]interface{})
	roleBindings := make([]ConfluentCloudRoleBinding, 0)
	for _, roleBinding := range responseData["data"].([]interface{}) {
		roleBindingData := roleBinding.(map[string]interface{})
		roleBindings = append(roleBindings, ConfluentCloudRoleBinding{
			Role:      roleBindingData["role_name"].(string),
			Resource:  roleBindingData["crn_pattern"].(string),
			Principal: roleBindingData["principal"].(string),
			Id:        roleBindingData["id"].(string),
		})
	}
	return roleBindings, nil
}

func (c *ConfluentCloudClient) DeleteRoleBindings(roleBindings []string) error {
	for _, roleBinding := range roleBindings {
		c.HTTPS.Endpoint = fmt.Sprintf(CONFLUENT_ENDPOINT+"/iam/v2/role-bindings/%s", roleBinding)
		_, err := c.HTTPS.Delete()
		if err != nil {
			fmt.Printf("\n Error deleting role bindings: %v", err)
			return err
		}
	}
	return nil
}

func (c *ConfluentCloudClient) GetConnectors() ([]ConfluentCloudConnector, error) {
	c.HTTPS.Endpoint = fmt.Sprintf("https://api.confluent.cloud/connect/v1/environments/%s/clusters/%s/connectors?expand=info,status,id", c.Environment, c.ClusterID)
	response, err := c.HTTPS.Get()
	if err != nil {
		fmt.Printf("\n Error getting connectors: %v", err)
		return nil, err
	}
	responseData := response.(map[string]interface{})
	connectors := make([]ConfluentCloudConnector, 0)
	for _, connector := range responseData {
		connectorData := connector.(map[string]interface{})
		info := connectorData["info"].(map[string]interface{})
		status := connectorData["status"].(map[string]interface{})
		id := connectorData["id"].(map[string]interface{})
		connector := status["connector"].(map[string]interface{})

		connectors = append(connectors, ConfluentCloudConnector{
			Name:  info["name"].(string),
			Id:    id["id"].(string),
			State: ConnectorState(connector["state"].(string)),
			Type:  connector["type"].(string),
		})
	}
	return connectors, nil
}
