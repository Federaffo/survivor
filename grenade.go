package main

import (
	"fmt"

	rl "github.com/gen2brain/raylib-go/raylib"
)

var (
	grenadeRadius        float32 = 20
	grenadeExplosionTime float64 = 2.0 // 2 seconds
	grenadeExplosionSize float32 = 150 // Larger explosion radius (was 150)
	grenadeDamage        float32 = 200 // Increased damage (was 100)
)

// Grenade represents a timed explosive device
type Grenade struct {
	pos           rl.Vector2
	placedTime    float64 // When the grenade was placed
	hasExploded   bool
	destroyed     bool
	explosionTime float64 // When the explosion starts
	currentTime   float64 // Current game time
}

// NewGrenade creates a new grenade
func NewGrenade(pos rl.Vector2, currentTime float64) *Grenade {
	return &Grenade{
		pos:           pos,
		placedTime:    currentTime,
		explosionTime: currentTime + grenadeExplosionTime,
		currentTime:   currentTime,
		hasExploded:   false,
		destroyed:     false,
	}
}

// Update updates the grenade state
func (g *Grenade) Update(currentTime float64, enemyList []*Enemy) {
	g.currentTime = currentTime

	// Check if it's time to explode
	if !g.hasExploded && g.currentTime >= g.explosionTime {
		g.hasExploded = true

		// Damage all enemies within explosion radius
		for _, enemy := range enemyList {
			if rl.Vector2Distance(g.pos, enemy.pos) <= grenadeExplosionSize {
				enemy.DealDamage(grenadeDamage)

				// We should NOT mark enemies as destroyed here
				// This breaks the game's enemy counting logic
				// The main game loop should handle this
			}
		}

		// Set a delayed destruction time (show explosion for 0.5 seconds)
		g.explosionTime = currentTime + 0.5
	}

	// Only destroy after showing explosion for a brief time
	if g.hasExploded && currentTime > g.explosionTime {
		g.destroyed = true
	}
}

// Render draws the grenade and explosion effect
func (g *Grenade) Render() {
	if g.hasExploded {
		// Draw explosion with multiple layers for better effect
		// Outer explosion layer
		rl.DrawCircle(int32(g.pos.X), int32(g.pos.Y), grenadeExplosionSize, rl.ColorAlpha(rl.Orange, 0.4))

		// Middle explosion layer
		rl.DrawCircle(int32(g.pos.X), int32(g.pos.Y), grenadeExplosionSize*0.7, rl.ColorAlpha(rl.Red, 0.6))

		// Inner explosion layer
		rl.DrawCircle(int32(g.pos.X), int32(g.pos.Y), grenadeExplosionSize*0.4, rl.ColorAlpha(rl.Yellow, 0.8))
	} else {
		// Draw grenade
		rl.DrawCircle(int32(g.pos.X), int32(g.pos.Y), grenadeRadius, rl.Gray)

		// Draw a small colored indicator
		rl.DrawCircle(int32(g.pos.X), int32(g.pos.Y), grenadeRadius*0.5, rl.Red)

		// Draw countdown timer
		timeLeft := g.explosionTime - g.currentTime
		countdown := fmt.Sprintf("%.1f", timeLeft)
		countdownWidth := rl.MeasureText(countdown, 20)
		rl.DrawText(countdown, int32(g.pos.X)-countdownWidth/2, int32(g.pos.Y)-10, 20, rl.White)
	}
}

// Destroyed checks if the grenade should be removed
func (g *Grenade) Destroyed() bool {
	return g.destroyed
}

// Position returns the grenade's position
func (g *Grenade) Position() rl.Vector2 {
	return g.pos
}

// CheckCollision checks for collision with other objects
func (g *Grenade) CheckCollision(other Collides) bool {
	return false // Grenades don't need to collide with other objects
}

// Rearrange handles collision rearrangement
func (g *Grenade) Rearrange(other Collides) {
	// No need to rearrange grenades
}
