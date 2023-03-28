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

	//have to be scaled based on screen size but it doesn't work :)
	playerSize = 20
	projSize = 3
	enemySize = 18

	rl.InitWindow(int32(w), int32(h), "Survivor")
	rl.SetTargetFPS(60)

	player := NewPlayer(100)

	var projList []*Projectile
	var enemyList []*Enemy
	var worldItems []WorldItem

	e := NewEnemy(rl.NewVector2(10, 10), 100, 10)
	enemyList = append(enemyList, e)
	worldItems = append(worldItems, e)

	e = NewEnemy(rl.NewVector2(100, 10), 100, 10)
	enemyList = append(enemyList, e)
	worldItems = append(worldItems, e)

	lastTime := rl.GetTime()

	for !rl.WindowShouldClose() {
		println(playerSize)
		println(int32(w))
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

		// move projectile
		for _, p := range projList {
			p.Update(dt)
		}
		// move enemy
		for _, p := range enemyList {
			p.Move(player.Pos, dt)
		}

		//check collision between proj and enemy
		for _, p := range projList {
			for _, e := range enemyList {

				if rl.CheckCollisionCircles(p.pos, 2, e.pos, 10) {
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
