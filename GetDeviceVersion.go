package main

// dmesg | tail -> port für serial device herausfinden

import (
	"flag"
	"fmt"
	"github.com/rzetterberg/elmobd"
)

func main() {
	serialPath := flag.String(
		"serial",
		"/dev/tty.usbserial-11130",
		"Path to the serial device to use",
	)

	flag.Parse()

	dev, err := elmobd.NewDevice(*serialPath, false)

	if err != nil {
		fmt.Println("Failed to create new device", err)
		return
	}

	version, err := dev.GetVersion()

	if err != nil {
		fmt.Println("Failed to get version", err)
		return
	}

	fmt.Println("Device has version", version)
}