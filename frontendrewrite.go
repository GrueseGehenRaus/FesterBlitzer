package main

import (
	"math"
	"strconv"

	rl "github.com/gen2brain/raylib-go/raylib"
)

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

func drawRPM(RPM int32) {
	screenWidth := rl.GetScreenWidth()
	screenHeight := rl.GetScreenHeight()

	// Base dimensions of the meter
	meterWidth := float32(80)
	meterHeight := float32(300)
	meterX := float32(screenWidth/2) - 240
	meterY := float32(screenHeight/2) - 150

	// Draw the meter background with rounded bottom corners
	rl.DrawRectangleRoundedLinesEx(rl.Rectangle{X: meterX, Y: meterY, Width: meterWidth, Height: meterHeight},
		0.5, 0, 2.0, rl.Fade(rl.Blue, 0.4))

	if RPM > 0 {
		// Compute percentage fill
		percent := float32(RPM) / 6000
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

func drawSpeed(speed int32) {
	rl.DrawText("69", int32(rl.GetScreenWidth()/2)-50, int32(rl.GetScreenHeight()/2)-50, 100, rl.White)
	rl.DrawText("km/h", int32(rl.GetScreenWidth()/2)-50, int32(rl.GetScreenHeight()/2)+50, 50, rl.White)
}

func drawBlitzer(distance float64) {
	centerY := 400.0
	topWidth := 175.0
	bottomWidth := 200.0
	height := 40.0

	fillCount := ((1 - distance/1000) * 5)

	for i := float64(0); i < 5; i++ {
		if i < math.Round(fillCount) {
			drawTrapezoid(600, float32(centerY), float32(topWidth), float32(bottomWidth), float32(height), rl.Green)
		} else {
			drawTrapezoid(600, float32(centerY), float32(topWidth), float32(bottomWidth), float32(height), rl.Gray)
		}
		centerY = centerY * 0.85
		topWidth = topWidth * 0.85
		bottomWidth = bottomWidth * 0.85
		height = height * 0.85
	}
	rl.DrawText(strconv.FormatFloat(distance, 'f', 0, 64), 584, 175, 20, rl.White)
}

func main() {
	screenWidth := 800.0
	screenHeight := 480.0
	rl.InitWindow(int32(screenWidth), int32(screenHeight), "FesterBlitzer")
	defer rl.CloseWindow()
	rl.SetTargetFPS(60)

	for !rl.WindowShouldClose() {
		rl.BeginDrawing()
		rl.DrawFPS(10, 10)
		rl.ClearBackground(rl.Black)

		drawSpeed(69)
		drawRPM(1500)
		drawBlitzer(999)

		rl.EndDrawing()
	}
}
