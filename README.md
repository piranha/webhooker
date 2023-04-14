# webhooker

An application to run shell commands on incoming WebHooks from Github.

[![Build Status](https://travis-ci.org/piranha/webhooker.png)](https://travis-ci.org/piranha/webhooker)

## Installation

```
cd /usr/local/bin
curl -sL https://github.com/piranha/webhooker/releases/download/1.2/webhooker-linux.gz | gunzip > webhooker && chmod +x webhooker
```

or something like this.

## Usage

You run it like this (see `webhooker --help` to get more help):

```
webhooker -p 3434 -i 127.0.0.1 piranha/webhooker:main='echo $COMMIT'
```

It runs every command in `sh`, so you can use more complex commands (with `&&`
and `|`).

`user/repo:branch` pattern is a regular expression, so you could do
`user/project:fix.*=cmd` or even `.*=cmd`.

You can put all your configuration in a file line-by-line and then run like
`webhooker -c this-file`.

## Running

Maybe you want (like I do) to run this with systemd, create
`/etc/systemd/system/webhooker.service` with the following content:

```
[Unit]
Description=webhooker

[Service]
StandardOutput=syslog
StandardError=syslog
SyslogIdentifier=webhooker
User=piranha
Group=piranha
ExecStart=/usr/local/bin/webhooker -p 3434 \
    user/repo:main='cd /opt/repo && git fetch -q && git reset --hard origin/main && restart it or something'

[Install]
WantedBy=default.target
```

And then tell systemd to pick it up: `systemctl daemon-reload && systemctl start
webhooker && systemctl enable webhooker`.

I expect you to run it behind your HTTP proxy of choice, and in my case it's
nginx and such config is used to protect it from unwanted requests:

```
    location /webhook {
        proxy_pass http://localhost:3434;
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
  reverse_proxy @webhook localhost:3434
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
