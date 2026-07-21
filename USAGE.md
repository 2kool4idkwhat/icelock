# Using icelock

`icelock [options] -- <command> [command options]`

For flags that can take multiple args, you can separate the args with a comma (`,`). Eg. `--rx /usr --rx /bin` is the same as `--rx /usr,/bin`

By default everything that icelock can restrict is denied and needs to be explicitly allowed, except for MDWE

## Filesystem

Use `--ro` to allow read access beneath a path, `--rw` for read/write access, `--rx` for read/execute access, and `--unix` for pathname unix socket access

The final allowed FS access is the sum of all rules, so if you run `icelock --rw=/aaa --rx=/aaa/bbb/ccc` then the app will have write access to `/aaa/bbb/ccc` because that path is below `/aaa`

Under the hood FS rules apply to file descriptors (not path strings), so to allow access to a path that might not exist you have to either 1. create it before running icelock (ie. `mkdir -p` or `touch`), or 2. allow access to the dir above it (but that will obviously worsen the sandbox security)

Since Landlock currently can't restrict chmod, chown, and writing extended attributes, these are blocked with seccomp. To allow them use `--syscalls=chmod,chown,xattr`

If you don't want to restrict FS access use `--unrestricted-fs` (this also allows chmod/chown/xattr syscalls). WARNING: This allows escaping the sandbox

`--unrestricted-fs` is needed to run apps that use mount namespaces for their own sandboxing, such as bubblewrap

## Network

If you don't want to restrict network access use `--unrestricted-net` (this disables Landlock TCP restrictions and allows AF_INET/AF_INET6 sockets)

For limited network access use `--af inet` to allow AF_INET/AF_INET6 sockets and `--bind`/`--bind-all`/`--connect`/`--connect-all` to allow binding/connecting to TCP ports. icelock currently can't restrict binding/connecting to UDP ports

If you need obscure socket families use `--af other`

## IPC

### Signals

To allow the app to send signals to processes outside the sandbox use `--signals`

### Abstract unix sockets

To allow connecting to abstract unix sockets created outside the sandbox use `--abstract-unix`

### Message queues

To allow POSIX and System V message queues use `--syscalls=mq`

## Capabilities

By default, icelock drops all capabilities (if it has any). To keep them use `--keep-caps`. WARNING: this is dangerous since capabilities are inherently dangerous, and some of them may allow sandbox escape

## User namespaces

To allow the app to create user namespaces use `--userns`

NOTE: as mentioned in the filesystem section, if filesystem access is restricted then the app effectively won't be able to use mount namespaces

## io_uring

To allow the app to use io_uring use `--io-uring`. WARNING: this may allow bypassing some restrictions

## Seccomp

In addition to syscall groups mentioned in previous sections, keyring, debugging, emulation, and some privileged syscalls are also blocked. You can allow them with `--syscalls=keyring,debug,privileged,emulation`

TIOCSTI and TIOCLINUX are also blocked since there's no legitimate reason to use them and they've been the source of many vulnerabilities

You can disable seccomp with `--no-seccomp`

## MDWE (Memory-Deny-Write-Execute)

You can use `--mdwe` to set the `PR_MDWE_REFUSE_EXEC_GAIN` prctl flag to block memory mappings that are both writable and executable

This is the only thing that's not restricted by default because it doesn't provide any isolation, it's only a hardening measure and is likely to cause breakage (mainly with JITs)

## Debugging

You can use `--log-level` or the `ICELOCK_LOG_LEVEL` env var to set the log level

While this is mainly useful for developing icelock, you can use `--seccomp-print` to view a human-readable version of the filter

### Audit subsystem logging

You can use `--audit` or the `ICELOCK_AUDIT` env var to log permission denials in the Audit subsystem

You can also use `--no-audit-subdomains` to turn off logging for future nested landlock domains that the app creates

See ["Landlock: system-wide management"](https://www.kernel.org/doc/html/latest/admin-guide/LSM/landlock.html) kernel docs for more info
