module icelock

go 1.26.0

require (
	github.com/landlock-lsm/go-landlock v0.11.0
	github.com/seccomp/libseccomp-golang v0.12.0
	github.com/urfave/cli/v3 v3.14.0
	golang.org/x/sys v0.49.0
	kernel.org/pub/linux/libs/security/libcap/cap v1.2.78
)

require kernel.org/pub/linux/libs/security/libcap/psx v1.2.78 // indirect
