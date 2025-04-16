package main

import (
	"fmt"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// WeaponLoot represents a weapon pickup
type WeaponLoot struct {
	weapon     weapon
	pos        rl.Vector2
	destroyed  bool
	createTime float64  // Time when the loot was created
	color      rl.Color // Color for visual distinction between weapons
}

// AmmoLoot represents an ammo pickup
type AmmoLoot struct {
	amount     int // Amount of ammo in this pickup
	pos        rl.Vector2
	destroyed  bool
	createTime float64 // Time when the loot was created
}

func NewWeaponLoot(weapon weapon, pos rl.Vector2, currentTime float64) *WeaponLoot {
	// Set color based on weapon type
	var color rl.Color
	switch weapon {
	case PISTOL:
		color = rl.Green
	case MITRA:
		color = rl.Blue
	case SHOTGUN:
		color = rl.Purple
	case MINIGUN:
		color = rl.Gold
	default:
		color = rl.Green
	}

	return &WeaponLoot{
		weapon:     weapon,
		pos:        pos,
		createTime: currentTime,
		color:      color,
	}
}

func NewAmmoLoot(amount int, pos rl.Vector2, currentTime float64) *AmmoLoot {
	return &AmmoLoot{
		amount:     amount,
		pos:        pos,
		createTime: currentTime,
	}
}

func (l *WeaponLoot) Destroyed() bool {
	return l.destroyed
}

func (l *WeaponLoot) Render() {
	rl.DrawRectangle(int32(l.pos.X), int32(l.pos.Y), int32(lootSize), int32(lootSize), l.color)

	// Draw weapon indicator
	label := l.weapon.weaponName

	fontSize := 20
	textSize := rl.MeasureText(label, int32(fontSize))
	rl.DrawText(label, int32(l.pos.X+lootSize/2-float32(textSize)/2), int32(l.pos.Y+lootSize/2-float32(fontSize)/2), int32(fontSize), rl.White)
}

func (l *WeaponLoot) Position() rl.Vector2 {
	return l.pos
}

func (l *WeaponLoot) CheckCollision(other Collides) bool {
	return false
}

func (l *WeaponLoot) Rearrange(other Collides) {
}

func (l *AmmoLoot) Destroyed() bool {
	return l.destroyed
}

func (l *AmmoLoot) Render() {
	// Draw ammo box
	rl.DrawRectangle(int32(l.pos.X), int32(l.pos.Y), int32(lootSize), int32(lootSize), rl.Yellow)

	// Draw ammo text
	label := fmt.Sprintf("%d", l.amount)
	fontSize := 20
	textSize := rl.MeasureText(label, int32(fontSize))
	rl.DrawText(label, int32(l.pos.X+lootSize/2-float32(textSize)/2), int32(l.pos.Y+lootSize/2-float32(fontSize)/2), int32(fontSize), rl.Black)
}

func (l *AmmoLoot) Position() rl.Vector2 {
	return l.pos
}

func (l *AmmoLoot) CheckCollision(other Collides) bool {
	return false
}

func (l *AmmoLoot) Rearrange(other Collides) {
	// No rearrangement needed for ammo
}
