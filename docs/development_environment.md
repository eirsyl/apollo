# Development Environment

The linux machine provided by idi is used for development. Paste the following
configuration into ~/.ssh/config and run `ssh sule` to connect with the
environment.

```
Host sule
    Hostname sule.sylliaas.no
    User eirsyl
    LogLevel QUIET
    RequestTTY yes
    RemoteCommand if [ $(tmux list-sessions | wc -l) -gt 0 ]; then tmux -CC attach -t 0; else tmux -CC; fi
    LocalForward 127.0.0.1:3000 127.0.0.1:3000
    LocalForward 127.0.0.1:9090 127.0.0.1:9090
    LocalForward 127.0.0.1:8080 127.0.0.1:8080
    LocalForward 127.0.0.1:8081 127.0.0.1:8081
    LocalForward 127.0.0.1:6000 127.0.0.1:6000
    LocalForward 127.0.0.1:7000 127.0.0.1:7000
    LocalForward 127.0.0.1:8000 127.0.0.1:8000
```
