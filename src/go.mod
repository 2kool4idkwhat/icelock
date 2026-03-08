module icelock

go 1.25

require (
	github.com/landlock-lsm/go-landlock v0.7.0
	github.com/seccomp/libseccomp-golang v0.11.1
	github.com/urfave/cli/v3 v3.7.0
	golang.org/x/sys v0.41.0
)

require kernel.org/pub/linux/libs/security/libcap/psx v1.2.77 // indirect
