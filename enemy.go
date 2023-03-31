package main

import (
	rl "github.com/gen2brain/raylib-go/raylib"
)

var (
	enemySpeed float32 = 70
)

type Enemy struct {
	pos               rl.Vector2
	health, maxHealth float32
	damage            float32
	destroyed         bool
}

func NewEnemy(pos rl.Vector2, maxHealth float32, damage float32) *Enemy {
	s := Enemy{
		pos:       pos,
		damage:    damage,
		health:    maxHealth,
		maxHealth: maxHealth, // Should be set based on level
	}
	return &s
}

func (e *Enemy) Move(playerPos rl.Vector2, dt float64) {
	dtspeed := dt * float64(enemySpeed)

	dir := rl.Vector2Subtract(playerPos, e.pos)
	dir = rl.Vector2Normalize(dir)
	dir = rl.Vector2Scale(dir, float32(dtspeed))

	e.pos = rl.Vector2Add(e.pos, dir)
}

func (e *Enemy) Render() {
	rl.DrawCircle(int32(e.pos.X), int32(e.pos.Y), enemySize, rl.Red)
}

func (e *Enemy) DealDamage(dmg float32) {
	e.health -= dmg
}

func (e *Enemy) Destroyed() bool {
	return e.destroyed
}
