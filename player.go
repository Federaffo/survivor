package main

import rl "github.com/gen2brain/raylib-go/raylib"

type player struct {
	totalHp   int
	currentHp int
	pos       rl.Vector2
}

var (
	playerSpeed float32 = 5
)

func NewPlayer(totalHp int) player {
	return player{
		totalHp:   totalHp,
		currentHp: totalHp,
		pos:       rl.NewVector2(200, 200),
	}
}

func (p *player) Move() {
	if rl.IsKeyDown(rl.KeyA) {
		p.pos.X -= playerSpeed
	}
	if rl.IsKeyDown(rl.KeyD) {
		p.pos.X += playerSpeed
	}

	if rl.IsKeyDown(rl.KeyS) {
		p.pos.Y += playerSpeed
	}

	if rl.IsKeyDown(rl.KeyW) {
		p.pos.Y -= playerSpeed
	}
}
