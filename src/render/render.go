package render

import (
	"fmt"
	"math/rand"

	"go-engine/src/load"
	"go-engine/src/pkg"
	"go-engine/src/world"

	gui "github.com/gen2brain/raylib-go/raygui"
	rl "github.com/gen2brain/raylib-go/raylib"
)

var waterColor rl.Color = world.BlockTypes["Water"].Color
var cloudColor rl.Color = world.BlockTypes["Cloud"].Color

func RenderVoxels(game *load.Game) {
	cam := game.Camera.Position

	rl.SetShaderValue(game.Shader, rl.GetShaderLocation(game.Shader, "viewPos"), []float32{cam.X, cam.Y, cam.Z}, rl.ShaderUniformVec3)

	// --- Round 1: solids ---
	game.ChunkCache.CacheMutex.RLock()
	for coord, chunk := range game.ChunkCache.Active {
		// Converts chunk coordinate to actual position
		chunkPos := rl.NewVector3(
			float32(coord.X*pkg.ChunkSize),
			0,
			float32(coord.Z*pkg.ChunkSize),
		)

		if chunk.IsOutdated {
			BuildChunkMesh(game, chunk)
			chunk.IsOutdated = false // reset flag → do not rebuild each frame
		}

		// If the chunk has mesh, draw directly
		if chunk.Model.MeshCount > 0 && chunk.Model.Meshes != nil {
			rl.DrawModel(chunk.Model, chunkPos, 1.0, rl.White)
		}
	}

	// --- Passage 2: plants per chunk (without global sorting) ---
	for _, chunk := range game.ChunkCache.Active {
		for _, plant := range chunk.Plants {
			pos := rl.NewVector3(
				plant.Position.X,
				plant.Position.Y,
				plant.Position.Z-0.8) // The plant model is a little off center so i just udjust it a little
			rl.DrawModel(pkg.PlantModels[plant.ModelID], pos, 0.4, rl.White)
		}
	}

	// --- Render Water Mesh ---

	rl.SetBlendMode(rl.BlendAlpha)
	rl.DisableDepthMask()

	for coord, chunk := range game.ChunkCache.Active {

		chunkPos := rl.NewVector3(
			float32(coord.X*pkg.ChunkSize),
			0,
			float32(coord.Z*pkg.ChunkSize),
		)

		if chunk.WaterModel.MeshCount > 0 && chunk.WaterModel.Meshes != nil {
			rl.DrawModel(
				chunk.WaterModel,
				chunkPos,
				1.0,
				waterColor,
			)
		}
	}

	//--- Render Cloud Mesh ---

	for coord, chunk := range game.ChunkCache.Active {

		chunkPos := rl.NewVector3(
			float32(coord.X*pkg.ChunkSize),
			float32(pkg.CloudHeight),
			float32(coord.Z*pkg.ChunkSize),
		)

		if chunk.CloudModel.MeshCount > 0 && chunk.CloudModel.Meshes != nil && ShowClouds {
			rl.DrawModel(
				chunk.CloudModel,
				chunkPos,
				1.0,
				cloudColor,
			)
		}
	}

	game.ChunkCache.CacheMutex.RUnlock()

	rl.EnableDepthMask()
	rl.SetBlendMode(rl.BlendMode(0))
}

