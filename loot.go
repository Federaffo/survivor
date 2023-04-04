package main

import rl "github.com/gen2brain/raylib-go/raylib"

type WeaponLoot struct {
	weapon    weapon
	pos       rl.Vector2
	destroyed bool
}

func NewWeaponLoot(weapon weapon, pos rl.Vector2) *WeaponLoot {
	return &WeaponLoot{
		weapon: weapon,
		pos:    pos,
	}
}

func (l *WeaponLoot) Destroyed() bool {
	return l.destroyed
}

func (l *WeaponLoot) Render() {
	rl.DrawRectangle(int32(l.pos.X), int32(l.pos.Y), int32(lootSize), int32(lootSize), rl.Green)
}
