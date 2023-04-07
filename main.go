package main

import (
	rl "github.com/gen2brain/raylib-go/raylib"
)

const (
	MAX_COLLISION_ORDERING_ITERS = 3
	SPACE_GRID_WIDTH             = 70
	SPACE_GRID_HEIGHT            = 70
)

var (
	playerSize float32
	enemySize  float32
	projSize   float32
	lootSize   float32
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

func RandomPointInCircle(radius float32) rl.Vector2 {
    x := float32(rl.GetRandomValue(-100, 100))
    y := float32(rl.GetRandomValue(-100, 100))
    vector := rl.Vector2Scale((rl.Vector2Normalize(rl.NewVector2(x, y))), radius)
    return vector
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
	var worldBodies []Collides
	var loots []*WeaponLoot

	lastTime := rl.GetTime()

	lastShoot := lastTime
	lastSpawn := lastTime
	spawnRate := 0.05

	w = rl.GetMonitorWidth(display)
	h = rl.GetMonitorHeight(display)

	// have to be scaled based on screen size
	playerSize = float32(w) / 150
	projSize = float32(w) / 1000
	enemySize = float32(w) / 200
	lootSize = 60

	spaceGrid := NewCollisionSpace(w, h, SPACE_GRID_WIDTH, SPACE_GRID_HEIGHT)

	for !rl.WindowShouldClose() {
		currentTime := rl.GetTime()
		dt := currentTime - lastTime

		mousePosition := rl.GetMousePosition()
		player.LookAt(mousePosition)
		player.Update(dt)

		// Spawn weapon
		{
			if rl.GetRandomValue(0, 1000) < 3 {
				x := rl.GetRandomValue(0, int32(w))
				y := rl.GetRandomValue(0, int32(h))
				w := NewWeaponLoot(MITRA, rl.NewVector2(float32(x), float32(y)))
				worldBodies = append(worldBodies, w)
				worldItems = append(worldItems, w)
				loots = append(loots, w)
			}
		}

		// Collision with loot
		{
			for _, l := range loots {
				if rl.CheckCollisionCircleRec(player.Pos, playerSize, rl.NewRectangle(l.pos.X, l.pos.Y, lootSize, lootSize)) {
					player.currentWeapon = l.weapon
					l.destroyed = true
				}
			}
		}

		// shoot
		{
			if rl.IsMouseButtonDown(0) {
				if currentTime > player.currentWeapon.shootingDelay+float64(lastShoot) {
					lastShoot = currentTime
					for _, p := range player.Shoot() {
						projList = append(projList, p)
						worldItems = append(worldItems, p)
					}
				}
			}
		}

		// Spwan enemys
		{
			if currentTime > lastSpawn+float64(spawnRate) {
				lastSpawn = currentTime

				respawn := true
				n := NewEnemy(rl.NewVector2(100,110), 100, 10, enemySize)
				for respawn{
					respawn = false
					spawnPosition := RandomPointInCircle(200)
					spawnPosition = rl.Vector2Add(spawnPosition, player.Pos)
					
					n = NewEnemy(rl.NewVector2(float32(spawnPosition.X), float32(spawnPosition.Y)), 100, 10, enemySize)
					if n.pos.X < 100 || n.pos.Y < 100{
						respawn = true
						continue
					}
					for _, e :=  range enemyList  {

						if rl.CheckCollisionCircles(n.pos,enemySize, e.pos,  enemySize){
							respawn = true
							break
						}
					}

					if !respawn{	
						enemyList = append(enemyList, n)
						worldItems = append(worldItems, n)
						worldBodies = append(worldBodies, n)
					}
				}
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

		spaceGrid.RearrangeBodies(MAX_COLLISION_ORDERING_ITERS, worldBodies, func() {
			spaceGrid.UpdateCells(worldBodies)
		})

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

			for _, x := range worldItems {
				x.Render()
			}

			/*
				for _, r := range projList {
					r.Render()
				}

				for _, r := range enemyList {
					r.Render()
				}
			*/

			player.Render()
			rl.DrawFPS(10, 10)
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
