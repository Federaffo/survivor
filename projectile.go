package main

import (
	"math"

	rl "github.com/gen2brain/raylib-go/raylib"
)

var (
	defaultDamage float32      = 100
	projSpeed     float32      = 400
	bulletTexture rl.Texture2D // Texture for bullet
)

// Initialize the bullet sprite
func InitBulletSprite() {
	// Try different possible paths for the bullet sprite
	path := "assets/bullet.png"

	// Print working directory for debugging
	rl.TraceLog(rl.LogWarning, "Loading bullet sprite...")

	// Try to load the sprite
	bulletTexture = rl.LoadTexture(path)

	// Check if sprite loaded successfully
	if bulletTexture.ID > 0 {
		rl.TraceLog(rl.LogInfo, "Successfully loaded bullet sprite from %s", path)
	}

	if bulletTexture.ID == 0 {
		rl.TraceLog(rl.LogWarning, "Failed to load bullet sprite! Will use fallback circle.")
	}
}

// Unload the bullet sprite
func UnloadBulletSprite() {
	if bulletTexture.ID > 0 {
		rl.UnloadTexture(bulletTexture)
	}
}

type Projectile struct {
	damage    float32
	dir       rl.Vector2
	pos       rl.Vector2
	destroyed bool
}

func NewProj(initialPos rl.Vector2, direction rl.Vector2, damage float32) *Projectile {
	dir := rl.Vector2Subtract(direction, initialPos)
	dir = rl.Vector2Normalize(dir)

	return &Projectile{
		damage: damage,
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

	// p.hitbox.X = p.pos.X
	// p.hitbox.Y = p.pos.Y
}

func (p *Projectile) Render() {
	// Check if bullet texture was loaded successfully
	if bulletTexture.ID > 0 {
		// Calculate rotation angle based on direction
		rotation := float32(math.Atan2(float64(p.dir.Y), float64(p.dir.X)) * 180.0 / math.Pi)

		// Size to draw the sprite
		bulletScale := projSize * 3.0 / float32(bulletTexture.Height)
		width := float32(bulletTexture.Width) * bulletScale
		height := float32(bulletTexture.Height) * bulletScale

		// Draw the bullet sprite with rotation
		rl.DrawTexturePro(
			bulletTexture,
			rl.NewRectangle(0, 0, float32(bulletTexture.Width), float32(bulletTexture.Height)),
			rl.NewRectangle(p.pos.X-width/2, p.pos.Y-height/2, width, height),
			rl.NewVector2(0, 0),
			rotation,
			rl.White,
		)
	} else {
		// Fallback to circle if texture not loaded
		rl.DrawCircle(int32(p.pos.X), int32(p.pos.Y), projSize, rl.Green)
	}
}

func (p *Projectile) Position() rl.Vector2 {
	return p.pos
}
