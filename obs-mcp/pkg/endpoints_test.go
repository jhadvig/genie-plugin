package pkg

import (
	"testing"
)

func TestGetEndpoint(t *testing.T) {
	tests := []struct {
		name        string
		config      *EndpointConfig
		path        string
		expectError bool
	}{
		{
			name:        "Valid endpoint in DefaultConfig",
			config:      DefaultEndpointConfig(),
			path:        PathListMetrics,
			expectError: false,
		},
		{
			name:        "Invalid endpoint in DefaultConfig",
			config:      DefaultEndpointConfig(),
			path:        PathQuery,
			expectError: true,
		},
		{
			name:        "Valid endpoint in FullConfig",
			config:      FullEndpointConfig(),
			path:        PathQuery,
			expectError: false,
		},
		{
			name:        "Non-existent endpoint",
			config:      DefaultEndpointConfig(),
			path:        "/api/v1/nonexistent",
			expectError: true,
		},
		{
			name:        "Empty path",
			config:      DefaultEndpointConfig(),
			path:        "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			endpoint, err := tt.config.GetEndpoint(tt.path)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if endpoint.Path != tt.path {
					t.Errorf("Expected path %s, got %s", tt.path, endpoint.Path)
				}
			}
		})
	}
}
