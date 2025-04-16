package main

import (
	"math"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type player struct {
	TotalHp   int
	CurrentHp int
	Pos       rl.Vector2

	currentWeapon weapon

	lookAt    rl.Vector2
	lookAtSet bool
}

type weapon struct {
	shootingDelay float64
	projDamage    float32
	nProj         int
}

var (
	PISTOL weapon = weapon{shootingDelay: 0.5, projDamage: 50, nProj: 1}
	MITRA  weapon = weapon{shootingDelay: 0.1, projDamage: 500, nProj: 1}
)

var playerSpeed float32 = 300

func NewPlayer(totalHp int) player {
	return player{
		TotalHp:       totalHp,
		CurrentHp:     totalHp,
		Pos:           rl.NewVector2(200, 200),
		lookAt:        rl.NewVector2(0, 0),
		currentWeapon: PISTOL,
		lookAtSet:     false,
	}
}

func (p *player) LookAt(lookAt rl.Vector2) {
	p.lookAtSet = true
	p.lookAt = rl.Vector2Normalize(rl.Vector2Subtract(lookAt, p.Pos))
}

func (p *player) Update(dt float64) {
	dtSpeed := playerSpeed * float32(dt)
	if rl.IsKeyDown(rl.KeyA) {
		p.Pos.X -= dtSpeed
	}

	if rl.IsKeyDown(rl.KeyD) {
		p.Pos.X += dtSpeed
	}

	if rl.IsKeyDown(rl.KeyS) {
		p.Pos.Y += dtSpeed
	}

	if rl.IsKeyDown(rl.KeyW) {
		p.Pos.Y -= dtSpeed
	}
}

func (p *player) Render() {
	if p.lookAtSet {
		directionRectangle := rl.NewRectangle(
			p.Pos.X+p.lookAt.X*10,
			p.Pos.Y+p.lookAt.Y*10,
			20,
			2,
		)
		rotation := float32(math.Atan2(float64(p.lookAt.Y), float64(p.lookAt.X)) * 180 / math.Pi)
		rl.DrawRectanglePro(directionRectangle, rl.NewVector2(0, 1), rotation, rl.Green)
	}
	rl.DrawCircle(int32(p.Pos.X), int32(p.Pos.Y), playerSize, rl.Red)

	// Draw health bar above player
	healthBarWidth := playerSize * 2
	healthBarHeight := 5.0
	healthPercentage := float32(p.CurrentHp) / float32(p.TotalHp)

	// Background of health bar
	rl.DrawRectangle(
		int32(p.Pos.X-healthBarWidth/2),
		int32(p.Pos.Y-playerSize-10),
		int32(healthBarWidth),
		int32(healthBarHeight),
		rl.DarkGray,
	)

	// Actual health
	rl.DrawRectangle(
		int32(p.Pos.X-healthBarWidth/2),
		int32(p.Pos.Y-playerSize-10),
		int32(healthBarWidth*healthPercentage),
		int32(healthBarHeight),
		rl.Red,
	)
}

func (p *player) Shoot() []*Projectile {
	var projs []*Projectile
	for i := 0; i < p.currentWeapon.nProj; i++ {
		noise := rl.GetRandomValue(-100, 100)
		noisedDirection := rl.Vector2Add(rl.GetMousePosition(), rl.NewVector2(float32(noise), float32(noise)))
		projs = append(projs, NewProj(p.Pos, noisedDirection, p.currentWeapon.projDamage))
	}
	return projs
}

func (p *player) Position() rl.Vector2 {
	return p.Pos
}

func (p *player) TakeDamage(damage float32) {
	p.CurrentHp -= int(damage)
	if p.CurrentHp < 0 {
		p.CurrentHp = 0
	}
}

func (p *player) CheckCollision(other Collides) bool {
	switch other.(type) {
	case *Enemy:
		enemy := other.(*Enemy)
		return rl.CheckCollisionCircles(p.Pos, playerSize, enemy.pos, enemy.bodyRadius)
	}
	return false
}

func (p *player) Rearrange(other Collides) {
	// Player doesn't need to rearrange since enemy collision will push the player back
}
