package main

import (
	"flag"
	"fmt"
	"os"

	"log/slog"

	"gopkg.in/yaml.v3"
)

type Service struct {
	Build   string   `yaml:"build,omitempty"`
	Image   string   `yaml:"image,omitempty"`
	Expose  []int    `yaml:"expose,omitempty"`
	Ports   []string `yaml:"ports,omitempty"`
	Volumes []string `yaml:"volumes,omitempty"`
}

type DockerCompose struct {
	Services map[string]Service `yaml:"services"`
}

func main() {
	// take a flag for how many levels we want
	var nFlag = flag.Int("n", 1, "help message for flag n")
	flag.Parse()
	slog.Info("inputs", "n", *nFlag)

	// if n 1
	// 80(lb) -> 8080(manager)

	// if n 2
	// 80(lb) -> 8081(lb2) -> 8080(manager)

	// if n 3
	// 80(lb) -> 8081(lb2) -> 8082(lb3) -> 8080(manager)

	// if n 4
	// 80(lb) -> 8081(lb2) -> 8082(lb3) -> 8083(lb4) -> 8080(manager)

	err := generateCaddyfile(*nFlag)
	if err != nil {
		slog.Error("failed to generate caddyfile", "err", err)
		return
	}
	err = generateDockerCompose(*nFlag)
	if err != nil {
		slog.Error("failed to generate docker-compose", "err", err)
		return
	}
}

func generateCaddyfile(n int) error {
	listenPort := 8080
	for i := 1; i < n; i++ {
		listenPort++
		slog.Info("generating caddyfile", "n", i, "listenPort", listenPort)
		config := fmt.Sprintf(`
:%d {
	# n = %d
	log {
	output stderr
	format console
	}
	reverse_proxy lb%d:%d
	}`, listenPort, i, i+1, listenPort+1)
		f, err := os.Create(fmt.Sprintf("Caddyfile%d", i))
		if err != nil {
			return err
		}
		defer f.Close()
		_, err = f.WriteString(config)
		if err != nil {
			return err
		}
	}

	// write the last one to point to manager
	config := fmt.Sprintf(`:%d {
	# n = %d
	log {
	output stderr
	format console
	}
	reverse_proxy manager:8080
	}`, listenPort+1, n)
	f, err := os.Create(fmt.Sprintf("Caddyfile%d", n))
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.WriteString(config)
	if err != nil {
		return err
	}
	return nil
}

func generateDockerCompose(n int) error {
	dockerCompose := DockerCompose{
		Services: map[string]Service{
			"manager": {
				Build:  "app/",
				Expose: []int{8080},
				Ports:  []string{"8080:8080"},
			},
			"lb": {
				Image:   "caddy:latest",
				Volumes: []string{"./Caddyfile1:/etc/caddy/Caddyfile"},
				Ports:   []string{"8081:8081"},
			},
		},
	}

	// when n > 1, we need to add more services
	for i := 2; i <= n; i++ {
		dockerCompose.Services[fmt.Sprintf("lb%d", i)] = Service{
			Image:   "caddy:latest",
			Volumes: []string{fmt.Sprintf("./Caddyfile%d:/etc/caddy/Caddyfile", i)},
			Expose:  []int{8080 + i - 1},
		}
	}

	data, err := yaml.Marshal(&dockerCompose)
	if err != nil {
		return err
	}
	f, err := os.Create("docker-compose.yml")
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.Write(data)
	return err
}
