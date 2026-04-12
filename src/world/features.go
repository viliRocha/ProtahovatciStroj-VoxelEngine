package world

import (
	"math"
	"math/rand"
	"strconv"
	"strings"

	"go-engine/src/pkg"

	"github.com/aquilax/go-perlin"
	rl "github.com/gen2brain/raylib-go/raylib"
)

// see https://github.com/adct-the-experimenter/Raylib_VoxelEngine/blob/main/blockfacehelper.c for inspiration
type BlockProperties struct {
	Color     rl.Color
	IsSolid   bool
	IsVisible bool
}

var BlockTypes = map[string]BlockProperties{
	"Grass": {
		Color:     rl.NewColor(72, 174, 34, 255), // Default green
		IsSolid:   true,
		IsVisible: true,
	},
	"Dirt": {
		Color:     rl.Brown,
		IsSolid:   true,
		IsVisible: true,
	},
	"Sand": {
		Color:     rl.NewColor(236, 221, 178, 255), //	Beige
		IsSolid:   true,
		IsVisible: true,
	},
	"Stone": {
		Color:     rl.Gray,
		IsSolid:   true,
		IsVisible: true,
	},
	"OakWood": {
		Color:     rl.NewColor(126, 90, 57, 255), // Light brown
		IsSolid:   true,
		IsVisible: true,
	},
	"Leaves": {
		Color:     rl.NewColor(73, 129, 49, 255), //	Default dark green
		IsSolid:   true,
		IsVisible: true,
	},
	"Plant": {
		Color:     rl.Red,
		IsSolid:   false,
		IsVisible: true,
	},
	"Water": {
		Color:     rl.NewColor(0, 0, 255, 110), // Transparent blue
		IsSolid:   false,
		IsVisible: true,
	},
	"Cloud": {
		Color:     rl.NewColor(249, 248, 248, 160),
		IsSolid:   false,
		IsVisible: true,
	},
	"Air": {
		Color:     rl.NewColor(0, 0, 0, 0), // Transparent
		IsSolid:   false,
		IsVisible: false,
	},
}

type TurtleState struct {
	Position  rl.Vector3
	Direction rl.Vector3
}

// Generate vegetation at random surface positions
func generatePlants(chunk *pkg.Chunk, chunkPos rl.Vector3, oldPlants []pkg.PlantData, reusePlants bool) {
	waterLevel := int(float64(pkg.WorldHeight) * pkg.WaterLevelFraction)

	if reusePlants && oldPlants != nil {
		for _, plant := range oldPlants {
			localX := int(plant.Position.X) - int(chunkPos.X)
			localY := int(plant.Position.Y)
			localZ := int(plant.Position.Z) - int(chunkPos.Z)

			chunk.Voxels[localX][localY][localZ] = pkg.VoxelData{
				Type:  "Plant",
				Model: pkg.PlantModels[plant.ModelID],
			}

			chunk.Plants = append(chunk.Plants, plant)
		}
		return
	}
	plantCount := pkg.ChunkSize / 4

	for i := 0; i < plantCount; i++ {
		x := rand.Intn(pkg.ChunkSize)
		z := rand.Intn(pkg.ChunkSize)

		biome := chunk.BiomeMap[x][z]

		height := chunk.HeightMap[x][z]

		if height < 0 || height+1 >= pkg.WorldHeight {
			continue // ignore invalid positions
		}

		// Ensure plants are only placed above layer 13 (water)
		if chunk.Voxels[x][height][z].Type == "Grass" &&
			chunk.Voxels[x][height+1][z].Type == "Air" &&
			height > waterLevel {

			if biome.PlantTypes != nil && len(biome.PlantTypes) > 0 {
				// Choose a random index from the list of PlantTypes.
				idx := rand.Intn(len(biome.PlantTypes))
				randomModel := biome.PlantTypes[idx]

				// plant_0: short grass - all biomes
				// plant_1: taller grass - all biomes
				// plant_2: orange flower -  savanna
				// plant_3: red flower - meadow and savanna
				// plant_4: light purple flower - meadow
				// plant_5: yellow flower - birchwood biome
				// plant_6: blooming blue grass - birchwood biome
				// plant_7: mushroom (I know they thecnically aren't plants 🤓) - birchwood biome
				chunk.Voxels[x][height+1][z] = pkg.VoxelData{
					Type:  "Plant",
					Model: pkg.PlantModels[randomModel],
				}
				plantPos := rl.NewVector3(chunkPos.X+float32(x), float32(height+1), chunkPos.Z+float32(z))
				chunk.Plants = append(chunk.Plants, pkg.PlantData{
					Position: plantPos,
					ModelID:  randomModel,
				})
			}
		}
	}
}

func parseLSystemRule(ruleStr string) map[rune]string {
	rules := make(map[rune]string)
	parts := strings.Split(ruleStr, "=")
	if len(parts) == 2 {
		rules[rune(parts[0][0])] = parts[1]
	}
	return rules
}

