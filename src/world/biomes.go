package world

import (
	"math"
	"math/rand"

	"go-engine/src/pkg"

	"github.com/aquilax/go-perlin"
	rl "github.com/gen2brain/raylib-go/raylib"
)

var BiomeTypes = map[string]*pkg.BiomeProperties{
	"Meadow": {
		Modifier:         meadowModifier,
		SurfaceBlock:     "Grass",
		UndergroundBlock: "Dirt",
		PlantTypes:       []int32{0, 1, 3, 4},
		TreeTypes: []string{
			"F=F[FA(3)L][FA(3)L][FA(3)L]A(3)",
			"F=F[F+A(5)L][−A(5)L][/A(4)L][\\A(4)L]",
			"F=F[A(3)L]F[-A(2)L]F[/A(2)L]F[+A(1)L]",
			"F=FFF[FA(2)L][FA(3)L][FA(4)L]",
		},
		TreeDensity: 0.2,
		GrassColor:  rl.NewColor(72, 174, 34, 255), // verde vivo
		LeavesColor: rl.NewColor(73, 129, 49, 255), // verde escuro
	},
	"Birchwood": {
		Modifier:         birchwoodModifier,
		SurfaceBlock:     "Grass",
		UndergroundBlock: "Dirt",
		PlantTypes:       []int32{0, 1, 5, 6, 7},
		TreeTypes: []string{
			"F=F[-A(2)L]F[/A(2)L]F[+A(1)L]",
			"F=FF[FA(2)L][FA(3)L][FA(4)L]",
			"F=F[+A(2)L][-A(2)L]FL",
		},
		TreeDensity: 0.4,
		GrassColor:  rl.NewColor(69, 143, 72, 255),
		LeavesColor: rl.NewColor(53, 105, 56, 255),
	},
	"Savanna": {
		Modifier:         savannaModifier,
		SurfaceBlock:     "Grass",
		UndergroundBlock: "Dirt",
		PlantTypes:       []int32{0, 1, 2, 3},
		TreeTypes: []string{
			"F=F[FA(3)L][FA(3)L][FA(3)L]A(3)",
			"F=F[+FA(4)L][-A(4)L][\\A(4)L]",
		},
		TreeDensity: 0.2,
		GrassColor:  rl.NewColor(134, 157, 36, 255),
		LeavesColor: rl.NewColor(102, 119, 23, 255),
	},
	"Desert": {
		Modifier:         desertModifier,
		SurfaceBlock:     "Sand",
		UndergroundBlock: "Sand",
		PlantTypes:       nil,
	},
}

type Cell struct {
	X, Z  int
	Biome pkg.BiomeProperties
}

type WorleyNoise struct {
	Seed     int64
	CellSize int
}

func NewWorleyNoise(seed int64, cellSize int) *WorleyNoise {
	return &WorleyNoise{Seed: seed, CellSize: cellSize}
}

func (w *WorleyNoise) featurePoint(cellX, cellZ int) (float64, float64) {
	h := int64(cellX*73856093^cellZ*19349663) ^ w.Seed
	r := rand.New(rand.NewSource(h))
	fx := float64(cellX*w.CellSize) + r.Float64()*float64(w.CellSize)
	fz := float64(cellZ*w.CellSize) + r.Float64()*float64(w.CellSize)
	return fx, fz
}

// Returns the nearest cell
func (w *WorleyNoise) Evaluate(x, z int) (float64, float64, int, int, int, int) {
	cellX := int(math.Floor(float64(x) / float64(w.CellSize)))
	cellZ := int(math.Floor(float64(z) / float64(w.CellSize)))

	d1, d2 := math.MaxFloat64, math.MaxFloat64
	nearestX, nearestZ := cellX, cellZ
	secondX, secondZ := cellX, cellZ

	for dx := -1; dx <= 1; dx++ {
		for dz := -1; dz <= 1; dz++ {
			fx, fz := w.featurePoint(cellX+dx, cellZ+dz)
			d := math.Hypot(float64(x)-fx, float64(z)-fz)
			if d < d1 {
				d2, secondX, secondZ = d1, nearestX, nearestZ
				d1, nearestX, nearestZ = d, cellX+dx, cellZ+dz
			} else if d < d2 {
				d2, secondX, secondZ = d, cellX+dx, cellZ+dz
			}
		}
	}
	return d1, d2, nearestX, nearestZ, secondX, secondZ
}

