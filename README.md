[Landlock]: https://landlock.io/
[Landrun]: https://github.com/Zouuup/landrun

# icelock 🧊🔒

Icelock is a small CLI tool for restricting programs with [Landlock] (and seccomp). You can use icelock to run programs with reduced privileges

Icelock currently requires Landlock v6 ABI to be supported by the kernel (v5 if you disable IPC scoping)

## Compiling

Just run `nix build`

You can also run `go build -v` in the `src/` dir, but then you'll need to ensure that libseccomp and pkg-config are installed

## Usage

`icelock [options] -- <command> [command options]`

By default, everything that icelock can restrict is denied and needs to be explicitly allowed

## Limitations

TODO

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
