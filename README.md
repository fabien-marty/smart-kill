# smart-kill

## What is it?

This is a little CLI tool written in [Golang](https://go.dev/) to:
- send a configurable signal (`SIGTERM` for example) to a process (given its pid)
- wait for it to stop for a set period of time
- send a `SIGKILL` if necessary

If the unix result code is `0`, the PID does not exist any more. 

## Non-goals

- Windows support (sorry 🤷‍♂️)

## How to install?

Go to [releases](https://github.com/fabien-marty/smart-kill/releases) and download the binary for your architecture.
Add the executable bit (`chmod +x`) and launch it.

Example without browser:

```console
OSARCH=linux-amd64
VERSION=v0.1.2
wget -O /usr/local/bin/smart-kill "https://github.com/fabien-marty/smart-kill/releases/download/${VERSION}/smart-kill-${VERSION}-${OSARCH}"
chmod +x /usr/local/bin/smart-kill
```

## How to use?

```console
$ ./smart-kill --help
NAME:
   smart-kill - Sends a signal to a process and waits for it to stop up to a certain length of time before sending a SIGKILL if necessary

USAGE:
   smart-kill [global options] command [command options] PROCESS_PID

COMMANDS:
   help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --log-level value  log level: DEBUG, INFO, WARN or ERROR (default: "INFO")
   --signal value     signal to sent (as integer) to the process (example: 15 for SIGTERM, 2 for SIGINT, 3 for SIGQUIT...) (default: 15)
   --wait-ms value    maximum number of milliseconds to wait after sending the signal (default: 5000)
   --help, -h         show help

EXIT CODES:
    - 0: the process PID does not exist any more
       (stopped or did not exist at program start)
    - 1: the process PID is still here after this program stopped :-(
    - 2: CLI error
```

## Roadmap

- [ ] multiple PIDs in parallel
- [ ] children handling
