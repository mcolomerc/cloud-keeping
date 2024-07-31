package config

// Config is a struct that holds the configuration for the client.
type Config struct {
	Props map[string]interface{}
}

func NewConfig(props map[string]interface{}) (*Config, error) {
	// some code
	return &Config{Props: props}, nil
}
