package main

import (
    "math"
    rl "github.com/gen2brain/raylib-go/raylib"
)

type player struct {
	TotalHp   int
	CurrentHp int
	Pos       rl.Vector2

	lookAt    rl.Vector2
    lookAtSet bool
}

var (
	playerSpeed float32 = 5
)

func NewPlayer(totalHp int) player {
	return player{
		TotalHp:   totalHp,
		CurrentHp: totalHp,
		Pos:       rl.NewVector2(200, 200),
        lookAt:    rl.NewVector2(0, 0),
        lookAtSet: false,
	}
}

func (p *player) LookAt(lookAt rl.Vector2) {
    p.lookAtSet = true
    p.lookAt = rl.Vector2Normalize(rl.Vector2Subtract(lookAt, p.Pos))
}

func (p *player) Move() {
	if rl.IsKeyDown(rl.KeyA) {
		p.Pos.X -= playerSpeed
	}

	if rl.IsKeyDown(rl.KeyD) {
		p.Pos.X += playerSpeed
	}

	if rl.IsKeyDown(rl.KeyS) {
		p.Pos.Y += playerSpeed
	}

	if rl.IsKeyDown(rl.KeyW) {
		p.Pos.Y -= playerSpeed
	}
}

func (p *player) Render() {
    if p.lookAtSet {
        directionRectangle := rl.NewRectangle(
            p.Pos.X + p.lookAt.X * 10,
            p.Pos.Y + p.lookAt.Y * 10,
            10,
            2,
        )
        rotation := float32(math.Atan2(float64(p.lookAt.Y), float64(p.lookAt.X)) * 180 / math.Pi)
        rl.DrawRectanglePro(directionRectangle, rl.NewVector2(0, 1), rotation, rl.Green)
    }
    rl.DrawCircle(int32(p.Pos.X), int32(p.Pos.Y), 10, rl.Red)
}
