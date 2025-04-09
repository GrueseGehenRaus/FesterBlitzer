package main

import (
	Blitzer "FesterBlitzer/Blitzer"
	"flag"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gen2brain/raylib-go/easings"
	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/rzetterberg/elmobd"
)

type Car struct {
	rpm   int32
	speed int32
}

func drawTriangle(centerX float32, centerY float32, width float32, height float32, color rl.Color) {
	top := rl.Vector2{X: centerX, Y: centerY - height/2}
	bottomLeft := rl.Vector2{X: centerX - width/2, Y: centerY + height/2}
	bottomRight := rl.Vector2{X: centerX + width/2, Y: centerY + height/2}

	rl.DrawTriangle(top, bottomLeft, bottomRight, color)
}

func drawTrapezoid(centerX, centerY, topWidth, bottomWidth, height float32, color rl.Color) {
	rectWidth := topWidth
	rectHeight := height
	rectX := centerX - topWidth/2
	rectY := centerY - height/2

	rl.DrawRectangleRec(rl.Rectangle{X: rectX, Y: rectY, Width: rectWidth, Height: rectHeight}, color)

	widthDiff := (bottomWidth - topWidth) / 2

	leftTriTop := rl.Vector2{X: rectX, Y: rectY}
	leftTriBottom := rl.Vector2{X: rectX, Y: rectY + rectHeight}
	leftTriOutside := rl.Vector2{X: rectX - widthDiff, Y: rectY + rectHeight} // Ensure bottom alignment

	rl.DrawTriangle(leftTriOutside, leftTriBottom, leftTriTop, color)

	rightTriTop := rl.Vector2{X: rectX + rectWidth, Y: rectY}
	rightTriBottom := rl.Vector2{X: rectX + rectWidth, Y: rectY + rectHeight}
	rightTriOutside := rl.Vector2{X: rectX + rectWidth + widthDiff, Y: rectY + rectHeight}

	rl.DrawTriangle(rightTriTop, rightTriBottom, rightTriOutside, color)
}

func drawrpm(rpm float32) {
	const (
		meterWidth  = float32(80)
		meterHeight = float32(300)
		maxRPM      = float32(5500)
		majorStep   = float32(1000) // 1k RPM per step
		minorStep   = float32(500)  // half-step ticks
		fontSize    = 20
	)

	meterX := float32(rl.GetScreenWidth()/2) - 240
	meterY := float32(rl.GetScreenHeight()/2) - meterHeight/2
	labelX := meterX - 40

	// Draw RPM meter outline
	rl.DrawRectangleRoundedLinesEx(rl.Rectangle{
		X:      meterX,
		Y:      meterY,
		Width:  meterWidth,
		Height: meterHeight,
	}, 0.5, 0, 2.0, rl.Fade(rl.Blue, 0.4))

	// Draw major (numbered) indicators: "1", "2", ..., "6"
	for val := float32(1000); val <= maxRPM; val += majorStep {
		posY := meterY + meterHeight - (val / maxRPM * meterHeight)
		label := fmt.Sprintf("%.0f", val/1000) // convert 1000 -> "1", 2000 -> "2", ...
		rl.DrawText(label, int32(labelX)+13, int32(posY)-fontSize/2, fontSize, rl.White)
		rl.DrawLine(int32(meterX-10), int32(posY), int32(meterX), int32(posY), rl.Gray)
	}

	// Draw minor (unlabeled) indicators every 500 RPM
	for val := float32(500); val < maxRPM; val += minorStep {
		if int(val)%int(majorStep) == 0 {
			continue // skip if it's a major tick
		}
		posY := meterY + meterHeight - (val / maxRPM * meterHeight)
		rl.DrawLine(int32(meterX-5), int32(posY), int32(meterX), int32(posY), rl.DarkGray)
	}

	if rpm > 0 {
		// Compute percentage fill
		percent := float32(rpm) / 6000
		fillHeight := meterHeight * percent
		if fillHeight > meterHeight {
			fillHeight = meterHeight
		}

		roundedHeight := float32(70) // Rounded part at the bottom

		// Ensure minimum fill height to cover the rounded part
		if fillHeight < roundedHeight {
			fillHeight = roundedHeight
		}

		// Draw the solid rectangular fill (flat top, increased overlap)
		rl.DrawRectangle(int32(meterX), int32(meterY+meterHeight-fillHeight),
			int32(meterWidth), int32(fillHeight-roundedHeight+20), rl.Fade(rl.Maroon, 1)) // Increased overlap

		// Draw the rounded bottom part
		rl.DrawRectangleRounded(rl.Rectangle{
			X:      meterX,
			Y:      meterY + meterHeight - roundedHeight, // Always at the bottom
			Width:  meterWidth,
			Height: roundedHeight,
		}, 0.5, 0, rl.Fade(rl.Maroon, 1))
	}
}

