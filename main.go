package main

import (
	"fmt"

	rl "github.com/gen2brain/raylib-go/raylib"
)

const (
	MAX_COLLISION_ORDERING_ITERS = 3
	SPACE_GRID_WIDTH             = 70
	SPACE_GRID_HEIGHT            = 70
)

var (
	playerSize        float32
	enemySize         float32
	projSize          float32
	lootSize          float32
	backgroundTexture rl.Texture2D // Background texture
	bloodTexture      rl.Texture2D // Blood texture
)

type WorldItem interface {
	Render()
	// Update()
	Destroyed() bool
}

type Blood struct {
	pos       rl.Vector2
	destroyed bool
	rotation  float32
	scale     float32
	opacity   float32
	fading    bool
}

func NewBlood(pos rl.Vector2) *Blood {
	// Random rotation between 0-360 degrees for variation
	rotation := float32(rl.GetRandomValue(0, 359))
	// Random scale between 0.8 and 1.2 for size variation
	scale := 0.8 + float32(rl.GetRandomValue(0, 40))/100.0
	// Full opacity
	opacity := float32(1.0)

	return &Blood{
		pos:       pos,
		destroyed: false,
		rotation:  rotation,
		scale:     scale,
		opacity:   opacity,
		fading:    false,
	}
}

func (b *Blood) Update(dt float64) {
	if b.fading {
		// Fade out gradually when level ends
		fadeSpeed := float32(dt) * 0.5 // Adjust speed as needed
		b.opacity -= fadeSpeed

		if b.opacity <= 0 {
			b.opacity = 0
			b.destroyed = true
		}
	}
}

func (b *Blood) Render() {
	if bloodTexture.ID > 0 {
		// Size to draw the blood
		baseSize := float32(50)
		width := baseSize * b.scale
		height := baseSize * b.scale

		// Draw the blood sprite with rotation and transparency
		rl.DrawTexturePro(
			bloodTexture,
			rl.NewRectangle(0, 0, float32(bloodTexture.Width), float32(bloodTexture.Height)),
			rl.NewRectangle(b.pos.X-width/2, b.pos.Y-height/2, width, height),
			rl.NewVector2(width/2, height/2),
			b.rotation,
			rl.ColorAlpha(rl.White, b.opacity),
		)
	} else {
		// Fallback if texture not loaded
		rl.DrawCircle(int32(b.pos.X), int32(b.pos.Y), 10, rl.ColorAlpha(rl.Maroon, b.opacity))
	}
}

func (b *Blood) Destroyed() bool {
	return b.destroyed
}

func RemoveIndex[T any](s []T, index int) []T {
	return append(s[:index], s[index+1:]...)
}

func UpdateWorldItems[T WorldItem](worldItems []T) []T {
	toRemove := make([]int, 0)
	for i := len(worldItems) - 1; i >= 0; i-- {
		if worldItems[i].Destroyed() {
			toRemove = append(toRemove, i)
		}
	}
	for _, i := range toRemove {
		worldItems = RemoveIndex(worldItems, i)
	}
	return worldItems
}

func RandomPointInCircle(radius float32) rl.Vector2 {
	x := float32(rl.GetRandomValue(-100, 100))
	y := float32(rl.GetRandomValue(-100, 100))
	vector := rl.Vector2Scale((rl.Vector2Normalize(rl.NewVector2(x, y))), radius)
	return vector
}

// Calculate enemies for level: base amount + level increment
func getEnemiesForLevel(level int) int {
	baseEnemies := 5
	enemiesPerLevel := 5
	return baseEnemies + (level-1)*enemiesPerLevel
}

// Calculate enemy health for level
func getEnemyHealthForLevel(level int) float32 {
	baseHealth := float32(100)
	healthIncreasePerLevel := float32(20)
	return baseHealth + float32(level-1)*healthIncreasePerLevel
}

// Calculate enemy damage for level
func getEnemyDamageForLevel(level int) float32 {
	baseDamage := float32(10)
	damageIncreasePerLevel := float32(5)
	return baseDamage + float32(level-1)*damageIncreasePerLevel
}

// Get the name of a weapon as a string
func getWeaponName(w weapon) string {
	return w.weaponName
}

