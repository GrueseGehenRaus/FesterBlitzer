package main

import (
	"flag"
	"fmt"

	"github.com/rzetterberg/elmobd"
)

func main() {
	serialPath := flag.String(
		"serial",
		"/dev/tty.usbserial-1130",
		"Path to the serial device to use",
	)

	flag.Parse()

	dev, err := elmobd.NewDevice(*serialPath, false)

	if err != nil {
		fmt.Println("Failed to create new device", err)
		return
	}

	for true {
		rpm, err := dev.RunOBDCommand(elmobd.NewEngineRPM())

		if err != nil {
			fmt.Println("Failed to get rpm", err)
		}

		fmt.Printf("Engine spins at %s RPMs\n", rpm.ValueAsLit())
	}

}
