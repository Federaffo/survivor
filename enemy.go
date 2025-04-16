package main

import (
	rl "github.com/gen2brain/raylib-go/raylib"
)

var enemySpeed float32 = 70
var enemySprite rl.Texture2D // Single shared texture for all enemies

// Initialize the enemy sprite
func InitEnemySprite() {
	// Try different possible paths for the zombie sprite
	path := "assets/zombie.png"

	// Print working directory for debugging
	rl.TraceLog(rl.LogWarning, "Loading zombie sprite...")

	// Try to load the sprite
	enemySprite = rl.LoadTexture(path)

	// Check if sprite loaded successfully
	if enemySprite.ID > 0 {
		rl.TraceLog(rl.LogInfo, "Successfully loaded zombie sprite from %s", path)
	}

	if enemySprite.ID == 0 {
		rl.TraceLog(rl.LogWarning, "Failed to load zombie sprite! Will use fallback circle.")
	}
}

// Unload the enemy sprite
func UnloadEnemySprite() {
	if enemySprite.ID > 0 {
		rl.UnloadTexture(enemySprite)
	}
}

type Enemy struct {
	pos               rl.Vector2
	bodyRadius        float32
	health, maxHealth float32
	damage            float32
	destroyed         bool
}

func NewEnemy(pos rl.Vector2, maxHealth, damage, bodyRadius float32) *Enemy {
	s := Enemy{
		pos:        pos,
		bodyRadius: bodyRadius,
		damage:     damage,
		health:     maxHealth,
		maxHealth:  maxHealth, // Should be set based on level
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

func (e *Enemy) Position() rl.Vector2 {
	return e.pos
}

func (e *Enemy) Render() {
	// Check if sprite was loaded successfully
	if enemySprite.ID > 0 {
		// Size to draw the sprite (scale it according to enemySize)
		spriteScale := enemySize / float32(enemySprite.Height) * 3.0
		width := float32(enemySprite.Width) * spriteScale
		height := float32(enemySprite.Height) * spriteScale

		// Draw the sprite centered on enemy position
		rl.DrawTexturePro(
			enemySprite,
			rl.NewRectangle(0, 0, float32(enemySprite.Width), float32(enemySprite.Height)),
			rl.NewRectangle(e.pos.X-width/2, e.pos.Y-height/2, width, height),
			rl.NewVector2(0, 0),
			0,
			rl.White,
		)

		// Draw health bar above enemy
		healthBarWidth := enemySize * 2
		healthBarHeight := 4.0
		healthPercentage := e.health / e.maxHealth

		// Background of health bar
		rl.DrawRectangle(
			int32(e.pos.X-healthBarWidth/2),
			int32(e.pos.Y-height/2-10),
			int32(healthBarWidth),
			int32(healthBarHeight),
			rl.DarkGray,
		)

		// Actual health
		rl.DrawRectangle(
			int32(e.pos.X-healthBarWidth/2),
			int32(e.pos.Y-height/2-10),
			int32(healthBarWidth*healthPercentage),
			int32(healthBarHeight),
			rl.Red,
		)
	} else {
		// Fallback to circle if sprite not loaded
		rl.DrawCircle(int32(e.pos.X), int32(e.pos.Y), enemySize, rl.Red)

		// Draw health bar above enemy
		healthBarWidth := enemySize * 2
		healthBarHeight := 4.0
		healthPercentage := e.health / e.maxHealth

		// Background of health bar
		rl.DrawRectangle(
			int32(e.pos.X-healthBarWidth/2),
			int32(e.pos.Y-enemySize-10),
			int32(healthBarWidth),
			int32(healthBarHeight),
			rl.DarkGray,
		)

		// Actual health
		rl.DrawRectangle(
			int32(e.pos.X-healthBarWidth/2),
			int32(e.pos.Y-enemySize-10),
			int32(healthBarWidth*healthPercentage),
			int32(healthBarHeight),
			rl.Red,
		)
	}
}

func (e *Enemy) DealDamage(dmg float32) {
	e.health -= dmg
}

func (e *Enemy) Destroyed() bool {
	return e.destroyed
}

func (e *Enemy) CheckCollision(other Collides) bool {
	switch (other).(type) {
	case *Enemy:
		enemy := other.(*Enemy)
		return rl.CheckCollisionCircles(e.pos, e.bodyRadius, enemy.pos, enemy.bodyRadius)
	}
	return false
}

func (e *Enemy) Rearrange(other Collides) {
	switch (other).(type) {
	case *Enemy:
		enemy := other.(*Enemy)
		dist := rl.Vector2Distance(e.pos, enemy.pos)

		// Only rearrange if we're actually overlapping
		if dist < e.bodyRadius+enemy.bodyRadius {
			desiredDist := e.bodyRadius + enemy.bodyRadius

			// Calculate overlap
			overlap := desiredDist - dist

			// Add a small random jitter to prevent perfect symmetry that can cause flickering
			jitterX := float32(rl.GetRandomValue(-10, 10)) * 0.01
			jitterY := float32(rl.GetRandomValue(-10, 10)) * 0.01
			jitter := rl.NewVector2(jitterX, jitterY)

			// Get direction vector from this enemy to the other
			dir := rl.Vector2Subtract(enemy.pos, e.pos)
			if dir.X == 0 && dir.Y == 0 {
				// If objects are perfectly overlapping, push in a random direction
				dir = rl.NewVector2(jitterX*10, jitterY*10)
			}

			// Normalize direction and apply dampening factor to make movement less aggressive
			dir = rl.Vector2Normalize(dir)
			dampening := float32(0.5) // Reduce the strength of the push

			// Calculate movement vectors with dampening
			moveAmount := overlap * dampening
			pToColliding := rl.Vector2Scale(dir, moveAmount/2)
			collidingToP := rl.Vector2Scale(rl.Vector2Negate(dir), moveAmount/2)

			// Apply jitter to both objects
			pToColliding = rl.Vector2Add(pToColliding, jitter)
			collidingToP = rl.Vector2Add(collidingToP, rl.Vector2Negate(jitter))

			// Move objects apart
			e.pos = rl.Vector2Add(e.pos, collidingToP)
			enemy.pos = rl.Vector2Add(enemy.pos, pToColliding)
		}
	}
}