func drawSpeed(speed int32, font rl.Font) {
	//70
	if speed < 10 {
		rl.DrawTextEx(font, strconv.FormatInt(int64(speed), 10), rl.Vector2{X: float32(rl.GetScreenWidth()/2) - 35, Y: float32(rl.GetScreenHeight()/2) - 50}, 125, 0, rl.White)
	} else if speed < 100 {
		rl.DrawTextEx(font, strconv.FormatInt(int64(speed), 10), rl.Vector2{X: float32(rl.GetScreenWidth()/2) - 70, Y: float32(rl.GetScreenHeight()/2) - 50}, 125, 0, rl.White)
	} else {
		rl.DrawTextEx(font, strconv.FormatInt(int64(speed), 10), rl.Vector2{X: float32(rl.GetScreenWidth()/2) - 105, Y: float32(rl.GetScreenHeight()/2) - 50}, 125, 0, rl.White)
	}
	rl.DrawTextEx(font, "km/h", rl.Vector2{X: float32(rl.GetScreenWidth()/2) - 55, Y: float32(rl.GetScreenHeight()/2) + 50}, 50, 0, rl.White)
}

func drawBlitzer(distance float64, vmax int32, speedTexture rl.Texture2D, infinityTexture rl.Texture2D, carSpeed int32, font rl.Font) {
	centerY := 400.0
	topWidth := 175.0
	bottomWidth := 200.0
	height := 40.0
	fillCount := float64(-1)

	// if carSpeed >= 15 {
	if carSpeed >= 0 {
		if vmax == 0 {
			fillCount = -1
			rl.DrawTexture(infinityTexture, 551, 80, rl.White)
		} else if vmax == -1 {
			fillCount = 6
			rl.DrawTexture(speedTexture, 551, 80, rl.White)
			rl.DrawTextEx(font, "0", rl.Vector2{X: 590, Y: 106}, 40, 0, rl.Black)
		} else {
			fillCount = ((1 - distance) * 5)

			// With Distance
			// rl.DrawTextEx(font, strconv.FormatFloat(distance*1000, 'f', 0, 64), rl.Vector2{X: 584, Y: 173}, 30, 0, rl.White)
			// rl.DrawTexture(speedTexture, 551, 65, rl.White)
			// rl.DrawTextEx(font, strconv.FormatInt(int64(vmax), 10), rl.Vector2{X: 581, Y: 98}, 50, 0, rl.Black)

			// Without Distance
			rl.DrawTexture(speedTexture, 551, 80, rl.White)
			rl.DrawTextEx(font, strconv.FormatInt(int64(vmax), 10), rl.Vector2{X: 575, Y: 106}, 50, 0, rl.Black)

		}

		for i := float64(0); i < 5; i++ {
			if i <= math.Round(fillCount) {
				drawTrapezoid(600, float32(centerY), float32(topWidth), float32(bottomWidth), float32(height), rl.Green)
			} else {
				drawTrapezoid(600, float32(centerY), float32(topWidth), float32(bottomWidth), float32(height), rl.Gray)
			}
			centerY = centerY * 0.85
			topWidth = topWidth * 0.85
			bottomWidth = bottomWidth * 0.85
			height = height * 0.85
		}
	}
}

