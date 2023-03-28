package main

import rl "github.com/gen2brain/raylib-go/raylib"

var (
	defaultDamage float32 = 10
	projSpeed     float32 = 10
)

type Projectile struct {
	damage float32
	dir    rl.Vector2
	pos    rl.Vector2
}

func NewProj(initialPos rl.Vector2, direction rl.Vector2) *Projectile {
	dir := rl.Vector2Subtract(direction, initialPos)
	dir = rl.Vector2Normalize(dir)
	dir = rl.Vector2Scale(dir, projSpeed)

	return &Projectile{
		damage: defaultDamage,
		pos:    initialPos,
		dir:    dir,
	}
}

func (p *Projectile) Update() {
	p.pos = rl.Vector2Add(p.pos, p.dir)

	//p.hitbox.X = p.pos.X
	//p.hitbox.Y = p.pos.Y
}

func (p *Projectile) Render() {
	rl.DrawCircle(int32(p.pos.X), int32(p.pos.Y), 2, rl.Green)
}
