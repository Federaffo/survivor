package main

import (
	rl "github.com/gen2brain/raylib-go/raylib"
)

var (
	defaultDamage float32 = 100
	projSpeed     float32 = 400
)

type Projectile struct {
	damage    float32
	dir       rl.Vector2
	pos       rl.Vector2
	destroyed bool
}

func NewProj(initialPos rl.Vector2, direction rl.Vector2) *Projectile {
	dir := rl.Vector2Subtract(direction, initialPos)
	dir = rl.Vector2Normalize(dir)

	return &Projectile{
		damage: defaultDamage,
		pos:    initialPos,
		dir:    dir,
	}
}

func (p *Projectile) Destroyed() bool {
	return p.destroyed
}

func (p *Projectile) Update(dt float64) {
	dtspeed := dt * float64(projSpeed)
	dir := rl.Vector2Scale(p.dir, float32(dtspeed))
	p.pos = rl.Vector2Add(p.pos, dir)

	//p.hitbox.X = p.pos.X
	//p.hitbox.Y = p.pos.Y
}

func (p *Projectile) Render() {
	rl.DrawCircle(int32(p.pos.X), int32(p.pos.Y), projSize, rl.Green)
}
