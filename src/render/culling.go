package render

import (
	"go-engine/src/load"
	"go-engine/src/pkg"
	"go-engine/src/world"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func BuildChunkMesh(game *load.Game, chunk *pkg.Chunk) {
	// Clears buffers and specials list
	chunk.Vertices = chunk.Vertices[:0]
	chunk.Indices = chunk.Indices[:0]
	chunk.Colors = chunk.Colors[:0]
	chunk.Normals = chunk.Normals[:0]

	chunk.WaterVertices = chunk.WaterVertices[:0]
	chunk.WaterIndices = chunk.WaterIndices[:0]
	chunk.WaterNormals = chunk.WaterNormals[:0]

	chunk.CloudVertices = chunk.CloudVertices[:0]
	chunk.CloudIndices = chunk.CloudIndices[:0]

	chunk.SpecialVoxels = chunk.SpecialVoxels[:0]

	Nx, Ny, Nz := int(pkg.ChunkSize), int(pkg.WorldHeight), int(pkg.ChunkSize)

	indexOffset := uint16(0)

	// Fixed table of normals by face
	faceNormals := [6][3]float32{
		{1, 0, 0}, {-1, 0, 0}, {0, 1, 0}, {0, -1, 0}, {0, 0, 1}, {0, 0, -1},
	}

	/* Multidimensional Arrays Linearization, docs and extras that may come in handy
	 * https://ic.unicamp.br/~bit/mc102/aulas/aula15.pdf (introdução)
	 * https://felippe.ubi.pt/texts3/contr_av_ppt01p.pdf (pág. 13)
	 * https://www.aussieai.com/book/ch36-linearized-multidimensional-arrays
	 * https://teotl.dev/vischunk/ (may be useful)
	 * (AI was used to help the interpretation of some of those docs)
	 */
	for i := 0; i < Nx*Ny*Nz; i++ {
		pos := pkg.Coords{
			X: i / (Ny * Nz),
			Y: (i / Nz) % Ny,
			Z: i % Nz,
		}

		voxel := chunk.Voxels[pos.X][pos.Y][pos.Z]
		block := world.BlockTypes[voxel.Type]

		// Special cases → not included in the terrain mesh
		switch voxel.Type {
		case "Plant":
			chunk.SpecialVoxels = append(chunk.SpecialVoxels, pkg.SpecialVoxel{
				Position: pos,
				Type:     voxel.Type,
				ModelID:  voxel.ModelID,
			})
			continue
		case "Water":
			addWater(chunk, pos)
			continue
		case "Cloud":
			addCloudFaces(chunk, pos)
			continue
		}

		if !block.IsSolid {
			continue
		}

		c := voxel.Color
		if c.R == 0 && c.G == 0 && c.B == 0 && c.A == 0 {
			// fallback to default color if not set
			c = block.Color
		}

		// Add color per block (RGBA) on a deterministic way (so it can be recalculated when a new chunk is created and still look the same)
		colorModifier := uint8(
			((pos.X*73856093 + pos.Y*19349663) ^
				(pos.Z*83492791 + pos.X*19349663) ^
				(pos.Y*83492791 + pos.Z*73856093)) % 16)

		for face := 0; face < 6; face++ {
			if !shouldDrawFace(chunk, pos, face) {
				continue
			}

			nx, ny, nz := faceNormals[face][0], faceNormals[face][1], faceNormals[face][2]

			/*
				voxelPosition := rl.NewVector3(
					chunkPos.X+float32(pos.X),
					chunkPos.Y+float32(pos.Y),
					chunkPos.Z+float32(pos.Z),
				)

				lightIntensity := calculateLightIntensity(voxelPosition, game.LightPosition)
				voxelColor := applyLighting(c, lightIntensity)
			*/

			//ao := calculateFaceAO(chunk, pos, face)

			for vertice := 0; vertice < 4; vertice++ {
				v := pkg.FaceVertices[face][vertice]
				chunk.Vertices = append(chunk.Vertices,
					float32(pos.X)+v[0],
					float32(pos.Y)+v[1],
					float32(pos.Z)+v[2],
				)
				/*

						colors = append(colors,
							uint8(float32(c.R+colorModifier)*ao),
							uint8(float32(c.G+colorModifier)*ao),
							uint8(float32(c.B+colorModifier)*ao),
							c.A,
						)
					colors = append(colors,
						uint8(c.R), uint8(c.G), uint8(c.B),
						uint8(ao*255.0), // store AO in alpha
					)
				*/

				chunk.Colors = append(chunk.Colors, c.R+colorModifier, c.G+colorModifier, c.B+colorModifier, c.A)

				// add normal for each face vertex
				chunk.Normals = append(chunk.Normals, nx, ny, nz)
			}

			//	Add the two triangles of the face
			chunk.Indices = append(chunk.Indices,
				indexOffset, indexOffset+1, indexOffset+2,
				indexOffset, indexOffset+2, indexOffset+3,
			)
			indexOffset += 4
		}
	}

	//  ---------- Terrain mesh ----------
	if len(chunk.Vertices) > 0 {
		mesh := rl.Mesh{
			VertexCount:   int32(len(chunk.Vertices) / 3),
			TriangleCount: int32(len(chunk.Indices) / 3),
		}

		mesh.Vertices = &chunk.Vertices[0]
		mesh.Indices = &chunk.Indices[0]
		mesh.Colors = &chunk.Colors[0]
		mesh.Normals = &chunk.Normals[0]

		chunk.Model = createChunkModel(mesh, game.Shader)
	}

	// ---------- Water mesh ----------
	if len(chunk.WaterVertices) > 0 {
		mesh := rl.Mesh{
			VertexCount:   int32(len(chunk.WaterVertices) / 3),
			TriangleCount: int32(len(chunk.WaterIndices) / 3),
		}

		mesh.Vertices = &chunk.WaterVertices[0]
		mesh.Indices = &chunk.WaterIndices[0]
		mesh.Normals = &chunk.WaterNormals[0]

		chunk.WaterModel = createChunkModel(mesh, game.Shader)
	}

	// ---------- Clouds mesh ----------
	if len(chunk.CloudVertices) > 0 {
		mesh := rl.Mesh{
			VertexCount:   int32(len(chunk.CloudVertices) / 3),
			TriangleCount: int32(len(chunk.CloudIndices) / 3),
		}

		mesh.Vertices = &chunk.CloudVertices[0]
		mesh.Indices = &chunk.CloudIndices[0]

		//	Clouds are not affected by the fog shader
		chunk.CloudModel = createChunkModel(mesh, rl.Shader{})
	}

	chunk.IsOutdated = false
}

func createChunkModel(mesh rl.Mesh, shader rl.Shader) rl.Model {
	rl.UploadMesh(&mesh, false)

	model := rl.LoadModelFromMesh(mesh)

	material := rl.LoadMaterialDefault()

	if shader.ID != 0 {
		material.Shader = shader
	}

	model.MaterialCount = 1
	model.Materials = &material

	return model
}

func unloadChunkModels(chunk *pkg.Chunk) {

	if chunk.Model.MeshCount > 0 {
		rl.UnloadModel(chunk.Model)
	}

	if chunk.WaterModel.MeshCount > 0 {
		rl.UnloadModel(chunk.WaterModel)
	}

	if chunk.CloudModel.MeshCount > 0 {
		rl.UnloadModel(chunk.CloudModel)
	}
}

func addWater(chunk *pkg.Chunk, pos pkg.Coords) {
	// only add if it's a surface
	isSurface := true
	if pos.Y+1 < pkg.WorldHeight {
		above := chunk.Voxels[pos.X][pos.Y+1][pos.Z]
		if above.Type == "Water" {
			isSurface = false
		}
	}
	if isSurface {
		addWaterQuad(chunk, pos)
	}
}

func addWaterQuad(chunk *pkg.Chunk, pos pkg.Coords) {
	indexOffset := uint16(len(chunk.WaterVertices) / 3)

	// normal pointing upwards
	for i := 0; i < 4; i++ {
		chunk.WaterNormals = append(chunk.WaterNormals, 0, 1, 0)
	}

	y := float32(pos.Y) + 0.50 // slightly below the top of the block

	verts := [4][3]float32{
		{float32(pos.X), y, float32(pos.Z)},
		{float32(pos.X + 1), y, float32(pos.Z)},
		{float32(pos.X + 1), y, float32(pos.Z + 1)},
		{float32(pos.X), y, float32(pos.Z + 1)},
	}

	for _, v := range verts {
		chunk.WaterVertices = append(chunk.WaterVertices,
			v[0], v[1], v[2],
		)
	}

	chunk.WaterIndices = append(chunk.WaterIndices,
		indexOffset, indexOffset+2, indexOffset+1,
		indexOffset, indexOffset+3, indexOffset+2,
	)
}

func addCloudFaces(chunk *pkg.Chunk, pos pkg.Coords) {
	addCloudFace(chunk, pos, true)  // top
	addCloudFace(chunk, pos, false) // bottom
}

func addCloudFace(chunk *pkg.Chunk, pos pkg.Coords, top bool) {
	indexOffset := uint16(len(chunk.CloudVertices) / 3)

	var y float32

	if top {
		y = float32(pos.Y) + 1.0
	} else {
		y = float32(pos.Y)
	}

	// Maintain the same order for both sides.
	verts := [4][3]float32{
		{float32(pos.X), y, float32(pos.Z)},
		{float32(pos.X + 1), y, float32(pos.Z)},
		{float32(pos.X + 1), y, float32(pos.Z + 1)},
		{float32(pos.X), y, float32(pos.Z + 1)},
	}

	for _, v := range verts {
		chunk.CloudVertices = append(chunk.CloudVertices,
			v[0], v[1], v[2],
		)
	}

	if top {
		chunk.CloudIndices = append(chunk.CloudIndices,
			indexOffset, indexOffset+2, indexOffset+1,
			indexOffset, indexOffset+3, indexOffset+2,
		)
	} else {
		// inverted to the lower face
		chunk.CloudIndices = append(chunk.CloudIndices,
			indexOffset, indexOffset+1, indexOffset+2,
			indexOffset, indexOffset+2, indexOffset+3,
		)
	}
}

func shouldDrawFace(chunk *pkg.Chunk, pos pkg.Coords, faceIndex int) bool {
	direction := pkg.FaceDirections[faceIndex]
	maxSize := int(pkg.ChunkSize - 1)
	maxHeight := int(pkg.WorldHeight - 1)

	// Calculates the new coordinates based on the face direction
	nx := pos.X + int(direction.X)
	ny := pos.Y + int(direction.Y)
	nz := pos.Z + int(direction.Z)

	// Case 1: Checks if the new coordinates are within the chunk bounds and does not render internal voxels
	if nx >= 0 && nx <= maxSize &&
		ny >= 0 && ny <= maxHeight &&
		nz >= 0 && nz <= maxSize {

		if chunk == nil {
			return false
		}

		voxelType := chunk.Voxels[nx][ny][nz].Type

		if block, ok := world.BlockTypes[voxelType]; ok {
			return !block.IsSolid
		}
		return false // unknown voxel → draw face
	}

	// Case 2: vertical faces (do not have chunk neighbors)
	if faceIndex == 2 {
		// +Y (top)
		return true
	}
	if faceIndex == 3 {
		// -Y (bottom)
		return false
	}

	// Case 3: Outside chunk boundries → depends on the neighbor
	var neighborIdx int
	switch faceIndex {
	case 0:
		neighborIdx = 0 // +X
	case 1:
		neighborIdx = 1 // -X
	case 4:
		neighborIdx = 2 // +Z
	case 5:
		neighborIdx = 3 // -Z
	}

	neighbor := chunk.Neighbors[neighborIdx]
	if neighbor == nil {
		return false // no neighbor → exposed face (still I'm not rendering it)
	}

	// Adjusts coordinates relative to the neighbor.
	nx = (nx + int(pkg.ChunkSize)) % int(pkg.ChunkSize)
	nz = (nz + int(pkg.ChunkSize)) % int(pkg.ChunkSize)

	if ny < 0 || ny > maxHeight {
		return true
	}

	voxelType := neighbor.Voxels[nx][ny][nz].Type
	if block, ok := world.BlockTypes[voxelType]; ok {
		return !block.IsSolid
	}
	return false
}
