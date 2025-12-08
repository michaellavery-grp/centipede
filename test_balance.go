// +build ignore

// Test harness for Centipede game balance analysis
// Simulates 1,000 games to analyze difficulty and player experience
// Build with: go run test_balance.go main.go
package main

import (
	"fmt"
	"math"
	"math/rand"
	"sort"
	"time"
)

// TestStats tracks metrics for a single game
type TestStats struct {
	score              int
	livesLost          int
	levelsCompleted    int
	segmentsDestroyed  int
	fliesHit           int
	mushroomsDestroyed int
	ticksAlive         int
	deathsByPoison     int
	bonusLivesEarned   int
	finalLevel         int
}

// AggregateStats summarizes 1,000 games
type AggregateStats struct {
	totalGames         int
	avgScore           float64
	medianScore        float64
	avgLivesLost       float64
	avgLevelsCompleted float64
	avgSurvivalTime    float64
	tooEasy            int // Games where player survived 10+ levels
	tooHard            int // Games where player died in level 1
	balanced           int // Games with 2-9 levels completed
	avgDeathsByPoison  float64
	poisonDeathRate    float64
	scores             []int
}

// SimulateGame runs a single automated game with AI player
func SimulateGame(gameNum int) TestStats {
	g := NewGame(50, 28)
	stats := TestStats{}

	// AI strategy parameters
	dodgeRange := 5        // How far to look ahead for threats
	shootChance := 0.7     // Probability to shoot when enemy nearby
	panicMode := false     // When centipede gets close

	maxTicks := 10000 // Prevent infinite games

	for tick := 0; tick < maxTicks && !g.gameOver; tick++ {
		stats.ticksAlive++

		// Check if we're in danger (centipede within dodgeRange rows)
		panicMode = false
		for _, seg := range g.segments {
			if seg.pos.Y >= g.height-dodgeRange {
				panicMode = true
				break
			}
		}

		// AI Decision Making
		if panicMode {
			// PANIC MODE: Focus on dodging
			aiPanicDodge(g, &stats)
			if rand.Float64() < 0.9 { // Shoot more aggressively
				g.Shoot()
			}
		} else {
			// NORMAL MODE: Balanced strategy
			aiNormalPlay(g, &stats, shootChance)
		}

		// Update game state
		g.Update()

		// Track statistics
		if g.level > stats.finalLevel {
			stats.levelsCompleted++
			stats.finalLevel = g.level
		}

		// Check for life loss
		if g.lives < 3-stats.livesLost {
			stats.livesLost++
			// Check if death was due to poison mushroom
			for _, seg := range g.segments {
				if seg.pos.Y >= g.height-3 {
					for _, mush := range g.mushrooms {
						if mush.poisoned && seg.pos.Y == mush.pos.Y {
							stats.deathsByPoison++
							break
						}
					}
				}
			}
		}

		// Prevent infinite loops
		if tick >= maxTicks-1 {
			break
		}
	}

	// Final stats
	stats.score = g.score
	stats.segmentsDestroyed = countDestroyedSegments(g)
	stats.bonusLivesEarned = (g.score / 10000)

	return stats
}

// AI strategy for panic mode - aggressive dodging
func aiPanicDodge(g *Game, stats *TestStats) {
	// Find nearest threat
	nearestDist := 999
	nearestX := -1

	for _, seg := range g.segments {
		if seg.pos.Y >= g.height-10 {
			dist := abs(seg.pos.X - g.player.pos.X)
			if dist < nearestDist {
				nearestDist = dist
				nearestX = seg.pos.X
			}
		}
	}

	if nearestX != -1 {
		// Move away from threat
		if g.player.pos.X < nearestX {
			g.MovePlayer(-1) // Move left
		} else if g.player.pos.X > nearestX {
			g.MovePlayer(1) // Move right
		}

		// Try to move up if possible
		if g.player.pos.Y > g.height-6 {
			g.MovePlayerY(-1)
		}
	}
}

