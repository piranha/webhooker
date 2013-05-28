// (c) 2013 Alexander Solovyov under terms of ISC License

package main

import (
	"encoding/json"
	"fmt"
	flags "github.com/jessevdk/go-flags"
	logmod "log"
	"log/syslog"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

/// Globals

var Version = "0.1"

var opts struct {
	Interface string `short:"i" long:"interface" default:"" description:"ip to listen on"`
	Port      string `short:"p" long:"port" default:"8000" description:"port to listen on"`
	Log       string `short:"l" long:"log" description:"path to file for logging (supply '-' for syslog)"`
	ShowHelp  bool   `long:"help" description:"show this help message"`
}

var log *logmod.Logger

/// Github types

type GithubUser struct {
	Name  string
	Email string
}

type GithubRepo struct {
	Name    string
	Url     string
	Private bool
	Owner   GithubUser
}

type GithubCommit struct {
	Id        string
	Message   string
	Timestamp string
	Url       string
	Author    GithubUser
}

type GithubPayload struct {
	Ref        string
	Repository GithubRepo
	Commits    []GithubCommit
}

/// Config types

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

func (c *Config) HandleRequest(w http.ResponseWriter, r *http.Request) {
	var data GithubPayload
	err := json.Unmarshal([]byte(r.FormValue("payload")), &data)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	name := fmt.Sprintf("%s/%s",
		data.Repository.Owner.Name, data.Repository.Name)
	rules, ok := (*c)[name]
	if !ok {
		return
	}

	branch := strings.TrimPrefix(data.Ref, "refs/heads/")

	for _, rule := range rules {
		if rule.Branch == branch {
			run(rule.Command, name, &data)
		}
	}
}

func run(command string, repo string, data *GithubPayload) error {
	cmd := exec.Command(command)

	commit := data.Commits[0]
	cmd.Env = []string{
		env("REPO", repo),
		env("REPO_URL", data.Repository.Url),
		env("PRIVATE", fmt.Sprintf("%t", data.Repository.Private)),
		env("BRANCH", data.Ref),
		env("COMMIT", commit.Id),
		env("COMMIT_MESSAGE", commit.Message),
		env("COMMIT_TIME", commit.Timestamp),
		env("COMMIT_AUTHOR", commit.Author.Name),
		env("COMMIT_URL", commit.Url),
	}

	out, err := cmd.CombinedOutput()
	log.Printf("'%s' for %s output: %s", command, repo, out)
	return err
}

func env(key string, value string) string {
	return fmt.Sprintf("%s=%s", key, value)
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

	configureLogging(opts.Log)

	cfg := make(Config)
	err = cfg.Parse(args)
	if err != nil {
		log.Println(err)
		// maybe it's better to print this error on stdout always
		//fmt.Println(err)
		os.Exit(1)
	}

	http.HandleFunc("/", cfg.HandleRequest)
	http.ListenAndServe(opts.Interface+":"+opts.Port, nil)
}

func configureLogging(dst string) {
	switch dst {
	case "":
		log = logmod.New(os.Stdout, "", logmod.LstdFlags)
	case "-": // syslog
		var err error
		log, err = syslog.NewLogger(syslog.LOG_NOTICE|syslog.LOG_LOCAL4,
			logmod.LstdFlags)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error: cannot connect to syslog!")
			os.Exit(1)
		}
	default:
		handler, err := os.OpenFile(opts.Log,
			os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error: cannot open log file!")
			os.Exit(1)
		}
		log = logmod.New(handler, "", logmod.LstdFlags)
	}
}
