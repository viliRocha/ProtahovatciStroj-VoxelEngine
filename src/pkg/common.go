package pkg

import (
	"github.com/aquilax/go-perlin"
	rl "github.com/gen2brain/raylib-go/raylib"
)

var PlantModels [4]rl.Model

var ChunkDistance int = 5
var CloudHeight int = 80

const (
	WorldHeight        int     = 112
	ChunkSize          int     = 16
	WaterLevelFraction float64 = 0.375 // 3/8
)

type VoxelData struct {
	Type  string
	Model rl.Model
	Color rl.Color
}

// Stores plant positions
type PlantData struct {
	Position rl.Vector3
	ModelID  int
}

type TreeData struct {
	Position     rl.Vector3
	StructureStr string
}

type SpecialVoxel struct {
	Position  Coords
	Type      string
	Model     rl.Model // for plants
	IsSurface bool     // for water
}

type TransparentItem struct {
	Position       rl.Vector3
	Type           string
	Color          rl.Color
	IsSurfaceWater bool
}

type Chunk struct {
	Voxels    [ChunkSize][WorldHeight][ChunkSize]VoxelData
	HeightMap [ChunkSize][ChunkSize]int // final height per terrain column
	BiomeMap  [ChunkSize][ChunkSize]BiomeProperties
	Neighbors [4]*Chunk // 0: +X, 1: -X, 2: 4: +Z, 5: -Z
	Plants    []PlantData
	Trees     []TreeData

	// Buffers reutilizáveis para mesh
	Vertices []float32
	Indices  []uint16
	Colors   []uint8
	Normals  []float32

	Mesh          rl.Mesh
	Model         rl.Model
	CloudModel    rl.Model
	SpecialVoxels []SpecialVoxel
	IsOutdated    bool // Flag to know if you need to update the mesh
}

type Coords struct {
	X, Y, Z int
}

// Function that calculates the height of a biome
type HeightFunc func(gx, gz int, p2, p3 *perlin.Perlin) float64

type BiomeProperties struct {
	Modifier         HeightFunc
	SurfaceBlock     string
	UndergroundBlock string
	TreeTypes        []string
	TreeDensity      float32
	GrassColor       rl.Color
	LeavesColor      rl.Color
}

type Particle struct {
	Position rl.Vector3
	Velocity rl.Vector3
	Active   bool
}

var RainDrops []Particle

var FaceDirections = []rl.Vector3{
	{1, 0, 0},  // Front
	{-1, 0, 0}, // Back
	{0, 1, 0},  // Left
	{0, -1, 0}, // Right
	{0, 0, 1},  // Top
	{0, 0, -1}, // Bottom
}

var HorizontalDirections = []rl.Vector3{
	{1, 0, 0},
	{-1, 0, 0},
	{0, 0, 1},
	{0, 0, -1},
}

var FaceVertices = [6][4][3]float32{
	// Face 0: Right (+X)
	{{1, 0, 0}, {1, 1, 0}, {1, 1, 1}, {1, 0, 1}},
	// Face 1: Left (-X)
	{{0, 0, 1}, {0, 1, 1}, {0, 1, 0}, {0, 0, 0}},
	// Face 2: Top (+Y)
	{{0, 1, 1}, {1, 1, 1}, {1, 1, 0}, {0, 1, 0}},
	// Face 3: Bottom (-Y)
	{{0, 0, 0}, {1, 0, 0}, {1, 0, 1}, {0, 0, 1}},
	// Face 4: Front (+Z)
	{{0, 0, 1}, {1, 0, 1}, {1, 1, 1}, {0, 1, 1}},
	// Face 5: Back (-Z)
	{{1, 0, 0}, {0, 0, 0}, {0, 1, 0}, {1, 1, 0}},
}

var VertexAOOffsets = [6][4][][]int{
	// face z- (front)
	{
		{{0, 0, -1}, {-1, 0, 0}, {-1, 0, -1}}, // bottom left corner
		{{0, 0, -1}, {1, 0, 0}, {1, 0, -1}},   // bottom right corner
		{{0, 0, -1}, {-1, 0, 0}, {-1, 0, -1}}, // top left corner
		{{0, 0, -1}, {1, 0, 0}, {1, 0, -1}},   // top right corner
	},
	//	face z+ (back)
	{
		{{0, 0, 1}, {-1, 0, 0}, {-1, 0, 1}},
		{{0, 0, 1}, {1, 0, 0}, {1, 0, 1}},
		{{0, 0, 1}, {-1, 0, 0}, {-1, 0, 1}},
		{{0, 0, 1}, {1, 0, 0}, {1, 0, 1}},
	},
	// Face -X (left)
	{
		{{-1, 0, 0}, {0, 0, -1}, {-1, 0, -1}},
		{{-1, 0, 0}, {0, 0, 1}, {-1, 0, 1}},
		{{-1, 0, 0}, {0, 0, -1}, {-1, 0, -1}},
		{{-1, 0, 0}, {0, 0, 1}, {-1, 0, 1}},
	},
	// Face +X (right)
	{
		{{1, 0, 0}, {0, 0, -1}, {1, 0, -1}},
		{{1, 0, 0}, {0, 0, 1}, {1, 0, 1}},
		{{1, 0, 0}, {0, 0, -1}, {1, 0, -1}},
		{{1, 0, 0}, {0, 0, 1}, {1, 0, 1}},
	},
	// Face -Y (bottom)
	{
		{{0, -1, 0}, {-1, 0, 0}, {-1, -1, 0}},
		{{0, -1, 0}, {1, 0, 0}, {1, -1, 0}},
		{{0, -1, 0}, {-1, 0, 0}, {-1, -1, 0}},
		{{0, -1, 0}, {1, 0, 0}, {1, -1, 0}},
	},
	// Face +Y (top)
	{
		{{0, 1, 0}, {-1, 0, 0}, {-1, 1, 0}},
		{{0, 1, 0}, {1, 0, 0}, {1, 1, 0}},
		{{0, 1, 0}, {-1, 0, 0}, {-1, 1, 0}},
		{{0, 1, 0}, {1, 0, 0}, {1, 1, 0}},
	},
}
