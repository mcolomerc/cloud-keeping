package confluent

import (
	"fmt"
	"mcolomer/cloud-keeping/pkg/commons"
	"mcolomer/cloud-keeping/pkg/outputs"
	"os"
	"slices"
	"strings"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

type ConfluentClean struct {
	MetricsAPI *ConfluentCloudMetricsClient
	CloudAPI   *ConfluentCloudClient
}

func NewConfluentClean(environment, cluster, cluster_api_key, cluster_api_secret, cloud_api_key, cloud_api_secret string) *ConfluentClean {
	cfltMetrics, err := NewConfluentCloudMetricsClient(cluster, cloud_api_key, cloud_api_secret)
	if err != nil {
		fmt.Println("Error creating Confluent Cloud Metrics Client")
		os.Exit(1)
	}
	confluentApi, err := NewConfluentCloudClient(environment, cluster, cluster_api_key, cluster_api_secret, cloud_api_key, cloud_api_secret)
	if err != nil {
		fmt.Println("Error getting active topics")
		os.Exit(1)
	}
	return &ConfluentClean{
		MetricsAPI: cfltMetrics,
		CloudAPI:   confluentApi,
	}
}

// Clean Topics
func (c *ConfluentClean) HandleInactiveTopics(confirm bool) {
	fmt.Println("\n Detecting inactive Topics...")

	activeTopicsCh := commons.AsyncCall(func() ([]string, error) {
		return c.MetricsAPI.GetActiveTopics()
	})

	topicsCh := commons.AsyncCall(func() ([]string, error) {
		return c.CloudAPI.GetTopics()
	})

	activeTopics := <-activeTopicsCh
	topics := <-topicsCh

	if activeTopics.Err != nil || topics.Err != nil {
		fmt.Println("Error getting active topics")
		os.Exit(1)
	}

	inactiveTopics := getInactiveTopics(activeTopics.Result, topics.Result)
	if len(inactiveTopics) > 0 {
		if commons.BuildConfirmationPrompt("Delete all inactive topics?", confirm) {
			deleted := c.CloudAPI.DeleteTopics(inactiveTopics)
			if len(deleted) > 0 {
				outputs.BuildList(deleted, "Topics Deleted")
			}
		}
	} else {
		fmt.Println("No inactive topics found.")
	}
}
func getInactiveTopics(activeTopics []string, topics []string) []string {
	fmt.Printf("\n Building Topic status, in the last 7 days... \n")
	allTopics := make([][]interface{}, len(topics))
	var inactiveTopics []string
	for i, topic := range topics {
		row := make([]interface{}, 2)
		row[0] = topic
		if slices.Contains(activeTopics, topic) {
			row[1] = ActiveStatus
		} else {
			row[1] = InactiveStatus
			inactiveTopics = append(inactiveTopics, topic)
		}
		allTopics[i] = row
	}
	outputs.NewTable([]string{"Topic", "Active (Last 7 Days)"}, allTopics)
	return inactiveTopics
}

// Clean ACLS
func (c *ConfluentClean) HandleInactiveACLs(confirm bool) {
	fmt.Println("Inactive ACLs")

	topicsCh := commons.AsyncCall(func() ([]string, error) {
		return c.CloudAPI.GetTopics()
	})

	aclsCh := commons.AsyncCall(func() ([]kafka.ACLBinding, error) {
		return c.CloudAPI.GetACLs()
	})

	topics := <-topicsCh
	acls := <-aclsCh

	if topics.Err != nil || acls.Err != nil {
		fmt.Println("Error getting resources")
		os.Exit(1)
	}

	allACLS := make([][]interface{}, len(acls.Result))
	inactiveACls := make([]kafka.ACLBinding, 0)

	for i, acl := range acls.Result {
		row := make([]interface{}, 8)
		aclBindingsToRow(row, acl)
		row[7] = ActiveStatus
		if acl.Type == kafka.ResourceTopic {
			if acl.ResourcePatternType == kafka.ResourcePatternTypeLiteral && !slices.Contains(topics.Result, acl.Name) {
				row[7] = TopicNotFoundStatus
				inactiveACls = append(inactiveACls, acl)
			}
			if acl.ResourcePatternType == kafka.ResourcePatternTypePrefixed && !commons.HasPrefix(topics.Result, acl.Name) {
				row[7] = TopicPrefixNotFoundStatus
				inactiveACls = append(inactiveACls, acl)
			}
		}
		allACLS[i] = row
	}
	outputs.NewTable([]string{TypeHeader, PrincipalHeader, NameHeader, PermissionHeader, OperationHeader, PatternHeader, HostHeader, StatusHeader}, allACLS)

	if len(inactiveACls) > 0 {
		if commons.BuildConfirmationPrompt("Delete all unused Topic ACLs?", confirm) {
			res, err := c.CloudAPI.DeleteACLs(inactiveACls)
			if err != nil {
				fmt.Println("Error deleting ACLs")
				os.Exit(1)
			}
			if len(res) > 0 {
				rows := aclBindingsToSlice(inactiveACls)
				// Convert [][]string to [][]interface{}
				interfaceSlice := make([][]interface{}, len(rows))
				for i, row := range rows {
					interfaceRow := make([]interface{}, len(row))
					for j, val := range row {
						interfaceRow[j] = val
					}
					interfaceSlice[i] = interfaceRow
				}
				headers := []string{TypeHeader, PrincipalHeader, NameHeader, PermissionHeader, OperationHeader, PatternHeader, HostHeader}
				outputs.NewTable(headers, interfaceSlice)
			}
		}
	} else {
		fmt.Println("No inactive ACLs found.")
	}
}

// Clean Service Accounts
func (c *ConfluentClean) HandleInactiveServiceAccounts(confirm bool) {
	fmt.Println("\n Get Service Accounts cluster connections (Last 7 Days)")

	principalWithKeysCh := commons.AsyncCall(func() (map[string][]string, error) {
		return c.CloudAPI.GetClusterApiKeys()
	})
	principalsConnectionsCh := commons.AsyncCall(func() (map[string]float64, error) {
		return c.MetricsAPI.GetConnectionsByServiceAccount()
	})

	principalWithKeys := <-principalWithKeysCh
	principalsCon := <-principalsConnectionsCh

	var principalWithClusterKeys map[string][]string
	var principalsConnections map[string]float64

	// Check the result and handle errors
	if principalWithKeys.Err != nil {
		fmt.Println("Error:", principalWithKeys.Err)
	} else {
		principalWithClusterKeys = principalWithKeys.Result
	}

	if principalsCon.Err != nil {
		fmt.Println("Error:", principalsCon.Err)
	} else {
		principalsConnections = principalsCon.Result
	}

	inactivePrincipals := make([]string, 0)
	inactiveRbacIds := make([]string, 0)
	y := make([][]interface{}, 0)
	for principal, values := range principalWithClusterKeys {
		if strings.Contains(principal, SERVICE_ACCOUNT) {
			row := make([]interface{}, 4)
			conn, ok := principalsConnections[principal]
			row[0] = principal
			row[3] = ""
			if ok {
				if conn > 0 {
					row[1] = ActiveStatus
				}
			} else {
				row[1] = InactiveStatus
				inactivePrincipals = append(inactivePrincipals, principal)
			}
			row[2] = values
			rolebindingsCh := commons.AsyncCall(func() ([]ConfluentCloudRoleBinding, error) {
				return c.CloudAPI.GetRoleBindings(principal)
			})
			roleBindings := <-rolebindingsCh
			if roleBindings.Err != nil {
				fmt.Println("Error getting role bindings")
				os.Exit(1)
			}
			for _, roleBinding := range roleBindings.Result {
				row[3] = roleBinding.Role
				inactiveRbacIds = append(inactiveRbacIds, roleBinding.Id)
			}
			y = append(y, row)
		}
	}

	outputs.NewTable([]string{"Service Account", "Active", "Cluster API KEYs", "Cluster Role Bindings"}, y)
	if len(inactivePrincipals) > 0 {
		if commons.BuildConfirmationPrompt("Delete inactive Service Accounts and cluster Role bindings", confirm) {
			fmt.Println("Delete Cluster API_KEYS")
			apikeysToDelete := make([]string, 0)
			for _, inactive := range inactivePrincipals {
				apikeysToDelete = append(apikeysToDelete, principalWithClusterKeys[inactive]...)
			}
			err := c.CloudAPI.DeleteApiKeys(apikeysToDelete)
			if err != nil {
				fmt.Println("Error deleting API keys")
				os.Exit(1)
			}
			fmt.Println("Delete Role Bindings")
			err = c.CloudAPI.DeleteRoleBindings(inactiveRbacIds)
			if err != nil {
				fmt.Println("Error deleting role bindings")
				os.Exit(1)
			}
		}
	} else {
		fmt.Println("No inactive service accounts found.")
	}
}

// Clean Connectors
func (c *ConfluentClean) InactiveConnectors() {
	fmt.Println("InactiveConnectors")
}

/** kafka.ACLBinding to slices */
func aclBindingsToSlice(acls []kafka.ACLBinding) [][]interface{} {
	allACLS := make([][]interface{}, len(acls))
	for i, acl := range acls {
		row := make([]interface{}, 7)
		allACLS[i] = aclBindingsToRow(row, acl)
	}
	return allACLS
}

/** kafka.ACLBinding to []string row */
func aclBindingsToRow(row []interface{}, acl kafka.ACLBinding) []interface{} {
	row[0] = acl.Type.String()
	row[1] = acl.Principal
	row[2] = acl.Name
	row[3] = acl.PermissionType.String()
	row[4] = acl.Operation.String()
	row[5] = acl.ResourcePatternType.String()
	row[6] = acl.Host
	return row
}
