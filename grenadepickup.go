package main

import (
	"fmt"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// GrenadePickup represents a grenade pickup item
type GrenadePickup struct {
	pos        rl.Vector2
	createTime float64
	destroyed  bool
	amount     int // Number of grenades to give
	size       int // Size of pickup icon
}

// NewGrenadePickup creates a new grenade pickup
func NewGrenadePickup(pos rl.Vector2, createTime float64) *GrenadePickup {
	return &GrenadePickup{
		pos:        pos,
		createTime: createTime,
		destroyed:  false,
		amount:     2, // Give 2 grenades
		size:       60,
	}
}

// Render draws the grenade pickup
func (g *GrenadePickup) Render() {
	// Draw a box background
	rl.DrawRectangle(int32(g.pos.X), int32(g.pos.Y), int32(g.size), int32(g.size), rl.DarkGray)

	// Draw a grenade icon
	centerX := g.pos.X + float32(g.size)/2
	centerY := g.pos.Y + float32(g.size)/2

	// Draw the grenade body
	radius := float32(g.size) / 4
	rl.DrawCircle(int32(centerX), int32(centerY), radius, rl.Gray)

	// Draw the red indicator
	rl.DrawCircle(int32(centerX), int32(centerY), radius/2, rl.Red)

	// Draw text showing how many grenades
	text := fmt.Sprintf("+%d", g.amount)
	textWidth := rl.MeasureText(text, 20)
	rl.DrawText(text, int32(g.pos.X)+int32(g.size)/2-textWidth/2,
		int32(g.pos.Y)+int32(g.size)-25, 20, rl.White)
}

// Destroyed checks if the pickup has been destroyed
func (g *GrenadePickup) Destroyed() bool {
	return g.destroyed
}

// Position returns the position of the pickup
func (g *GrenadePickup) Position() rl.Vector2 {
	return g.pos
}

// CheckCollision checks for collision with other objects
func (g *GrenadePickup) CheckCollision(other Collides) bool {
	return false // Grenade pickups don't need to collide with other objects
}

// Rearrange handles collision rearrangement
func (g *GrenadePickup) Rearrange(other Collides) {
	// No need to rearrange grenade pickups
}