func main() {
	display := rl.GetCurrentMonitor()

	w := rl.GetMonitorWidth(display)
	h := rl.GetMonitorHeight(display)

	rl.InitWindow(int32(w), int32(h), "Survivor")
	rl.SetTargetFPS(60)

	// Print debugging info about sprite loading
	rl.TraceLog(rl.LogWarning, "Looking for sprite files: player_left.png, player_right.png, and zombie.png")

	// Initialize enemy sprite
	InitEnemySprite()

	// Initialize bullet sprite
	InitBulletSprite()

	// Load background texture
	backgroundTexture = rl.LoadTexture("assets/background-dark.png")
	rl.TraceLog(rl.LogInfo, "Loaded background texture: %dx%d", backgroundTexture.Width, backgroundTexture.Height)

	// Load blood texture
	bloodTexture = rl.LoadTexture("assets/blod.png")
	rl.TraceLog(rl.LogInfo, "Loaded blood texture: %dx%d", bloodTexture.Width, bloodTexture.Height)

	// Check if texture loaded correctly
	if backgroundTexture.Width == 0 || backgroundTexture.Height == 0 {
		rl.TraceLog(rl.LogError, "Failed to load background texture or dimensions are zero")
		// Set default texture size to prevent division by zero
		backgroundTexture.Width = 8
		backgroundTexture.Height = 8
	}

	player := NewPlayer(1000)

	var projList []*Projectile
	var enemyList []*Enemy
	var grenadeList []*Grenade
	var worldItems []WorldItem
	var worldBodies []Collides
	var loots []*WeaponLoot
	var ammoLoots []*AmmoLoot
	var grenadePickups []*GrenadePickup
	var bloodList []*Blood

	// Level system variables
	currentLevel := 1
	enemiesRemaining := getEnemiesForLevel(currentLevel)
	enemiesInPlay := 0
	maxConcurrentEnemies := 5 // Maximum enemies on screen at once
	levelCompleted := false
	levelCompletedTime := 0.0
	levelCompletedDuration := 2.0 // Show level complete message for 2 seconds

	lastTime := rl.GetTime()
	lastEnemySpawn := lastTime
	enemySpawnDelay := 1.0 // Seconds between enemy spawns

	lastAmmoSpawn := lastTime
	ammoSpawnDelay := 3.0 // Spawn ammo every 3 seconds

	lastShoot := lastTime
	lastGrenade := lastTime
	grenadeDelay := 1.0

	lastGrenadePickupSpawn := lastTime
	grenadePickupDelay := 10.0 // Spawn grenade pickup every 10 seconds

	w = rl.GetMonitorWidth(display)
	h = rl.GetMonitorHeight(display)

	// have to be scaled based on screen size
	playerSize = float32(w) / 120
	projSize = float32(w) / 1000
	enemySize = float32(w) / 120
	lootSize = 60

	spaceGrid := NewCollisionSpace(w, h, SPACE_GRID_WIDTH, SPACE_GRID_HEIGHT)

	// Add player to worldBodies so it can be used for collisions
	worldBodies = append(worldBodies, &player)

	for !rl.WindowShouldClose() {
		currentTime := rl.GetTime()
		dt := currentTime - lastTime

		mousePosition := rl.GetMousePosition()
		player.LookAt(mousePosition)
		player.Update(dt, currentTime)

		// Check if player is dead
		if player.CurrentHp <= 0 {
			rl.BeginDrawing()
			rl.ClearBackground(rl.Black)
			gameOverText := "GAME OVER"
			textWidth := rl.MeasureText(gameOverText, 60)
			rl.DrawText(gameOverText, int32(w)/2-textWidth/2, int32(h)/2-30, 60, rl.Red)

			restartText := "Press R to restart"
			restartWidth := rl.MeasureText(restartText, 30)
			rl.DrawText(restartText, int32(w)/2-restartWidth/2, int32(h)/2+40, 30, rl.White)
			rl.EndDrawing()

			if rl.IsKeyPressed(rl.KeyR) {
				// Reset the game
				player = NewPlayer(1000)
				enemyList = make([]*Enemy, 0)
				projList = make([]*Projectile, 0)
				grenadeList = make([]*Grenade, 0)
				worldItems = make([]WorldItem, 0)
				worldBodies = make([]Collides, 0)
				loots = make([]*WeaponLoot, 0)
				ammoLoots = make([]*AmmoLoot, 0)
				grenadePickups = make([]*GrenadePickup, 0)
				bloodList = make([]*Blood, 0)
				currentLevel = 1
				enemiesRemaining = getEnemiesForLevel(currentLevel)
				enemiesInPlay = 0
				levelCompleted = false
				worldBodies = append(worldBodies, &player)
			}

			lastTime = currentTime
			continue
		}

		// Check if level is completed
		if len(enemyList) == 0 && enemiesRemaining == 0 && !levelCompleted {
			levelCompleted = true
			levelCompletedTime = currentTime
			currentLevel++
			enemiesRemaining = getEnemiesForLevel(currentLevel)
		}

		// Handle level transition
		if levelCompleted {
			// Reset when transition time is over
			if currentTime > levelCompletedTime+levelCompletedDuration {
				levelCompleted = false

				// Start fading out all blood when level ends
				for i := range bloodList {
					bloodList[i].fading = true
				}
			}
		}

		// Spawn weapon
		{
			if rl.GetRandomValue(0, 1000) < 1 {
				x := rl.GetRandomValue(0, int32(w))
				y := rl.GetRandomValue(0, int32(h))

				// Pick a random weapon
				weaponType := rl.GetRandomValue(0, 2)
				var selectedWeapon weapon
				switch weaponType {
				case 0:
					selectedWeapon = MITRA
				case 1:
					selectedWeapon = SHOTGUN
				case 2:
					selectedWeapon = MINIGUN
				default:
					selectedWeapon = PISTOL
				}

				loot := NewWeaponLoot(selectedWeapon, rl.NewVector2(float32(x), float32(y)), currentTime)
				worldBodies = append(worldBodies, loot)
				worldItems = append(worldItems, loot)
				loots = append(loots, loot)
			}
		}

		// Spawn ammo
		{
			if currentTime > lastAmmoSpawn+ammoSpawnDelay {
				lastAmmoSpawn = currentTime

				if rl.GetRandomValue(0, 100) < 30 { // 30% chance to spawn ammo
					x := rl.GetRandomValue(0, int32(w))
					y := rl.GetRandomValue(0, int32(h))

					// Random ammo amount between 50-200
					ammoAmount := rl.GetRandomValue(50, 200)

					ammo := NewAmmoLoot(int(ammoAmount), rl.NewVector2(float32(x), float32(y)), currentTime)
					worldBodies = append(worldBodies, ammo)
					worldItems = append(worldItems, ammo)
					ammoLoots = append(ammoLoots, ammo)
				}
			}
		}

		// Collision with loot
		{
			// Check for loot timeout (10 seconds)
			for i := len(loots) - 1; i >= 0; i-- {
				// Check if this loot has been around for more than 10 seconds
				if currentTime-loots[i].createTime > 10.0 && !loots[i].destroyed {
					loots[i].destroyed = true
				}
			}

			for _, l := range loots {
				if rl.CheckCollisionCircleRec(player.Pos, playerSize*0.7, rl.NewRectangle(l.pos.X, l.pos.Y, lootSize, lootSize)) {
					player.currentWeapon = l.weapon
					player.isReloading = false // Cancel any reload in progress

					// Initialize magazine for the new weapon
					if player.currentWeapon.usesAmmo {
						// If we have enough ammo, fill the magazine
						if player.ammo >= player.currentWeapon.magazineSize {
							player.currentMagazine = player.currentWeapon.magazineSize
							player.ammo -= player.currentWeapon.magazineSize
						} else {
							// Otherwise use what we have
							player.currentMagazine = player.ammo
							player.ammo = 0
						}
					} else {
						// Pistol always has full magazine
						player.currentMagazine = player.currentWeapon.magazineSize
					}

					l.destroyed = true

					// Store pickup message details
					player.weaponPickupTime = currentTime
					player.weaponPickupName = getWeaponName(l.weapon)
				}
			}
		}

		// Collision with ammo
		{
			// Check for ammo timeout (10 seconds)
			for i := len(ammoLoots) - 1; i >= 0; i-- {
				// Check if this ammo has been around for more than 10 seconds
				if currentTime-ammoLoots[i].createTime > 10.0 && !ammoLoots[i].destroyed {
					ammoLoots[i].destroyed = true
				}
			}

			// Handle ammo pickup
			for _, a := range ammoLoots {
				if !a.destroyed && rl.CheckCollisionCircleRec(player.Pos, playerSize*0.7, rl.NewRectangle(a.pos.X, a.pos.Y, lootSize, lootSize)) {
					player.ammo += a.amount
					a.destroyed = true

					// Store pickup message details
					player.ammoPickupTime = currentTime
					player.ammoPickupAmount = a.amount
				}
			}
		}

		// shoot
		{
			if rl.IsMouseButtonDown(0) {
				if currentTime > player.currentWeapon.shootingDelay+float64(lastShoot) {
					lastShoot = currentTime
					for _, p := range player.Shoot() {
						projList = append(projList, p)
						worldItems = append(worldItems, p)
					}
				}
			}
		}

		// Place grenade when 'E' is pressed
		{
			if rl.IsKeyPressed(rl.KeyE) && currentTime > lastGrenade+grenadeDelay && player.grenades > 0 {
				lastGrenade = currentTime

				// Create new grenade at player position
				grenade := NewGrenade(player.Pos, currentTime)
				grenadeList = append(grenadeList, grenade)
				worldItems = append(worldItems, grenade)

				// Decrease player's grenade count
				player.grenades--
			}
		}

		// Spawn enemies for current level
		{
			if !levelCompleted && enemiesRemaining > 0 && enemiesInPlay < maxConcurrentEnemies && currentTime > lastEnemySpawn+enemySpawnDelay {
				lastEnemySpawn = currentTime

				respawn := true
				enemyHealth := getEnemyHealthForLevel(currentLevel)
				enemyDamage := getEnemyDamageForLevel(currentLevel)
				n := NewEnemy(rl.NewVector2(100, 110), enemyHealth, enemyDamage, enemySize)

				for respawn {
					respawn = false
					spawnPosition := RandomPointInCircle(200)
					spawnPosition = rl.Vector2Add(spawnPosition, player.Pos)

					n = NewEnemy(rl.NewVector2(float32(spawnPosition.X), float32(spawnPosition.Y)), enemyHealth, enemyDamage, enemySize)
					if n.pos.X < 100 || n.pos.Y < 100 {
						respawn = true
						continue
					}
					for _, e := range enemyList {
						if rl.CheckCollisionCircles(n.pos, enemySize, e.pos, enemySize) {
							respawn = true
							break
						}
					}

					if !respawn {
						enemyList = append(enemyList, n)
						worldItems = append(worldItems, n)
						worldBodies = append(worldBodies, n)
						enemiesRemaining--
						enemiesInPlay++
					}
				}
			}
		}

		// move projectile
		for _, p := range projList {
			p.Update(dt)
		}

		// Update grenades
		for _, g := range grenadeList {
			g.Update(currentTime, enemyList)
		}

		// Check for enemies killed by grenades
		for i := len(enemyList) - 1; i >= 0; i-- {
			if enemyList[i].health <= 0 && !enemyList[i].destroyed {
				enemyList[i].destroyed = true
				enemiesInPlay--

				// Add blood at enemy position when it dies from projectile hit
				blood := NewBlood(enemyList[i].pos)
				worldItems = append(worldItems, blood)
				bloodList = append(bloodList, blood)
			}
		}

		// move enemy
		for _, p := range enemyList {
			p.Move(player.Pos, dt)
		}

		// Spawn grenade pickups
		{
			if currentTime > lastGrenadePickupSpawn+grenadePickupDelay {
				lastGrenadePickupSpawn = currentTime

				if rl.GetRandomValue(0, 100) < 40 { // 40% chance to spawn grenade pickup
					x := rl.GetRandomValue(0, int32(w))
					y := rl.GetRandomValue(0, int32(h))

					pickup := NewGrenadePickup(rl.NewVector2(float32(x), float32(y)), currentTime)
					worldItems = append(worldItems, pickup)
					grenadePickups = append(grenadePickups, pickup)
				}
			}
		}

		// Check for grenades timeout (similar to ammo pickup timeout)
		{
			for i := len(grenadePickups) - 1; i >= 0; i-- {
				// Check if this pickup has been around for more than 10 seconds
				if currentTime-grenadePickups[i].createTime > 10.0 && !grenadePickups[i].destroyed {
					grenadePickups[i].destroyed = true
				}
			}

			// Handle grenade pickup
			for _, g := range grenadePickups {
				if !g.destroyed && rl.CheckCollisionCircleRec(player.Pos, playerSize*0.7,
					rl.NewRectangle(g.pos.X, g.pos.Y, float32(g.size), float32(g.size))) {
					player.grenades += g.amount
					g.destroyed = true
				}
			}
		}

		type CollisionPair struct {
			first  *Enemy
			second *Enemy
		}

		spaceGrid.RearrangeBodies(MAX_COLLISION_ORDERING_ITERS, worldBodies, func() {
			spaceGrid.UpdateCells(worldBodies)
		})

		// check collision between proj and enemy
		for _, p := range projList {
			for _, e := range enemyList {
				if rl.CheckCollisionCircles(p.pos, projSize, e.pos, enemySize) {
					e.DealDamage(p.damage)
					if e.health <= 0 {
						e.destroyed = true
						enemiesInPlay--

						// Add blood at enemy position when it dies from projectile hit
						blood := NewBlood(e.pos)
						worldItems = append(worldItems, blood)
						bloodList = append(bloodList, blood)
					}
					p.destroyed = true
				}
			}
		}

		// Check collision between enemies and player
		for _, e := range enemyList {
			if rl.CheckCollisionCircles(player.Pos, playerSize*0.7, e.pos, enemySize) {
				// Apply damage to player based on enemy's damage stat
				player.TakeDamage(e.damage)

				// Simple invulnerability frame mechanic by slightly pushing enemy away
				dir := rl.Vector2Subtract(e.pos, player.Pos)
				if dir.X == 0 && dir.Y == 0 {
					dir = rl.NewVector2(float32(rl.GetRandomValue(-10, 10))*0.1,
						float32(rl.GetRandomValue(-10, 10))*0.1)
				}
				dir = rl.Vector2Normalize(dir)
				pushDistance := float32(10.0) // Slight push
				pushVector := rl.Vector2Scale(dir, pushDistance)
				e.pos = rl.Vector2Add(e.pos, pushVector)
			}
		}

		rl.BeginDrawing()
		{
			rl.ClearBackground(rl.Black)

			// Draw tiled background - only if texture was loaded properly
			if backgroundTexture.ID > 0 {
				tileWidth := backgroundTexture.Width
				tileHeight := backgroundTexture.Height

				// Ensure tile dimensions are not zero to avoid division by zero
				if tileWidth > 0 && tileHeight > 0 {
					tilesX := int(w)/int(tileWidth) + 1
					tilesY := int(h)/int(tileHeight) + 1

					for y := 0; y < tilesY; y++ {
						for x := 0; x < tilesX; x++ {
							rl.DrawTexture(backgroundTexture,
								int32(x)*tileWidth,
								int32(y)*tileHeight,
								rl.White)
						}
					}
				}
			} else {
				// Just draw a background color if texture failed to load
				rl.ClearBackground(rl.DarkGray)
			}

			// Draw blood first (so it's underneath everything else)
			for _, blood := range bloodList {
				blood.Render()
			}

			// Then draw other world items (but skip blood which we already drew)
			for _, item := range worldItems {
				if _, isBlood := item.(*Blood); !isBlood {
					item.Render()
				}
			}

			player.Render()
			rl.DrawFPS(10, 10)

			// Draw player HP in the top-right corner of the screen
			healthText := fmt.Sprintf("HP: %d/%d", player.CurrentHp, player.TotalHp)
			textWidth := rl.MeasureText(healthText, 20)
			rl.DrawText(healthText, int32(w)-textWidth-20, 20, 20, rl.White)

			// Draw ammo count
			var ammoText string
			if player.currentWeapon.usesAmmo {
				ammoText = fmt.Sprintf("Ammo: %d / %d", player.currentMagazine, player.ammo)
			} else {
				ammoText = fmt.Sprintf("Ammo: ∞ / %d", player.ammo) // Infinite for pistol
			}
			ammoWidth := rl.MeasureText(ammoText, 20)
			rl.DrawText(ammoText, int32(w)-ammoWidth-20, 50, 20, rl.White)

			// Show current weapon
			weaponText := fmt.Sprintf("Weapon: %s", player.currentWeapon.weaponName)
			weaponWidth := rl.MeasureText(weaponText, 20)
			rl.DrawText(weaponText, int32(w)-weaponWidth-20, 80, 20, rl.White)

			// Show reload key hint if magazine not full
			if player.currentMagazine < player.currentWeapon.magazineSize &&
				!player.isReloading &&
				(player.ammo > 0 || !player.currentWeapon.usesAmmo) {
				reloadText := "Press R to reload"
				reloadWidth := rl.MeasureText(reloadText, 18)
				rl.DrawText(reloadText, int32(w)-reloadWidth-20, 110, 18, rl.Gray)
			}

			// Show reload state if reloading
			if player.isReloading {
				reloadProgress := (currentTime - player.reloadStartTime) / player.currentWeapon.reloadTime * 100
				reloadText := fmt.Sprintf("RELOADING... %.0f%%", reloadProgress)
				reloadWidth := rl.MeasureText(reloadText, 25)
				rl.DrawText(reloadText, int32(w)/2-reloadWidth/2, int32(h)-120, 25, rl.Yellow)
			}

			// Draw level information
			levelText := fmt.Sprintf("Level: %d", currentLevel)
			rl.DrawText(levelText, 10, 40, 20, rl.White)

			enemiesText := fmt.Sprintf("Enemies remaining: %d", enemiesRemaining+len(enemyList))
			rl.DrawText(enemiesText, 10, 70, 20, rl.White)

			// Show level complete message
			if levelCompleted {
				levelCompleteText := fmt.Sprintf("LEVEL %d COMPLETE!", currentLevel-1)
				completeTextWidth := rl.MeasureText(levelCompleteText, 40)
				rl.DrawText(levelCompleteText, int32(w)/2-completeTextWidth/2, int32(h)/2-20, 40, rl.Yellow)

				nextLevelText := fmt.Sprintf("NEXT LEVEL: %d", currentLevel)
				nextLevelWidth := rl.MeasureText(nextLevelText, 30)
				rl.DrawText(nextLevelText, int32(w)/2-nextLevelWidth/2, int32(h)/2+30, 30, rl.Green)
			}

			// Show weapon pickup message for 2 seconds
			if player.weaponPickupName != "" && currentTime-player.weaponPickupTime < 2.0 {
				pickupText := fmt.Sprintf("Acquired: %s", player.weaponPickupName)
				textWidth := rl.MeasureText(pickupText, 30)
				rl.DrawText(pickupText, int32(w)/2-textWidth/2, int32(h)-50, 30, rl.Yellow)
			}

			// Show ammo pickup message for 2 seconds
			if player.ammoPickupAmount > 0 && currentTime-player.ammoPickupTime < 2.0 {
				ammoText := fmt.Sprintf("Ammo +%d", player.ammoPickupAmount)
				textWidth := rl.MeasureText(ammoText, 30)
				rl.DrawText(ammoText, int32(w)/2-textWidth/2, int32(h)-90, 30, rl.Yellow)
			}

			// Show grenade key hint
			grenadeText := "Press E to place grenade"
			grenadeWidth := rl.MeasureText(grenadeText, 18)
			rl.DrawText(grenadeText, int32(w)-grenadeWidth-20, 140, 18, rl.Gray)

			// Show grenade count
			grenadeCountText := fmt.Sprintf("Grenades: %d", player.grenades)
			grenadeCountWidth := rl.MeasureText(grenadeCountText, 20)
			rl.DrawText(grenadeCountText, int32(w)-grenadeCountWidth-20, 170, 20, rl.White)
		}
		rl.EndDrawing()

		{
			worldItems = UpdateWorldItems(worldItems)
			projList = UpdateWorldItems(projList)
			enemyList = UpdateWorldItems(enemyList)
			loots = UpdateWorldItems(loots)
			ammoLoots = UpdateWorldItems(ammoLoots)
			grenadeList = UpdateWorldItems(grenadeList)
			grenadePickups = UpdateWorldItems(grenadePickups)

			// Update blood (for fading effect)
			for _, blood := range bloodList {
				blood.Update(dt)
			}
			bloodList = UpdateWorldItems(bloodList)
		}
		lastTime = currentTime
	}

	// Unload textures before closing
	rl.UnloadTexture(player.spriteLeft)
	rl.UnloadTexture(player.spriteRight)
	rl.UnloadTexture(backgroundTexture)
	rl.UnloadTexture(bloodTexture)
	UnloadEnemySprite()
	UnloadBulletSprite()

	rl.CloseWindow()
}
