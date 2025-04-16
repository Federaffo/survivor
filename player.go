package main

import (
	"math"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type player struct {
	TotalHp   int
	CurrentHp int
	Pos       rl.Vector2

	currentWeapon   weapon
	ammo            int     // Total ammo in inventory
	currentMagazine int     // Current ammo in magazine
	isReloading     bool    // Whether player is currently reloading
	reloadStartTime float64 // When reload started

	grenades int // Number of grenades player has

	lookAt    rl.Vector2
	lookAtSet bool

	// Player sprite textures
	spriteLeft  rl.Texture2D
	spriteRight rl.Texture2D
	facingLeft  bool // Track player direction for sprite selection

	// Weapon pickup notification
	weaponPickupTime float64
	weaponPickupName string

	// Ammo pickup notification
	ammoPickupTime   float64
	ammoPickupAmount int
}

type weapon struct {
	weaponName    string
	shootingDelay float64
	projDamage    float32
	nProj         int
	usesAmmo      bool    // Whether this weapon uses ammo
	magazineSize  int     // How many bullets in a full magazine
	reloadTime    float64 // How long it takes to reload in seconds
}

var (
	PISTOL  weapon = weapon{shootingDelay: 0.5, projDamage: 50, nProj: 1, weaponName: "Pistol", usesAmmo: false, magazineSize: 12, reloadTime: 1.0}
	MITRA   weapon = weapon{shootingDelay: 0.1, projDamage: 500, nProj: 1, weaponName: "Mitra", usesAmmo: true, magazineSize: 30, reloadTime: 1.5}
	SHOTGUN weapon = weapon{shootingDelay: 0.8, projDamage: 30, nProj: 5, weaponName: "Shotgun", usesAmmo: true, magazineSize: 8, reloadTime: 2.0}    // Shoots multiple projectiles
	MINIGUN weapon = weapon{shootingDelay: 0.05, projDamage: 15, nProj: 1, weaponName: "Minigun", usesAmmo: true, magazineSize: 100, reloadTime: 3.0} // Very fast firing rate
)

var playerSpeed float32 = 300

func NewPlayer(totalHp int) player {
	// Load player sprites
	var spriteLeft, spriteRight rl.Texture2D

	// Try different possible paths for the sprites
	path := "assets/player_left.png"

	// Print working directory for debugging
	rl.TraceLog(rl.LogWarning, "Loading player sprites...")

	spriteLoaded := false
	// Try to load the left sprite
	spriteLeft = rl.LoadTexture(path)

	// Check if left sprite loaded successfully
	if spriteLeft.ID > 0 {
		// Now try to load the right sprite
		rightPath := "assets/player_right.png"
		spriteRight = rl.LoadTexture(rightPath)

		// Check if right texture loaded successfully
		if spriteRight.ID > 0 {
			rl.TraceLog(rl.LogInfo, "Successfully loaded sprites from %s and %s", path, rightPath)
			spriteLoaded = true
		}
	}

	if !spriteLoaded {
		rl.TraceLog(rl.LogWarning, "Failed to load player sprites! Will use fallback circle.")
	}

	return player{
		TotalHp:          totalHp,
		CurrentHp:        totalHp,
		Pos:              rl.NewVector2(500, 500),
		lookAt:           rl.NewVector2(0, 0),
		currentWeapon:    PISTOL,
		lookAtSet:        false,
		weaponPickupName: "",
		weaponPickupTime: 0,
		ammo:             10000,
		currentMagazine:  PISTOL.magazineSize, // Start with full magazine
		isReloading:      false,
		ammoPickupTime:   0,
		ammoPickupAmount: 0,
		grenades:         3, // Start with 3 grenades
		spriteLeft:       spriteLeft,
		spriteRight:      spriteRight,
		facingLeft:       false,
	}
}

func (p *player) LookAt(lookAt rl.Vector2) {
	p.lookAtSet = true
	p.lookAt = rl.Vector2Normalize(rl.Vector2Subtract(lookAt, p.Pos))
}

func (p *player) Update(dt float64, currentTime float64) {
	dtSpeed := playerSpeed * float32(dt)

	// Track movement for sprite direction
	moving := false

	if rl.IsKeyDown(rl.KeyA) {
		p.Pos.X -= dtSpeed
		p.facingLeft = true
		moving = true
	}

	if rl.IsKeyDown(rl.KeyD) {
		p.Pos.X += dtSpeed
		p.facingLeft = false
		moving = true
	}

	if rl.IsKeyDown(rl.KeyS) {
		p.Pos.Y += dtSpeed
		moving = true
	}

	if rl.IsKeyDown(rl.KeyW) {
		p.Pos.Y -= dtSpeed
		moving = true
	}

	// Screen boundaries - get current screen dimensions
	screenWidth := float32(rl.GetScreenWidth())
	screenHeight := float32(rl.GetScreenHeight())

	// Calculate player's effective size (taking into account the sprite scaling)
	var playerBoundarySize float32
	spritesLoaded := p.spriteLeft.ID > 0 && p.spriteRight.ID > 0
	if spritesLoaded {
		// Use the standard size for boundary calculations
		playerBoundarySize = playerSize * 1.6 // A bit smaller than the actual sprite for better feel
	} else {
		// Fallback circle size
		playerBoundarySize = playerSize * 1.6
	}

	// Apply boundary constraints
	// Left boundary
	if p.Pos.X < playerBoundarySize {
		p.Pos.X = playerBoundarySize
	}
	// Right boundary
	if p.Pos.X > screenWidth-playerBoundarySize {
		p.Pos.X = screenWidth - playerBoundarySize
	}
	// Top boundary
	if p.Pos.Y < playerBoundarySize {
		p.Pos.Y = playerBoundarySize
	}
	// Bottom boundary
	if p.Pos.Y > screenHeight-playerBoundarySize {
		p.Pos.Y = screenHeight - playerBoundarySize
	}

	// Handle reload key press
	if rl.IsKeyPressed(rl.KeyR) {
		p.Reload(currentTime)
	}

	// Update reload progress
	if p.isReloading {
		// Check if reload is complete
		if currentTime >= p.reloadStartTime+p.currentWeapon.reloadTime {
			p.isReloading = false

			// Calculate how many bullets to add to magazine
			bulletsNeeded := p.currentWeapon.magazineSize - p.currentMagazine

			if p.currentWeapon.usesAmmo {
				// If we have enough ammo, add full magazine
				if p.ammo >= bulletsNeeded {
					p.ammo -= bulletsNeeded
					p.currentMagazine = p.currentWeapon.magazineSize
				} else {
					// Otherwise add whatever we have left
					p.currentMagazine += p.ammo
					p.ammo = 0
				}
			} else {
				// If weapon doesn't use ammo, just fill the magazine
				p.currentMagazine = p.currentWeapon.magazineSize
			}
		}
	}

	// Determine sprite direction based on mouse position if not moving
	if !moving && p.lookAtSet {
		// If looking more left than right, face left
		if p.lookAt.X < 0 {
			p.facingLeft = true
		} else {
			p.facingLeft = false
		}
	}
}

// New method that handles everything in Update except for movement
func (p *player) UpdateWithoutMovement(dt float64, currentTime float64) {
	// Handle reload key press
	if rl.IsKeyPressed(rl.KeyR) {
		p.Reload(currentTime)
	}

	// Update reload progress
	if p.isReloading {
		// Check if reload is complete
		if currentTime >= p.reloadStartTime+p.currentWeapon.reloadTime {
			p.isReloading = false

			// Calculate how many bullets to add to magazine
			bulletsNeeded := p.currentWeapon.magazineSize - p.currentMagazine

			if p.currentWeapon.usesAmmo {
				// If we have enough ammo, add full magazine
				if p.ammo >= bulletsNeeded {
					p.ammo -= bulletsNeeded
					p.currentMagazine = p.currentWeapon.magazineSize
				} else {
					// Otherwise add whatever we have left
					p.currentMagazine += p.ammo
					p.ammo = 0
				}
			} else {
				// If weapon doesn't use ammo, just fill the magazine
				p.currentMagazine = p.currentWeapon.magazineSize
			}
		}
	}

	// Note: Sprite direction logic is now handled in the main game loop
}

func (p *player) Render() {
	// Check if sprites were loaded successfully
	spritesLoaded := p.spriteLeft.ID > 0 && p.spriteRight.ID > 0

	if spritesLoaded {
		// Render player sprite
		sprite := p.spriteRight
		if p.facingLeft {
			sprite = p.spriteLeft
		}

		// Size to draw the sprite (scale it according to playerSize)
		spriteScale := playerSize / float32(sprite.Height) * 4.0
		width := float32(sprite.Width) * spriteScale
		height := float32(sprite.Height) * spriteScale

		// Draw the sprite centered on player position
		rl.DrawTexturePro(
			sprite,
			rl.NewRectangle(0, 0, float32(sprite.Width), float32(sprite.Height)),
			rl.NewRectangle(p.Pos.X-width/2, p.Pos.Y-height/2, width, height),
			rl.NewVector2(0, 0),
			0,
			rl.White,
		)

		// Draw health bar above player
		healthBarY := p.Pos.Y - height/2 - 10
		drawHealthBar(p, healthBarY)
	} else {
		// Fallback to circle if sprites not loaded
		rl.DrawCircle(int32(p.Pos.X), int32(p.Pos.Y), playerSize*1.6, rl.Red)

		// Draw health bar above circle
		healthBarY := p.Pos.Y - playerSize*1.6 - 10
		drawHealthBar(p, healthBarY)
	}

	// Draw direction indicator if needed
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
}

// Helper function to draw health bar
func drawHealthBar(p *player, yPosition float32) {
	healthBarWidth := playerSize * 3 // Increased from 2 to 3 for wider health bar
	healthBarHeight := 6.0           // Increased from 5.0 to 6.0 for taller health bar
	healthPercentage := float32(p.CurrentHp) / float32(p.TotalHp)

	// Background of health bar
	rl.DrawRectangle(
		int32(p.Pos.X-healthBarWidth/2),
		int32(yPosition),
		int32(healthBarWidth),
		int32(healthBarHeight),
		rl.DarkGray,
	)

	// Actual health
	rl.DrawRectangle(
		int32(p.Pos.X-healthBarWidth/2),
		int32(yPosition),
		int32(healthBarWidth*healthPercentage),
		int32(healthBarHeight),
		rl.Red,
	)
}

func (p *player) Shoot() []*Projectile {
	// Can't shoot while reloading
	if p.isReloading {
		return nil
	}

	// Switch to pistol if out of ammo and trying to use a weapon that requires ammo
	if p.currentMagazine <= 0 && p.ammo <= 0 && p.currentWeapon.usesAmmo {
		p.currentWeapon = PISTOL
		p.currentMagazine = PISTOL.magazineSize
		p.weaponPickupTime = rl.GetTime()
		p.weaponPickupName = "Pistol (Out of ammo!)"
	}

	var projs []*Projectile

	// Only shoot if we have ammo in magazine
	if p.currentMagazine > 0 || !p.currentWeapon.usesAmmo {
		for i := 0; i < p.currentWeapon.nProj; i++ {
			noise := rl.GetRandomValue(-100, 100)
			noisedDirection := rl.Vector2Add(rl.GetMousePosition(), rl.NewVector2(float32(noise), float32(noise)))
			projs = append(projs, NewProj(p.Pos, noisedDirection, p.currentWeapon.projDamage))
		}

		// Consume ammo from magazine if this weapon uses it
		if p.currentWeapon.usesAmmo && len(projs) > 0 {
			p.currentMagazine--
		}
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
		// Use a slightly smaller collision radius than the sprite size
		// This makes the collision feel more fair
		collisionRadius := playerSize * 0.7
		return rl.CheckCollisionCircles(p.Pos, collisionRadius, enemy.pos, enemy.bodyRadius)
	}
	return false
}

func (p *player) Rearrange(other Collides) {
	// Player doesn't need to rearrange since enemy collision will push the player back
}

// Attempt to reload the current weapon
func (p *player) Reload(currentTime float64) bool {
	// Don't reload if already reloading
	if p.isReloading {
		return false
	}

	// Don't reload if magazine is full
	if p.currentMagazine >= p.currentWeapon.magazineSize {
		return false
	}

	// Don't reload if pistol (infinite ammo)
	if !p.currentWeapon.usesAmmo {
		p.currentMagazine = p.currentWeapon.magazineSize
		return true
	}

	// Don't reload if no ammo in inventory
	if p.ammo <= 0 {
		return false
	}

	// Start reload
	p.isReloading = true
	p.reloadStartTime = currentTime
	return true
}
