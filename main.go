package main

import (
	Blitzer "FesterBlitzer/Blitzer"
	"flag"
	"fmt"
	"image/color"
	"log"
	"math"
	"net/http"
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

func drawKMH(speed int) {
	if speed < 10 {
		rl.DrawText(strconv.Itoa(speed), int32(rl.GetScreenWidth()/2)-40, int32(rl.GetScreenHeight()/2)-30, 100, rl.White)
	} else if speed < 100 {
		rl.DrawText(strconv.Itoa(speed), int32(rl.GetScreenWidth()/2)-50, int32(rl.GetScreenHeight()/2)-30, 100, rl.White)
	} else if speed < 220 {
		rl.DrawText(strconv.Itoa(speed), int32(rl.GetScreenWidth()/2)-80, int32(rl.GetScreenHeight()/2)-30, 100, rl.White)
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
		time.Sleep(time.Millisecond * 16)
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

func getBlitzer(BlitzerChannel chan <- Blitzer.Blitzer) {
	for {
		// getPos() braucht man halt und noch lastpos speichern vor schreiben vom Blitzer in den channel
		// Blitzer types noch casen (6 ist z.B. Abstandsmessung)
		
		// HEK nach Norden
		//lastPos := [2]float64{49.0161, 8.3980}
		//currPos := [2]float64{49.0189, 8.3974}
	
		// Hailfingen nach Seebron
		lastPos := [2]float64{48.515966, 8.869765}
		currPos := [2]float64{48.515276, 8.870355}
		
		scanBox := Blitzer.GetScanBox(lastPos, currPos)
		boxStart, boxEnd := Blitzer.GetBoundingBox(scanBox)
	
		url := fmt.Sprintf("https://cdn2.atudo.net/api/4.0/pois.php?type=22,26,20,101,102,103,104,105,106,107,108,109,110,111,112,113,115,117,114,ts,0,1,2,3,4,5,6,21,23,24,25,29,vwd,traffic&z=17&box=%f,%f,%f,%f",
			boxStart[0], boxStart[1], boxEnd[0], boxEnd[1])
		print(url)
	
		resp, err := http.Get(url)
		if err != nil {
			BlitzerChannel <- Blitzer.Blitzer{Vmax: -1}
			time.Sleep(time.Second)
			return
		}
		response := Blitzer.Decode(resp)
		Blitzers := Blitzer.GetBlitzer(response, currPos)
		if len(Blitzers) == 0 {
			// auch hier UI handling wenn es keine Blitzer hat? Maybe Autobahn keine Begrenzung Schild xd
			time.Sleep(time.Second)
			return
		}
		BlitzerChannel <- Blitzer.GetClosestBlitzer(Blitzers)
		time.Sleep(time.Second)
	}
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

	// Init OBD2 Reader with path/to/usb
	//device := initDevice("/dev/tty.usbserial-11310")
	device := initDevice("test://")

	// RPM definitions
	RPMChannel := make(chan float32, 2048)
	go getEngineRPM(*device, RPMChannel)
	RPMStart := 120
	RPMMax := 260
	rpm := float32(0)

	// Speed definitions
	SpeedChannel := make(chan int, 2048)
	go getCarSpeed(*device, SpeedChannel, file)
	speed := 0

	// Throttle definitions
	ThrottleChannel := make(chan float32, 2048)
	go getThrottlePos(*device, ThrottleChannel)
	ThrottleStart := 110
	ThrottleMax := float32(70.0)
	throttlePos := float32(0)

	// Blitzer definitions
	BlitzerChannel := make(chan Blitzer.Blitzer, 2048)
	go getBlitzer(BlitzerChannel)
	blitzer := Blitzer.Blitzer{Vmax: 0}
	
	// Runtime definitions
	RuntimeChannel := make(chan float32, 2048)
	// go getRuntimeSinceEngineStart(*device, RuntimeChannel)

	// Eco definitions
	EcoChannel := make(chan Eco, 2048)
	go getEcoScore(EcoChannel, RPMChannel, ThrottleChannel, RuntimeChannel)

	// Define Circle Positions
	circleCenter := rl.NewVector2(float32(rl.GetScreenWidth()/2), float32(rl.GetScreenHeight()/2))
	circleInnerRadius := 350.0
	circleOuterRadius := 400.0
	
	// Load all speed signs
	// Consider: Only load sign and write numbers urself
	// Also load texture outside of for loop!
	limitNeg1 := rl.LoadImage("Assets/-1.png")
	rl.ImageResize(limitNeg1, 200, 200)
	limit30 := rl.LoadImage("Assets/30.png")
	rl.ImageResize(limit30, 200, 200)
	limit50 := rl.LoadImage("Assets/50.png")
	rl.ImageResize(limit50, 200, 200)
	limit60 := rl.LoadImage("Assets/60.png")
	rl.ImageResize(limit60, 200, 200)
	limit70 := rl.LoadImage("Assets/70.png")
	rl.ImageResize(limit70, 200, 200)
	limit80 := rl.LoadImage("Assets/80.png")
	rl.ImageResize(limit80, 200, 200)
	limit100 := rl.LoadImage("Assets/100.png")
	rl.ImageResize(limit100, 200, 200)

	for !rl.WindowShouldClose() {
		// things to add:
		// engine load
		// coolant temp
		// runtime_since_engine_start
		// correct error handling von Durak klauen
		// check if device.RunOBDCommand(elmobd.NewEngineOilTemperature()) is a thing (wäre cool)

		
		// Get new value form channels
		select {
			case rpm = <-RPMChannel:
			default:
		}
		select {
			case speed = <-SpeedChannel:
			default:
		}
		select {
			case throttlePos = <-ThrottleChannel:
			default:
		}
		select {
			case blitzer = <-BlitzerChannel:
			default:
		}

		RPMEnd := getRPMDegrees(rpm)
		RPMColor := getRPMColor(rpm)
		ThrottleMax = getThrottleDegrees(throttlePos)
		
		rl.BeginDrawing()
		rl.DrawFPS(10, 10)
		rl.ClearBackground(rl.Black)
		
		switch blitzer.Vmax {
			case -1:
				texture := rl.LoadTextureFromImage(limitNeg1)
				rl.DrawTexture(texture, int32(rl.GetScreenWidth())/2-texture.Width/2, int32(rl.GetScreenHeight())/2-texture.Height/2-200, rl.White)
			case 0:
				// do nothing
			case 30:
				texture := rl.LoadTextureFromImage(limit30)
				rl.DrawTexture(texture, int32(rl.GetScreenWidth())/2-texture.Width/2, int32(rl.GetScreenHeight())/2-texture.Height/2-200, rl.White)
			case 50:
				texture := rl.LoadTextureFromImage(limit50)
				rl.DrawTexture(texture, int32(rl.GetScreenWidth())/2-texture.Width/2, int32(rl.GetScreenHeight())/2-texture.Height/2-200, rl.White)
			case 60:
				texture := rl.LoadTextureFromImage(limit60)
				rl.DrawTexture(texture, int32(rl.GetScreenWidth())/2-texture.Width/2, int32(rl.GetScreenHeight())/2-texture.Height/2-200, rl.White)
			case 70:
				texture := rl.LoadTextureFromImage(limit70)
				rl.DrawTexture(texture, int32(rl.GetScreenWidth())/2-texture.Width/2, int32(rl.GetScreenHeight())/2-texture.Height/2-200, rl.White)
			case 80:
				texture := rl.LoadTextureFromImage(limit80)
				rl.DrawTexture(texture, int32(rl.GetScreenWidth())/2-texture.Width/2, int32(rl.GetScreenHeight())/2-texture.Height/2-200, rl.White)
			case 90:
				// braucht noch Bild
				texture := rl.LoadTextureFromImage(limit80)
				rl.DrawTexture(texture, int32(rl.GetScreenWidth())/2-texture.Width/2, int32(rl.GetScreenHeight())/2-texture.Height/2-200, rl.White)
			case 100:
				texture := rl.LoadTextureFromImage(limit100)
				rl.DrawTexture(texture, int32(rl.GetScreenWidth())/2-texture.Width/2, int32(rl.GetScreenHeight())/2-texture.Height/2-200, rl.White)
			case 120:
				// Bild braucht man noch!
				return
			default:
				log.Fatal(blitzer.Vmax, " hat noch kein Image!")
		}
		
		// Draw Speed
		drawKMH(speed)
		
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
		
		// Draw Needle
		rl.DrawLineEx(rl.Vector2{X: needleStart.X + float32(rl.GetScreenWidth()/2), Y: needleStart.Y + float32(rl.GetScreenHeight()/2)}, rl.Vector2{needleEnd.X + float32(rl.GetScreenWidth()/2), needleEnd.Y + float32(rl.GetScreenHeight()/2)}, 5, rl.Red)

		rl.EndDrawing()
	}
}
