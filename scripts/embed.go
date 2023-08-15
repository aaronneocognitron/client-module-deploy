package scripts

import _ "embed"

//go:embed install-docker.sh
var InstallDocker string

//go:embed init-docker-compose.sh
var InitDockerCompose string
