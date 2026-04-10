package render

import (
	"fmt"
	"math"

	"go-engine/src/load"
	"go-engine/src/pkg"

	gui "github.com/gen2brain/raylib-go/raygui"
	rl "github.com/gen2brain/raylib-go/raylib"
	"golang.org/x/exp/constraints"
)

var ShowMenu bool = false
var ShowFPS bool = true
var ShowPosition bool = true
var ShowClouds bool = true

var FrameLimit int = 60
var prevFrameLimit int = FrameLimit

var menuScroll rl.Vector2
var menuView rl.Rectangle

func renderMenu(menuX, menuY, width int32) {
	rl.BeginScissorMode(
		int32(menuView.X),
		int32(menuView.Y),
		int32(menuView.Width),
		int32(menuView.Height),
	)

	offsetY := int32(menuScroll.Y)

	newButton(menuX+20, menuY+40+offsetY, float32(width-40), 40.0, &ShowPosition, "Show Player Position")

	newButton(menuX+20, menuY+90+offsetY, float32(width-40), 40.0, &ShowFPS, "Show FPS")

	//Y = Y + 60 + 30
	newGuiSlider(menuX+20, menuY+140+offsetY, float32(width-40), 40.0,
		&FrameLimit, 30, 120,
		fmt.Sprintf("FPS Limit: %d", FrameLimit),
	)

	newButton(menuX+20, menuY+230+offsetY, float32(width-40), 40.0, &ShowClouds, "Clouds")

	newGuiSlider(menuX+20, menuY+290+offsetY, float32(width-40), 40.0,
		&pkg.CloudHeight, 30, 120,
		fmt.Sprintf("Cloud Height: %d", pkg.CloudHeight),
	)

	newGuiSlider(menuX+20, menuY+380+offsetY, float32(width-40), 40.0,
		&pkg.ChunkDistance, 1, 10,
		fmt.Sprintf("View Distance: %d", pkg.ChunkDistance),
	)

	newGuiSlider(menuX+20, menuY+470+offsetY, float32(width-40), 40.0,
		&load.FogCoefficient, 0.0, 0.1,
		fmt.Sprintf("Fog Density: %.3f", load.FogCoefficient),
	)

	newGuiSlider(menuX+20, menuY+560+offsetY, float32(width-40), 40.0,
		&baseVolume, 0.0, 1.0,
		fmt.Sprintf("Sound FX Volume: %.3f", baseVolume),
	)

	rl.EndScissorMode()
}

func newButton(menuX, menuY int32, buttonWidth, buttonHeight float32, isOn *bool, text string) {
	// raygui button
	if gui.Button(rl.NewRectangle(float32(menuX), float32(menuY), buttonWidth, buttonHeight),
		func() string {
			if *isOn {
				return fmt.Sprintf("%s: ON", text)
			}
			return fmt.Sprintf("%s: OFF", text)
		}()) {
		*isOn = !*isOn
	}
}

func newGuiSlider[T constraints.Integer | constraints.Float](menuX, menuY int32, barWidth, barHeight float32, value *T, minVal, maxVal float32, text string) {
	floatVal := float32(*value)

	rl.DrawText(text, menuX, menuY, 20, rl.DarkGray)

	floatVal = gui.Slider(rl.NewRectangle(float32(menuX), float32(menuY+30), barWidth, barHeight),
		"", "",
		floatVal, minVal, maxVal,
	)

	switch any(value).(type) {
	case *int:
		*value = T(int(math.Floor(float64(floatVal))))
	case *float32:
		*value = T(floatVal)
	}
}
