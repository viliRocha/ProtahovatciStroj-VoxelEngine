package render

import (
	"fmt"
	"math/rand"
	"sort"

	"go-engine/src/load"
	"go-engine/src/pkg"
	"go-engine/src/world"

	gui "github.com/gen2brain/raylib-go/raygui"
	rl "github.com/gen2brain/raylib-go/raylib"
)

func RenderVoxels(game *load.Game) {
	cam := game.Camera.Position

	rl.SetShaderValue(game.Shader, rl.GetShaderLocation(game.Shader, "viewPos"), []float32{cam.X, cam.Y, cam.Z}, rl.ShaderUniformVec3)

	// --- Round 1: solids ---
	for coord, chunk := range game.ChunkCache.Active {
		// Converts chunk coordinate to actual position
		chunkPos := rl.NewVector3(
			float32(coord.X*pkg.ChunkSize),
			0,
			float32(coord.Z*pkg.ChunkSize),
		)

		if chunk.IsOutdated {
			BuildChunkMesh(game, chunk, chunkPos)
			chunk.IsOutdated = false // reset flag → do not rebuild each frame
		}

		// If the chunk has mesh, draw directly
		if chunk.Model.MeshCount > 0 && chunk.Model.Meshes != nil {
			rl.DrawModel(chunk.Model, chunkPos, 1.0, rl.White)
		}
	}

	// --- Passage 2: plants per chunk (without global sorting) ---
	for coord, chunk := range game.ChunkCache.Active {
		chunkPos := rl.NewVector3(
			float32(coord.X*pkg.ChunkSize),
			0,
			float32(coord.Z*pkg.ChunkSize),
		)

		for _, voxel := range chunk.SpecialVoxels {
			if voxel.Type == "Plant" {
				pos := rl.NewVector3(
					chunkPos.X+float32(voxel.Position.X),
					chunkPos.Y+float32(voxel.Position.Y),
					chunkPos.Z+float32(voxel.Position.Z))
				rl.DrawModel(voxel.Model, pos, 0.4, rl.White)
			}
		}
	}

	// --- Global collection of transparencies ---
	var transparentItems []pkg.TransparentItem

	for coord, chunk := range game.ChunkCache.Active {
		chunkPos := rl.NewVector3(
			float32(coord.X*pkg.ChunkSize),
			0,
			float32(coord.Z*pkg.ChunkSize),
		)

		for _, voxel := range chunk.SpecialVoxels {
			pos := rl.NewVector3(
				chunkPos.X+float32(voxel.Position.X),
				chunkPos.Y+float32(voxel.Position.Y),
				chunkPos.Z+float32(voxel.Position.Z),
			)

			transparentItems = append(transparentItems, pkg.TransparentItem{
				Position:       pos,
				Type:           voxel.Type,
				Color:          world.BlockTypes[voxel.Type].Color,
				IsSurfaceWater: voxel.Type == "Water" && voxel.IsSurface,
			})
		}
	}

	// --- Back-to-front sorting ---
	if game.Camera.Position.Y >= float32(pkg.CloudHeight) {
		sort.Slice(transparentItems, func(i, j int) bool {
			di := rl.Vector3Length(rl.Vector3Subtract(transparentItems[i].Position, cam))
			dj := rl.Vector3Length(rl.Vector3Subtract(transparentItems[j].Position, cam))
			return di > dj // furthest first
		})
	}

	// --- Round 3: transparent ---
	rl.SetBlendMode(rl.BlendAlpha)
	rl.DisableDepthMask()
	//rl.BeginShaderMode(game.Shader)
	for _, it := range transparentItems {
		switch it.Type {
		case "Water":
			p := rl.NewVector3(it.Position.X+0.5, it.Position.Y+0.5, it.Position.Z+0.5)

			/*
				// calculates light intensity for this position
				lightIntensity := calculateLightIntensity(p, game.LightPosition)
				litColor := applyLighting(it.Color, lightIntensity)
			*/

			rl.DrawPlane(p, rl.NewVector2(1.0, 1.0), it.Color)
		case "Cloud":
			/*
				lightIntensity := calculateLightIntensity(it.Position, game.LightPosition)
				litColor := applyLighting(it.Color, lightIntensity)
			*/
			if ShowClouds {
				rl.DrawCube(it.Position, 1.0, 0.0, 1.0, it.Color)
			}
		}
	}

	//rl.EndShaderMode()
	rl.EnableDepthMask()
	rl.SetBlendMode(rl.BlendMode(0))
}

func applyUnderwaterEffect(game *load.Game) {
	waterLevel := int(float64(pkg.WorldHeight)*pkg.WaterLevelFraction) + 1

	coord := world.ToChunkCoord(game.Camera.Position)
	chunk := game.ChunkCache.Active[coord]

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
			}
		}
	}
}

var shouldRain = 0
var nextWeatherChange = 0

var stormSkyColor = rl.NewColor(120, 120, 120, 255) // grey
var clearSkyColor = rl.NewColor(150, 208, 233, 255) // normal blue

var stormFog float32 = 0.2
var stormFogColor = []float32{0.5, 0.5, 0.5, 1.0} // grey

var clearFog float32 = 0.072
var clearFogColor = []float32{0.588, 0.816, 0.914, 1.0} // grey

func RenderGame(game *load.Game) {
	rl.BeginDrawing()
	rl.ClearBackground(skyColor)
	//rl.ClearBackground(rl.NewColor(134, 13, 13, 255))  Red

	load.UpdateTimer()

	// Only change the climate when the scheduled time comes
	if load.ElapsedSeconds >= nextWeatherChange {
		shouldRain = rand.Intn(3)
		nextWeatherChange = load.ElapsedSeconds + 20 // shedules next change
	}

	updateSkyColor()

	rl.BeginMode3D(game.Camera)

	//	Begin drawing solid blocks and then transparent ones (avoid flickering)
	RenderVoxels(game)

	if shouldRain == 1 {
		targetSkyColor = stormSkyColor

		updateFog(game, stormFog, stormFogColor)

		// only initializes if there are not enough particles yet.
		if len(pkg.RainDrops) < 300 {
			initRain(game, 300, 30)
		}

		updateRain(game, 30)

		drawRain()
	} else {
		targetSkyColor = clearSkyColor

		updateFog(game, clearFog, clearFogColor)

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

		contentHeight := float32(600) // Actual height of the content, including what is not visible.
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
