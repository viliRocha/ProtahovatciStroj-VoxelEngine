package render

import (
	"go-engine/src/load"
	"go-engine/src/pkg"
	"math/rand"

	rl "github.com/gen2brain/raylib-go/raylib"
)

var skyColor = clearSkyColor
var targetSkyColor = clearSkyColor

func updateSkyColor() {
	blendFactor := float32(0.03) // transition speed

	skyColor.R = uint8(float32(skyColor.R) + blendFactor*(float32(targetSkyColor.R)-float32(skyColor.R)))
	skyColor.G = uint8(float32(skyColor.G) + blendFactor*(float32(targetSkyColor.G)-float32(skyColor.G)))
	skyColor.B = uint8(float32(skyColor.B) + blendFactor*(float32(targetSkyColor.B)-float32(skyColor.B)))
}

var currentFogColor = load.FogColor

var currentFog float32 = 0.0

func updateFog(game *load.Game, targetFog float32, targetFogColor []float32) {
	// The final value is a mix between what the player set and what the weather suggests.
	desiredFog := (load.FogCoefficient + targetFog) / 2.0

	var step float32 = 0.01
	if currentFog < desiredFog {
		currentFog += step
		if currentFog > desiredFog {
			currentFog = desiredFog
		}
	} else if currentFog > desiredFog {
		currentFog -= step
		if currentFog < desiredFog {
			currentFog = desiredFog
		}
	}

	// --- Update color smoothly ---
	blendStep := float32(0.01)
	for i := 0; i < 4; i++ {
		diff := targetFogColor[i] - currentFogColor[i]
		currentFogColor[i] += blendStep * diff
	}

	fogDensity := float32(currentFog * (1.0 / float32(pkg.ChunkDistance)))

	locFogDensity := rl.GetShaderLocation(game.Shader, "fogDensity")
	rl.SetShaderValue(game.Shader, locFogDensity, []float32{fogDensity}, rl.ShaderUniformFloat)

	locFogColor := rl.GetShaderLocation(game.Shader, "fogColor")
	rl.SetShaderValue(game.Shader, locFogColor, currentFogColor, rl.ShaderUniformVec4)
}

func initRain(game *load.Game, density int, areaSize float32) {
	current := len(pkg.RainDrops)

	if density > current {
		// adds new particles
		for i := 0; i < density-current; i++ {
			pkg.RainDrops = append(pkg.RainDrops, pkg.Particle{
				Position: rl.NewVector3(
					game.Camera.Position.X+rand.Float32()*areaSize-areaSize/2,
					game.Camera.Position.Y+rand.Float32()*20+10,
					rand.Float32()*areaSize-areaSize/2,
				),
				Velocity: rl.NewVector3(0, -rand.Float32()*0.5-0.2, 0),
				Active:   true,
			})
		}
	} else if density < current {
		// deactivates extra particles
		pkg.RainDrops = pkg.RainDrops[:density]
	}
}

func updateRain(game *load.Game, areaSize float32) {
	for i := range pkg.RainDrops {
		if pkg.RainDrops[i].Active {
			pkg.RainDrops[i].Position = rl.Vector3Add(pkg.RainDrops[i].Position, pkg.RainDrops[i].Velocity)

			pos := pkg.RainDrops[i].Position

			// It reappears at the top when it falls from the water level.
			if pos.Y < 42 {
				pos.Y = rand.Float32()*40 + 50
				pos.X = game.Camera.Position.X + rand.Float32()*areaSize - areaSize/2
				pos.Z = game.Camera.Position.Z + rand.Float32()*areaSize - areaSize/2

				pkg.RainDrops[i].Position = pos
			}
		}
	}
}

func drawRain() {
	for _, drop := range pkg.RainDrops {
		if drop.Active {
			size := rl.NewVector3(0.02, 0.3, 0.02)
			rl.DrawCubeV(drop.Position, size, rl.NewColor(156, 154, 154, 255))
		}
	}
}
