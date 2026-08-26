[Landlock]: https://landlock.io/
[Landrun]: https://github.com/Zouuup/landrun
[Island]: https://github.com/landlock-lsm/island

# icelock 🧊🔒

Icelock is a small Linux CLI tool for restricting programs with [Landlock] (and seccomp). You can use icelock to run programs with reduced privileges

Run `icelock --help` for a list of options

## Current limitations (non-exhaustive)

- execute permission only covers direct file execution, so [it can be bypassed](https://github.com/landlock-lsm/linux/issues/37)

- if filesystem access is restricted, the sandboxed processes can't modify the filesystem topology of their mount namespace
  - this breaks bubblewrap

- icelock doesn't stop the sandboxed processes from using too much resources (memory, CPU time, etc), so it won't protect you from eg. a fork bomb

- [port restrictions don't apply to Multipath-TCP](https://github.com/landlock-lsm/linux/issues/54)

- reading file metadata (`stat(2)`) isn't restricted

- file locking (`flock(2)`) isn't restricted

- changing file access/modify times (`utime(2)`) isn't restricted

## Notes

The final allowed filesystem access is the sum of all rules, so if you run `icelock --rw=/aaa --rx=/aaa/bbb/ccc` then the sandbox will have write access to `/aaa/bbb/ccc` because that path is below `/aaa`

Under the hood filesystem rules are created using file descriptors (not path strings), so to allow access to a path that might not exist you have to either:

1. create it before running icelock (ie. `mkdir -p` or `touch`)

2. allow access to the dir above it instead (but that will obviously weaken the sandbox)

### Kernel compatibility

Icelock currently targets Landlock v10 ABI

You can use the `--best-effort` flag to run icelock on older kernels, but this will weaken the sandbox

Landlock ABI version | Minimum upstream kernel version | New features
---------------------|---------------------------------|-------------
10 | 7.2  | UDP port restrictions
9  | 7.1  | pathname unix socket restrictions (**prevents sandbox escape via D-bus**)
7  | 6.15 | audit logging
6  | 6.12 | signal scoping, abstract unix socket scoping

## Related projects

### Landrun

[Landrun] was the initial inspiration for icelock, and what got me interested in Landlock in the first place. That being said, there are some major differences. As of landrun version 0.1.17:

- landrun only passes the env vars that you explicitly specify, which makes it very annoying to use

- icelock uses seccomp (in addition to Landlock)

- icelock drops capabilities by default

- landrun has flags for automatically adding the app executable/libraries to RX paths

- various flags are subtly different, for example:
  - in icelock, the read/execute flag is `--rx`. In landrun, it's `--rox`

  - icelock doesn't have a `--rwx` flag

  - in icelock, the `--unix` flag only grants the `RESOLVE_UNIX` access right. In landrun, `--unix` also grants `READ_FILE` and `READ_DIR`

### Island

[Island] is the official Landlock sandboxing tool. Island is designed around workspaces, and as such is very different from icelock
