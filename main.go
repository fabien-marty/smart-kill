package main

import (
	"fmt"
	"os"
	"strconv"
	"syscall"
	"time"

	"github.com/urfave/cli/v2"
	"golang.org/x/exp/slog"
)

const waitMsAfterEachLoopIteration = 100

// GetProcess returns an os.Process object corresponding to an existing process (given its pid) or nil
func GetProcess(pid int) *os.Process {
	process, err := os.FindProcess(pid)
	if err != nil {
		return nil
	}
	// note: On Unix systems, FindProcess always succeeds and returns a Process for the given pid,
	//       regardless of whether the process exists. To test whether the process actually exists,
	//       see whether p.Signal(syscall.Signal(0)) reports an error.
	// => see https://pkg.go.dev/os#FindProcess
	err = process.Signal(syscall.Signal(0))
	if err != nil {
		return nil
	}
	return process
}

func SmartKill(logger *slog.Logger, pid int, signal syscall.Signal, maxWaitMs int) bool {
	clogger := logger.With(
		slog.Int("process", pid),
	)
	process := GetProcess(pid)
	if process == nil {
		clogger.Warn("can't find process")
		return true
	}
	err := process.Signal(signal)
	if err != nil {
		if GetProcess(pid) != nil {
			clogger.Warn(fmt.Sprintf("can't send signal: %d", int(signal)))
			return false
		}
		clogger.Warn("the process stopped without sending signal to it!")
		return true
	}
	clogger.Info(fmt.Sprintf("signal: %d sent", int(signal)))
	start := time.Now()
	for {
		time.Sleep(waitMsAfterEachLoopIteration * time.Millisecond)
		process = GetProcess(pid)
		elapsed := time.Since(start)
		if process == nil {
			clogger.Info(fmt.Sprintf("process stopped after %s", elapsed))
			return true
		}
		logger.Debug(fmt.Sprintf("process still here after %s", elapsed))
		if elapsed.Milliseconds() >= int64(maxWaitMs) {
			err = process.Signal(syscall.SIGKILL)
			if err != nil {
				if GetProcess(pid) != nil {
					clogger.Warn(fmt.Sprintf("can't send SIGKILL to process: %v", err))
					break
				}
			}
			clogger.Info("signal: 9 (SIGKILL) sent")
			break
		}
	}
	// SIGKILL sent => let's wait up to 1s to be sure than the process is gone
	before := time.Now()
	for {
		time.Sleep(waitMsAfterEachLoopIteration * time.Millisecond)
		process = GetProcess(pid)
		elapsed := time.Since(start)
		if process == nil {
			clogger.Info(fmt.Sprintf("process stopped after %s", elapsed))
			return true
		}
		if time.Since(before).Milliseconds() >= 1000 {
			break
		}
	}
	// SIGKILL sent but process still here after 1s
	clogger.Warn("SIGKILL sent but process still here after 1s => giving up")
	return false
}

func main() {
	cli.AppHelpTemplate += `
EXIT CODES:
    - 0: the process PID does not exist any more 
	     (stopped or did not exist at program start)
    - 1: the process PID is still here after this program stopped :-(
    - 2: CLI error
`
	app := &cli.App{
		Name:      "smart-kill",
		Usage:     "Sends a signal to a process and waits for it to stop up to a certain length of time before sending a SIGKILL if necessary",
		ArgsUsage: "PROCESS_PID",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "log-level",
				Value: "INFO",
				Usage: "log level: DEBUG, INFO, WARN or ERROR",
			},
			&cli.IntFlag{
				Name:  "signal",
				Value: 15,
				Usage: "signal to sent (as integer) to the process (example: 15 for SIGTERM, 2 for SIGINT, 3 for SIGQUIT...)",
			},
			&cli.IntFlag{
				Name:  "wait-ms",
				Value: 5000,
				Usage: "maximum number of milliseconds to wait after sending the signal",
			},
		},
		Action: func(context *cli.Context) error {
			var logLevel slog.Leveler
			switch context.String("log-level") {
			case "DEBUG":
				logLevel = slog.LevelDebug
			case "INFO":
				logLevel = slog.LevelInfo
			case "WARN":
				logLevel = slog.LevelWarn
			case "ERROR":
				logLevel = slog.LevelError
			default:
				return cli.Exit(fmt.Sprintf("ERROR: unsupported log-level: %s\n", context.String("log-leel")), 2)
			}
			opts := &slog.HandlerOptions{
				Level: logLevel,
			}
			logger := slog.New(slog.NewTextHandler(os.Stdout, opts))
			args := context.Args()
			if args.Len() != 1 {
				return cli.Exit("ERROR: you have to provide exactly one argument: the process PID", 2)
			}
			pid, err := strconv.Atoi(args.Get(0))
			if err != nil {
				return cli.Exit(fmt.Sprintf("ERROR: the process PID must be an integer: %v\n", err), 2)
			}
			res := SmartKill(logger, pid, syscall.Signal(context.Int("signal")), context.Int("wait-ms"))
			if !res {
				os.Exit(1)
			}
			os.Exit(0)
			return nil
		},
	}
	if err := app.Run(os.Args); err != nil {
		fmt.Printf("ERROR: %v\n", err)
	}
}
