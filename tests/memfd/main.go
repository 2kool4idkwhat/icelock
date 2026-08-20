package main

import (
	"flag"
	"log"

	"golang.org/x/sys/unix"
)

func main() {
	memfdCreateFlags := unix.MFD_CLOEXEC

	exec := flag.Bool("exec", false, "adds MFD_EXEC")
	noexecSeal := flag.Bool("noexec-seal", false, "adds MFD_NOEXEC_SEAL")
	flag.Parse()

	if *exec {
		memfdCreateFlags |= unix.MFD_EXEC
	}
	if *noexecSeal {
		memfdCreateFlags |= unix.MFD_NOEXEC_SEAL
	}

	_, err := unix.MemfdCreate("test", memfdCreateFlags)
	if err != nil {
		log.Fatalf("Failed to create memfd: %v", err)
	}
	log.Print("Successfully created memfd")
}
