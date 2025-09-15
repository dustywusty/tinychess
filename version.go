package main

import (
	"os/exec"
	"runtime/debug"
	"strings"
)

var commit = "dev"

func init() {
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, s := range info.Settings {
			switch s.Key {
			case "vcs.revision":
				if commit == "dev" && s.Value != "" {
					commit = s.Value
					if len(commit) > 7 {
						commit = commit[:7]
					}
				}
			}
		}
	}
	if commit == "dev" {
		if c, err := exec.Command("git", "rev-parse", "--short", "HEAD").Output(); err == nil {
			commit = strings.TrimSpace(string(c))
		}
	}
}