func applyLSystem(axiom string, rules map[rune]string, iterations int) string {
	result := axiom
	for range iterations {
		var builder strings.Builder

		// reserve approximate space to avoid relocations
		builder.Grow(len(result) * 2)

		for _, char := range result {
			if replacement, ok := rules[char]; ok {
				builder.WriteString(replacement)
			} else {
				builder.WriteRune(char)
			}
		}

		// converts the builder's content into a string
		result = builder.String()
	}
	return result
}

func placeTree(chunkCache *ChunkCache, position rl.Vector3, treeStructure string, biome pkg.BiomeProperties) {
	stack := []TurtleState{}
	currentPos := position
	direction := rl.Vector3{0, 1, 0} //	Initial direction (upwards)

	// Variables to Extend Angular Spacing
	angleIncrement := 45.0 * (1.0 + rand.Float64()*0.2) // Initial separation angle between branches
	currentAngle := 0.0

	for i := 0; i < len(treeStructure); i++ {
		char := treeStructure[i]

		switch char {
		case 'F': // Create wood blocks for tree tunks

			if currentPos.Y >= 0 && int(currentPos.Y) < pkg.WorldHeight {
				setVoxelGlobal(chunkCache, currentPos, pkg.VoxelData{Type: "OakWood"})
			}

			// Moving in the current direction
			currentPos = rl.Vector3{
				currentPos.X + direction.X,
				currentPos.Y + direction.Y,
				currentPos.Z + direction.Z,
			}

			//	if the structure left the world, interrupt
			if int(currentPos.Y) < 0 || int(currentPos.Y) >= pkg.WorldHeight {
				return
			}

		case '+': // Turn right (around the Y-axis)
			direction = rl.Vector3{1, 0, 0}

		case '-': // Turn left (around the Y-axis)
			direction = rl.Vector3{-1, 0, 0}

		case '/': // Move forward (positive Z-axis)
			direction = rl.Vector3{0, 0, 1}

		case '\\': //  Move back (negative Z-axis)
			direction = rl.Vector3{0, 0, -1}

		case '[': // Save the current position and direction
			stack = append(stack, TurtleState{Position: currentPos, Direction: direction})

		case ']': // Restore the last saved position
			state := stack[len(stack)-1]
			stack = stack[:len(stack)-1] //	Removes the last item from the stack
			currentPos = state.Position
			direction = state.Direction

		case 'A': // Create simple branch (diagonal movement without rotation)
			// Checks if the next character is '('
			if i+1 < len(treeStructure) && treeStructure[i+1] == '(' {
				// Reads the numeric value in parentheses
				end := i + 2
				for end < len(treeStructure) && treeStructure[end] != ')' {
					end++
				}

				if end < len(treeStructure) && treeStructure[end] == ')' {
					// Extracts the number as string and converts to integer
					branchDepthStr := treeStructure[i+2 : end]
					branchDepth, err := strconv.Atoi(branchDepthStr)
					if err == nil {
						// Handles the branch with the defined depth
						for range branchDepth {
							// Calculates new direction based on current angle
							radians := currentAngle * (math.Pi / 180.0) // Converts to radians
							// Adjusts X and Z based on angle and keeps Y rising
							newDir := rl.Vector3{
								float32(math.Cos(radians)),
								1.0,
								float32(math.Sin(radians)),
							}

							newPos := rl.Vector3{
								currentPos.X + newDir.X,
								currentPos.Y + newDir.Y,
								currentPos.Z + newDir.Z,
							}

							setVoxelGlobal(chunkCache, newPos, pkg.VoxelData{Type: "OakWood"})

							currentPos = newPos
							// Increases the angle to open the next branch
							currentAngle += angleIncrement

						}
					}
					// Updates the index to continue after
					i = end
				}
			}

		case 'L':
			numLeaves := 32
			radius := 2.0 // Radius of the leaf circle

			for i := range numLeaves {
				angle := float64(i) * (2 * math.Pi / float64(numLeaves)) // Calculate leaf cluster angle

				// "l": leafPos
				lx := currentPos.X + float32(radius*math.Cos(angle))
				ly := currentPos.Y + float32(rand.Intn(2)) // vertical variation
				lz := currentPos.Z + float32(radius*math.Sin(angle))

				if int(ly) >= 0 && int(ly) < pkg.WorldHeight {
					leafPos := rl.Vector3{lx, ly, lz}
					setVoxelGlobal(chunkCache, leafPos, pkg.VoxelData{
						Type:  "Leaves",
						Color: biome.LeavesColor,
					})
				}
			}
		}
	}
}

