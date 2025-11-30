package docker

import "testing"

func Test_YamlParser(t *testing.T) {
	composeFile, err := ParseDockerComposeFile("docker-compose.yaml")
	if err != nil {
		t.Fatalf("Failed to parse YAML: %v", err)
	}

	tests := map[string]struct {
		expectedPorts []string
		serviceName   string
	}{
		"kafka service ports": {
			expectedPorts: []string{"9092", "9093"},
			serviceName:   "kafka",
		},
		"timescale service ports": {
			expectedPorts: []string{"5432"},
			serviceName:   "timescale",
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			service, exists := composeFile.Services[tt.serviceName]
			if !exists {
				t.Fatalf("Service '%s' not found in compose file", tt.serviceName)
			}

			if len(service.Ports) != len(tt.expectedPorts) {
				t.Fatalf("Expected %d ports, got %d", len(tt.expectedPorts), len(service.Ports))
			}

			for i, port := range service.Ports {
				if port.Port != tt.expectedPorts[i] {
					t.Errorf("Expected port %s, got %s", tt.expectedPorts[i], port.Port)
				}
			}
		})
	}
}
