# webhooker

An application to run shell commands on incoming WebHooks from Github.

[![Build Status](https://travis-ci.org/piranha/webhooker.png)](https://travis-ci.org/piranha/webhooker)

## Installation

Install it with `go get github.com/piranha/webhooker` or download a binary from
[releases page](https://github.com/piranha/webhooker/releases).

## Usage

You run it like this (run webhooker without arguments to get help - you could
also put all rules in a separate config file):

```
./webhooker -p 3456 -i 127.0.0.1 piranha/webhooker:master='echo $COMMIT'
```

It runs every command in `sh`, so you can use more complex commands (with `&&`
and `|`).

`user/repo:branch` pattern is a regular expression, so you could do
`user/project:fix.*=cmd` or even `.*=cmd`.

## Running

I expect you to run it behind your HTTP proxy of choice, and in my case it's
nginx and such config is used to protect it from unwanted requests:

```
    location /webhook {
        proxy_pass http://localhost:3456;
        allow 204.232.175.64/27;
        allow 192.30.252.0/22;
        deny all;
    }
```

Or in caddy:

```
  @webhook {
    remote_ip 192.30.252.0/22 185.199.108.0/22 140.82.112.0/20
    path /webhook
  }
  reverse_proxy @webhook localhost:3456
```

After that I can put `http://domain.my/webhook` in Github's repo settings
WebHook URLs and press 'Test Hook' to check if it works.

## Environment

webhooker provides your commands with some variables in case you need them:

- `$REPO` - repository name in "user/name" format
- `$REPO_URL` - full repository url
- `$PRIVATE` - strings "true" or "false" if repository is private or not
- `$BRANCH` - branch name
- `$COMMIT` - last commit hash id
- `$COMMIT_MESSAGE` - last commit message
- `$COMMIT_TIME` - last commit timestamp
- `$COMMIT_AUTHOR` - username of author of last commit
- `$COMMIT_URL` - full url to commit

And, of course, it passes through some common variables: `$PATH`, `$HOME`,
`$USER`.

## Example

I render my own sites using `webhooker`. Using systemd you can put that in `/etc/systemd/system/webhooker.service`:

```
[Unit]
Description=webhooker

[Service]
StandardOutput=syslog
StandardError=syslog
SyslogIdentifier=webhooker
User=piranha
Group=piranha
Environment="CACHE_DIR=/home/piranha/bin"
ExecStart=/usr/local/bin/webhooker -p 3456 -i 127.0.0.1 \
	piranha/solovyov.net:master='cd /opt/solovyov.net && git pull -q && make'

[Install]
WantedBy=default.target
```

You can see that it updates and renders site on push. Run `systemctl
daemon-reload && systemctl start webhooker` to run this stuff.
