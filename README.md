[Landlock]: https://landlock.io/
[Landrun]: https://github.com/Zouuup/landrun
[Island]: https://github.com/landlock-lsm/island

# icelock 🧊🔒

Icelock is a small CLI tool for restricting programs with [Landlock] (and seccomp). You can use icelock to run programs with reduced privileges

Run `icelock --help` for a list of options, and see [USAGE.md](./USAGE.md) for details

## Compiling

Just run `nix build`

You can also run `go build -v` in the `src/` dir, but then you'll need to ensure that libseccomp and pkg-config are installed

## Current limitations (non-exhaustive)

- if unix sockets are allowed (`--af unix`) the sandbox can be escaped via D-bus (and potentially any other service that has a pathname unix socket, such as Docker daemon)
  - this will be addressed in the future using Landlock v9 ABI's pathname unix socket restrictions (not yet in a released kernel version)

- execute permission only covers direct file execution, so [it can be bypassed](https://github.com/landlock-lsm/linux/issues/37)

- if filesystem access is restricted the app can't modify filesystem topology, which breaks bubblewrap and other sandboxing solutions that use mount namespaces

- icelock doesn't stop the app from using too much resources (memory, CPU time, etc), so it won't protect you from eg. a fork bomb

- [landlock TCP port restrictions don't apply to Multipath-TCP](https://github.com/landlock-lsm/linux/issues/54)

- reading file metadata (`stat(2)`) isn't restricted

- file locking (`flock(2)`) isn't restricted

- changing file access/modify times (`utime(2)`) isn't restricted

## Related projects

### Landrun

[Landrun] was the initial inspiration for icelock, and what got me interested in Landlock in the first place. That being said, there are some major differences. As of landrun version 0.1.15:

- landrun only passes the env vars that you explicitly specify, which makes it very annoying to use

- icelock uses seccomp to block some dangerous things that Landlock can't restrict yet. Namely unix sockets as they allow escaping the sandbox via D-bus

- icelock has support for signal/abstract unix socket scoping

- landrun has flags for automatically adding the app executable/libraries to RX paths

- landrun has a best-effort mode

- icelock doesn't have a `--rwx` flag because you very rarely want to have a path that is both writable and executable, and if you do then you can just combine the `--rx` and `--rw` flags

- in icelock the RX paths flag is called `--rx`, in landrun it's `--rox`

### Island

[Island] is the official Landlock sandboxing tool. Island is designed around workspaces, and as such is very different from icelock
