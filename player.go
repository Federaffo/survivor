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

	lookAt    rl.Vector2
	lookAtSet bool

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
	return player{
		TotalHp:          totalHp,
		CurrentHp:        totalHp,
		Pos:              rl.NewVector2(200, 200),
		lookAt:           rl.NewVector2(0, 0),
		currentWeapon:    PISTOL,
		lookAtSet:        false,
		weaponPickupName: "",
		weaponPickupTime: 0,
		ammo:             0,
		currentMagazine:  PISTOL.magazineSize, // Start with full magazine
		isReloading:      false,
		ammoPickupTime:   0,
		ammoPickupAmount: 0,
	}
}

func (p *player) LookAt(lookAt rl.Vector2) {
	p.lookAtSet = true
	p.lookAt = rl.Vector2Normalize(rl.Vector2Subtract(lookAt, p.Pos))
}

func (p *player) Update(dt float64, currentTime float64) {
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
			noise := rl.GetRandomValue(-50, 50)
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
		return rl.CheckCollisionCircles(p.Pos, playerSize, enemy.pos, enemy.bodyRadius)
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
