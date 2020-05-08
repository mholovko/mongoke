package mongoke

import (
	yaml "gopkg.in/yaml.v2"
)

type Config struct {
	schemaString string
	mongoDbUri   string
	Schema       string                `yaml:"schema"`
	SchemaPath   string                `yaml:"schema_path"`
	Types        map[string]TypeConfig `yaml:"types"`
	Relations    []RelationConfig      `yaml:"relations"`
}

type TypeConfig struct {
	Exposed    bool   `yaml:"exposed"`
	Collection string `yaml:"collection"`
}

type RelationConfig struct {
	From         string                 `yaml:"from"`
	To           string                 `yaml:"to"`
	RelationType string                 `yaml:"relation_type"`
	where        map[string]interface{} `yaml:"where"`
}

// MakeConfigFromYaml parses the config from yaml
func MakeConfigFromYaml(data string) (Config, error) {
	t := Config{}

	err := yaml.Unmarshal([]byte(data), &t)
	if err != nil {
		return Config{}, err
	}

	return t, nil
}
