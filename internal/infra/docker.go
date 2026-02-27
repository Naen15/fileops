package infra

import (
	"bytes"
	"encoding/json"
	"errors"
	"os/exec"
	"strings"
)

// Container résume ce que l’on veut afficher.
type Container struct {
	ID     string
	Name   string
	Image  string
	Status string
}

// List retourne tous les conteneurs actifs (docker ps).
func List() ([]Container, error) {
	cmd := exec.Command("docker", "ps", "--format", "{{json .}}")
	out, err := cmd.Output()
	if err != nil {
		return nil, errors.New("docker non installé ou pas dans PATH")
	}
	var res []Container
	for _, line := range bytes.Split(out, []byte{'\n'}) {
		if len(line) == 0 {
			continue
		}
		var raw map[string]string
		if err := json.Unmarshal(line, &raw); err != nil {
			continue
		}
		res = append(res, Container{
			ID:     raw["ID"],
			Name:   raw["Names"],
			Image:  raw["Image"],
			Status: raw["Status"],
		})
	}
	return res, nil
}

// Stats renvoie une ligne « CPU %   MEM % » via docker stats --no-stream.
func Stats(containerID string) (string, error) {
	out, err := exec.Command("docker", "stats", "--no-stream",
		"--format", "{{.CPUPerc}} {{.MemPerc}}", containerID).Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}
