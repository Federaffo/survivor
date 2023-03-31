package main

import (
	rl "github.com/gen2brain/raylib-go/raylib"
)

const (
    MAX_COLLISION_ORDERING_ITERS = 15
    SPACE_GRID_WIDTH = 100
    SPACE_GRID_HEIGHT = 100
)

var (
	playerSize float32
	enemySize  float32
	projSize   float32
)

type WorldItem interface {
	Render()
	// Update()
	Destroyed() bool
}

func RemoveIndex[T any](s []T, index int) []T {
	return append(s[:index], s[index+1:]...)
}

func UpdateWorldItems[T WorldItem](worldItems []T) []T {
	toRemove := make([]int, 0)
	for i := len(worldItems) - 1; i >= 0; i-- {
		if worldItems[i].Destroyed() {
			toRemove = append(toRemove, i)
		}
	}
	for _, i := range toRemove {
		worldItems = RemoveIndex(worldItems, i)
	}
	return worldItems
}

func main() {
	display := rl.GetCurrentMonitor()

	w := rl.GetMonitorWidth(display)
	h := rl.GetMonitorHeight(display)

	rl.InitWindow(int32(w), int32(h), "Survivor")
	rl.SetTargetFPS(60)

	player := NewPlayer(100)

	var projList []*Projectile
	var enemyList []*Enemy
	var worldItems []WorldItem

	lastTime := rl.GetTime()

	lastSpawn := lastTime
	spawnRate := 0.1

	w = rl.GetMonitorWidth(display)
	h = rl.GetMonitorHeight(display)

	// have to be scaled based on screen size
	playerSize = float32(w) / 150
	projSize = float32(w) / 1000
	enemySize = float32(w) / 200

    gridCellWidth := w / SPACE_GRID_WIDTH
    gridCellHeight:= h / SPACE_GRID_HEIGHT

    spaceGrid := make([][][]int, SPACE_GRID_HEIGHT + 2)
    for i := 0; i < SPACE_GRID_HEIGHT + 2; i++ {
        spaceGrid[i] = make([][]int, SPACE_GRID_WIDTH + 2)
    }

	for !rl.WindowShouldClose() {
		currentTime := rl.GetTime()
		dt := currentTime - lastTime

		mousePosition := rl.GetMousePosition()
		player.LookAt(mousePosition)
		player.Update(dt)

		// shoot
		{
			if rl.IsMouseButtonPressed(0) {
				p := NewProj(player.Pos, mousePosition)
				projList = append(projList, p)
				worldItems = append(worldItems, p)

			}
		}

		// Spwan enemys
		{
			if currentTime > lastSpawn+float64(spawnRate) {
				lastSpawn = currentTime

				x := rl.GetRandomValue(0, int32(w))
				y := rl.GetRandomValue(0, int32(h))

				e := NewEnemy(rl.NewVector2(float32(x), float32(y)), 100, 10)
				enemyList = append(enemyList, e)
				worldItems = append(worldItems, e)
			}
		}

        updateSpaceGrid := func () {
            for y := range spaceGrid {
                for x := range spaceGrid[y] {
                    spaceGrid[y][x] = []int{}
                }
            }

            for i, enemy := range enemyList {
                gy := ((int(enemy.pos.Y) / gridCellHeight) % SPACE_GRID_HEIGHT) + 1
                gx := ((int(enemy.pos.X) / gridCellWidth) % SPACE_GRID_WIDTH) + 1
                spaceGrid[gy][gx] = append(spaceGrid[gy][gx], i)
            }
        }


		// move projectile
		for _, p := range projList {
			p.Update(dt)
		}
		// move enemy

		for _, p := range enemyList {
			p.Move(player.Pos, dt)
		}

		type CollisionPair struct {
			first  *Enemy
			second *Enemy
		}

		anyColliding := true
        for iters := 0; anyColliding && iters < MAX_COLLISION_ORDERING_ITERS; iters++ {
            // Update space grid
            updateSpaceGrid()

			anyColliding = false
	        collisions := []CollisionPair{}

            for y := 1; y < len(spaceGrid) - 2; y++ {
                for x := 1; x < len(spaceGrid[y]) - 2; x++ {
                    central := spaceGrid[y][x]
                    for yy := y-1; yy < y + 2; yy++ {
                        for xx := x-1; xx < y + 2; xx++ {
                            around := spaceGrid[yy][xx]
                            for _, enemyId := range central {
                                p := enemyList[enemyId]
                                for _, nearbyEnemyId := range around {
                                    pp := enemyList[nearbyEnemyId]
                                    if rl.CheckCollisionCircles(p.pos, enemySize, pp.pos, enemySize) && nearbyEnemyId != enemyId {
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
                dist := rl.Vector2Distance(collision.first.pos, collision.second.pos)
                desiredDist := enemySize * 2
                diff := desiredDist - dist
                pToColliding := rl.Vector2Scale(rl.Vector2Normalize(rl.Vector2Subtract(collision.second.pos, collision.first.pos)), diff/2)
                collidingToP := rl.Vector2Scale(rl.Vector2Normalize(rl.Vector2Subtract(collision.first.pos, collision.second.pos)), diff/2)
                collision.first.pos = rl.Vector2Add(collision.first.pos, collidingToP)
                collision.second.pos = rl.Vector2Add(collision.second.pos, pToColliding)
            }
        }

		// check collision between proj and enemy
		for _, p := range projList {
			for _, e := range enemyList {
				if rl.CheckCollisionCircles(p.pos, projSize, e.pos, enemySize) {
					e.DealDamage(p.damage)
					if e.health <= 0 {
						e.destroyed = true
					}
					p.destroyed = true
				}
			}
		}

		rl.BeginDrawing()
		{
			rl.ClearBackground(rl.Black)
			for _, r := range projList {
				r.Render()
			}

			for _, r := range enemyList {
				r.Render()
			}

			player.Render()
		}
		rl.EndDrawing()

		{
			worldItems = UpdateWorldItems(worldItems)
			projList = UpdateWorldItems(projList)
			enemyList = UpdateWorldItems(enemyList)
		}
		lastTime = currentTime
	}

	rl.CloseWindow()
}
