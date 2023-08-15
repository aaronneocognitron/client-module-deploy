package dockercompose

import (
	"asterizm/builder/config"
	"fmt"
	"strings"
)

const (
	DbHost          = "asterizm-db"
	AsterizmConsole = "asterizm-console"
	AsterizmCron    = "asterizm-cron"
	AsterizmScanner = "asterizm-scanner-%s"
)

type Service struct {
	Image         string         `yaml:"image"`
	ContainerName string         `yaml:"container_name"`
	Volumes       []string       `yaml:"volumes,omitempty"`
	Ports         []string       `yaml:"ports,omitempty"`
	Networks      []string       `yaml:"networks,omitempty"`
	Command       []string       `yaml:"command,omitempty"`
	DependsOn     map[string]any `yaml:"depends_on,omitempty"`
	Environment   map[string]any `yaml:"environment,omitempty"`
	HealthCheck   map[string]any `yaml:"healthcheck,omitempty"`
	Restart       string         `yaml:"restart,omitempty"`
}

type DockerCompose struct {
	Version  string                       `yaml:"version"`
	Networks map[string]map[string]string `yaml:"networks"`
	Volumes  map[string]map[string]string `yaml:"volumes"`
	Services map[string]Service           `yaml:"services"`
}

func InitFromConfig(configPath string, config *config.Config) *DockerCompose {
	asterizmNetwork := "asterizm"
	dbDataVolume := "dbdata"

	asterizmImage := "asterizm/client-server:latest"
	configVolume := configPath + ":" + "/app/config.yml:rw"

	dockerCompose := &DockerCompose{
		Version: "3.9",
		Networks: map[string]map[string]string{
			asterizmNetwork: {"name": asterizmNetwork, "driver": "bridge"},
		},
		Volumes: map[string]map[string]string{
			dbDataVolume: {"name": dbDataVolume, "driver": "local"},
		},
		Services: make(map[string]Service),
	}

	asterizmDependOn := make(map[string]any)

	if config.Utils.Db.Host == DbHost {
		dockerCompose.Services[DbHost] = Service{
			ContainerName: DbHost,
			Image:         "postgres:15-alpine",
			Networks:      []string{asterizmNetwork},
			Volumes:       []string{dbDataVolume + ":/var/lib/postgresql/data"},
			Environment: map[string]any{
				"POSTGRES_USER":     config.Utils.Db.User,
				"POSTGRES_PASSWORD": config.Utils.Db.Password,
				"POSTGRES_DB":       config.Utils.Db.Name,
				"POSTGRES_PORT":     config.Utils.Db.Port,
			},
			Restart: "always",
			HealthCheck: map[string]any{
				"test":     []string{"CMD-SHELL", "pg_isready -U postgres"},
				"interval": "5s",
				"retries":  3,
			},
		}

		asterizmDependOn[DbHost] = map[string]string{
			"condition": "service_healthy",
		}
	}

	dockerCompose.Services[AsterizmConsole] = Service{
		ContainerName: AsterizmConsole,
		Image:         asterizmImage,
		Volumes:       []string{configVolume},
		Networks:      []string{asterizmNetwork},
		DependsOn:     asterizmDependOn,
		Restart:       "always",
	}

	dockerCompose.Services[AsterizmCron] = Service{
		ContainerName: AsterizmCron,
		Image:         asterizmImage,
		Volumes:       []string{configVolume},
		Networks:      []string{asterizmNetwork},
		Command:       []string{"cron/process"},
		DependsOn:     asterizmDependOn,
		Restart:       "always",
	}

	for key, _ := range config.Nodes.List {
		containerName := fmt.Sprintf(AsterizmScanner, strings.ToLower(key))

		dockerCompose.Services[containerName] = Service{
			ContainerName: containerName,
			Image:         asterizmImage,
			Volumes:       []string{configVolume},
			Networks:      []string{asterizmNetwork},
			DependsOn:     asterizmDependOn,
			Command:       []string{"node/scan", strings.ToUpper(key)},
			Restart:       "always",
		}
	}

	return dockerCompose
}
