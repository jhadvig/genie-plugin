package pkg

import (
	"testing"
)

func TestToolEnumConstraintsLLM(t *testing.T) {
	defaultConfig := DefaultEndpointConfig()
	tool := CreateUniversalPrometheusQueryTool(defaultConfig)

	endpointProp := tool.InputSchema.Properties["endpoint"].(map[string]interface{})

	var enumSlice []string
	switch v := endpointProp["enum"].(type) {
	case []string:
		enumSlice = v
	case []interface{}:
		enumSlice = make([]string, len(v))
		for i, val := range v {
			enumSlice[i] = val.(string)
		}
	}

	allowedInEnum := make(map[string]bool)
	for _, val := range enumSlice {
		allowedInEnum[val] = true
	}

	allowedEndpoints := defaultConfig.AllowedEndpoints
	for path := range allowedEndpoints {
		if !allowedInEnum[path] {
			t.Errorf("Expected path %s to be in enum constraints, but it was not found", path)
		}
	}

	disallowedEndpoints := map[string]PrometheusEndpoint{}
	for path := range FullEndpointConfig().AllowedEndpoints {
		if _, ok := allowedEndpoints[path]; !ok {
			disallowedEndpoints[path] = FullEndpointConfig().AllowedEndpoints[path]
		}
	}

	for path := range disallowedEndpoints {
		if allowedInEnum[path] {
			t.Errorf("Did not expect path %s to be in enum constraints, but it was found", path)
		}
	}
}
