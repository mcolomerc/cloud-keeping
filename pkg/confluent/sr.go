package confluent

import "mcolomer/cloud-keeping/pkg/client"

type SchemaRegistryClient struct {
	Endpoint  string
	ApiKey    string
	ApiSecret string
	SR_API    client.HTTPS
}

func NewSchemaRegistryClient(endpoint, api_key, api_secret string) *SchemaRegistryClient {
	return &SchemaRegistryClient{
		Endpoint:  endpoint,
		ApiKey:    api_key,
		ApiSecret: api_secret,
		SR_API:    *client.NewHTTPS(endpoint, api_key, api_secret),
	}
}

func (s *SchemaRegistryClient) GetSubjects() ([]string, error) {
	//TODO: Implement
	return nil, nil
}

func (s *SchemaRegistryClient) DeleteSubject(subject string) error {
	return nil 
}

