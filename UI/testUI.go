package main

import (
	"FesterBlitzer/Blitzer"
	"fmt"
	"image/color"
	"math"
	"strconv"
	"net/http"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func getNeedlePos(radius float32, degree float32) rl.Vector2 {
	return rl.Vector2{radius * float32(math.Cos(float64(degree*math.Pi/180.0))), radius * float32(math.Sin(float64(degree*math.Pi/180.0)))}
}

func getRPMDegrees(rpm float32) float32 {
	return float32(rpm/6000*140 + 120)
}

func getThrottleDegrees(throttle float32) float32 {
	return float32(throttle*40 + 70)
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

func drawKMH(speed int) {
	if speed < 10 {
		rl.DrawText(strconv.Itoa(speed), int32(rl.GetScreenWidth()/2)-40, int32(rl.GetScreenHeight()/2)-30, 100, rl.White)
	} else if speed < 100 {
		rl.DrawText(strconv.Itoa(speed), int32(rl.GetScreenWidth()/2)-50, int32(rl.GetScreenHeight()/2)-30, 100, rl.White)
	} else if speed < 220 {
		rl.DrawText(strconv.Itoa(speed), int32(rl.GetScreenWidth()/2)-80, int32(rl.GetScreenHeight()/2)-30, 100, rl.White)
	}
}
func getBlitzer() {
	// Karlsruhe nach Norden
	//lastPos := [2]float64{49.0161, 8.3980}
	//currPos := [2]float64{49.0189, 8.3974}
	//lastPos := [2]float64{49.01880678328532, 8.389688331453078}

	// Hailfingen nach Seebron
	lastPos := [2]float64{48.515966, 8.869765}
	currPos := [2]float64{48.515276, 8.870355}
	
	scanBox := blitzer.GetScanBox(lastPos, currPos)
	// print(scanBox[0][0], scanBox[0][1], scanBox[1][0], scanBox[1][1], scanBox[2][0], scanBox[2][1], scanBox[3][0], scanBox[3][1], "\n")
	boxStart, boxEnd := blitzer.GetBoundingBox(scanBox)

	url := fmt.Sprintf("https://cdn2.atudo.net/api/4.0/pois.php?type=22,26,20,101,102,103,104,105,106,107,108,109,110,111,112,113,115,117,114,ts,0,1,2,3,4,5,6,21,23,24,25,29,vwd,traffic&z=17&box=%f,%f,%f,%f",
		boxStart[0], boxStart[1], boxEnd[0], boxEnd[1])

	resp, err := http.Get(url)
	print(url, "\n")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	response := blitzer.Decode(resp)
	Blitzers := blitzer.GetBlitzer(response, currPos)
	if len(Blitzers) == 0 {
		println("No Blitzers found")
		return
	}
	ClosestBlitzer := blitzer.GetClosestBlitzer(Blitzers)
	println(fmt.Sprintf("%d limit in %s %s in %fkm", ClosestBlitzer.Vmax, ClosestBlitzer.Street, ClosestBlitzer.City, ClosestBlitzer.Distance))
}

func main() {
	rl.InitWindow(int32(rl.GetScreenWidth()), int32(rl.GetScreenHeight()), "Fester Blitzer in 500m!")
	defer rl.CloseWindow()
	rl.SetTargetFPS(60)
	rl.ToggleFullscreen()
	circleCenter := rl.NewVector2(float32(rl.GetScreenWidth()/2), float32(rl.GetScreenHeight()/2))

	RPMStart := 120
	RPMMax := 260

	ThrottleStart := float32(110.0)
	ThrottleMax := float32(70.0)

	circleInnerRadius := 350.0
	circleOuterRadius := 400.5

	rl.DrawFPS(1, 1)

	rpm := float32(100)
	speed := 0
	throttle := 1.0
	ecoStart := 61
	
	blitzer := blitzer.Blitzer{Vmax: 80, City: "Lehm", Street: "xdStraße", Distance: 69.420}
	
	for !rl.WindowShouldClose() {

		if blitzer.Vmax != 0 {
			image := rl.LoadImage(fmt.Sprintf("Assets/%d.png", blitzer.Vmax))
			rl.ImageResize(image, 200, 200)
			texture := rl.LoadTextureFromImage(image)
			rl.DrawTexture(texture, int32(rl.GetScreenWidth())/2-texture.Width/2, int32(rl.GetScreenHeight())/2-texture.Height/2-200, rl.White)
			getBlitzer()
		}
		
		// time.Sleep(1000 * time.Millisecond)
		rpm += 100
		speed += 1

		if rpm > 6000 {
			rpm = 100
		}
		if speed > 150 {
			speed = 0
		}

		RPMEnd := getRPMDegrees(rpm)
		RPMColor := getRPMColor(rpm)

		ecoStart = ecoStart - 5
		if ecoStart <= -80 {
			ecoStart = 61
		}
		throttle = throttle - 0.1
		if throttle < 0 {
			throttle = 1
		}

		ThrottleMax = getThrottleDegrees(float32(throttle))

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

		rl.DrawRing(circleCenter, float32(circleInnerRadius)+1, float32(circleOuterRadius)-1, float32(ThrottleStart), float32(ThrottleMax), int32(0.0), rl.Red)

		rl.DrawRing(circleCenter, float32(circleInnerRadius)+1, float32(circleOuterRadius)-1, float32(60), float32(ecoStart), int32(0.0), rl.Green)
		// 120
		// 260
		// Draw Actual Circle outline
		// rl.DrawRingLines(circleCenter, float32(circleInnerRadius), float32(circleOuterRadius), float32(RPMStart), float32(RPMMax), int32(0.0), rl.Black)

		// Draw Black bars above
		rl.DrawRing(circleCenter, float32(circleInnerRadius)+4.5, float32(circleInnerRadius), float32(RPMStart), float32(RPMMax), int32(0.0), rl.Blue)
		rl.DrawRing(circleCenter, float32(circleOuterRadius)-4.5, float32(circleOuterRadius), float32(RPMStart), float32(RPMMax), int32(0.0), rl.Blue)

		// Draw KM/H
		drawKMH(speed)
		needleStart := getNeedlePos(float32(circleInnerRadius)-15, RPMEnd)
		needleEnd := getNeedlePos(float32(circleOuterRadius)+15, RPMEnd)

		// Draw Needle
		rl.DrawLineEx(rl.Vector2{needleStart.X + float32(rl.GetScreenWidth()/2), needleStart.Y + float32(rl.GetScreenHeight()/2)}, rl.Vector2{needleEnd.X + float32(rl.GetScreenWidth()/2), needleEnd.Y + float32(rl.GetScreenHeight()/2)}, 5, rl.Red)

		// Draw Fake Needle
		// rl.DrawCircleSectorLines(circleCenter, float32(circleOuterRadius), RPMEnd, RPMEnd, int32(0.0), rl.Black)
		rl.EndDrawing()
	}
}
