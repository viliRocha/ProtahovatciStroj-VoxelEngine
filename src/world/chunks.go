package world

import (
	"math"
	"runtime"
	"sync"

	"go-engine/src/pkg"

	"github.com/aquilax/go-perlin"
	rl "github.com/gen2brain/raylib-go/raylib"
)

const MaxChunksPerFrame = 2

type PendingWrite struct {
	Pos   [3]int
	Voxel pkg.VoxelData
}

type ChunkCache struct {
	Active        map[pkg.Coords]*pkg.Chunk      // chunks that are loaded, meshed, and ready to render.
	Loading       map[pkg.Coords]bool            // avoid duplication when enqueuing chunks
	PlantsCache   map[pkg.Coords][]pkg.PlantData // persistent plants by chunk
	TreesCache    map[pkg.Coords][]pkg.TreeData
	PendingVoxels map[pkg.Coords][]PendingWrite // queue of voxel modifications that haven’t yet been applied to the chunk
	CacheMutex    sync.RWMutex                  // Synchronization primitive to protect concurrent access to the cache maps. Multiple goroutines may read chunk data in parallel, but writes (adding/removing chunks, applying voxel changes) must be exclusive
}

func NewChunkCache() *ChunkCache {
	// Creates a hash map to store voxel data
	return &ChunkCache{
		Active:        make(map[pkg.Coords]*pkg.Chunk),
		Loading:       make(map[pkg.Coords]bool),
		PlantsCache:   make(map[pkg.Coords][]pkg.PlantData),
		TreesCache:    make(map[pkg.Coords][]pkg.TreeData),
		PendingVoxels: make(map[pkg.Coords][]PendingWrite),
	}
}

func ToChunkCoord(pos rl.Vector3) pkg.Coords {
	return pkg.Coords{
		X: int(math.Floor(float64(pos.X) / float64(pkg.ChunkSize))),
		Y: 0, //	Fixed Height
		Z: int(math.Floor(float64(pos.Z) / float64(pkg.ChunkSize))),
	}
}

func (cc *ChunkCache) GetChunk(worley *WorleyNoise, biomeSel *BiomeSelector, position rl.Vector3, p1, p2, p3 *perlin.Perlin) *pkg.Chunk {
	coord := ToChunkCoord(position)

	cc.CacheMutex.RLock()
	_, exists := cc.Active[coord]
	oldPlants, hasPlants := cc.PlantsCache[coord]
	oldTrees, hasTrees := cc.TreesCache[coord]
	cc.CacheMutex.RUnlock()

	var newChunk *pkg.Chunk

	if exists {
		newChunk = GenerateChunk(worley, biomeSel, position, p1, p2, p3, cc, oldPlants, true, oldTrees, true)

		newChunk.Coord = coord
	} else {
		// First time the chunk is generated
		// If there are saved plants, reuse them; if not, create new ones
		if (hasPlants && len(oldPlants) > 0) || (hasTrees && len(oldTrees) > 0) {
			newChunk = GenerateChunk(worley, biomeSel, position, p1, p2, p3, cc, oldPlants, true, oldTrees, true)

			newChunk.Coord = coord
		} else {
			newChunk = GenerateChunk(worley, biomeSel, position, p1, p2, p3, cc, nil, false, nil, false)

			newChunk.Coord = coord
		}
	}

	newChunk.IsOutdated = true // reset flag after reconstruction --> ensures that the mesh is rebuilt and the plant voxels are reapplied

	// Update caches
	cc.CacheMutex.Lock()
	cc.Active[coord] = newChunk

	// applies pending issues after the core terrain generation so tree voxels aren’t overwritten.
	if writes, ok := cc.PendingVoxels[coord]; ok {
		for _, w := range writes {
			if w.Pos[0] >= 0 && w.Pos[0] < pkg.ChunkSize &&
				w.Pos[1] >= 0 && w.Pos[1] < pkg.WorldHeight &&
				w.Pos[2] >= 0 && w.Pos[2] < pkg.ChunkSize {
				newChunk.Voxels[w.Pos[0]][w.Pos[1]][w.Pos[2]] = w.Voxel
				//cc.MarkOutdated(coord)
			}
		}
		newChunk.IsOutdated = true
		delete(cc.PendingVoxels, coord)
	}

	if len(newChunk.Plants) > 0 {
		// Always save when generating/rebuilding
		cc.PlantsCache[coord] = newChunk.Plants
	} else if !hasPlants {
		// Ensures key exists even if empty to avoid future nil checks
		cc.PlantsCache[coord] = nil
	}

	if len(newChunk.Trees) > 0 {
		cc.TreesCache[coord] = newChunk.Trees
	} else if !hasTrees {
		cc.TreesCache[coord] = nil
	}

	cc.CacheMutex.Unlock()

	return newChunk
}

func (cc *ChunkCache) CleanUp(playerPosition rl.Vector3) {
	cc.CacheMutex.Lock()
	defer cc.CacheMutex.Unlock()

	playerCoord := ToChunkCoord(playerPosition)
	chDist := int(pkg.ChunkDistance)

	for coord := range cc.Active {
		if Abs(coord.X-playerCoord.X) > chDist ||
			Abs(coord.Z-playerCoord.Z) > chDist {
			delete(cc.Active, coord)
			// DO NOT delete cc.PlantsCache[coord] — plants remain stored
		}
	}
}

// global channel for ready-made chunks
var readyChunks = make(chan *pkg.Chunk, 200)

