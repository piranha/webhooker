// (c) 2013 Alexander Solovyov under terms of ISC License

package main

import (
	"fmt"
	flags "github.com/jessevdk/go-flags"
	"log"
	"os"
	"regexp"
)

/// Globals

var Version = "0.1"

var opts struct {
	Port     string `short:"p" long:"port" default:"8000" description:"port to listen on"`
	ShowHelp bool   `long:"help" description:"show this help message"`
}

/// Types

type Rule struct {
	Branch  string
	Command string
}

type Rules []Rule

type Config map[string]Rules

/// Config

func (c *Config) Parse(input []string) error {
	RuleRe := regexp.MustCompile("(.+/.+):(.+)=(.+)")

	for _, line := range input {
		bits := RuleRe.FindStringSubmatch(line)
		if len(bits) != 4 {
			return fmt.Errorf("Can't parse line '%s'", line)
		}

		name := bits[1]
		if _, ok := (*c)[name]; !ok {
			(*c)[name] = make(Rules, 0)
		}
		(*c)[name] = append((*c)[name], Rule{bits[2], bits[3]})
	}

	return nil
}

/// Main

func main() {
	argparser := flags.NewParser(&opts, flags.PrintErrors|flags.PassDoubleDash)

	args, err := argparser.Parse()
	if err != nil {
		return
	}

	argparser.Usage = `[OPTIONS] user/repo:branch=command [more rules...]`

	if opts.ShowHelp || len(args) == 0 {
		argparser.WriteHelp(os.Stdout)
		return
	}

	cfg := make(Config)
	errhandle(cfg.Parse(args))

	fmt.Printf("%v", cfg)
}

func errhandle(err error) {
	if err == nil {
		return
	}
	log.Fatalf("ERR %s\n", err)
	os.Exit(1)
}
