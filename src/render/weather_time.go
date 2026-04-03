package render

import (
	"go-engine/src/load"
	"go-engine/src/pkg"
	"math/rand"

	rl "github.com/gen2brain/raylib-go/raylib"
)

var dayLength = float32(160.0) // day duration in seconds

var stormSkyColor = rl.NewColor(120, 120, 120, 255) // grey

var morningSky = rl.NewColor(200, 220, 255, 255) // light Blue
var noonSky = rl.NewColor(150, 208, 233, 255)    // intense blue
var eveningSky = rl.NewColor(255, 180, 120, 255) // orange ish
var nightSky = rl.NewColor(12, 12, 36, 255)      // dark blue

var rainBlend = float32(0.0)

var skyColor = morningSky
var targetSkyColor = morningSky

var currentVolume float32 = 0.0
var baseVolume float32 = 0.5 // global volume a little lower (the audio is very intense)

var wasRaining = false

func lerpColor(a, b rl.Color, f float32) rl.Color {
	return rl.NewColor(
		uint8(float32(a.R)+f*(float32(b.R)-float32(a.R))),
		uint8(float32(a.G)+f*(float32(b.G)-float32(a.G))),
		uint8(float32(a.B)+f*(float32(b.B)-float32(a.B))),
		255,
	)
}

func getDayColor() rl.Color {
	//	"%" what is left from the division
	t := float32(load.ElapsedSeconds%int(dayLength)) / dayLength

	if t < 0.20 { // morning → noon..
		return lerpColor(morningSky, noonSky, t/0.20)
	} else if t < 0.40 {
		return lerpColor(noonSky, eveningSky, (t-0.20)/0.20)
	} else if t < 0.60 {
		return lerpColor(eveningSky, nightSky, (t-0.40)/0.20)
	} else {
		return lerpColor(nightSky, morningSky, (t-0.60)/0.40)
	}
}

func updateSkyColor() {
	//base color of the day cycle
	dayColor := getDayColor()

	// adjust rainBlend smoothly
	step := float32(0.01) // transition speed
	if shouldRain == 1 {
		rainBlend += step
		if rainBlend > 0.5 {
			rainBlend = 0.5
		}
	} else {
		rainBlend -= step
		if rainBlend < 0.0 {
			rainBlend = 0.0
		}
	}

	// mixes the color of the day with the color of the rain (stormSkyColor is 0 if it isn't raining)
	targetSkyColor = lerpColor(dayColor, stormSkyColor, rainBlend)

	// smooth interpolation between the current color and the target color
	skyColor = lerpColor(skyColor, targetSkyColor, 0.05)
}

// Ambient light configured via shader
func updateAmbient() {
	t := float32(load.ElapsedSeconds%int(dayLength)) / dayLength

	if t < 0.20 { // morning
		load.Ambient = 0.4 + 0.1*(t/0.20) // sobe de 0.1 até 0.4
	} else if t < 0.40 { // midday
		load.Ambient = 0.6 // max
	} else if t < 0.60 { // afternoon
		load.Ambient = 0.4 - 0.2*((t-0.4)/0.20) // reduce ambient
	} else { // night
		load.Ambient = 0.1 + 0.2*((t-0.6)/0.40)
	}
}

var currentFogColor = load.FogColor

var currentFog float32 = 0.0

func updateFog(game *load.Game, targetFog float32, targetFogColor []float32) {
	desiredFog := load.FogCoefficient * targetFog

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

func updateRainAudio(targetVolume float32) {
	// Update the stream every frame.
	rl.UpdateMusicStream(load.RainSound)

	// Detects the start of the rain.
	if shouldRain == 1 && !wasRaining {
		rl.SeekMusicStream(load.RainSound, 0.0) // resets the rain sound whenever it starts to rain
		wasRaining = true
	} else if shouldRain == 0 {
		wasRaining = false
	}

	// gently interpolates
	step := float32(0.01)
	if currentVolume < targetVolume {
		currentVolume += step
		if currentVolume > targetVolume {
			currentVolume = targetVolume
		}
	} else if currentVolume > targetVolume {
		currentVolume -= step
		if currentVolume < targetVolume {
			currentVolume = targetVolume
		}
	}

	rl.SetMusicVolume(load.RainSound, currentVolume*baseVolume)
}
