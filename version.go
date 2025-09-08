package main

import (
	"os/exec"
	"runtime/debug"
	"strings"
)

var commit = "dev"

func init() {
	if commit != "dev" {
		return
	}
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, s := range info.Settings {
			if s.Key == "vcs.revision" && s.Value != "" {
				commit = s.Value
				if len(commit) > 7 {
					commit = commit[:7]
				}
				return
			}
		}
	}
	if c, err := exec.Command("git", "rev-parse", "--short", "HEAD").Output(); err == nil {
		commit = strings.TrimSpace(string(c))
	}
}
