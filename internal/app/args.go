package app

import (
	"errors"
	"fmt"
	"strings"
)

type Options struct {
	Command        string
	Branch         string
	Base           string
	Skill          string
	ProviderFlag   string
	FormatFlag     string
	LangFlag       string
	ConfigProvider string
}

func ParseArgs(args []string) (Options, error) {
	opts := Options{Command: "review", Base: "main"}
	if len(args) == 0 {
		opts.Command = "help"
		return opts, nil
	}
	if args[0] == "--help" || args[0] == "-h" {
		opts.Command = "help"
		return opts, nil
	}
	if args[0] == "--version" {
		opts.Command = "version"
		return opts, nil
	}
	if args[0] == "update" {
		opts.Command = "update"
		return opts, nil
	}
	if len(args) == 2 && args[0] == "providers" && args[1] == "doctor" {
		opts.Command = "providers-doctor"
		return opts, nil
	}
	if len(args) == 4 && args[0] == "config" && args[1] == "set" && args[2] == "provider" {
		opts.Command = "config-set-provider"
		opts.ConfigProvider = args[3]
		return opts, nil
	}
	if strings.HasPrefix(args[0], "--") {
		return Options{}, errors.New("branch name is required")
	}
	opts.Branch = args[0]
	for _, arg := range args[1:] {
		switch {
		case strings.HasPrefix(arg, "--base="):
			opts.Base = strings.TrimPrefix(arg, "--base=")
		case strings.HasPrefix(arg, "--skill="):
			opts.Skill = strings.TrimPrefix(arg, "--skill=")
		case strings.HasPrefix(arg, "--provider="):
			opts.ProviderFlag = strings.TrimPrefix(arg, "--provider=")
		case strings.HasPrefix(arg, "--format="):
			opts.FormatFlag = strings.TrimPrefix(arg, "--format=")
		case strings.HasPrefix(arg, "--lang="):
			opts.LangFlag = strings.TrimPrefix(arg, "--lang=")
		default:
			return Options{}, fmt.Errorf("unknown option %q", arg)
		}
	}
	return opts, nil
}
