package main

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

	supported, err := dev.CheckSupportedCommands()

	if err != nil {
		fmt.Println("Failed to check supported commands", err)
		return
	}

	allCommands := elmobd.GetSensorCommands()
	carCommands := supported.FilterSupported(allCommands)

	fmt.Printf("%d of %d commands supported:\n", len(carCommands), len(allCommands))

	for _, cmd := range carCommands {
		fmt.Printf("%s supported!\n", cmd.Key())

		value, err := dev.RunOBDCommand(cmd)

		if err != nil {
			fmt.Println("Failed to get value", err)
		}

		fmt.Printf("%s outputs: %s\n", cmd.Key(), value.ValueAsLit())
	}

}
