module icelock

go 1.25

require (
	github.com/landlock-lsm/go-landlock v0.0.0-20260203220723-9199ef6ceb31
	github.com/seccomp/libseccomp-golang v0.11.1
	github.com/urfave/cli/v3 v3.6.2
	golang.org/x/sys v0.40.0
)

require kernel.org/pub/linux/libs/security/libcap/psx v1.2.77 // indirect
