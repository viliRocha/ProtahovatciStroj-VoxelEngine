package main

import (
	"go-engine/src/load"
	"go-engine/src/render"
	"go-engine/src/world"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func main() {
	game := load.InitGame()

	// Main game loop
	for !rl.WindowShouldClose() {
		// Toggle menu
		if rl.IsKeyPressed(rl.KeyTab) {
			render.ShowMenu = !render.ShowMenu

			if render.ShowMenu {
				rl.EnableCursor()
			}
		}

		if rl.IsMouseButtonPressed(rl.MouseLeftButton) && !render.ShowMenu {
			rl.DisableCursor()
		}

		// Update the camera only when it is not in the menu.
		if !render.ShowMenu {
			rl.UpdateCamera(&game.Camera, game.CameraMode)
		}

		// Manage chunks based on player's position
		world.ManageChunks(game.Worley, game.BiomeSelector, game.Camera.Position, game.ChunkCache, game.Perlin1, game.Perlin2, game.Perlin3)

		//  Draw
		render.RenderGame(&game)
	}
	rl.UnloadShader(game.Shader)

	// After the loop ends:
	defer rl.CloseWindow()
}
