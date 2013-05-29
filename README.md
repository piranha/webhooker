# webhooker

An application to run shell commands on incoming WebHooks from Github.

## Installation

Install it with `go get github.com/piranha/webhooker` or use 64 bit binary:

 - [Linux](http://solovyov.net/files/webhooker-linux)
 - [OS X](http://solovyov.net/files/webhooker-osx)
 - [Windows](http://solovyov.net/files/webhooker-win.exe)

## Usage

You run it like this (run webhooker without arguments to get help - you could
also put all rules in a separate config file):

```
./webhooker -p 3456 -i 127.0.0.1 piranha/webhooker:master='echo $COMMIT'
```

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

After that I can put `http://domain.my/webhook/` in Github's repo settings
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
