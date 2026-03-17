package main

import (
	"golang.org/x/sys/unix"
	"os"

	"icelock/log"

	"github.com/seccomp/libseccomp-golang"
)

type seccompRule struct {
	Action   seccomp.ScmpAction
	Syscall  seccomp.ScmpSyscall
	Arg      uint
	Op       seccomp.ScmpCompareOp
	OpValue1 uint64
	OpValue2 uint64
}

var (
	actionAllow        = seccomp.ActAllow
	actionEperm        = seccomp.ActErrno.SetReturnCode(int16(unix.EPERM))
	actionEnosys       = seccomp.ActErrno.SetReturnCode(int16(unix.ENOSYS))
	actionEafnosupport = seccomp.ActErrno.SetReturnCode(int16(unix.EAFNOSUPPORT))
)

func setupSeccomp(cfg *config) {

	if cfg.SeccompEnabled {
		filter, err := seccomp.NewFilter(actionAllow)
		if err != nil {
			log.Error("Failed to create a seccomp filter context: %v", err)
			os.Exit(1)
		}

		blockedSyscallsEperm := []string{}
		blockedSyscallsEnosys := []string{}

		rules := []seccompRule{
			// modern distros should disable TIOCSTI, but we block it in case they don't.
			// See https://wiki.gnoack.org/TiocstiTioclinuxSecurityProblems
			//
			// NOTE: if TIOCSTI is blocked by the sysctl, the error will be EIO
			{
				Action:   actionEperm,
				Syscall:  unix.SYS_IOCTL,
				Arg:      1,
				Op:       seccomp.CompareMaskedEqual,
				OpValue1: 0xFFFFFFFF,
				OpValue2: unix.TIOCSTI,
			},

			// TIOCLINUX needs CAP_SYS_ADMIN, but we block it anyway as there's no
			// legitimate reason to use it
			{
				Action:   actionEperm,
				Syscall:  unix.SYS_IOCTL,
				Arg:      1,
				Op:       seccomp.CompareMaskedEqual,
				OpValue1: 0xFFFFFFFF,
				OpValue2: unix.TIOCLINUX,
			},
		}

		var sysKeyring, sysMq, sysChmod, sysChown, sysXattr, sysEmulation, sysPrivileged bool

		for _, group := range cfg.Syscalls {
			switch group {
			case "keyring":
				sysKeyring = true

			case "mq":
				sysMq = true

			case "chmod":
				sysChmod = true

			case "chown":
				sysChown = true

			case "xattr":
				sysXattr = true

			case "emulation":
				sysEmulation = true

			case "privileged", "priv":
				sysPrivileged = true

			default:
				log.Warn("Unknown syscall group '%s', skipping", group)
			}
		}

		// blocking kernel keyring access would have prevented CVE-2024-42318 (landlock bypass),
		// and most things don't need it anyway
		if !sysKeyring {
			blockedSyscallsEperm = append(blockedSyscallsEperm,
				"add_key",
				"keyctl",
				"request_key",
			)
		}

		// block message queues since that's ipc which landlock can't (fully) restrict yet,
		// and they're very rarely used anyway
		// see https://github.com/landlock-lsm/linux/issues/29
		// and https://github.com/landlock-lsm/linux/issues/30
		if !sysMq {
			blockedSyscallsEperm = append(blockedSyscallsEperm,
				// posix
				"mq_getsetattr",
				"mq_notify",
				"mq_open",
				"mq_timedreceive",
				"mq_timedsend",
				"mq_unlink",

				// system v
				"msgget",
				"msgsnd",
				"msgrcv",
				"msgctl",
			)
		}

		// landlock can't restrict chmod as of ABI v7
		// see https://github.com/landlock-lsm/linux/issues/11
		if !sysChmod && cfg.FsRestricted {
			blockedSyscallsEperm = append(blockedSyscallsEperm,
				"chmod",
				"fchmod",
				"fchmodat",
				"fchmodat2",
			)
		}

		// landlock can't restrict chown as of ABI v7
		if !sysChown && cfg.FsRestricted {
			blockedSyscallsEperm = append(blockedSyscallsEperm,
				"chown",
				"chown32",
				"fchown",
				"fchown32",
				"fchownat",
				"lchown",
				"lchown32",
			)
		}

		// landlock can't restrict xattrs as of ABI v7, so block changing them (but not reading them)
		if !sysXattr && cfg.FsRestricted {
			blockedSyscallsEperm = append(blockedSyscallsEperm,
				"setxattr",
				"setxattrat",
				"lsetxattr",
				"fsetxattr",

				"removexattr",
				"removexattrat",
				"lremovexattr",
				"fremovexattr",
			)
		}

		// reduce the exposed kernel surface
		if !sysEmulation {
			blockedSyscallsEperm = append(blockedSyscallsEperm,
				"modify_ldt",
			)

			// only allow PER_LINUX (0). Ideally we would also allow 0xFFFFFFFF (special value
			// that's used for checking the current personality without changing it) but I don't
			// think that's possible with libseccomp on a denylist filter
			rules = append(rules,
				seccompRule{
					Action:   actionEperm,
					Syscall:  unix.SYS_PERSONALITY,
					Arg:      0,
					Op:       seccomp.CompareNotEqual,
					OpValue1: 0,
				},
			)
		}

		// block some syscalls that unprivileged processes shouldn't be able to use anyway
		// to reduce the exposed kernel surface
		if !sysPrivileged {
			blockedSyscallsEperm = append(blockedSyscallsEperm,
				// @module systemd group
				"delete_module",
				"finit_module",
				"init_module",

				// @reboot systemd group
				"kexec_file_load",
				"kexec_load",
				"reboot",

				// @swap systemd group
				"swapoff",
				"swapon",

				// these *could* be used by unprivileged processes, but most distros
				// disable that
				"bpf",
				"perf_event_open",
				"syslog", // dmesg

				// misc
				"acct",
				"lookup_dcookie",
				"vhangup",
			)
		}

		if !cfg.UserNamespaces {
			rules = append(rules,
				// TODO: test on arm64 (since apparently the clone() syscall args order is
				//  different there)
				seccompRule{
					Action:   actionEperm,
					Syscall:  unix.SYS_CLONE,
					Arg:      0,
					Op:       seccomp.CompareMaskedEqual,
					OpValue1: unix.CLONE_NEWUSER,
					OpValue2: unix.CLONE_NEWUSER,
				},

				seccompRule{
					Action:   actionEperm,
					Syscall:  unix.SYS_UNSHARE,
					Arg:      0,
					Op:       seccomp.CompareMaskedEqual,
					OpValue1: unix.CLONE_NEWUSER,
					OpValue2: unix.CLONE_NEWUSER,
				},
			)

			// seccomp can't look into clone3() args, so we block it with ENOSYS so that the app
			// falls back to clone(). Flatpak also does this
			blockedSyscallsEnosys = append(blockedSyscallsEnosys, "clone3")
		}

		// io_uring can be used to create sockets without using the socket() syscall [1]
		// which in theory allows escaping the sandbox via unix sockets
		//
		// io_uring has also caused vulnerabilities in the past [2]
		//
		// [1]: https://gist.github.com/rusty-snake/4fc70c99b9ec53292c90546b790f7429
		// [2]: https://security.googleblog.com/2023/06/learnings-from-kctf-vrps-42-linux.html
		//
		// ENOSYS so that the app falls back to something else
		if !cfg.IoUring {
			blockedSyscallsEnosys = append(blockedSyscallsEnosys,
				"io_uring_enter",
				"io_uring_register",
				"io_uring_setup",
			)
		}

		for _, syscall := range blockedSyscallsEperm {
			err := filter.AddRule(getSyscall(syscall), actionEperm)
			if err != nil {
				panic(err)
			}
		}
		for _, syscall := range blockedSyscallsEnosys {
			err := filter.AddRule(getSyscall(syscall), actionEnosys)
			if err != nil {
				panic(err)
			}
		}
		log.Debug("Blocked syscalls: %v (EPERM), %v (ENOSYS)", blockedSyscallsEperm, blockedSyscallsEnosys)

		var afNetlink, afUnix, afInet, afOther bool

		for _, af := range cfg.SocketFamilies {
			switch af {
			case "netlink":
				afNetlink = true
			case "unix":
				afUnix = true
			case "inet":
				afInet = true
			case "other":
				afOther = true

			case "inet6":
				log.Warn("AF_INET6 is included with '--af inet', use that instead of '--af inet6'")

			default:
				log.Warn("Unknown address family '%s', skipping", af)
			}
		}

		blockedAf := []uint64{}

		if !afNetlink {
			blockedAf = append(blockedAf, unix.AF_NETLINK)
		}

		if !afUnix {
			blockedAf = append(blockedAf, unix.AF_UNIX)
		}

		if !afInet && cfg.NetRestricted {
			blockedAf = append(blockedAf, unix.AF_INET, unix.AF_INET6)
		}

		// block everything other than AF_UNIX (1), AF_INET (2), AF_INET6 (10), and
		// AF_NETLINK (16)
		if !afOther {
			rules = append(rules, seccompRule{
				Action:   actionEafnosupport,
				Syscall:  unix.SYS_SOCKET,
				Arg:      0,
				Op:       seccomp.CompareGreater,
				OpValue1: 16,
			})

			for i := range 16 {
				switch i {
				case unix.AF_UNIX, unix.AF_INET, unix.AF_INET6, unix.AF_NETLINK:
					continue
				default:
					blockedAf = append(blockedAf, uint64(i))
				}
			}
		}

		for _, af := range blockedAf {
			rules = append(rules, seccompRule{
				Action:   actionEafnosupport,
				Syscall:  unix.SYS_SOCKET,
				Arg:      0,
				Op:       seccomp.CompareEqual,
				OpValue1: af,
			})
		}

		for _, rule := range rules {
			condition, err := seccomp.MakeCondition(rule.Arg, rule.Op, rule.OpValue1, rule.OpValue2)
			if err != nil {
				panic(err)
			}
			err = filter.AddRuleConditional(rule.Syscall, rule.Action, []seccomp.ScmpCondition{condition})
			if err != nil {
				panic(err)
			}
		}

		if cfg.SeccompPrint {
			err := filter.ExportPFC(os.Stdout)
			if err != nil {
				log.Error("Failed to export seccomp filter: %v", err)
				os.Exit(1)
			}
			os.Exit(0)
		} else {
			err := filter.Load()
			if err != nil {
				log.Error("Failed to apply seccomp filter: %v", err)
				os.Exit(1)
			}
			log.Info("Applied seccomp filter")
		}

	}
}

func getSyscall(name string) seccomp.ScmpSyscall {

	syscall, err := seccomp.GetSyscallFromName(name)
	if err != nil {
		log.Error("Failed to get syscall '%s': %v", name, err)
		os.Exit(1)
	}

	return syscall
}
