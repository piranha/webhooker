// (c) 2013 Alexander Solovyov under terms of ISC License

package main

import (
	"encoding/json"
	"fmt"
	flags "github.com/jessevdk/go-flags"
	"log"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"io/ioutil"
)

/// Globals

var Version = "0.1"

var opts struct {
	Interface string `short:"i" long:"interface" default:"" description:"ip to listen on"`
	Port      string `short:"p" long:"port" default:"8000" description:"port to listen on"`
	Log       string `short:"l" long:"log" description:"path to file for logging"`
	Config    string `short:"c" long:"config" description:"read rules from this file"`
	Dump      bool   `short:"d" long:"dump" description:"dump rules to console"`
	ShowHelp  bool   `long:"help" description:"show this help message"`
}

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

func (c Config) Parse(input []string) error {
	RuleRe := regexp.MustCompile("([^:/]+/[^:/]+?):([^=]+?)=(.+)")

	for _, line := range input {
		bits := RuleRe.FindStringSubmatch(line)
		if len(bits) != 4 {
			return fmt.Errorf("Can't parse line '%s'", line)
		}

		name := bits[1]
		if _, ok := c[name]; !ok {
			c[name] = make(Rules, 0)
		}
		c[name] = append(c[name], Rule{bits[2], bits[3]})
	}

	return nil
}

func (c Config) HandleRequest(w http.ResponseWriter, r *http.Request) {
	var data GithubPayload
	err := json.Unmarshal([]byte(r.FormValue("payload")), &data)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	name := fmt.Sprintf("%s/%s",
		data.Repository.Owner.Name, data.Repository.Name)
	rules, ok := c[name]
	if !ok {
		log.Printf("No handlers for %s", name)
		return
	}

	branch := strings.TrimPrefix(data.Ref, "refs/heads/")

	executed := 0
	for _, rule := range rules {
		if rule.Branch == branch {
			run(rule.Command, name, &data)
			executed += 1
		}
	}
	if executed == 0 {
		log.Printf("No handlers for %s:%s", name, branch)
	}
}

func run(command string, repo string, data *GithubPayload) error {
	cmd := exec.Command("sh", "-c", command)

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

	argparser.Usage = `[OPTIONS] user/repo:branch=command [more rules...]

Runs specified shell commands on incoming webhook from Github. Shell command
environment contains:

  $REPO - repository name in "user/name" format
  $REPO_URL - full repository url
  $PRIVATE - strings "true" or "false" if repository is private or not
  $BRANCH - branch name
  $COMMIT - last commit hash id
  $COMMIT_MESSAGE - last commit message
  $COMMIT_TIME - last commit timestamp
  $COMMIT_AUTHOR - username of author of last commit
  $COMMIT_URL - full url to commit
`

	if opts.ShowHelp || (len(args) == 0 && opts.Config == "") {
		argparser.WriteHelp(os.Stdout)
		return
	}

	configureLogging(opts.Log)

	cfg := make(Config)
	if len(args) > 0 {
		errhandle(cfg.Parse(args), "")
	}
	if opts.Config != "" {
		data, err := ioutil.ReadFile(opts.Config)
		errhandle(err, "")
		bits := strings.Split(strings.TrimSpace(string(data)), "\n")
		errhandle(cfg.Parse(bits), "")
	}

	if opts.Dump {
		for repo, rules := range cfg {
			for _, rule := range rules {
				fmt.Printf("%s:%s='%s'\n", repo, rule.Branch, rule.Command)
			}
		}
		return
	}

	http.HandleFunc("/", cfg.HandleRequest)
	http.ListenAndServe(opts.Interface+":"+opts.Port, nil)
}

func configureLogging(dst string) {
	if dst == "" || dst == "-" {
		log.SetOutput(os.Stdout)
		return
	}

	file, err := os.OpenFile(opts.Log,
		os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	errhandle(err, "Error: cannot open log file!")
	log.SetOutput(file)
}

func errhandle(err error, msg string) {
	if err == nil {
		return
	}
	if msg == "" {
		msg = err.Error()
	}
	fmt.Fprintln(os.Stderr, msg)
	os.Exit(1)
}
