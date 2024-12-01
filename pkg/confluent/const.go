package confluent

const (
	ActiveStatus = "ACTIVE"
	// Topic status constants
	InactiveStatus = "INACTIVE"

	// ACL status constants
	TopicNotFoundStatus       = "TOPIC_NOT_FOUND"
	TopicPrefixNotFoundStatus = "TOPIC_PREFIX_NOT_FOUND"

	// Service account status constants
	ActiveServiceAccount   = "YES"
	InactiveServiceAccount = "NO"

	TypeHeader       = "Type"
	PrincipalHeader  = "Principal"
	NameHeader       = "Name"
	PermissionHeader = "Permission"
	OperationHeader  = "Operation"
	PatternHeader    = "Pattern"
	HostHeader       = "Host"

	StatusHeader = "Status"

	// The Confluent Cloud API endpoint
	CONFLUENT_ENDPOINT = "https://api.confluent.cloud"
	//CLUSTER
	KAFKA_ENDPOINT  = "%s/kafka/v3/clusters/%s/topics"
	INTERNAL_PREFIX = "__"
	DATA            = "data"
	TOPIC_NAME      = "topic_name"
	SASL_SSL        = "SASL_SSL"
	PLAIN           = "PLAIN"
	//ACL
	ACL_ENDPOINT = "%s/kafka/v3/clusters/%s/acls"
	//OPERATIONS
	CLUSTER = "/cmk/v2/clusters/%s?environment=%s"
	//API KEYS
	API_KEYS         = "/iam/v2/api-keys"
	CLUSTER_API_KEYS = API_KEYS + "?spec.resource=%s"
	//RBAC
	RBAC_ENDPOINT = "/iam/v2/role-bindings?principal=User:%s&crn_pattern=%s"
)
