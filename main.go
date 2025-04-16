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
	playerSize float32
	enemySize  float32
	projSize   float32
	lootSize   float32
)

type WorldItem interface {
	Render()
	// Update()
	Destroyed() bool
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
	enemiesPerLevel := 2
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

func main() {
	display := rl.GetCurrentMonitor()

	w := rl.GetMonitorWidth(display)
	h := rl.GetMonitorHeight(display)

	rl.InitWindow(int32(w), int32(h), "Survivor")
	rl.SetTargetFPS(60)

	player := NewPlayer(1000)

	var projList []*Projectile
	var enemyList []*Enemy
	var worldItems []WorldItem
	var worldBodies []Collides
	var loots []*WeaponLoot

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

	lastShoot := lastTime

	w = rl.GetMonitorWidth(display)
	h = rl.GetMonitorHeight(display)

	// have to be scaled based on screen size
	playerSize = float32(w) / 150
	projSize = float32(w) / 1000
	enemySize = float32(w) / 200
	lootSize = 60

	spaceGrid := NewCollisionSpace(w, h, SPACE_GRID_WIDTH, SPACE_GRID_HEIGHT)

	// Add player to worldBodies so it can be used for collisions
	worldBodies = append(worldBodies, &player)

	for !rl.WindowShouldClose() {
		currentTime := rl.GetTime()
		dt := currentTime - lastTime

		mousePosition := rl.GetMousePosition()
		player.LookAt(mousePosition)
		player.Update(dt)

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
				worldItems = make([]WorldItem, 0)
				worldBodies = make([]Collides, 0)
				loots = make([]*WeaponLoot, 0)
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
			}
		}

		// Spawn weapon
		{
			if rl.GetRandomValue(0, 1000) < 3 {
				x := rl.GetRandomValue(0, int32(w))
				y := rl.GetRandomValue(0, int32(h))
				w := NewWeaponLoot(MITRA, rl.NewVector2(float32(x), float32(y)))
				worldBodies = append(worldBodies, w)
				worldItems = append(worldItems, w)
				loots = append(loots, w)
			}
		}

		// Collision with loot
		{
			for _, l := range loots {
				if rl.CheckCollisionCircleRec(player.Pos, playerSize, rl.NewRectangle(l.pos.X, l.pos.Y, lootSize, lootSize)) {
					player.currentWeapon = l.weapon
					l.destroyed = true
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
		// move enemy

		for _, p := range enemyList {
			p.Move(player.Pos, dt)
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
					}
					p.destroyed = true
				}
			}
		}

		// Check collision between enemies and player
		for _, e := range enemyList {
			if rl.CheckCollisionCircles(player.Pos, playerSize, e.pos, enemySize) {
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

			for _, x := range worldItems {
				x.Render()
			}

			player.Render()
			rl.DrawFPS(10, 10)

			// Draw player HP in the top-right corner of the screen
			healthText := fmt.Sprintf("HP: %d/%d", player.CurrentHp, player.TotalHp)
			textWidth := rl.MeasureText(healthText, 20)
			rl.DrawText(healthText, int32(w)-textWidth-20, 20, 20, rl.White)

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
		}
		rl.EndDrawing()

		{
			worldItems = UpdateWorldItems(worldItems)
			projList = UpdateWorldItems(projList)
			enemyList = UpdateWorldItems(enemyList)
		}
		lastTime = currentTime
	}

	rl.CloseWindow()
}
