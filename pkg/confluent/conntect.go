package confluent

//https://api.confluent.cloud/connect/v1/environments/env-zmz2zd/clusters/lkc-q8dr5m/connectors?expand=info,status,id

type ConfluentCloudConnector struct {
	Name  string         `json:"name"`
	Id    string         `json:"id"`
	State ConnectorState `json:"state"`
	Type  string         `json:"type"`
}

type ConnectorState string

const (
	RUNNING      ConnectorState = "RUNNING"
	PAUSED       ConnectorState = "PAUSED"
	FAILED       ConnectorState = "FAILED"
	PROVISIONING ConnectorState = "PROVISIONING"
)

type ConnectorType string

const (
	SINK   ConnectorType  = "sink"
	SOURCE ConnectorState = "source"
)
