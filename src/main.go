package main

import (
	"context"
	"golang.org/x/sys/unix"
	"os"
	"os/exec"
	"syscall"

	"icelock/log"

	llsys "github.com/landlock-lsm/go-landlock/landlock/syscall"
	"github.com/urfave/cli/v3"
)

const version = "26.06.1"

const (
	flagLogLevel   = "log-level"
	flagBestEffort = "best-effort"

	flagUnrestrictedFs = "unrestricted-fs"
	flagRO             = "ro"
	flagRX             = "rx"
	flagRW             = "rw"
	flagUnix           = "unix"

	flagUnrestrictedNet = "unrestricted-net"
	flagBind            = "bind"
	flagBindAll         = "bind-all"
	flagConnect         = "connect"
	flagConnectAll      = "connect-all"

	flagSignals      = "signals"
	flagAbstractUnix = "abstract-unix"

	flagAudit             = "audit"
	flagNoAuditSubdomains = "no-audit-subdomains"

	flagNoSeccomp       = "no-seccomp"
	flagSeccompPrint    = "seccomp-print"
	flagSeccompPrintBPF = "seccomp-print-bpf"
	flagSyscalls        = "syscalls"
	flagAF              = "af"
	flagUserns          = "userns"
	flagIoUring         = "io-uring"

	flagKeepCaps = "keep-caps"

	flagMdwe = "mdwe"
)

const (
	categoryFilesystem = "Filesystem"
	categoryNetwork    = "Network"
	categoryIpcScoping = "IPC Scoping"
	categoryAudit      = "Audit Subsystem Logging"
	categorySeccomp    = "Seccomp"
)

type config struct {
	LogLevel   string
	BestEffort bool

	FsRestricted bool
	FsRO         []string
	FsRX         []string
	FsRW         []string
	FsUnix       []string

	NetRestricted bool
	NetBind       []int
	NetBindAll    bool
	NetConnect    []int
	NetConnectAll bool

	SignalsScoped      bool
	AbstractUnixScoped bool

	AuditLogNewExecOn     bool
	AuditLogSubdomainsOff bool

	SeccompEnabled  bool
	SeccompPrint    bool
	SeccompPrintBPF bool
	Syscalls        []string
	SocketFamilies  []string
	UserNamespaces  bool
	IoUring         bool

	KeepCaps bool

	Mdwe bool
}

