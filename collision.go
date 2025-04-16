package main

import rl "github.com/gen2brain/raylib-go/raylib"

type Collides interface {
	Position() rl.Vector2
	Rearrange(other Collides)
	CheckCollision(other Collides) bool
}

type CollisionPair struct {
	first, second Collides
}

type CollisionSpaceGrid [][][]int

type CollisionSpace struct {
	CollisionSpaceGrid
	Cols, Rows            int
	CellWidth, CellHeight int
}

func NewCollisionSpace(pw, ph, cols, rows int) (out CollisionSpace) {
	out.CollisionSpaceGrid = make(CollisionSpaceGrid, rows+2)
	for i := 0; i < rows+2; i++ {
		out.CollisionSpaceGrid[i] = make([][]int, cols+2)
	}
	out.Cols = cols
	out.Rows = rows
	out.CellWidth = pw / cols
	out.CellHeight = ph / rows
	return
}

func (cs *CollisionSpace) UpdateCells(bodies []Collides) {
	// Clear the grid
	for y := range cs.CollisionSpaceGrid {
		for x := range cs.CollisionSpaceGrid[y] {
			cs.CollisionSpaceGrid[y][x] = nil
		}
	}

	// Place each body in the appropriate grid cell
	for i, body := range bodies {
		pos := body.Position()

		// Calculate grid coordinates, ensuring they stay within bounds
		gx := int(pos.X / float32(cs.CellWidth))
		gy := int(pos.Y / float32(cs.CellHeight))

		// Apply bounds checking
		if gx < 0 {
			gx = 0
		} else if gx >= cs.Cols {
			gx = cs.Cols - 1
		}

		if gy < 0 {
			gy = 0
		} else if gy >= cs.Rows {
			gy = cs.Rows - 1
		}

		// Add 1 for the border cells
		gx += 1
		gy += 1

		// Add the body to the grid
		cs.CollisionSpaceGrid[gy][gx] = append(cs.CollisionSpaceGrid[gy][gx], i)
	}
}

func (cs *CollisionSpace) RearrangeBodies(maxIters int, collidables []Collides, each func()) {
	anyColliding := true
	for iters := 0; anyColliding && iters < maxIters; iters++ {
		collisions := []CollisionPair{}

		each()

		for y := 1; y < cs.Rows; y++ {
			for x := 1; x < cs.Cols; x++ {
				central := cs.CollisionSpaceGrid[y][x]

				if len(central) == 0 {
					continue
				}

				for yy := y - 1; yy < y+2; yy++ {
					for xx := x - 1; xx < x+2; xx++ {
						if yy < 0 || yy >= len(cs.CollisionSpaceGrid) ||
							xx < 0 || xx >= len(cs.CollisionSpaceGrid[yy]) {
							continue
						}

						around := cs.CollisionSpaceGrid[yy][xx]

						if len(around) == 0 {
							continue
						}

						for _, entityId := range central {
							entity := collidables[entityId]

							for _, nearbyEntityId := range around {
								if nearbyEntityId == entityId {
									continue
								}

								nearbyEntity := collidables[nearbyEntityId]

								if entity.CheckCollision(nearbyEntity) {
									collisions = append(collisions, CollisionPair{entity, nearbyEntity})
								}
							}
						}
					}
				}
			}
		}

		anyColliding = len(collisions) > 0

		for _, collision := range collisions {
			collision.first.Rearrange(collision.second)
		}
	}
}
