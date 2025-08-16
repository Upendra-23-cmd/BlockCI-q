package core

import (
	"os"
	"gopkg.in/yaml.v3"
)

// ParsePipeline parses YAML content into a Pipeline object
func ParsePipeline(data []byte) (*Pipeline, error) {
    var pipeline Pipeline
    err := yaml.Unmarshal(data, &pipeline)
    if err != nil {
        return nil, err
    }
    return &pipeline, nil
}


// LoadPipeline reads pipeline.yaml and returns a Pipeline object
func LoadPipeline(path string) (*Pipeline, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var pipeline Pipeline
	err = yaml.Unmarshal(data, &pipeline)
	if err != nil {
		return nil , err
	}

	return &pipeline, nil
}