func generateTrees(chunk *pkg.Chunk, chunkCache *ChunkCache, chunkOrigin rl.Vector3, oldTrees []pkg.TreeData, reuseTrees bool) {
	waterLevel := int(float64(pkg.WorldHeight) * pkg.WaterLevelFraction)

	if reuseTrees && oldTrees != nil {
		for _, tree := range oldTrees {
			x := int(tree.Position.X) - int(chunkOrigin.X)
			z := int(tree.Position.Z) - int(chunkOrigin.Z)
			biome := chunk.BiomeMap[x][z]

			placeTree(chunkCache, tree.Position, tree.StructureStr, biome)
			chunk.Trees = append(chunk.Trees, tree)
		}
		return
	}

	//	Max amount of trees in the chunk
	treeCount := pkg.ChunkSize / 4

	for range treeCount {
		x := rand.Intn(pkg.ChunkSize)
		z := rand.Intn(pkg.ChunkSize)

		biome := chunk.BiomeMap[x][z]

		// densidade do bioma (0.0 a 1.0)
		density := biome.TreeDensity

		// chance de gerar árvore nesta tentativa
		if rand.Float32() > density {
			continue
		}

		// find the surface
		height := chunk.HeightMap[x][z]

		if height < 0 || height >= pkg.WorldHeight {
			continue
		}

		if chunk.Voxels[x][height][z].Type != "Grass" {
			continue
		}

		// Make sure the surface is valid and not in the water
		if height < waterLevel {
			continue
		}

		// It only generates a tree if the biome has defined types
		if len(biome.TreeTypes) == 0 {
			continue
		}

		// Choose a tree from the biome
		treeType := biome.TreeTypes[rand.Intn(len(biome.TreeTypes))]

		// Builds the tree based on the type
		rules := parseLSystemRule(treeType)
		treeStructure := applyLSystem("F", rules, 2)

		treePosGlobal := rl.NewVector3(
			chunkOrigin.X+float32(x),
			chunkOrigin.Y+float32(height+1),
			chunkOrigin.Z+float32(z),
		)

		// Build the tree with the generated structure
		placeTree(chunkCache, treePosGlobal, treeStructure, biome)

		chunk.Trees = append(chunk.Trees, pkg.TreeData{
			Position:     treePosGlobal,
			StructureStr: treeStructure,
		})
	}
}

func genClouds(chunk *pkg.Chunk, position rl.Vector3, x, z int, p *perlin.Perlin) {
	threshold := 0.05 // Intensity of the cloud formation
	cloudFrequency := 0.05

	// Global coordinates
	globalX := int(position.X) + x
	globalZ := int(position.Z) + z

	noise := p.Noise2D(float64(globalX)*cloudFrequency, float64(globalZ)*cloudFrequency)

	if noise > threshold {
		if chunk.Voxels[x][pkg.CloudHeight][z].Type == "Air" {
			chunk.Voxels[x][pkg.CloudHeight][z] = pkg.VoxelData{Type: "Cloud"}
		} else {
			return // meets trees or mauntain
		}
	}
}

func genWaterFormations(chunk *pkg.Chunk, x, z int) {
	waterLevel := int(float64(pkg.WorldHeight) * pkg.WaterLevelFraction)

	topWaterY := waterLevel

	for y := waterLevel; y >= 0; y-- {
		if chunk.Voxels[x][y][z].Type == "Air" {
			//	Water shouldn't replace solid blocks (go through them)
			chunk.Voxels[x][y][z] = pkg.VoxelData{Type: "Water"}
		} else {
			break // meets ground → stops
		}
	}

	genSandFormations(chunk, topWaterY, x, z)
}

func genSandFormations(chunk *pkg.Chunk, ylevel, x, z int) {
	// Creates a Perlin Noise generator
	perlinNoise := perlin.NewPerlin(2, 2, 4, 0)

	// Only generates sand near the water's surface
	for y := ylevel - 2; y <= ylevel+1; y++ {
		// Checks adjacent blocks for generating sand
		for dy := -3; dy <= 1; dy++ {
			for dx := -3; dx <= 3; dx++ {
				adjX := x + dx
				adjZ := z + dy

				//  adjX and adjZ >= 0 ensures that it does not access negative indices.
				//  adjX and adjZ < pkg.ChunkSize ensures that it does not exceed the chunk size.
				if adjX < 0 || adjX >= pkg.ChunkSize || adjZ < 0 || adjZ >= pkg.ChunkSize {
					continue
				}
				// Generate a Perlin Noise value
				noiseValue := perlinNoise.Noise2D(float64(adjX)/8, float64(adjZ)/8)
				voxel := chunk.Voxels[adjX][y][adjZ].Type

				// Replaces dirt and grass with sand
				above := chunk.Voxels[adjX][y+1][adjZ]

				if (voxel == "Grass" || voxel == "Dirt") && noiseValue > 0.32 && (above.Type == "Water" || above.Type == "Air") {
					chunk.Voxels[adjX][y][adjZ] = pkg.VoxelData{Type: "Sand"}
				}
			}
		}
	}
}
