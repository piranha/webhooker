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

var Version = "0.2"

var opts struct {
	Interface string `short:"i" long:"interface" default:"127.0.0.1" description:"ip to listen on"`
	Port      string `short:"p" long:"port"      default:"8000"      description:"port to listen on"`
	Log       string `short:"l" long:"log"                           description:"path to file for logging"`
	Config    string `short:"c" long:"config"                        description:"read rules from this file"`
	Dump      bool   `short:"d" long:"dump"                          description:"dump rules to console"`
	Version   bool   `short:"V" long:"version"                       description:"show version and exit"`
	ShowHelp  bool   `          long:"help"                          description:"show this help message"`
}

/// Core interfaces

type Payload interface {
	RepoName() string
	BranchName() string
	EnvData() []string
}

func GetPath(p Payload) string {
	return p.RepoName() + ":" + p.BranchName()
}

type Rule interface {
	Match(path string) bool
	Run(data Payload) error
	String() string
}

type Config []Rule

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

func (g *GithubPayload) RepoName() string {
	return g.Repository.Owner.Name+"/"+g.Repository.Name
}

func (g *GithubPayload) BranchName() string {
	return strings.TrimPrefix(g.Ref, "refs/heads/")
}

func (g *GithubPayload) EnvData() []string {
	commit := g.Commits[0]

	return []string{
		env("REPO", g.RepoName()),
		env("REPO_URL", g.Repository.Url),
		env("PRIVATE", fmt.Sprintf("%t", g.Repository.Private)),
		env("BRANCH", g.Ref),
		env("COMMIT", commit.Id),
		env("COMMIT_MESSAGE", commit.Message),
		env("COMMIT_TIME", commit.Timestamp),
		env("COMMIT_AUTHOR", commit.Author.Name),
		env("COMMIT_URL", commit.Url),
	}
}

/// Rule implementation

type PatRule struct {
	Pattern string
	Command string
	Re      *regexp.Regexp
}

func (r *PatRule) Match(path string) bool {
	return r.Re.MatchString(path)
}

func (r *PatRule) String() string {
	return fmt.Sprintf("%s='%s'", r.Pattern, r.Command)
}

func (r *PatRule) Run(data Payload) error {
	cmd := exec.Command("sh", "-c", r.Command)

	cmd.Env = append(data.EnvData(),
		env("PATH", os.Getenv("PATH")),
		env("HOME", os.Getenv("HOME")),
		env("USER", os.Getenv("USER")),
		)

	out, err := cmd.CombinedOutput()
	log.Printf("'%s' for %s output: %s", r.Command, data.RepoName(), out)
	return err
}

/// actual work

func (c *Config) ParsePatterns(input []string) error {
	for _, line := range input {
		bits := strings.SplitN(line, "=", 2)
		if len(bits) != 2 {
			return fmt.Errorf("Can't parse line '%s'", line)
		}

		re, err := regexp.Compile(bits[0])
		if err != nil {
			return fmt.Errorf("Line '%s' is not a valid regular expression",
				line)
		}

		*c = append(*c, &PatRule{bits[0], bits[1], re})
	}

	return nil
}

func (c Config) ExecutePayload(data Payload) error {
	path := GetPath(data)

	for _, rule := range c {
		if rule.Match(path) {
			return rule.Run(data)
		}
	}

	log.Printf("No handlers for %s", path)
	return nil
}

func (c Config) HandleRequest(w http.ResponseWriter, r *http.Request) {
	data := new(GithubPayload)
	err := json.Unmarshal([]byte(r.PostFormValue("payload")), data)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	c.ExecutePayload(data)
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

  $PATH - proxied from parent environment
  $HOME - proxied from parent environment
  $USER - proxied from parent environment

  $REPO - repository name in "user/name" format
  $REPO_URL - full repository url
  $PRIVATE - strings "true" or "false" if repository is private or not
  $BRANCH - branch name
  $COMMIT - last commit hash id
  $COMMIT_MESSAGE - last commit message
  $COMMIT_TIME - last commit timestamp
  $COMMIT_AUTHOR - username of author of last commit
  $COMMIT_URL - full url to commit

'user/repo:branch' pattern is a regular expression, so you could do
'user/project:fix.*=cmd' or even '.*=cmd'.

Also never forget to properly escape your rule, if you pass it through command
line: usually enclosing it in single quotes (') is enough.
`

	if opts.ShowHelp || (len(args) == 0 && opts.Config == "") {
		argparser.WriteHelp(os.Stdout)
		return
	}

	configureLogging(opts.Log)

	cfg := Config{}
	if len(args) > 0 {
		errhandle(cfg.ParsePatterns(args), "")
	}
	if opts.Config != "" {
		data, err := ioutil.ReadFile(opts.Config)
		errhandle(err, "")
		bits := strings.Split(strings.TrimSpace(string(data)), "\n")
		errhandle(cfg.ParsePatterns(bits), "")
	}

	if opts.Dump {
		for _, rule := range cfg {
			fmt.Println(rule)
		}
		return
	}

	http.HandleFunc("/", cfg.HandleRequest)
	http.ListenAndServe(opts.Interface+":"+opts.Port, nil)
}

/// Utils

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

func env(key string, value string) string {
	return fmt.Sprintf("%s=%s", key, value)
}
