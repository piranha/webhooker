#!/usr/bin/env prysk

webhooker tests init:

  $ TARGET="$PWD"
  $ (cd $(dirname "$TESTDIR") && go build && mv webhooker "$TARGET")
  $ POST() {
  >   curl -s --data-urlencode payload@$1 http://localhost:1234/
  > }
  $ LINES=0
  $ LOGS() {
  > awk "NR > $LINES" logs
  > LINES=$(cat logs | wc -l)
  > }
  $ ./webhooker -p 1234 \
  > 'octokitty/testing:main=echo OTM' \
  > 'hellothere/.+=echo $REPO' \
  > 'invalid/command:=does-not-exist' \
  > > logs &

Usage:

  $ ./webhooker
  Usage:
    webhooker [OPTIONS] user/repo:branch=command [more rules...]
  
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
  
  
  Application Options:
    -i, --interface= ip to listen on (default: 127.0.0.1)
    -p, --port=      port to listen on (default: 3434)
    -l, --log=       path to file for logging
    -c, --config=    read rules linewise from this file
    -d, --dump       dump rules to console
    -V, --version    show version and exit
        --help       show this help message

Check that it works:

  $ POST $TESTDIR/example.json
  'echo OTM' for 'octokitty/testing:main' output:
  OTM
  $ POST $TESTDIR/other.json
  'echo $REPO' for 'hellothere/other:lolster' output:
  hellothere/other
  $ LOGS
  [\d/: ]+ 'echo OTM' for 'octokitty/testing:main' output: OTM (re)
  [\d/: ]+ 'echo \$REPO' for 'hellothere/other:lolster' output: hellothere/other (re)

Check that errors are correctly processed:

  $ POST nothing
  unexpected end of JSON input
  $ POST $TESTDIR/nocommit.json
  No handlers for 'piranha/unknown:'
  $ POST $TESTDIR/cmderror.json
  'does-not-exist' for 'invalid/command:' output:
  sh: does-not-exist: command not found
  
  $ LOGS
  [\d/: ]+ unexpected end of JSON input (re)
  [\d/: ]+ No handlers for 'piranha/unknown:' (re)
  [\d/: ]+ 'does-not-exist' for 'invalid/command:' output: sh: (1:)? does-not-exist: command not found (re)
  [\d/: ]+ exit status 127 (re)
 

Cool down:

  $ kill $!
