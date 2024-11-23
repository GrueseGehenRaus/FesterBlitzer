package main

import (
	"flag"
	"image/color"
	"log"
	"math"
	"strconv"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/rzetterberg/elmobd"
)

func getNeedlePos(radius float32, degree float32) rl.Vector2 {
	return rl.Vector2{radius * float32(math.Cos(float64(degree*math.Pi/180.0))), radius * float32(math.Sin(float64(degree*math.Pi/180.0)))}
}

func getRPMDegrees(rpm float32) float32 {
	return float32(rpm/6000*140 + 120)
}

func getRPMColor(rpm float32) rl.Color {
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

func InitDevice() *elmobd.Device {
	serialPath := flag.String(
		"serial",
		"/dev/tty.usbserial-11130",
		"Path to the serial device to use",
	)
	flag.Parse()
	device, err := elmobd.NewDevice(*serialPath, false)

	if err != nil {
		log.Fatalf("Failed to create new device", err)
		return nil
	}
	return device
}

func getEngineRPM(device elmobd.Device, RPMChannel chan <- float32 ) {
	response, err := device.RunOBDCommand(elmobd.NewEngineRPM())
	if err != nil {
		log.Println("Failed to get rpm", err)
	}
	// wenn mehr als 10 errors in a row, dann erstmal pause?
	MonthDate, err := strconv.ParseFloat(response.ValueAsLit(), 32)
	
	select {
    case RPMChannel <- float32(MonthDate):
        log.Println("sent message", MonthDate)
    default:
        log.Println("no")
    }
	return 
}

func main() {
	rl.InitWindow(int32(rl.GetScreenWidth()), int32(rl.GetScreenHeight()), "Fester Blitzer in 500m!")
	defer rl.CloseWindow()
	rl.SetTargetFPS(60)
	rl.ToggleFullscreen()
	circleCenter := rl.NewVector2(float32(rl.GetScreenWidth()/2), float32(rl.GetScreenHeight()/2))

	// Create channel
	RPMChannel := make(chan float32)

	
	RPMStart := 120
	RPMMax := 260

	circleInnerRadius := 350.0
	circleOuterRadius := 400.5

	rl.DrawFPS(1, 1)

	speed := 0

	device := InitDevice()

	for !rl.WindowShouldClose() {
		go getEngineRPM(*device, RPMChannel)
		
		rpm := float32(0)
		
		select {
    	case lastKnownRPM := <-RPMChannel:
    	log.Println("rpm updated",lastKnownRPM)
    	rpm = lastKnownRPM
        default:
        	log.Println("no RPM update")
    	}
		
		speed += 1

		if speed > 150 {
			speed = 0
		}

		RPMEnd := getRPMDegrees(rpm)
		RPMColor := getRPMColor(rpm)

		rl.BeginDrawing()

		rl.DrawFPS(10, 10)

		rl.ClearBackground(rl.Black)

		// Speedometer
		// Draw Fake Circle outline
		// rl.DrawRing(circleCenter, float32(circleInnerRadius)-5, float32(circleOuterRadius)+5, float32(RPMStart), float32(RPMMax), int32(0.0), rl.Black)
		// rl.DrawRing(circleCenter, float32(circleInnerRadius), float32(circleOuterRadius), float32(RPMStart), float32(RPMMax), int32(0.0), rl.White)

		// Draw current RPM
		rl.DrawRing(circleCenter, float32(circleInnerRadius), float32(circleOuterRadius), 0, 360, int32(0.0), rl.Gray)
		rl.DrawRing(circleCenter, float32(circleInnerRadius)+1, float32(circleOuterRadius)-1, float32(RPMStart), RPMEnd, int32(0.0), RPMColor)

		// 120
		// 260
		// Draw Actual Circle outline
		// rl.DrawRingLines(circleCenter, float32(circleInnerRadius), float32(circleOuterRadius), float32(RPMStart), float32(RPMMax), int32(0.0), rl.Black)

		// Draw Black bars above
		rl.DrawRing(circleCenter, float32(circleInnerRadius)+4.5, float32(circleInnerRadius), float32(RPMStart), float32(RPMMax), int32(0.0), rl.Blue)
		rl.DrawRing(circleCenter, float32(circleOuterRadius)-4.5, float32(circleOuterRadius), float32(RPMStart), float32(RPMMax), int32(0.0), rl.Blue)

		// Draw KM/H
		DrawKMH(speed)
		needleStart := getNeedlePos(float32(circleInnerRadius)-15, RPMEnd)
		needleEnd := getNeedlePos(float32(circleOuterRadius)+15, RPMEnd)

		// Draw Needle
		rl.DrawLineEx(rl.Vector2{needleStart.X + float32(rl.GetScreenWidth()/2), needleStart.Y + float32(rl.GetScreenHeight()/2)}, rl.Vector2{needleEnd.X + float32(rl.GetScreenWidth()/2), needleEnd.Y + float32(rl.GetScreenHeight()/2)}, 5, rl.Red)

		// Draw Fake Needle
		// rl.DrawCircleSectorLines(circleCenter, float32(circleOuterRadius), RPMEnd, RPMEnd, int32(0.0), rl.Black)
		rl.EndDrawing()
	}
}
