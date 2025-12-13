package main

import (
	"context"
	"os"
	"os/exec"
	"syscall"

	"icelock/log"

	llsys "github.com/landlock-lsm/go-landlock/landlock/syscall"
	"github.com/urfave/cli/v3"
)

const version = "25.12.4"

type config struct {
	LogLevel string

	FsRestricted bool
	FsRO         []string
	FsRX         []string
	FsRW         []string

	NetRestricted bool
	NetBindTCP    []int
	NetConnectTCP []int

	IpcScoped bool

	SeccompEnabled     bool
	SeccompPrint       bool
	SeccompKillBlocked bool

	Syscalls       []string
	SocketFamilies []string
}

func main() {
	cmd := &cli.Command{
		Name:                  "icelock",
		Usage:                 "tool for restricting programs with landlock",
		Version:               version,
		EnableShellCompletion: true,

		// disable the help subcommand ($ icelock help) since it's unintuitive
		HideHelpCommand: true,

		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "log-level",
				Usage:   "set the log level",
				Value:   "warn",
				Sources: cli.EnvVars("ICELOCK_LOG_LEVEL"),
			},

			&cli.BoolFlag{
				Name:     "unrestricted-fs",
				Usage:    "don't restrict filesystem access",
				Category: "Filesystem",
			},
			&cli.StringSliceFlag{
				Name:      "ro",
				Usage:     "allow read access beneath this path",
				Category:  "Filesystem",
				TakesFile: true,
			},
			&cli.StringSliceFlag{
				Name:      "rx",
				Usage:     "allow read/execute access beneath this path",
				Category:  "Filesystem",
				TakesFile: true,
			},
			&cli.StringSliceFlag{
				Name:      "rw",
				Usage:     "allow read/write access beneath this path",
				Category:  "Filesystem",
				TakesFile: true,
			},

			&cli.BoolFlag{
				Name:     "unrestricted-net",
				Usage:    "don't restrict network access",
				Category: "Network",
			},
			&cli.IntSliceFlag{
				Name:     "bind-tcp",
				Usage:    "allow binding to this TCP port",
				Category: "Network",
			},
			&cli.IntSliceFlag{
				Name:     "connect-tcp",
				Usage:    "allow connecting to this TCP port",
				Category: "Network",
			},

			&cli.BoolFlag{
				Name:  "unscoped-ipc",
				Usage: "don't scope IPC (signals and abstract unix sockets)",
			},

			&cli.BoolFlag{
				Name:     "no-seccomp",
				Usage:    "don't filter syscalls",
				Category: "Seccomp",
			},
			&cli.BoolFlag{
				Name:     "seccomp-print",
				Usage:    "print a human-readable version of the filter and exit",
				Category: "Seccomp",
			},
			&cli.BoolFlag{
				Name:     "seccomp-kill",
				Usage:    "if a syscall is blocked, kill the process",
				Category: "Seccomp",
			},

			&cli.StringSliceFlag{
				Name:     "syscalls",
				Usage:    `extra allowed syscall groups ("keyring", "chmod", "chown", "xattr", "privileged")`,
				Category: "Seccomp",
			},
			&cli.StringSliceFlag{
				Name:     "af",
				Usage:    `allowed socket address families ("netlink", "unix", "inet", "other")`,
				Category: "Seccomp",
			},
		},

		Action: func(ctx context.Context, cmd *cli.Command) error {
			args := cmd.Args().Slice()

			if len(args) == 0 {
				cli.ShowRootCommandHelpAndExit(cmd, 1)
			}

			appExe, err := exec.LookPath(args[0])
			if err != nil {
				log.Error("Failed to find the app exe: %v", err)
				os.Exit(1)
			}

			cfg := config{
				LogLevel: cmd.String("log-level"),

				FsRestricted: !cmd.Bool("unrestricted-fs"),
				FsRO:         cmd.StringSlice("ro"),
				FsRX:         cmd.StringSlice("rx"),
				FsRW:         cmd.StringSlice("rw"),

				NetRestricted: !cmd.Bool("unrestricted-net"),
				NetBindTCP:    cmd.IntSlice("bind-tcp"),
				NetConnectTCP: cmd.IntSlice("connect-tcp"),

				IpcScoped: !cmd.Bool("unscoped-ipc"),

				SeccompEnabled:     !cmd.Bool("no-seccomp"),
				SeccompPrint:       cmd.Bool("seccomp-print"),
				SeccompKillBlocked: cmd.Bool("seccomp-kill"),

				Syscalls:       cmd.StringSlice("syscalls"),
				SocketFamilies: cmd.StringSlice("af"),
			}

			log.SetLevel(cfg.LogLevel)
			log.Debug("Icelock config: %+v", cfg)

			landlockAbi, err := llsys.LandlockGetABIVersion()
			if err != nil {
				log.Error("Kernel doesn't have landlock enabled")
				os.Exit(1)
			}
			log.Debug("Landlock ABI version: %d", landlockAbi)

			setupLandlock(&cfg)

			// separate function for now so it's easy to remove it in case it breaks
			// things, since we're getting ipc scoping support from the "scoped" branch
			// of go-landlock
			setupLandlockIpc(&cfg)

			setupSeccomp(&cfg)

			log.Info("Executing: %s, args: %v", appExe, getAppArgs(args))
			err = syscall.Exec(appExe, args, os.Environ())
			if err != nil {
				log.Error("Failed to run the app: %v", err)
				os.Exit(1)
			}

			return nil
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Error("%s", err.Error())
	}
}

func getAppArgs(args []string) []string {
	var appArgs []string

	for i, arg := range args {
		if i == 0 {
			continue
		}
		appArgs = append(appArgs, arg)
	}

	return appArgs
}
