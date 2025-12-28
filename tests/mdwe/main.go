package main

import (
	"fmt"
	"os"

	"golang.org/x/sys/unix"
)

func main() {
	fmt.Println("Trying to make a WX mapping...")

	_, err := unix.Mmap(-1, 0, 4096, unix.PROT_WRITE|unix.PROT_EXEC, unix.MAP_ANONYMOUS|unix.MAP_PRIVATE)
	if err != nil {
		fmt.Printf("Failed to create a WX mapping: %v", err)
		os.Exit(1)
	}

	fmt.Println("Successfully created a WX mapping")
}
