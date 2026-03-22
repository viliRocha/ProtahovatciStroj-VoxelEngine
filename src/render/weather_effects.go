package render

import (
	"go-engine/src/load"

	rl "github.com/gen2brain/raylib-go/raylib"
)

var skyColor = targetSkyColor

func updateSkyColor(fromColor, toColor rl.Color) {
	// smooth interpolation
	var blendFactor float32 = 0.01 // transition speed
	skyColor.R = uint8(float32(skyColor.R) + blendFactor*float32(int(toColor.R)-int(fromColor.R)))
	skyColor.G = uint8(float32(skyColor.G) + blendFactor*float32(int(toColor.G)-int(fromColor.G)))
	skyColor.B = uint8(float32(skyColor.B) + blendFactor*float32(int(toColor.B)-int(fromColor.B)))
}

var targetFogColor = []float32{0.5, 0.5, 0.5, 1.0} // grey
var currentFogColor = load.FogColor
var targetFog float32 = 0.1
var currentFog float32 = load.FogCoefficient

func updateFog(game *load.Game) {
	var step float32 = 0.005
	if currentFog < targetFog {
		currentFog += step
	} else if currentFog > targetFog {
		currentFog -= step
	}

	// --- Update color smoothly ---
	blendStep := float32(0.01)
	for i := 0; i < 4; i++ {
		diff := targetFogColor[i] - currentFogColor[i]
		currentFogColor[i] += blendStep * diff
	}

	locFogDensity := rl.GetShaderLocation(game.Shader, "fogDensity")
	rl.SetShaderValue(game.Shader, locFogDensity, []float32{currentFog}, rl.ShaderUniformFloat)

	locFogColor := rl.GetShaderLocation(game.Shader, "fogColor")
	rl.SetShaderValue(game.Shader, locFogColor, currentFogColor, rl.ShaderUniformVec4)
}
