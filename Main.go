package main

import (
	"flag"
	"image/color"
	"log"
	"math"
	"strconv"
	"time"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/rzetterberg/elmobd"
)

func getNeedlePos(radius float32, degree float32) rl.Vector2 {
	return rl.Vector2{radius * float32(math.Cos(float64(degree*math.Pi/180.0))), radius * float32(math.Sin(float64(degree*math.Pi/180.0)))}
}

func getRPMDegrees(rpm float32) float32 {
	return float32(rpm/6000*140 + 120)
}

func getThrottleDegrees(throttle float32) float32 {
	return float32(throttle*40 + 71)
}

func getRPMColor(rpm float32) rl.Color {
	// TODO: add gradient
	switch rm := int(rpm); {
	case rm < 1500:
		return color.RGBA{R: 144, G: 238, B: 144, A: 255}
	case rm < 3500:
		return color.RGBA{R: 0, G: 128, B: 0, A: 255}
	case rm < 5000:
		return color.RGBA{R: 255, G: 200, B: 87, A: 255}
	case rm < 6000:
		return color.RGBA{R: 200, G: 0, B: 0, A: 255}
	}
	return rl.Gray
}

func DrawKMH(speed int) {
	if speed < 10 {
		rl.DrawText(strconv.Itoa(speed), int32(rl.GetScreenWidth()/2)-40, int32(rl.GetScreenHeight()/2)-30, 100, rl.White)
	} else if speed < 100 {
		rl.DrawText(strconv.Itoa(speed), int32(rl.GetScreenWidth()/2)-50, int32(rl.GetScreenHeight()/2)-30, 100, rl.White)
	} else if speed < 220 {
		rl.DrawText(strconv.Itoa(speed), int32(rl.GetScreenWidth()/2)-80, int32(rl.GetScreenHeight()/2)-30, 100, rl.White)
	}
}

func InitDevice(path string) *elmobd.Device {
	serialPath := flag.String(
		"serial",
		path,
		"Path to the serial device to use",
	)
	flag.Parse()
	device, err := elmobd.NewDevice(*serialPath, false)

	if err != nil {
		log.Fatalf("Check switch? && port")
	}

	return device
}

func getEngineRPM(device elmobd.Device, RPMChannel chan<- float32) {
	for {
		response, err := device.RunOBDCommand(elmobd.NewEngineRPM())
		if err != nil {
			log.Fatal(err)
		}

		MonthDate, err := strconv.ParseFloat(response.ValueAsLit(), 32)

		RPMChannel <- float32(MonthDate)
		time.Sleep(time.Millisecond * 16)
	}
}

func getCarSpeed(device elmobd.Device, SpeedChannel chan<- int) {
	for {
		response, err := device.RunOBDCommand(elmobd.NewVehicleSpeed())
		if err != nil {
			log.Fatal(err)
		}

		MonthDate, err := strconv.ParseInt(response.ValueAsLit(), 0, 32)
		SpeedChannel <- int(MonthDate)
		// 60 FPS = 16ms
		time.Sleep(time.Millisecond * 16)
	}
}

func getThrottlePos(device elmobd.Device, ThrottleChannel chan<- float32) {
	for {
		response, err := device.RunOBDCommand(elmobd.NewThrottlePosition())
		if err != nil {
			log.Fatal(err)
		}
		MonthDate, err := strconv.ParseInt(response.ValueAsLit(), 0, 32)

		ThrottleChannel <- float32(MonthDate)
		time.Sleep(time.Millisecond * 16)
	}
}

func main() {
	// Setting up the window
	rl.InitWindow(int32(rl.GetScreenWidth()), int32(rl.GetScreenHeight()), "Fester Blitzer in 500m!")
	defer rl.CloseWindow()
	rl.SetTargetFPS(60)
	rl.ToggleFullscreen()
	rl.DrawFPS(1, 1)

	// Init OBD2 Reader with path to usb port
	device := InitDevice("/dev/tty.usbserial-11130")

	// RPM goroutine and channel
	RPMChannel := make(chan float32, 2048)
	go getEngineRPM(*device, RPMChannel)
	// RPM Degrees
	RPMStart := 120
	RPMMax := 260
	// RPM Value for UI
	rpm := float32(0)

	// Speed goroutine and channel
	SpeedChannel := make(chan int, 2048)
	go getCarSpeed(*device, SpeedChannel)
	// Speed value for UI
	speed := 0

	// Throttle goroutine and channel
	ThrottleChannel := make(chan float32, 2048)
	go getThrottlePos(*device, ThrottleChannel)
	// Throttle Degrees
	ThrottleStart := 110
	ThrottleMax := float32(70.0)
	// Throttle value for UI
	throttlePos := float32(0)

	// Define Circle Positions
	circleCenter := rl.NewVector2(float32(rl.GetScreenWidth()/2), float32(rl.GetScreenHeight()/2))
	circleInnerRadius := 350.0
	circleOuterRadius := 400.0

	for !rl.WindowShouldClose() {
		// things to add:
		// engine load
		// coolant temp
		// runtime_since_engine_start
		// if gettingData == false, paint error!
		// correct error handling von Durak klauen
		// check if device.RunOBDCommand(elmobd.NewEngineOilTemperature()) is a thing (wäre cool)

		// Get new value form channels
		rpm = <-RPMChannel
		speed = <-SpeedChannel
		throttlePos = <-ThrottleChannel

		RPMEnd := getRPMDegrees(rpm)
		RPMColor := getRPMColor(rpm)
		ThrottleMax = getThrottleDegrees(throttlePos)

		rl.BeginDrawing()
		rl.DrawFPS(10, 10)
		rl.ClearBackground(rl.Black)

		// Draw gray circle background
		rl.DrawRing(circleCenter, float32(circleInnerRadius), float32(circleOuterRadius), 0, 360, int32(0.0), rl.Gray)		
		
		// Draw current RPM
		rl.DrawRing(circleCenter, float32(circleInnerRadius)+1, float32(circleOuterRadius)-1, float32(RPMStart), RPMEnd, int32(0.0), RPMColor)

		// Draw current Throttle
		rl.DrawRing(circleCenter, float32(circleInnerRadius)+1, float32(circleOuterRadius)-1, float32(ThrottleStart), float32(ThrottleMax), int32(0.0), rl.Red)
		
		// Draw Black bars above
		rl.DrawRing(circleCenter, float32(circleInnerRadius)+4.5, float32(circleInnerRadius), float32(RPMStart), float32(RPMMax), int32(0.0), rl.Blue)
		rl.DrawRing(circleCenter, float32(circleOuterRadius)-4.5, float32(circleOuterRadius), float32(RPMStart), float32(RPMMax), int32(0.0), rl.Blue)

		// Draw Speed
		DrawKMH(speed)
		needleStart := getNeedlePos(float32(circleInnerRadius)-15, RPMEnd)
		needleEnd := getNeedlePos(float32(circleOuterRadius)+15, RPMEnd)

		// Draw 1v9 Mathing Needle
		rl.DrawLineEx(rl.Vector2{X: needleStart.X + float32(rl.GetScreenWidth()/2), Y: needleStart.Y + float32(rl.GetScreenHeight()/2)}, rl.Vector2{needleEnd.X + float32(rl.GetScreenWidth()/2), needleEnd.Y + float32(rl.GetScreenHeight()/2)}, 5, rl.Red)

		rl.EndDrawing()
	}
}
