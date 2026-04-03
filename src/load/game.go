package load

import (
	"fmt"
	"math/rand"
	"time"

	"go-engine/src/pkg"
	"go-engine/src/world"

	"github.com/aquilax/go-perlin"
	gui "github.com/gen2brain/raylib-go/raygui"
	rl "github.com/gen2brain/raylib-go/raylib"
)

const (
	ScreenWidth  int32 = 1000
	ScreenHeight int32 = 480

	//  Dimension of the space in which Perlin Noise is being calculated. For example, in 3D, it would be 3.
	perlinN = int32(2)

	//  Control of the intensity/amplitude of the noise
	perlinAlpha = 3
	//  Adjust the frequency of noise, affecting the amount of detail present in the noise by controlling the scale of the variations.
	perlinBeta = 1.5
)

var FogColor = []float32{0.588, 0.816, 0.914, 1.0} // clearSky

var FogCoefficient float32 = 0.072

var Ambient float32 = 0.4

var startTime time.Time
var ElapsedSeconds int

func initTimer() {
	startTime = time.Now()
}

func UpdateTimer() {
	elapsed := time.Since(startTime)
	ElapsedSeconds = int(elapsed.Seconds())
}

var RainSound rl.Music

func initAudio() {
	RainSound = rl.LoadMusicStream("./assets/sounds/drizzle.mp3")
	rl.SetMusicVolume(RainSound, 0.0)
	rl.PlayMusicStream(RainSound)
}

type Game struct {
	Camera        rl.Camera
	CameraMode    rl.CameraMode
	ChunkCache    *world.ChunkCache
	Perlin1       *perlin.Perlin
	Perlin2       *perlin.Perlin
	Perlin3       *perlin.Perlin
	Worley        *world.WorleyNoise
	BiomeSelector *world.BiomeSelector
	Shader        rl.Shader
	//LightPosition rl.Vector3
}

func InitGame() Game {
	rl.SetConfigFlags(rl.FlagWindowResizable)
	rl.InitWindow(ScreenWidth, ScreenHeight, "Protahovatsi Stroj - Voxel Game")

	// Disable INFO logs, only show errors
	rl.SetTraceLogLevel(rl.LogError)

	camera := rl.Camera{
		Position:   rl.NewVector3(2.79, 62.0, 10.0),
		Target:     rl.NewVector3(0.0, 0.0, 0.0),
		Up:         rl.NewVector3(0.0, 1.0, 0.0),
		Fovy:       45.0,
		Projection: rl.CameraPerspective,
	}
	cameraMode := rl.CameraFree

	// Initializes Perlin noise
	seed1 := rand.Int63()
	seed2 := rand.Int63()
	seed3 := rand.Int63()

	perlin1 := perlin.NewPerlin(perlinAlpha, perlinBeta, perlinN, seed1)
	perlin2 := perlin.NewPerlin(perlinAlpha, perlinBeta, perlinN, seed2)
	perlin3 := perlin.NewPerlin(perlinAlpha, perlinBeta, perlinN, seed3)
	/*
		seed1 := rand.Int63()
		perlin1 := perlin.NewPerlin(perlinAlpha, perlinBeta, perlinN, seed1)
		perlin2 := perlin1
		perlin3 := perlin2
	*/

	worley := world.NewWorleyNoise(seed1, 128)
	biomeSel := world.NewBiomeSelector(seed1, 128)

	Shader := rl.LoadShader("shaders/shader.vs", "shaders/shader.fs")

	// Locations (do not index shader.Locs)
	locFogDensity := rl.GetShaderLocation(Shader, "fogDensity")

	// Fog density calculation
	// Inverse relationship: more chunks → less fog
	fogDensity := float32(FogCoefficient * (1.0 / float32(pkg.ChunkDistance)))
	//fmt.Println(fogDensity)
	rl.SetShaderValue(Shader, locFogDensity, []float32{fogDensity}, rl.ShaderUniformFloat)

	locAmbient := rl.GetShaderLocation(Shader, "ambient")
	rl.SetShaderValue(Shader, locAmbient, []float32{Ambient}, rl.ShaderUniformFloat)

	locFogColor := rl.GetShaderLocation(Shader, "fogColor")
	rl.SetShaderValue(Shader, locFogColor, FogColor, rl.ShaderUniformVec4)

	lightLoc := rl.GetShaderLocation(Shader, "lightDir")
	rl.SetShaderValue(Shader, lightLoc, []float32{-1, -1, -0.5}, rl.ShaderUniformVec3)

	//	Start weather & time cycle
	initTimer()

	initAudio()

	// Load .vox models
	for i := 0; i < len(pkg.PlantModels); i++ {
		pkg.PlantModels[i] = rl.LoadModel(fmt.Sprintf("assets/plants/plant_%d.vox", i))

		// Ensure the material is valid and apply shader
		if pkg.PlantModels[i].MaterialCount == 0 || pkg.PlantModels[i].Materials == nil {
			def := rl.LoadMaterialDefault()
			pkg.PlantModels[i].MaterialCount = 1
			pkg.PlantModels[i].Materials = &def
		}
		(*pkg.PlantModels[i].Materials).Shader = Shader
	}

	//LightPosition := rl.NewVector3(0, 34, 0)

	chunkCache := world.NewChunkCache() // Initialize ChunkCache

	// Creates the first chunk at the origin
	originCoord := pkg.Coords{X: 0, Y: 0, Z: 0}
	originPos := rl.NewVector3(0, 0, 0)

	chunkCache.Active[originCoord] = world.GenerateChunk(worley, biomeSel, originPos, perlin1, perlin2, perlin3, chunkCache, nil, false, nil, false)

	//	That value can later be changed in game, it is just initialized here
	rl.SetTargetFPS(60)

	defaultFont := rl.GetFontDefault()

	gui.SetFont(defaultFont)

	gui.SetStyle(gui.DEFAULT, gui.TEXT_SIZE, 20) // bigger text font for buttons

	// Color when mouse hovers over the button (RRGGBBAA)
	gui.SetStyle(gui.BUTTON, gui.BORDER_COLOR_FOCUSED, 0x888888ff)
	gui.SetStyle(gui.BUTTON, gui.BASE_COLOR_FOCUSED, 0xaaaaaaff)
	gui.SetStyle(gui.BUTTON, gui.TEXT_COLOR_FOCUSED, 0xffffffff)

	// Color when pressed
	gui.SetStyle(gui.BUTTON, gui.BORDER_COLOR_PRESSED, 0x333333ff)
	gui.SetStyle(gui.BUTTON, gui.BASE_COLOR_PRESSED, 0x444444ff)
	gui.SetStyle(gui.BUTTON, gui.TEXT_COLOR_PRESSED, 0xffffffff)

	//	Slider Bar style:
	gui.SetStyle(gui.SLIDER, gui.BORDER_COLOR_NORMAL, 0x888888ff)

	gui.SetStyle(gui.SLIDER, gui.BORDER_COLOR_FOCUSED, 0x444444ff)

	gui.SetStyle(gui.SLIDER, gui.BORDER_COLOR_PRESSED, 0x333333ff)

	return Game{
		Camera:        camera,
		CameraMode:    cameraMode,
		ChunkCache:    chunkCache,
		Perlin1:       perlin1,
		Perlin2:       perlin2,
		Perlin3:       perlin3,
		Worley:        worley,
		BiomeSelector: biomeSel,
		Shader:        Shader,
		//LightPosition: LightPosition,
	}
}
