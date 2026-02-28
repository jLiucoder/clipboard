// +build ignore

// Run this to regenerate icon_1024.png:
//   go run generate_icon.go
//
// Then run ./update_icon.sh to convert to .icns format

package main

import (
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
)

func main() {
	// Create 1024x1024 icon
	size := 1024
	img := image.NewRGBA(image.Rect(0, 0, size, size))

	// Black background
	draw.Draw(img, img.Bounds(), &image.Uniform{color.Black}, image.Point{}, draw.Src)

	// Draw clipboard outline (white)
	clipboardColor := color.White
	
	// Board rectangle
	boardMargin := 220
	boardTop := 300
	boardBottom := size - boardMargin
	boardLeft := boardMargin
	boardRight := size - boardMargin
	
	// Draw board body (rounded corners simulated with rectangles)
	for y := boardTop + 40; y < boardBottom; y++ {
		for x := boardLeft; x < boardRight; x++ {
			img.Set(x, y, clipboardColor)
		}
	}
	
	// Draw board sides (corner rounding)
	cornerSize := 40
	for i := 0; i < cornerSize; i++ {
		// Top left corner
		for x := boardLeft; x < boardLeft+cornerSize-i; x++ {
			img.Set(x, boardTop+40+i, clipboardColor)
		}
		// Top right corner
		for x := boardRight-cornerSize+i; x < boardRight; x++ {
			img.Set(x, boardTop+40+i, clipboardColor)
		}
		// Bottom left corner
		for x := boardLeft; x < boardLeft+cornerSize-i; x++ {
			img.Set(x, boardBottom-1-i, clipboardColor)
		}
		// Bottom right corner
		for x := boardRight-cornerSize+i; x < boardRight; x++ {
			img.Set(x, boardBottom-1-i, clipboardColor)
		}
	}
	
	// Draw clip at top (rounded rectangle shape)
	clipWidth := 200
	clipHeight := 80
	clipLeft := (size - clipWidth) / 2
	clipTop := boardTop - 20
	
	for y := clipTop; y < clipTop+clipHeight; y++ {
		for x := clipLeft; x < clipLeft+clipWidth; x++ {
			img.Set(x, y, clipboardColor)
		}
	}
	
	// Draw clip rounded top
	clipRound := 30
	for i := 0; i < clipRound; i++ {
		for x := clipLeft + clipRound - i; x < clipLeft+clipWidth-clipRound+i; x++ {
			img.Set(x, clipTop-1-i, clipboardColor)
		}
	}
	
	// Draw inner lines (document lines)
	lineColor := color.Black
	lineMargin := 80
	lineStart := boardTop + 100
	lineSpacing := 60
	lineWidth := 8
	
	for i := 0; i < 5; i++ {
		y := lineStart + i*lineSpacing
		for x := boardLeft + lineMargin; x < boardRight-lineMargin; x++ {
			for dy := 0; dy < lineWidth; dy++ {
				img.Set(x, y+dy, lineColor)
			}
		}
	}

	// Save as PNG
	f, _ := os.Create("icon_1024.png")
	defer f.Close()
	png.Encode(f, img)
	
	println("Created icon_1024.png")
}
