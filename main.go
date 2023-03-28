package main

import (
	rl "github.com/gen2brain/raylib-go/raylib"
)

func main() {
	display := rl.GetCurrentMonitor()

	w := rl.GetMonitorWidth(display)
	h := rl.GetMonitorHeight(display)
	rl.InitWindow(int32(w), int32(h), "Survivor")
	rl.SetTargetFPS(60)

	player := NewPlayer(100)

	var projList []*Projectile

	for !rl.WindowShouldClose() {

		player.Move()

		//shoot
		{
			if rl.IsMouseButtonPressed(0) {
				p := NewProj(player.pos, rl.GetMousePosition())
				projList = append(projList, p)
			}
		}

		// move projectile
		for _, p := range projList {
			p.Update()
		}

		rl.BeginDrawing()
		{

			for _, r := range projList {
				r.Render()
			}

			rl.ClearBackground(rl.Black)
			rl.DrawCircle(int32(player.pos.X), int32(player.pos.Y), 10, rl.Red)
		}
		rl.EndDrawing()

	}

	rl.CloseWindow()

}