func applyUnderwaterEffect(game *load.Game) {
	waterLevel := int(float64(pkg.WorldHeight)*pkg.WaterLevelFraction) + 1

	coord := world.ToChunkCoord(game.Camera.Position)

	//	Protect cache reading
	game.ChunkCache.CacheMutex.RLock()
	chunk := game.ChunkCache.Active[coord]
	game.ChunkCache.CacheMutex.RUnlock()

	if chunk != nil {
		localX := int(game.Camera.Position.X) - coord.X*pkg.ChunkSize
		localY := int(game.Camera.Position.Y) - coord.Y*pkg.WorldHeight
		localZ := int(game.Camera.Position.Z) - coord.Z*pkg.ChunkSize

		if localX >= 0 && localX < pkg.ChunkSize &&
			localY >= 0 && localY < pkg.WorldHeight &&
			localZ >= 0 && localZ < pkg.ChunkSize {

			voxel := chunk.Voxels[localX][localY][localZ]
			if voxel.Type == "Water" && game.Camera.Position.Y < float32(waterLevel)-0.5 {
				// apply blue overlay
				rl.SetBlendMode(rl.BlendMode(0))
				rl.DrawRectangle(0, 0, int32(rl.GetScreenWidth()), int32(rl.GetScreenHeight()), rl.NewColor(0, 0, 255, 100))

				if shouldRain == 1 {
					rl.SetMusicVolume(load.RainSound, currentVolume*0.3)
				}
			}
		}
	}
}

var shouldRain = 0
var nextWeatherChange = 0

var targetFogDensity = float32(1)

func colorToVec4(c rl.Color) []float32 {
	return []float32{
		float32(c.R) / 255.0,
		float32(c.G) / 255.0,
		float32(c.B) / 255.0,
		float32(c.A) / 255.0,
	}
}

func RenderGame(game *load.Game) {
	rl.BeginDrawing()
	rl.ClearBackground(skyColor)
	//rl.ClearBackground(rl.NewColor(134, 13, 13, 255))  Red

	load.UpdateTimer()

	// Only change the climate when the scheduled time comes
	if load.ElapsedSeconds >= nextWeatherChange {
		shouldRain = rand.Intn(4)
		nextWeatherChange = load.ElapsedSeconds + 60 // shedules next change
	}

	updateSkyColor()

	updateFog(game, targetFogDensity, colorToVec4(skyColor))

	updateAmbient()

	locAmbient := rl.GetShaderLocation(game.Shader, "ambient")
	rl.SetShaderValue(game.Shader, locAmbient, []float32{load.Ambient}, rl.ShaderUniformFloat)

	rl.BeginMode3D(game.Camera)

	//	Begin drawing solid blocks and then transparent ones (avoid flickering)
	RenderVoxels(game)

	if shouldRain == 1 {
		targetFogDensity = 1.5

		updateRainAudio(1.0)

		// only initializes if there are not enough particles yet.
		if len(pkg.RainDrops) < 400 {
			initRain(game, 400, 40)
		}

		updateRain(game, 40)

		drawRain()
	} else {
		targetFogDensity = 1

		updateRainAudio(0.0)

		// clean up particles when it stops raining.
		pkg.RainDrops = nil
	}

	rl.EndMode3D()

	applyUnderwaterEffect(game)

	if ShowMenu {
		menuWidth := int32(rl.GetScreenWidth()) / 2
		menuHeight := int32(rl.GetScreenHeight()) / 2
		menuX := (int32(rl.GetScreenWidth()) - menuWidth) / 2
		menuY := (int32(rl.GetScreenHeight()) - menuHeight) / 2

		// Visible menu area
		menuBounds := rl.NewRectangle(float32(menuX), float32(menuY), float32(menuWidth), float32(menuHeight))

		contentHeight := float32(750) // Actual height of the content, including what is not visible.
		contentBounds := rl.NewRectangle(0, 0, float32(menuWidth-20), contentHeight)

		gui.ScrollPanel(menuBounds, "Game settings", contentBounds, &menuScroll, &menuView)

		renderMenu(menuX, menuY, menuWidth)
	}

	if ShowFPS {
		rl.DrawFPS(10, 30)
	}

	if FrameLimit != prevFrameLimit {
		rl.SetTargetFPS(int32(FrameLimit))
	}

	if ShowPosition {
		positionText := fmt.Sprintf("Player's position: (%.2f, %.2f, %.2f)", game.Camera.Position.X, game.Camera.Position.Y, game.Camera.Position.Z)
		rl.DrawText(positionText, 10, 5, 20, rl.DarkGreen)
	}

	rl.EndDrawing()
}
