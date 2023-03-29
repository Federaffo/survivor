package main

import (
	rl "github.com/gen2brain/raylib-go/raylib"
)

var (
	playerSize float32
	enemySize  float32
	projSize   float32
)

type WorldItem interface {
	Render()
	//Update()
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
	spawnRate := 1

	w = rl.GetMonitorWidth(display)
	h = rl.GetMonitorHeight(display)

	//have to be scaled based on screen size
	playerSize = float32(w) / 150
	projSize = float32(w) / 1000
	enemySize = float32(w) / 200

	for !rl.WindowShouldClose() {
		currentTime := rl.GetTime()
		dt := currentTime - lastTime

		mousePosition := rl.GetMousePosition()
		player.LookAt(mousePosition)
		player.Update(dt)

		//shoot
		{
			if rl.IsMouseButtonPressed(0) {
				p := NewProj(player.Pos, mousePosition)
				projList = append(projList, p)
				worldItems = append(worldItems, p)

			}
		}

		//Spwan enemys
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

		// move projectile
		for _, p := range projList {
			p.Update(dt)
		}
		// move enemy
		for _, p := range enemyList {

			var dir rl.Vector2
			collision := false

			for _, pp := range enemyList {
				if p.pos != pp.pos {

					dtspeed := dt * float64(enemySpeed)

					dir = rl.Vector2Subtract(player.Pos, p.pos)
					dir = rl.Vector2Normalize(dir)
					dir = rl.Vector2Scale(dir, float32(dtspeed))

					if rl.CheckCollisionCircles(rl.Vector2Add(p.pos, dir), enemySize, pp.pos, enemySize) {
						collision = true
						break
					}

				}
			}
			if !collision {
				//p.pos = rl.Vector2Add(p.pos, dir)
				p.Move(player.Pos, dt)
			}

		}

		//check collision between proj and enemy
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
