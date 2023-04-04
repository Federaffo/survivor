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
	for y := range cs.CollisionSpaceGrid {
		for x := range cs.CollisionSpaceGrid[y] {
			cs.CollisionSpaceGrid[y][x] = nil
		}
	}

	for i, body := range bodies {
		gy := ((int(body.Position().Y) / cs.CellHeight) % cs.Cols) + 1
		gx := ((int(body.Position().X) / cs.CellWidth) % cs.Rows) + 1
		cs.CollisionSpaceGrid[gy][gx] = append(cs.CollisionSpaceGrid[gy][gx], i)
	}
}

func (cs *CollisionSpace) RearrangeBodies(maxIters int, collidables []Collides, each func()) {
	collisions := []CollisionPair{}

	anyColliding := true
	for iters := 0; anyColliding && iters < MAX_COLLISION_ORDERING_ITERS; iters++ {
		// Update space grid
		each()

		for y := 1; y < cs.Rows; y++ {
			for x := 1; x < cs.Cols; x++ {
				central := cs.CollisionSpaceGrid[y][x]
				for yy := y - 1; yy < y+2; yy++ {
					for xx := x - 1; xx < x+2; xx++ {
						around := cs.CollisionSpaceGrid[yy][xx]
						for _, enemyId := range central {
							p := collidables[enemyId]
							for _, nearbyEnemyId := range around {
								pp := collidables[nearbyEnemyId]
								if rl.CheckCollisionCircles(p.Position(), enemySize, pp.Position(), enemySize) && nearbyEnemyId != enemyId {
									collisions = append(collisions, CollisionPair{p, pp})
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
