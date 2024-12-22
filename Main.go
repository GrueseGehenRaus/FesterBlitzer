package main

import (
	"flag"
	"fmt"
	"image/color"
	"log"
	"math"
	"os"
	"strconv"
	"time"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/rzetterberg/elmobd"
)

type Eco struct {
	rpm      float32
	throttle float32
}

func getNeedlePos(radius float32, degree float32) rl.Vector2 {
	return rl.Vector2{radius * float32(math.Cos(float64(degree*math.Pi/180.0))), radius * float32(math.Sin(float64(degree*math.Pi/180.0)))}
}

func getRPMDegrees(rpm float32) float32 {
	if rpm > 6000 {
		return 260
	}
	return float32(rpm/6000*140 + 120)
}

func getThrottleDegrees(throttle float32) float32 {
	fmt.Print(strconv.FormatFloat(float64(throttle*40) + 71, 'f', 6, 32))
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
		log.Fatalf("Check switch and port")
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

func getCarSpeed(device elmobd.Device, SpeedChannel chan<- int, file *os.File) {
	log.SetOutput(file)
	log.Print("lehm")
	for {
		response, err := device.RunOBDCommand(elmobd.NewVehicleSpeed())
		if err != nil {
			log.Print(err)
		}

		MonthDate, err := strconv.ParseFloat(response.ValueAsLit(), 32)
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
		
		//log.Print(response.ValueAsLit())
		//log.Print(elmobd.NewThrottlePosition().Value)
		
		MonthDate, err := strconv.ParseFloat(response.ValueAsLit(), 32)

		// log.Print("Throttle%: ", MonthDate)
		
		ThrottleChannel <- float32(MonthDate)
		time.Sleep(time.Millisecond * 16)
	}
}

func getRuntimeSinceEngineStart(device elmobd.Device, RuntimeChannel chan<- float32) {
	for {
		response, err := device.RunOBDCommand(elmobd.NewRuntimeSinceStart())
		if err != nil {
			log.Fatal(err)
		}
		MonthDate, err := strconv.ParseFloat(response.ValueAsLit(), 32)
		
		RuntimeChannel <- float32(MonthDate)
		time.Sleep(time.Second * 15)
	}
}

func getEcoScore(EcoChannel chan Eco, RPMChannel <-chan float32, ThrottleChannel <-chan float32, RuntimeChannel <-chan float32) {
	rpm := float32(0)
	throttle := float32(0)
	count := float32(0)
	for {
		for count < 60 {
			// man muss schauen wie sehr sich das verschiebt. ggf 0.96 * time.Second um sicher die 60s zu erreichen
			for range time.Tick(time.Second) {
				rpm += evalRPM(<-RPMChannel)
				throttle += evalThrottle(<-ThrottleChannel)
				count++
			}
		}
		currentEco := <-EcoChannel
		runtime := <-RuntimeChannel
		minutes := float32(int32(runtime / 60))
		rpm = rpm/count*(1/minutes) + currentEco.rpm*(1-(1/minutes))
		throttle = throttle/count*(1/minutes) + currentEco.throttle*(1-(1/minutes))
		EcoChannel <- Eco{rpm, throttle}
		rpm = 0
		throttle = 0
		count = 0
	}
}

func evalRPM(rpm float32) float32 {
	if rpm <= 2500 {
		return 100.0
	}
	return 100 - ((rpm - 2500) / (6000 - 2500) * 100)
}

func evalThrottle(throttle float32) float32 {
	if throttle <= 50 {
		return 100.0
	}
	return 100 - ((throttle - 50) / 50 * 100)
}

func main() {
	file, err := os.OpenFile("log.txt", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal(err)
	}
	// Setting up the window
	rl.InitWindow(int32(rl.GetScreenWidth()), int32(rl.GetScreenHeight()), "Fester Blitzer in 500m!")
	defer rl.CloseWindow()
	rl.SetTargetFPS(60)
	rl.ToggleFullscreen()
	rl.DrawFPS(1, 1)

	// Init OBD2 Reader with path to usb port
	//device := InitDevice("/dev/tty.usbserial-11310")
	device := InitDevice("test://")

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
	go getCarSpeed(*device, SpeedChannel, file)
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

	// Runtime goroutine and channel
	RuntimeChannel := make(chan float32, 2048)
	// go getRuntimeSinceEngineStart(*device, RuntimeChannel)

	EcoChannel := make(chan Eco, 2048)
	go getEcoScore(EcoChannel, RPMChannel, ThrottleChannel, RuntimeChannel)

	// Define Circle Positions
	circleCenter := rl.NewVector2(float32(rl.GetScreenWidth()/2), float32(rl.GetScreenHeight()/2))
	circleInnerRadius := 350.0
	circleOuterRadius := 400.0

	for !rl.WindowShouldClose() {
		// things to add:
		// engine load
		// coolant temp
		// runtime_since_engine_start
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
		
		// log.Print(ThrottleMax)
		
		// Draw Speed
		DrawKMH(speed)

		// Draw gray circle background
		rl.DrawRing(circleCenter, float32(circleInnerRadius), float32(circleOuterRadius), 0, 360, int32(0.0), rl.Gray)

		// Draw current RPM
		rl.DrawRing(circleCenter, float32(circleInnerRadius)+1, float32(circleOuterRadius)-1, float32(RPMStart), RPMEnd, int32(0.0), RPMColor)

		// Draw current Throttle
		rl.DrawRing(circleCenter, float32(circleInnerRadius)+1, float32(circleOuterRadius)-1, float32(ThrottleStart), float32(ThrottleMax), int32(0.0), rl.Red)

		// Draw Black bars above
		rl.DrawRing(circleCenter, float32(circleInnerRadius)+4.5, float32(circleInnerRadius), float32(RPMStart), float32(RPMMax), int32(0.0), rl.Blue)
		rl.DrawRing(circleCenter, float32(circleOuterRadius)-4.5, float32(circleOuterRadius), float32(RPMStart), float32(RPMMax), int32(0.0), rl.Blue)

		
		needleStart := getNeedlePos(float32(circleInnerRadius)-15, RPMEnd)
		needleEnd := getNeedlePos(float32(circleOuterRadius)+15, RPMEnd)
		
		// eco := <-EcoChannel
		// rl.DrawText(strconv.Itoa(int(eco.rpm)), int32(rl.GetScreenWidth()/2)-40, int32(rl.GetScreenHeight()/2)+300, 100, rl.White)
		
		// Draw 1v9 Mathing Needle
		rl.DrawLineEx(rl.Vector2{X: needleStart.X + float32(rl.GetScreenWidth()/2), Y: needleStart.Y + float32(rl.GetScreenHeight()/2)}, rl.Vector2{needleEnd.X + float32(rl.GetScreenWidth()/2), needleEnd.Y + float32(rl.GetScreenHeight()/2)}, 5, rl.Red)

		rl.EndDrawing()
	}
}
