package main

import (
	"asterizm/builder/config"
	"asterizm/builder/dockercompose"
	"asterizm/builder/scripts"
	"bufio"
	"errors"
	"flag"
	"fmt"
	"golang.org/x/sys/unix"
	"gopkg.in/yaml.v3"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"unicode"
)

func main() {
	help := flag.Bool("help", false, "Show help")
	configPath := flag.String("f", "", "Config file path")
	isTest := flag.Bool("test", false, "Use test networks")
	flag.Parse()

	if *help {
		flag.Usage()
		os.Exit(1)
	}

	if *configPath == "" {
		fmt.Println("Config path is required")
		flag.Usage()
		os.Exit(1)
	}

	if err := checkConfigFileAndDir(*configPath); err != nil {
		fmt.Println(capitalize(err.Error()))
		os.Exit(1)
	}

	if err := processScript(scripts.InstallDocker); err != nil {
		fmt.Println("Please, install docker and docker compose manually")
		os.Exit(1)
	}

	refreshedConfig, err := config.ParseAndRefreshConfig(dockercompose.DbHost, *configPath)
	if err != nil {
		fmt.Printf("Parse config error: %v \n", err)
		os.Exit(1)
	}

	yml, err := yaml.Marshal(refreshedConfig)
	if err != nil {
		fmt.Printf("Marshal config error: %v \n", err)
		os.Exit(1)
	}

	err = os.WriteFile(*configPath, yml, 0644)
	if err != nil {
		fmt.Printf("Write config error: %v \n", err)
		os.Exit(1)
	}

	dockerComposePath := path.Dir(*configPath) + "/docker-compose.yml"
	if err := processScript(scripts.InitDockerCompose, path.Dir(*configPath)); err != nil {
		fmt.Printf("Please, create %s file manually \n", dockerComposePath)
		os.Exit(1)
	}

	generatedDockerCompose := dockercompose.InitFromConfig("./"+path.Base(*configPath), refreshedConfig)
	dockerComposeYml, err := yaml.Marshal(generatedDockerCompose)
	if err != nil {
		fmt.Printf("Marshal docker-compose.yml error: %v \n", err)
		os.Exit(1)
	}

	err = os.WriteFile(dockerComposePath, dockerComposeYml, 0644)
	if err != nil {
		fmt.Printf("Write docker-compose.yml error: %v \n", err)
		os.Exit(1)
	}

	var commands []string
	if _, ok := generatedDockerCompose.Services[dockercompose.DbHost]; ok {
		commands = append(commands, fmt.Sprintf("docker compose -f %s up %s -d --wait", dockerComposePath, dockercompose.DbHost))
	}

	commands = append(commands, fmt.Sprintf("docker compose -f %s up %s -d --wait", dockerComposePath, dockercompose.AsterizmConsole))
	commands = append(commands, fmt.Sprintf("docker exec -t %s ./main migrations/up", dockercompose.AsterizmConsole))
	commands = append(commands, fmt.Sprintf("docker exec -t %s ./main db/seed", dockercompose.AsterizmConsole))
	if *isTest {
		commands[len(commands)-1] += " --test" // add test to db/seed
	}
	commands = append(commands, fmt.Sprintf("docker compose -f %s up -d", dockerComposePath))

	for i, command := range commands {
		if err := processScript(command); err != nil {
			printCommandsError(commands[i:])
			os.Exit(1)
		}
	}

	fmt.Println("Finish!")
}

func capitalize(str string) string {
	runes := []rune(str)
	runes[0] = unicode.ToUpper(runes[0])
	return string(runes)
}

func printCommandsError(commands []string) {
	for _, command := range commands {
		printCommandError(command)
	}
}

func printCommandError(command string) {
	fmt.Printf("Please, run %q manually \n", command)
}

func checkConfigFileAndDir(configPath string) error {
	if path.Ext(configPath) != ".yml" && path.Ext(configPath) != ".yaml" {
		return errors.New("config extension is not supported")
	}

	if _, err := os.Stat(configPath); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return errors.New("config is not exists")
		}

		return fmt.Errorf("check config errors: %w", err)
	}

	if unix.Access(configPath, unix.W_OK) != nil {
		return errors.New("config is not readable")
	}

	if unix.Access(path.Dir(configPath), unix.W_OK) != nil {
		return errors.New("directory is not writable")
	}

	return nil
}

func processScript(script string, params ...string) error {
	for i, param := range params {
		script = strings.Replace(script, "$"+strconv.Itoa(i+1), param, -1)
	}

	cmd := exec.Command("bash")
	cmd.Stdin = strings.NewReader(script)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		fmt.Println(err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	err = cmd.Start()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// print the output of the subprocess
	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			m := scanner.Text()
			fmt.Println(m)
		}
	}()

	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			m := scanner.Text()
			fmt.Println(m)
		}
	}()

	return cmd.Wait()
}
