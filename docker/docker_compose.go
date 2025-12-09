package docker

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
	"testing"
	"time"
)

const composeLockPath = "/tmp/golang-app-test.compose.lock"

func lock(path string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for {
		f, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o644)
		if err == nil {
			// write pid for observability
			_, _ = fmt.Fprintf(f, "%d\n", os.Getpid())
			_ = f.Close()
			return nil
		}
		if !os.IsExist(err) {
			return fmt.Errorf("lock: unexpected error: %w", err)
		}
		if time.Now().After(deadline) {
			// read current owner for diagnostics
			b, _ := os.ReadFile(path)
			return fmt.Errorf("lock: timeout acquiring %s (owner pid: %s)", path, strings.TrimSpace(string(b)))
		}
		time.Sleep(300 * time.Millisecond)
	}
}

func unlock(path string) {
	_ = os.Remove(path)
}

// Container tracks information about the docker container started for tests.
type Container struct {
	Name      string
	HostPorts map[string]string
}

type portMapping struct {
	HostIP   string
	HostPort string
}

var (
	startupCount int
	composeMu    sync.Mutex
)

// ComposeUp starts the specified docker-compose services and returns their container information.
// If no service names are provided, all services in the compose file are started.
// It returns a map of service names to their corresponding Container information.
func ComposeUp(t *testing.T, composeFile string, serviceNames ...string) (map[string]Container, error) {
	composeMu.Lock()
	defer composeMu.Unlock()

	// inter-process lock
	if err := lock(composeLockPath, 30*time.Second); err != nil {
		return nil, err
	}
	defer unlock(composeLockPath)

	t.Logf("Starting docker-compose services from %s (%d)", composeFile, startupCount)
	_, err := os.Stat(composeFile)
	if err != nil {
		return nil, fmt.Errorf("compose file not found: %s", composeFile)
	}

	args := []string{"-f", composeFile, "up", "-d"}
	if len(serviceNames) > 0 {
		args = append(args, serviceNames...)
	}

	var out bytes.Buffer
	cmd := exec.Command("docker", append([]string{"compose"}, args...)...)
	cmd.Stdout = &out
	cmd.Stderr = &out
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("could not start docker compose %s: %v; output: %s", composeFile, err, strings.TrimSpace(out.String()))
	}

	containers, err := fromCompose(composeFile, serviceNames)
	if err != nil {
		return nil, fmt.Errorf("could not get docker-compose containers: %w", err)
	}

	return containers, nil
}

// ComposeDown stops the specified docker-compose services.
// If no service names are provided, all services in the compose file are stopped.
func ComposeDown(t *testing.T, composeFile string, serviceNames ...string) error {
	composeMu.Lock()
	defer composeMu.Unlock()

	// inter-process lock
	if err := lock(composeLockPath, 30*time.Second); err != nil {
		return err
	}
	defer unlock(composeLockPath)

	t.Logf("Stopping docker-compose services from %s (%d)", composeFile, startupCount)

	args := []string{"-f", composeFile, "down"}
	if len(serviceNames) > 0 {
		args = append(args, serviceNames...)
	}

	var out bytes.Buffer
	cmd := exec.Command("docker", append([]string{"compose"}, args...)...)
	cmd.Stdout = &out
	cmd.Stderr = &out
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("could not stop docker compose %s: %v; output: %s", composeFile, err, strings.TrimSpace(out.String()))
	}
	t.Logf("waiting for 3s...")
	time.Sleep(3 * time.Second)
	return nil
}

// fromCompose retrieves container information from the specified docker-compose file.
// It parses the compose file to get service definitions and their ports,
// then inspects each service's container to get the actual host port mappings.
// It returns a map of service names to their corresponding Container information.
func fromCompose(composeFile string, serviceNames []string) (map[string]Container, error) {
	compose, err := ParseDockerComposeFile(composeFile)
	if err != nil {
		return nil, fmt.Errorf("could not parse docker-compose file: %s", composeFile)
	}
	containers := make(map[string]Container, len(compose.Services))
	for serviceName, service := range compose.Services {
		if notFound(serviceNames, serviceName) {
			continue
		}

		portMappings, err := getServicePortsFromCompose(serviceName, service)
		if err != nil {
			return nil, fmt.Errorf("could not get ports for service %s: %w", serviceName, err)
		}

		hostPorts := make(map[string]string, len(portMappings))
		for i, pm := range portMappings {
			hostPorts[service.Ports[i].Port] = pm.HostPort
		}

		c := Container{
			Name:      serviceName,
			HostPorts: hostPorts,
		}
		containers[serviceName] = c
		waitForHealthy(serviceName, 20*time.Second)
	}

	return containers, nil
}

// notFound checks if an item is not present in a slice of strings.
// It returns true if the item is not found, and false otherwise.
func notFound(slice []string, item string) bool {
	if len(slice) == 0 {
		return false
	}

	for _, s := range slice {
		if s == item {
			return false
		}
	}

	return true
}

// getServicePortsFromCompose retrieves the port mappings for a given service from the docker-compose setup.
// It uses 'docker inspect' to get the actual host ports mapped to the service's defined ports.
// It returns a slice of portMapping structs containing the HostIP and HostPort for each defined port.
func getServicePortsFromCompose(serviceName string, service Service) ([]portMapping, error) {
	portMappings := make([]portMapping, len(service.Ports))

	for i, port := range service.Ports {
		// inspect port mapping for each port from docker
		tmpl := fmt.Sprintf("[{{range $i,$v := (index .NetworkSettings.Ports \"%s/%s\")}}{{if $i}},{{end}}{{json $v}}{{end}}]", port.Port, strings.ToLower(port.Protocol))

		var out bytes.Buffer
		cmd := exec.Command("docker", "inspect", "-f", tmpl, serviceName)
		cmd.Stdout = &out
		if err := cmd.Run(); err != nil {
			return []portMapping{}, fmt.Errorf("could not inspect container %s: %w", serviceName, err)
		}

		var docs []struct {
			HostIP   string `json:"HostIp"`
			HostPort string `json:"HostPort"`
		}
		if err := json.Unmarshal(out.Bytes(), &docs); err != nil {
			return []portMapping{}, fmt.Errorf("could not decode json: %w", err)
		}

		if len(docs) < 1 {
			return []portMapping{}, fmt.Errorf("could not find port mappings for service %s", serviceName)
		}

		for _, doc := range docs {
			if doc.HostIP != "::" {

				// Podman keeps HostIP empty instead of using 0.0.0.0.
				// - https://github.com/containers/podman/issues/17780
				if doc.HostIP == "" {
					portMappings[i] = portMapping{
						HostIP:   "localhost",
						HostPort: doc.HostPort,
					}
				}

				portMappings[i] = portMapping{
					HostIP:   doc.HostIP,
					HostPort: doc.HostPort,
				}
			}
		}
	}

	return portMappings, nil
}

func waitForHealthy(containerName string, timeout time.Duration) {
	fmt.Printf("Checking if %s is healthy\n", containerName)
	for i := 0; i < int(timeout.Seconds()); i++ {
		cmd := exec.Command("docker", "inspect", "--format", "{{if .State.Health}}{{.State.Health.Status}}{{else}}{{.State.Status}}{{end}}", containerName)
		out, err := cmd.Output()
		status := strings.TrimSpace(string(out))
		if err == nil && status == "healthy" {
			fmt.Printf("%s is healthy\n", containerName)

			return
		}
		time.Sleep(1 * time.Second)
	}
	panic(fmt.Sprintf("Container %s did not become healthy in time", containerName))
}
