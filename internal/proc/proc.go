package proc

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
)

type Process struct {
	PID  int
	Name string
}

// LISTE DES PROCESSUS

func List() ([]Process, error) {
	switch runtime.GOOS {
	case "windows":
		return listWindows()
	case "darwin":
		return listMac()
	default:
		return nil, errors.New("OS non supporté pour ProcessOps")
	}
}

func listWindows() ([]Process, error) {
	out, err := exec.Command("tasklist", "/FO", "CSV").Output()
	if err != nil {
		return nil, err
	}
	lines := bytes.Split(out, []byte{'\n'})
	var procs []Process
	for _, raw := range lines[1:] {
		fields := parseCSV(raw)
		if len(fields) < 2 {
			continue
		}
		pid, _ := strconv.Atoi(fields[1])
		procs = append(procs, Process{PID: pid, Name: strings.Trim(fields[0], "\"")})
	}
	return procs, nil
}

func listMac() ([]Process, error) {
	out, err := exec.Command("ps", "-Ao", "pid,comm").Output()
	if err != nil {
		return nil, err
	}
	lines := strings.Split(string(out), "\n")
	var procs []Process
	for _, l := range lines[1:] {
		l = strings.TrimSpace(l)
		if l == "" {
			continue
		}
		parts := strings.Fields(l)
		if len(parts) < 2 {
			continue
		}
		pid, _ := strconv.Atoi(parts[0])
		name := parts[1]
		procs = append(procs, Process{PID: pid, Name: name})
	}
	return procs, nil
}

func Filter(procs []Process, kw string) []Process {
	kw = strings.ToLower(kw)
	var out []Process
	for _, p := range procs {
		if strings.Contains(strings.ToLower(p.Name), kw) {
			out = append(out, p)
		}
	}
	return out
}

func Kill(pid int, force bool) error {
	switch runtime.GOOS {
	case "windows":
		args := []string{"/PID", fmt.Sprint(pid), "/T"}
		if force {
			args = append(args, "/F")
		}
		return exec.Command("taskkill", args...).Run()
	case "darwin":
		signal := "-15"
		if force {
			signal = "-9"
		}
		return exec.Command("kill", signal, fmt.Sprint(pid)).Run()
	default:
		return errors.New("OS non supporté")
	}
}

func parseCSV(b []byte) []string {
	var res []string
	cur := bytes.Buffer{}
	inQuotes := false
	for _, c := range b {
		switch c {
		case '"':
			inQuotes = !inQuotes
		case ',':
			if inQuotes {
				cur.WriteByte(c)
			} else {
				res = append(res, cur.String())
				cur.Reset()
			}
		default:
			cur.WriteByte(c)
		}
	}
	res = append(res, cur.String())
	return res
}