func getBlitzer(BlitzerChannel chan<- Blitzer.Blitzer) {
	count := 0
	for {
		// getPos() braucht man halt und noch LastPos speichern vor schreiben vom Blitzer in den channel
		// Blitzer types noch casen (6 ist z.B. Abstandsmessung)

		// HEK nach Norden
		// lastPos := [2]float64{49.0161, 8.3980}
		// currPos := [2]float64{49.0189, 8.3974}

		positions := [][2]float64{
			{48.521266, 8.868477},
			{48.518950, 8.869078},
			{48.517287, 8.868842},
			{48.515966, 8.869765},
			{48.515276, 8.870355},
			{48.512568, 8.871718},
			{48.510862, 8.871846},
			{48.509369, 8.872361},
			{48.508445, 8.873134},
		}

		// Hailfingen nach Seebron
		lastPos := positions[count]
		currPos := positions[count+1]

		scanBox := Blitzer.GetScanBox(lastPos, currPos)
		boxStart, boxEnd := Blitzer.GetBoundingBox(scanBox)

		url := fmt.Sprintf("https://cdn2.atudo.net/api/4.0/pois.php?type=22,26,20,101,102,103,104,105,106,107,108,109,110,111,112,113,115,117,114,ts,0,1,2,3,4,5,6,21,23,24,25,29,vwd,traffic&z=17&box=%f,%f,%f,%f",
			boxStart[0], boxStart[1], boxEnd[0], boxEnd[1])

		client := http.Client{
			Timeout: time.Second * 15,
		}

		resp, err := client.Get(url)
		if err != nil {
			print("INTERNET OFF \n")
			BlitzerChannel <- Blitzer.Blitzer{Vmax: -1}
			time.Sleep(time.Second)
			continue
		}
		response := Blitzer.Decode(resp)

		if response == nil {
			print("Error decoding \n")
			BlitzerChannel <- Blitzer.Blitzer{Vmax: -1}
			time.Sleep(time.Second)
			continue
		}

		Blitzers := Blitzer.GetBlitzer(*response, currPos)
		if len(Blitzers) == 0 {
			print("No Blitzer found \n")
			BlitzerChannel <- Blitzer.Blitzer{Vmax: 0}
			time.Sleep(time.Second)
		} else {
			BlitzerChannel <- Blitzer.GetClosestBlitzer(Blitzers)
			time.Sleep(time.Second)
		}
		if count == 7 {
			count = 0
		} else {
			count += 1
		}
	}
}

func initDevice(path string) *elmobd.Device {
	serialPath := flag.String(
		"serial",
		path,
		"Path to the serial device to use",
	)
	flag.Parse()
	device, err := elmobd.NewDevice(*serialPath, false)

	if err != nil {
		print("Check switch and port \n")
		os.Exit(0)
	}

	return device
}

func getCarStats(CarChannel chan<- Car, device *elmobd.Device) {
	for {
		response, err := device.RunManyOBDCommands([]elmobd.OBDCommand{elmobd.NewEngineRPM(), elmobd.NewVehicleSpeed()})
		if err != nil {
			log.Fatal(err)
		}

		rpm, err := strconv.ParseFloat(response[0].ValueAsLit(), 64)
		speed, err := strconv.ParseInt(response[1].ValueAsLit(), 10, 32)

		// print("RPM: ", rpm, "\n")

		CarChannel <- Car{rpm: int32(rpm), speed: int32(speed)}
		time.Sleep(time.Millisecond * 160)
	}
}

func main() {
	screenWidth := 800.0
	screenHeight := 480.0
	rl.InitWindow(int32(screenWidth), int32(screenHeight), "FesterBlitzer 🫴🫳")
	defer rl.CloseWindow()
	rl.SetTargetFPS(60)

	// Init OBD2 Reader with path/to/usb
	// device := initDevice("/dev/tty.usbserial-11340")
	device := initDevice("test://")

	CarStatsChannel := make(chan Car, 2048)
	go getCarStats(CarStatsChannel, device)
	carStats := Car{}

	BlitzerChannel := make(chan Blitzer.Blitzer, 2048)
	go getBlitzer(BlitzerChannel)
	closestBlitzer := Blitzer.Blitzer{}

	limitOutline := rl.LoadImage("Assets/SpeedSign.png")
	rl.ImageResize(limitOutline, 100, 100)
	speedTexture := rl.LoadTextureFromImage(limitOutline)

	infinity := rl.LoadImage("Assets/infinity.png")
	rl.ImageResize(infinity, 100, 100)
	infinityTexture := rl.LoadTextureFromImage(infinity)

	font := rl.LoadFontEx("Assets/AzeretMono-SemiBold.ttf", 125, nil, 0)

	framesCounter := 0
	oldRPM := int32(0)
	rpm := float32(0)

	for !rl.WindowShouldClose() {
		rl.BeginDrawing()
		rl.DrawFPS(10, 10)
		rl.ClearBackground(rl.Black)

		select {
		case closestBlitzer = <-BlitzerChannel:
		default:
		}
		select {
		case carStats = <-CarStatsChannel:
		default:
		}
		rpm = easings.LinearIn(float32(framesCounter), float32(oldRPM), float32(carStats.rpm)-float32(oldRPM), 30)

		drawSpeed(carStats.speed, font)
		drawrpm(rpm)
		drawBlitzer(closestBlitzer.Distance, closestBlitzer.Vmax, speedTexture, infinityTexture, carStats.speed, font)

		rl.EndDrawing()

		framesCounter += 1
		if framesCounter >= 30 {
			framesCounter = 0
			oldRPM = carStats.rpm
		}
	}
}