// AI strategy for normal play - balanced offense/defense
func aiNormalPlay(g *Game, stats *TestStats, shootChance float64) {
	// Target priority: Head > Flies > Body segments
	targetX := -1
	targetValue := 0

	// Look for head
	for _, seg := range g.segments {
		if seg.isHead && seg.pos.X == g.player.pos.X {
			if targetValue < 100 {
				targetX = seg.pos.X
				targetValue = 100
			}
		}
	}

	// Look for flies
	for _, fly := range g.flies {
		if fly.active && abs(fly.pos.X-g.player.pos.X) < 3 {
			if targetValue < 50 {
				targetX = fly.pos.X
				targetValue = 50
			}
		}
	}

	// Look for any segment above us
	if targetValue == 0 {
		for _, seg := range g.segments {
			if seg.pos.X == g.player.pos.X {
				targetX = seg.pos.X
				targetValue = 10
				break
			}
		}
	}

	// Move toward target or hunt
	if targetValue > 0 {
		if g.player.pos.X < targetX {
			g.MovePlayer(1)
		} else if g.player.pos.X > targetX {
			g.MovePlayer(-1)
		}

		// Shoot if aligned
		if rand.Float64() < shootChance {
			g.Shoot()
		}
	} else {
		// Hunt mode - random walk with shooting
		if rand.Float64() < 0.3 {
			if rand.Float64() < 0.5 {
				g.MovePlayer(1)
			} else {
				g.MovePlayer(-1)
			}
		}
		if rand.Float64() < 0.4 {
			g.Shoot()
		}
	}
}

