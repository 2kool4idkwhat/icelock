[Landlock]: https://landlock.io/

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
