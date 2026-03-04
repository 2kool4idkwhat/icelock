package main

import (
	"flag"
	"fmt"
	"os"

	"golang.org/x/sys/unix"
)

func main() {
	var sock int
	m := map[string]int{
		"unix":  unix.AF_UNIX,
		"inet":  unix.AF_INET,
		"inet6": unix.AF_INET6,
		"vsock": unix.AF_VSOCK,
	}

	sockType := flag.String("af", "", "socket type to try to create")
	flag.Parse()

	if *sockType == "" {
		fmt.Println("Socket type not specified")
		os.Exit(1)
	}
	sock = m[*sockType]

	_, err := unix.Socket(sock, unix.SOCK_STREAM|unix.SOCK_CLOEXEC, 0)
	if err != nil {
		fmt.Printf("Failed to create %s socket: %v\n", *sockType, err)
		os.Exit(1)
	}
	fmt.Printf("Successfully created %s socket\n", *sockType)
}