var chunkRequests = make(chan pkg.Coords, 200)

// Initializes the worker pool once at startup
func StartChunkWorkers(worley *WorleyNoise, biomeSel *BiomeSelector, cache *ChunkCache, p1, p2, p3 *perlin.Perlin) {
	for i := 0; i < runtime.NumCPU(); i++ {
		go func() {
			for coord := range chunkRequests {
				pos := rl.NewVector3(float32(coord.X*pkg.ChunkSize), 0, float32(coord.Z*pkg.ChunkSize))
				chunk := cache.GetChunk(worley, biomeSel, pos, p1, p2, p3)
				readyChunks <- chunk
			}
		}()
	}
}

func ManageChunks(worley *WorleyNoise, biomeSel *BiomeSelector, playerPosition rl.Vector3, chunkCache *ChunkCache, p1, p2, p3 *perlin.Perlin) {
	playerCoord := ToChunkCoord(playerPosition)

	// Counter to limit how many chunks are enqueued
	chunksQueued := 0

	// Send the chunk positions to be loaded
	for x := playerCoord.X - pkg.ChunkDistance; x <= playerCoord.X+pkg.ChunkDistance; x++ {
		for z := playerCoord.Z - pkg.ChunkDistance; z <= playerCoord.Z+pkg.ChunkDistance; z++ {
			coord := pkg.Coords{X: x, Y: 0, Z: z}

			chunkCache.CacheMutex.RLock()
			_, exists := chunkCache.Active[coord]
			loading := chunkCache.Loading[coord]
			chunkCache.CacheMutex.RUnlock()

			if !exists && !loading {
				//	Mark chunk as beeing loaded that way it is lined up in the worker pool only once
				chunkCache.CacheMutex.Lock()
				chunkCache.Loading[coord] = true
				chunkCache.CacheMutex.Unlock()

				chunkRequests <- coord
				chunksQueued++
				if chunksQueued >= MaxChunksPerFrame {
					break
				}
			}
		}
	}

	for i := 0; i < MaxChunksPerFrame; i++ {
		select {
		case pc := <-readyChunks:
			chunkCache.CacheMutex.Lock()
			chunkCache.Active[pc.Coord] = pc
			delete(chunkCache.Loading, pc.Coord) // remove da lista de carregamento
			chunkCache.CacheMutex.Unlock()
		default:
			return
		}
	}

	// Updates neighbors
	chunkCache.CacheMutex.Lock()
	// Ensures that each chunk on the chunkCache.chunks map has up-to-date references to its neighboring chunks on X and Z directions
	for coord, chunk := range chunkCache.Active {
		for i, direction := range pkg.HorizontalDirections {
			neighborCoord := pkg.Coords{
				X: coord.X + int(direction.X),
				Y: 0,
				Z: coord.Z + int(direction.Z),
			}
			if neighbor, exists := chunkCache.Active[neighborCoord]; exists {
				if chunk.Neighbors[i] != neighbor {
					// Update reference
					chunk.Neighbors[i] = neighbor
					// Mark both as outdated to rebuild the mesh (exposed/hidden faces are recalculated, fixing the 'holes' in the terrain)
					chunk.IsOutdated = true
					neighbor.IsOutdated = true
				}
			} else {
				if chunk.Neighbors[i] != nil {
					// Neighbor was removed → mark chunk as outdated
					chunk.Neighbors[i] = nil
					chunk.IsOutdated = true
				}
			}
		}
	}
	chunkCache.CacheMutex.Unlock()

	// Remove chunks outside the range
	chunkCache.CleanUp(playerPosition)
}

func setVoxelGlobal(chunkCache *ChunkCache, globalPos rl.Vector3, voxel pkg.VoxelData) {
	coord := ToChunkCoord(globalPos)

	// math.Floor prevents inconsistent rounding that throws blocks into the wrong chunk
	localX := int(math.Floor(float64(globalPos.X))) - coord.X*pkg.ChunkSize
	localY := int(math.Floor(float64(globalPos.Y)))
	localZ := int(math.Floor(float64(globalPos.Z))) - coord.Z*pkg.ChunkSize

	// If the voxel is outside the local boundaries → just skip it
	if localX < 0 || localX >= pkg.ChunkSize ||
		localY < 0 || localY >= pkg.WorldHeight ||
		localZ < 0 || localZ >= pkg.ChunkSize {

		return
	}

	// Protect map reading
	chunkCache.CacheMutex.RLock()
	chunk := chunkCache.Active[coord]
	chunkCache.CacheMutex.RUnlock()

	if chunk == nil {
		chunkCache.CacheMutex.Lock()
		chunkCache.PendingVoxels[coord] = append(chunkCache.PendingVoxels[coord], PendingWrite{
			Pos:   [3]int{localX, localY, localZ},
			Voxel: voxel,
		})
		chunkCache.CacheMutex.Unlock()
		return
	} else {
		// if multiple goroutines write to the same chunk, consider a mutex per chunk
		chunkCache.CacheMutex.Lock()
		chunk.Voxels[localX][localY][localZ] = voxel
		chunk.IsOutdated = true
		chunkCache.CacheMutex.Unlock()
	}
}

// Function to calculate the absolute value
// https://stackoverflow.com/questions/664852/which-is-the-fastest-way-to-get-the-absolute-value-of-a-number#2074403
func Abs(x int) int {
	mask := x >> 31
	return (x + mask) ^ mask
}