func main() {
	cmd := &cli.Command{
		Name:                  "icelock",
		Usage:                 "tool for restricting programs with landlock",
		Version:               version,
		EnableShellCompletion: true,

		// disable the help subcommand ($ icelock help) since it's unintuitive
		HideHelpCommand: true,

		MutuallyExclusiveFlags: []cli.MutuallyExclusiveFlags{
			{
				Category: categorySeccomp,
				Flags: [][]cli.Flag{
					{
						&cli.BoolFlag{
							Name:  flagSeccompPrint,
							Usage: "print a human-readable version of the filter and exit",
						},
					},
					{
						&cli.BoolFlag{
							Name:  flagSeccompPrintBPF,
							Usage: "print the raw BPF filter and exit",
						},
					},
				},
			},
		},

		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    flagLogLevel,
				Usage:   `set the log level ("debug", "info", "warn", "error")`,
				Value:   "warn",
				Sources: cli.EnvVars("ICELOCK_LOG_LEVEL"),
			},
			&cli.BoolFlag{
				Name:  flagBestEffort,
				Usage: "disable unsupported landlock features",
			},

			&cli.BoolFlag{
				Name:     flagUnrestrictedFs,
				Usage:    "don't restrict filesystem access",
				Category: categoryFilesystem,
			},
			&cli.StringSliceFlag{
				Name:      flagRO,
				Usage:     "allow read access beneath this path",
				Category:  categoryFilesystem,
				TakesFile: true,
			},
			&cli.StringSliceFlag{
				Name:      flagRX,
				Usage:     "allow read/execute access beneath this path",
				Category:  categoryFilesystem,
				TakesFile: true,
			},
			&cli.StringSliceFlag{
				Name:      flagRW,
				Usage:     "allow read/write access beneath this path",
				Category:  categoryFilesystem,
				TakesFile: true,
			},
			&cli.StringSliceFlag{
				Name:      flagUnix,
				Usage:     "allow unix socket access beneath this path",
				Category:  categoryFilesystem,
				TakesFile: true,
			},

			&cli.BoolFlag{
				Name:     flagUnrestrictedNet,
				Usage:    "don't restrict network access",
				Category: categoryNetwork,
			},
			&cli.IntSliceFlag{
				Name:     flagBind,
				Usage:    "allow binding to this TCP port",
				Category: categoryNetwork,
			},
			&cli.BoolFlag{
				Name:     flagBindAll,
				Usage:    "allow binding to all TCP ports",
				Category: categoryNetwork,
			},
			&cli.IntSliceFlag{
				Name:     flagConnect,
				Usage:    "allow connecting to this TCP port",
				Category: categoryNetwork,
			},
			&cli.BoolFlag{
				Name:     flagConnectAll,
				Usage:    "allow connecting to all TCP ports",
				Category: categoryNetwork,
			},

			&cli.BoolFlag{
				Name:     flagSignals,
				Usage:    "don't scope signals",
				Category: categoryIpcScoping,
			},
			&cli.BoolFlag{
				Name:     flagAbstractUnix,
				Usage:    "don't scope abstract unix sockets",
				Category: categoryIpcScoping,
			},

			&cli.BoolFlag{
				Name:     flagAudit,
				Usage:    "turn on logging for the app",
				Category: categoryAudit,
				Sources:  cli.EnvVars("ICELOCK_AUDIT"),
			},
			&cli.BoolFlag{
				Name:     flagNoAuditSubdomains,
				Usage:    "turn off logging for nested landlock domains",
				Category: categoryAudit,
				Sources:  cli.EnvVars("ICELOCK_NO_AUDIT_SUBDOMAINS"),
			},

			&cli.BoolFlag{
				Name:     flagNoSeccomp,
				Usage:    "don't filter syscalls",
				Category: categorySeccomp,
			},
			&cli.StringSliceFlag{
				Name:     flagSyscalls,
				Usage:    `extra allowed syscall groups ("keyring", "debug", "mq", "chmod", "chown", "xattr", "emulation", "privileged")`,
				Category: categorySeccomp,
			},
			&cli.StringSliceFlag{
				Name:     flagAF,
				Usage:    `allowed socket address families ("netlink", "inet", "other")`,
				Category: categorySeccomp,
			},
			&cli.BoolFlag{
				Name:     flagUserns,
				Usage:    "allow creating user namespaces",
				Category: categorySeccomp,
			},
			&cli.BoolFlag{
				Name:     flagIoUring,
				Usage:    "allow using io_uring",
				Category: categorySeccomp,
			},

			&cli.BoolFlag{
				Name:  flagKeepCaps,
				Usage: "don't drop capabilities",
			},

			&cli.BoolFlag{
				Name:  flagMdwe,
				Usage: "block W&X memory with the PR_MDWE_REFUSE_EXEC_GAIN prctl flag",
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
				LogLevel:   cmd.String(flagLogLevel),
				BestEffort: cmd.Bool(flagBestEffort),

				FsRestricted: !cmd.Bool(flagUnrestrictedFs),
				FsRO:         cmd.StringSlice(flagRO),
				FsRX:         cmd.StringSlice(flagRX),
				FsRW:         cmd.StringSlice(flagRW),
				FsUnix:       cmd.StringSlice(flagUnix),

				NetRestricted: !cmd.Bool(flagUnrestrictedNet),
				NetBind:       cmd.IntSlice(flagBind),
				NetBindAll:    cmd.Bool(flagBindAll),
				NetConnect:    cmd.IntSlice(flagConnect),
				NetConnectAll: cmd.Bool(flagConnectAll),

				SignalsScoped:      !cmd.Bool(flagSignals),
				AbstractUnixScoped: !cmd.Bool(flagAbstractUnix),

				AuditLogNewExecOn:     cmd.Bool(flagAudit),
				AuditLogSubdomainsOff: cmd.Bool(flagNoAuditSubdomains),

				SeccompEnabled:  !cmd.Bool(flagNoSeccomp),
				SeccompPrint:    cmd.Bool(flagSeccompPrint),
				SeccompPrintBPF: cmd.Bool(flagSeccompPrintBPF),
				Syscalls:        cmd.StringSlice(flagSyscalls),
				SocketFamilies:  cmd.StringSlice(flagAF),
				UserNamespaces:  cmd.Bool(flagUserns),
				IoUring:         cmd.Bool(flagIoUring),

				KeepCaps: cmd.Bool(flagKeepCaps),

				Mdwe: cmd.Bool(flagMdwe),
			}

			log.SetLevel(cfg.LogLevel)
			log.Debug("Icelock config: %+v", cfg)

			landlockAbi, err := llsys.LandlockGetABIVersion()
			if err != nil {
				log.Error("Kernel doesn't have landlock enabled")
				os.Exit(1)
			}
			log.Debug("Landlock ABI version: %d", landlockAbi)

			// libcap uses libpsx which needs to read /proc/pid/task, so we have
			// to drop caps before setting up landlock
			// (go-landlock used to implicitly add a rule to allow that, but now it
			// only does that if LANDLOCK_RESTRICT_SELF_TSYNC isn't available)
			setupCaps(&cfg)

			setupLandlock(&cfg)
			setupSeccomp(&cfg)
			setupMdwe(&cfg)

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

func setupMdwe(cfg *config) {
	if cfg.Mdwe {
		err := unix.Prctl(unix.PR_SET_MDWE, unix.PR_MDWE_REFUSE_EXEC_GAIN, 0, 0, 0)
		if err != nil {
			log.Error("Failed to set PR_MDWE_REFUSE_EXEC_GAIN: %v", err)
			os.Exit(1)
		}
		log.Info("Set PR_MDWE_REFUSE_EXEC_GAIN")
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
