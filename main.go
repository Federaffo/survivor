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

        mousePosition := rl.GetMousePosition()
        player.LookAt(mousePosition)

		//shoot
		{
			if rl.IsMouseButtonPressed(0) {
				p := NewProj(player.Pos, mousePosition)
				projList = append(projList, p)
			}
		}

		// move projectile
		for _, p := range projList {
			p.Update()
		}

		rl.BeginDrawing()
		{
			rl.ClearBackground(rl.Black)
			for _, r := range projList {
				r.Render()
			}

            player.Render()
		}
		rl.EndDrawing()
	}

	rl.CloseWindow()
}
