package main

import (
	"os/exec"
	"runtime/debug"
	"strings"
	"time"
)

var commit = "dev"
var buildDate = ""

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
			case "vcs.time":
				if buildDate == "" && s.Value != "" {
					if t, err := time.Parse(time.RFC3339, s.Value); err == nil {
						buildDate = t.Format("2006-01-02")
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
	if buildDate == "" {
		buildDate = time.Now().Format("2006-01-02")
	}
}