type BiomeSelector struct {
	Seed     int64
	CellSize int
}

func NewBiomeSelector(seed int64, cellSize int) *BiomeSelector {
	return &BiomeSelector{Seed: seed, CellSize: cellSize}
}

func (b *BiomeSelector) biomeForCell(cellX, cellZ int) *pkg.BiomeProperties {
	h := int64(cellX*83492791^cellZ*1234567) ^ b.Seed
	r := rand.New(rand.NewSource(h))

	biomeKeys := []string{"Meadow", "Birchwood", "Savanna", "Desert"}
	key := biomeKeys[r.Intn(len(biomeKeys))]
	return BiomeTypes[key]
}

func globalHeight(gx, gz int, p *perlin.Perlin) float64 {
	freq := 0.002
	n := p.Noise2D(float64(gx)*freq, float64(gz)*freq) // [-1,1]
	return (n + 1.0) / 2.0                             // normalizes to [0,1]
}

func meadowModifier(gx, gz int, p2, p3 *perlin.Perlin) float64 {
	n2 := p2.Noise2D(float64(gx)*0.09, float64(gz)*0.09)
	n3 := p3.Noise2D(float64(gx)*0.01, float64(gz)*0.01)

	// Narrower mountains
	n2 = math.Pow(math.Abs(n2), 4)

	// Combines both (can be done with sum, average, or another function)
	return (n2*1.2 + n3*0.3) * float64(pkg.WorldHeight/2) // Normalizes the noise value to [0, worldHeight / 2]
}

func desertModifier(gx, gz int, p2, p3 *perlin.Perlin) float64 {
	n := p2.Noise2D(float64(gx)*0.003, float64(gz)*0.003)
	base := n * 0.7 * float64(pkg.WorldHeight/3)
	return base
}

/*
func mauntainModifier(gx, gz int, p2, p3 *perlin.Perlin) float64 {
	// Ruído base
	n := p2.Noise2D(float64(gx)*0.01, float64(gz)*0.01)

	// Normaliza para positivo
	n = (n + 1) / 2.0

	// Cria formato convexo: topo largo, base fina
	// sqrt(n) dá mais área no topo, pow(n, >1) afina a base
	shaped := math.Sqrt(n) * math.Pow(n, 2.0)

	// Escala para altura
	return shaped * float64(pkg.WorldHeight/3)
}
*/

func birchwoodModifier(gx, gz int, p2, p3 *perlin.Perlin) float64 {
	n2 := p2.Noise2D(float64(gx)*0.007, float64(gz)*0.007)
	n3 := p3.Noise2D(float64(gx)*0.01, float64(gz)*0.01)

	return (n2*0.6 + n3*0.3) * float64(pkg.WorldHeight/3)
}

func savannaModifier(gx, gz int, p2, p3 *perlin.Perlin) float64 {
	n2 := p2.Noise2D(float64(gx)*0.007, float64(gz)*0.007)
	n3 := p3.Noise2D(float64(gx)*0.02, float64(gz)*0.02)

	return (n2*0.4 + n3*0.3) * float64(pkg.WorldHeight/3)
}

/*
func mountainModifier(gx, gz int, p2, p3 *perlin.Perlin) float64 {
	// Defina um centro fixo ou derivado do Worley
	centerX, centerZ := 0, 0 // pode ser ajustado dinamicamente
	dx := float64(gx - centerX)
	dz := float64(gz - centerZ)
	dist := math.Sqrt(dx*dx + dz*dz)

	// Raio da montanha
	radius := 80.0
	height := 1.0 - (dist / radius)

	if height < 0 {
		height = 0
	}

	// Elevação máxima
	peak := math.Pow(height, 2.0) * float64(pkg.WorldHeight-8)

	// Combina com globalHeight para suavizar entorno
	//base := globalHeight(gx, gz, p2) * float64(pkg.WorldHeight/4)

	return peak
}
*/
