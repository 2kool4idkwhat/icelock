module icelock

go 1.25

require (
	github.com/landlock-lsm/go-landlock v0.0.0-20241109072704-b981810c71ce
	github.com/seccomp/libseccomp-golang v0.11.1
	github.com/urfave/cli/v3 v3.6.1
	golang.org/x/sys v0.39.0
)

require kernel.org/pub/linux/libs/security/libcap/psx v1.2.70 // indirect