func countDestroyedSegments(g *Game) int {
	// Estimate from score (10 per body, 100 per head)
	return g.score / 10
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// AnalyzeBalance processes all test results
func AnalyzeBalance(results []TestStats) AggregateStats {
	agg := AggregateStats{
		totalGames: len(results),
		scores:     make([]int, len(results)),
	}

	totalScore := 0
	totalLives := 0
	totalLevels := 0
	totalTicks := 0
	totalPoisonDeaths := 0
	totalDeaths := 0

	for i, stat := range results {
		totalScore += stat.score
		totalLives += stat.livesLost
		totalLevels += stat.levelsCompleted
		totalTicks += stat.ticksAlive
		totalPoisonDeaths += stat.deathsByPoison
		totalDeaths += stat.livesLost

		agg.scores[i] = stat.score

		// Categorize difficulty
		if stat.levelsCompleted >= 10 {
			agg.tooEasy++
		} else if stat.levelsCompleted <= 1 {
			agg.tooHard++
		} else {
			agg.balanced++
		}
	}

	agg.avgScore = float64(totalScore) / float64(len(results))
	agg.avgLivesLost = float64(totalLives) / float64(len(results))
	agg.avgLevelsCompleted = float64(totalLevels) / float64(len(results))
	agg.avgSurvivalTime = float64(totalTicks) / float64(len(results))
	agg.avgDeathsByPoison = float64(totalPoisonDeaths) / float64(len(results))

	if totalDeaths > 0 {
		agg.poisonDeathRate = float64(totalPoisonDeaths) / float64(totalDeaths)
	}

	// Calculate median score
	sort.Ints(agg.scores)
	agg.medianScore = float64(agg.scores[len(agg.scores)/2])

	return agg
}

// CalculateBalanceScore rates game balance from 0-100
func CalculateBalanceScore(agg AggregateStats) (float64, string) {
	score := 100.0
	feedback := []string{}

	// Ideal: 60-80% of games in balanced range
	balancedPct := float64(agg.balanced) / float64(agg.totalGames) * 100
	if balancedPct < 50 {
		penalty := (50 - balancedPct) / 2
		score -= penalty
		feedback = append(feedback, fmt.Sprintf("⚠️  Only %.1f%% balanced games (target: 60-80%%)", balancedPct))
	} else if balancedPct > 90 {
		feedback = append(feedback, fmt.Sprintf("✓ Excellent balance: %.1f%% games in 2-9 level range", balancedPct))
	}

	// Too easy check (should be < 15%)
	easyPct := float64(agg.tooEasy) / float64(agg.totalGames) * 100
	if easyPct > 15 {
		penalty := (easyPct - 15)
		score -= penalty
		feedback = append(feedback, fmt.Sprintf("⚠️  Too easy: %.1f%% reach 10+ levels (target: <15%%)", easyPct))
	}

	// Too hard check (should be < 20%)
	hardPct := float64(agg.tooHard) / float64(agg.totalGames) * 100
	if hardPct > 20 {
		penalty := (hardPct - 20) / 2
		score -= penalty
		feedback = append(feedback, fmt.Sprintf("⚠️  Too hard: %.1f%% die in level 1 (target: <20%%)", hardPct))
	}

	// Survival time (ideal: 200-400 ticks per life)
	avgTicksPerLife := agg.avgSurvivalTime / (agg.avgLivesLost + 1)
	if avgTicksPerLife < 150 {
		penalty := (150 - avgTicksPerLife) / 10
		score -= penalty
		feedback = append(feedback, fmt.Sprintf("⚠️  Deaths too quick: %.0f ticks/life (target: 200-400)", avgTicksPerLife))
	} else if avgTicksPerLife > 500 {
		penalty := (avgTicksPerLife - 500) / 20
		score -= penalty
		feedback = append(feedback, fmt.Sprintf("⚠️  Lives too long: %.0f ticks/life (target: 200-400)", avgTicksPerLife))
	}

	// Poison death rate (should be 15-30% of deaths)
	poisonPct := agg.poisonDeathRate * 100
	if poisonPct < 10 {
		feedback = append(feedback, fmt.Sprintf("⚠️  Poison mushrooms underutilized: %.1f%% of deaths", poisonPct))
		score -= 5
	} else if poisonPct > 40 {
		feedback = append(feedback, fmt.Sprintf("⚠️  Poison mushrooms too deadly: %.1f%% of deaths", poisonPct))
		score -= 10
	} else {
		feedback = append(feedback, fmt.Sprintf("✓ Poison mushrooms well-balanced: %.1f%% of deaths", poisonPct))
	}

	// Score variance (check if games feel different)
	variance := calculateVariance(agg.scores)
	stdDev := math.Sqrt(variance)
	if stdDev < agg.avgScore*0.3 {
		feedback = append(feedback, "⚠️  Games too similar - needs more randomness")
		score -= 5
	}

	feedbackStr := ""
	for _, f := range feedback {
		feedbackStr += f + "\n"
	}

	return score, feedbackStr
}

func calculateVariance(scores []int) float64 {
	sum := 0
	for _, s := range scores {
		sum += s
	}
	mean := float64(sum) / float64(len(scores))

	variance := 0.0
	for _, s := range scores {
		diff := float64(s) - mean
		variance += diff * diff
	}
	return variance / float64(len(scores))
}

func runBalanceTest() {
	rand.Seed(time.Now().UnixNano())

	fmt.Println("🐛 CENTIPEDE BALANCE TEST HARNESS")
	fmt.Println("==================================")
	fmt.Println("Simulating 1,000 games with AI player...")
	fmt.Println()

	results := make([]TestStats, 1000)

	// Progress bar
	for i := 0; i < 1000; i++ {
		results[i] = SimulateGame(i)
		if (i+1)%100 == 0 {
			fmt.Printf("Progress: %d/1000 games completed\n", i+1)
		}
	}

	fmt.Println()
	fmt.Println("Analyzing results...")
	fmt.Println()

	agg := AnalyzeBalance(results)

	// Print detailed report
	fmt.Println("📊 AGGREGATE STATISTICS")
	fmt.Println("========================")
	fmt.Printf("Total Games Simulated:  %d\n", agg.totalGames)
	fmt.Printf("Average Score:          %.0f\n", agg.avgScore)
	fmt.Printf("Median Score:           %.0f\n", agg.medianScore)
	fmt.Printf("Average Lives Lost:     %.2f / 3\n", agg.avgLivesLost)
	fmt.Printf("Average Levels Done:    %.2f\n", agg.avgLevelsCompleted)
	fmt.Printf("Avg Survival Time:      %.0f ticks (~%.1f seconds)\n",
		agg.avgSurvivalTime, agg.avgSurvivalTime*0.08)
	fmt.Println()

	fmt.Println("🎯 DIFFICULTY DISTRIBUTION")
	fmt.Println("===========================")
	fmt.Printf("Too Easy (10+ levels):  %d games (%.1f%%)\n",
		agg.tooEasy, float64(agg.tooEasy)/float64(agg.totalGames)*100)
	fmt.Printf("Balanced (2-9 levels):  %d games (%.1f%%)\n",
		agg.balanced, float64(agg.balanced)/float64(agg.totalGames)*100)
	fmt.Printf("Too Hard (0-1 levels):  %d games (%.1f%%)\n",
		agg.tooHard, float64(agg.tooHard)/float64(agg.totalGames)*100)
	fmt.Println()

	fmt.Println("☠️  DEATH ANALYSIS")
	fmt.Println("===================")
	fmt.Printf("Avg Deaths by Poison:   %.2f\n", agg.avgDeathsByPoison)
	fmt.Printf("Poison Death Rate:      %.1f%% of all deaths\n", agg.poisonDeathRate*100)
	fmt.Println()

	fmt.Println("📈 SCORE DISTRIBUTION")
	fmt.Println("=====================")
	percentiles := []int{10, 25, 50, 75, 90, 95, 99}
	for _, p := range percentiles {
		idx := (p * len(agg.scores)) / 100
		fmt.Printf("%2dth percentile:        %d\n", p, agg.scores[idx])
	}
	fmt.Println()

	// Calculate overall balance score
	balanceScore, feedback := CalculateBalanceScore(agg)

	fmt.Println("⚖️  BALANCE SCORE")
	fmt.Println("=================")
	fmt.Printf("Overall Rating: %.1f / 100\n\n", balanceScore)
	fmt.Println(feedback)

	// Recommendations
	fmt.Println("💡 RECOMMENDATIONS")
	fmt.Println("===================")

	balancedPct := float64(agg.balanced) / float64(agg.totalGames) * 100
	if balancedPct < 60 {
		fmt.Println("❌ Game needs difficulty tuning")

		hardPct := float64(agg.tooHard) / float64(agg.totalGames) * 100
		if hardPct > 25 {
			fmt.Println("   → Reduce centipede speed")
			fmt.Println("   → Increase initial lives to 4")
			fmt.Println("   → Reduce poison mushroom spawn rate")
		}

		easyPct := float64(agg.tooEasy) / float64(agg.totalGames) * 100
		if easyPct > 15 {
			fmt.Println("   → Increase centipede spawn rate")
			fmt.Println("   → Add more mushrooms per level")
			fmt.Println("   → Increase fly spawn rate")
		}
	} else if balancedPct >= 60 && balancedPct <= 80 {
		fmt.Println("✅ Game balance is GOOD - within target range!")
		fmt.Println("   → Minor tweaks may still improve player experience")
	} else {
		fmt.Println("✅ Game balance is EXCELLENT!")
		fmt.Println("   → Current difficulty curve is well-tuned")
	}

	// Frustration vs Ease analysis
	fmt.Println()
	fmt.Println("😤 FRUSTRATION VS EASE ANALYSIS")
	fmt.Println("=================================")

	avgTicksPerLife := agg.avgSurvivalTime / (agg.avgLivesLost + 1)

	if avgTicksPerLife < 150 {
		fmt.Println("⚠️  HIGH FRUSTRATION - Deaths feel unfair/too quick")
		fmt.Println("   Players don't have time to react to threats")
	} else if avgTicksPerLife > 500 {
		fmt.Println("⚠️  TOO EASY - Game feels childish/unchallenging")
		fmt.Println("   Players can survive too long without skill")
	} else if avgTicksPerLife >= 200 && avgTicksPerLife <= 350 {
		fmt.Println("✅ OPTIMAL - Good balance of challenge and fairness")
		fmt.Println("   Players have time to react but must stay alert")
	} else {
		fmt.Println("✓ ACCEPTABLE - Balance is decent but could be better")
	}

	fmt.Println()
	fmt.Printf("Target: 200-350 ticks/life | Actual: %.0f ticks/life\n", avgTicksPerLife)

	// Final verdict
	fmt.Println()
	fmt.Println("==================================================")
	if balanceScore >= 80 {
		fmt.Println("🏆 VERDICT: Game is well-balanced and ready!")
	} else if balanceScore >= 60 {
		fmt.Println("👍 VERDICT: Game is playable but needs minor tuning")
	} else {
		fmt.Println("⚠️  VERDICT: Game needs significant rebalancing")
	}
	fmt.Println("==")
}
