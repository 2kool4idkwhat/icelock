# Using icelock

`icelock [options] -- <command> [command options]`

For flags that can take multiple args, you can separate the args with a comma (`,`). Eg. `--rx /usr --rx /bin` is the same as `--rx /usr,/bin`

By default, everything that icelock can restrict is denied and needs to be explicitly allowed

## Filesystem

Use `--ro` to allow read access beneath a path, `--rw` for read/write access, and `--rx` for read/execute access

The final allowed FS access is the sum of all rules, so if you run `icelock --rw=/aaa --rx=/aaa/bbb/ccc` then the app will have write access to `/aaa/bbb/ccc` because that path is below `/aaa`

Under the hood FS rules apply to file descriptors (not path strings), so to allow access to a path that might not exist you have to either 1. create it before running icelock (ie. `mkdir -p` or `touch`), or 2. allow access to the dir above it (but that will obviously worsen the sandbox security)

Since Landlock currently can't restrict chmod, chown, and writing extended attributes, these are blocked with seccomp. To allow them use `--syscalls=chmod,chown,xattr`

If you don't want to restrict FS access use `--unrestricted-fs` (this also allows chmod/chown/xattr syscalls). WARNING: This allows escaping the sandbox by writing to one of the many dangerous files like `~/.bashrc`

`--unrestricted-fs` is needed to run apps that use mount namespaces for their own sandboxing, such as bubblewrap

## Network

If you don't want to restrict network access use `--unrestricted-net` (this disables Landlock TCP restrictions and allows AF_INET/AF_INET6 sockets)

For limited network access use `--af inet` to allow AF_INET/AF_INET6 sockets and `--bind-tcp`/`--connect-tcp` to allow binding/connecting to TCP ports. Landlock currently can't restrict binding/connecting to UDP ports

If you need obscure socket families use `--af other`

## IPC

### Signals

To allow the app to send signals to processes outside the sandbox use `--unscoped-ipc`

### Unix sockets

To allow unix sockets use `--af unix`. WARNING: This allows escaping the sandbox via D-bus, since Landlock currently can't restrict pathname unix sockets

If the app needs to connect to abstract unix sockets created outside the sandbox also use `--unscoped-ipc`

## Seccomp

In addition to syscall groups mentioned in previous sections, keyring syscalls and some privileged syscalls are also blocked. You can allow them with `--syscalls=keyring,privileged`

TIOCSTI and TIOCLINUX are also blocked since there's no legitimate reason to use them and they've been the source of many vulnerabilities

By default, if seccomp blocks something then the syscall will return `EPERM` or `EAFNOSUPPORT`. If you use `--seccomp-kill`, the app process will instead be terminated

You can disable seccomp with `--no-seccomp`. WARNING: This allows escaping the sandbox via D-bus since unix socket restrictions are currently implemented with seccomp

## Debugging

You can use `--log-level` or the `ICELOCK_LOG_LEVEL` env var to set the log level

Currently icelock doesn't set `LANDLOCK_RESTRICT_SELF_LOG_NEW_EXEC_ON`, so permission denials won't be logged in the audit subsystem (and you'd need a kernel with Landlock v7 ABI for that anyway). You can use `strace --status=failed` though

While this is mainly useful for developing icelock, you can use `--seccomp-print` to view a human-readable version of the filter
