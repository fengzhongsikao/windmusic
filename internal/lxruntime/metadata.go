package lxruntime

import (
	"regexp"
	"strings"
)

type ScriptMeta struct {
	Name        string
	Description string
	Version     string
	Author      string
	Homepage    string
}

var metaPatterns = map[string]*regexp.Regexp{
	"name":        regexp.MustCompile(`(?m)@name\s+(.+)`),
	"description": regexp.MustCompile(`(?m)@description\s+(.+)`),
	"version":     regexp.MustCompile(`(?m)@version\s+(.+)`),
	"author":      regexp.MustCompile(`(?m)@author\s+(.+)`),
	"homepage":    regexp.MustCompile(`(?m)@homepage\s+(.+)`),
}

func ParseScriptMeta(script string) ScriptMeta {
	meta := ScriptMeta{}
	for key, pattern := range metaPatterns {
		match := pattern.FindStringSubmatch(script)
		if len(match) < 2 {
			continue
		}
		value := strings.TrimSpace(match[1])
		switch key {
		case "name":
			meta.Name = value
		case "description":
			meta.Description = value
		case "version":
			meta.Version = value
		case "author":
			meta.Author = value
		case "homepage":
			meta.Homepage = value
		}
	}
	return meta
}
