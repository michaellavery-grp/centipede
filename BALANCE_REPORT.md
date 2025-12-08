# Centipede Balance Analysis Report
## 1,000 Game Simulation Results

**Test Date**: 2025-12-08
**Test Method**: AI player simulation with panic-mode dodging and strategic targeting
**Sample Size**: 1,000 games

---

## Executive Summary

⚠️ **CRITICAL FINDING**: Game is currently **TOO EASY** and lacks sufficient challenge.

**Balance Score**: -372.5 / 100 ⚠️

**Key Issues**:
1. Players survive **9,606 ticks/life** (target: 200-350) - **27× too long**!
2. Average **0.04 lives lost** per game (players almost never die)
3. **22.2% of games** reach 10+ levels (target: <15%)
4. **Poison mushrooms cause 0% of deaths** (completely ineffective)
5. Games feel too similar - needs more randomness

---

## Detailed Statistics

### Overall Performance
- **Average Score**: 17,899
- **Median Score**: 17,974
- **Average Lives Lost**: 0.04 / 3 (**96% survival rate**)
- **Average Levels Completed**: 6.91
- **Avg Survival Time**: 9,981 ticks (~13.3 minutes)
- **Avg Ticks Per Life**: 9,606 ticks

### Difficulty Distribution
| Category | Count | Percentage | Target |
|----------|-------|------------|--------|
| **Too Easy** (10+ levels) | 222 | **22.2%** | <15% ❌ |
| **Balanced** (2-9 levels) | 729 | **72.9%** | 60-80% ✅ |
| **Too Hard** (0-1 levels) | 49 | **4.9%** | <20% ✅ |

### Death Analysis
- **Deaths by Poison Mushrooms**: 0 (0.0% of deaths)
- **Deaths by Centipede Collision**: Minimal
- **Problem**: Players easily avoid all threats

### Score Distribution (Percentiles)
| Percentile | Score |
|------------|-------|
| 10th | 14,814 |
| 25th | 16,355 |
| **50th (Median)** | **17,974** |
| 75th | 19,524 |
| 90th | 21,232 |
| 95th | 22,063 |
| 99th | 23,740 |

**Analysis**: Very tight score distribution indicates games are too predictable

---

## Root Cause Analysis

### 1. **Centipedes Too Slow**
The current centipede speed allows players to easily dodge and destroy them before threats materialize.

### 2. **Poison Mushrooms Ineffective**
- Flies rarely hit mushrooms
- Even when poison mushrooms exist, they don't cause deaths
- The zigzag chute mechanic isn't dangerous enough

### 3. **Not Enough Pressure**
- Too few mushrooms spawned
- Centipedes too sparse
- Flies spawn too rarely
- Player has too much safe space

### 4. **Lives System Too Generous**
- 3 starting lives + bonus lives every 10k points
- Players accumulate lives faster than they lose them
- Almost no one reaches game over

---

## Recommended Difficulty Adjustments

### Priority 1: **Increase Threat Frequency** (Critical)

```go
// main.go changes needed:

// 1. Faster centipede movement - update every other tick
if tick % 2 == 0 {
    updateCentipedePositions()
}

// 2. More aggressive fly spawning
func (g *Game) spawnFly() {
    if rand.Float64() < 0.05 { // Was 0.02, now 0.05 (2.5× more flies)
        ...
    }
}

// 3. Spawn more mushrooms per level
g.spawnMushrooms(10) // Was 5, now 10

// 4. Start with MORE mushrooms initially
g.spawnMushrooms(25) // Was 15, now 25
```

### Priority 2: **Make Poison Mushrooms Deadlier**

```go
// When centipede hits poison mushroom, drop 3 rows instead of 1
if mush.poisoned {
    seg.pos.Y += 3  // Was 1, now 3 (true "chute" effect)
    seg.direction *= -1
}
```

### Priority 3: **Reduce Lives Generosity**

```go
// Option A: Bonus life every 20k instead of 10k
if g.score >= g.lastLifeScore+20000 {
    g.lives++
    g.lastLifeScore = g.score - (g.score % 20000)
}

// Option B: Start with 2 lives instead of 3
lives: 2,  // Was 3
```

### Priority 4: **Add Speed Scaling**

```go
// Centipedes get faster each level
tickDelay := 80 - (g.level * 5) // Subtract 5ms per level
if tickDelay < 30 {
    tickDelay = 30 // Cap at 30ms (very fast)
}
```

---

## Target Metrics After Tuning

| Metric | Current | Target | Status |
|--------|---------|--------|--------|
| Avg Lives Lost | 0.04 | 2.5 | ❌ Fix |
| Avg Ticks/Life | 9,606 | 200-350 | ❌ Fix |
| Too Easy % | 22.2% | <15% | ❌ Fix |
| Balanced % | 72.9% | 60-80% | ✅ Good |
| Poison Deaths % | 0.0% | 15-30% | ❌ Fix |
| Avg Levels Completed | 6.91 | 3-5 | ⚠️ Adjust |

---

## Frustration vs Ease Assessment

### Current State: **TOO EASY / CHILDISH** ⚠️

**Player Experience Problems**:
- ❌ No sense of danger or urgency
- ❌ Easy to survive indefinitely
- ❌ Poison mushrooms feel like decoration
- ❌ Lives accumulate faster than lost
- ❌ Games feel repetitive and predictable

**What Players Will Say**:
- "This is boring, I can't die"
- "Where's the challenge?"
- "The arcade version was way harder"
- "Poison mushrooms don't do anything"
- "I beat 10 levels without trying"

### Target State: **BALANCED CHALLENGE** ✅

**Desired Player Experience**:
- ✅ Constant sense of danger
- ✅ Deaths feel fair but frequent
- ✅ Poison mushrooms are SCARY
- ✅ Must use skill to survive
- ✅ Each life feels valuable
- ✅ Reaching level 5+ feels like achievement

---

## Implementation Plan

### Phase 1: **Quick Wins** (Immediate)
1. Increase mushroom spawn counts (15→25 initial, 5→10 per level)
2. Increase fly spawn rate (2%→5%)
3. Make poison mushroom chute drop 3 rows instead of 1
4. Reduce bonus life frequency (10k→20k)

### Phase 2: **Speed Tuning** (Next)
5. Add level-based speed scaling
6. Make centipedes update every 2 ticks instead of every tick
7. Reduce player movement delay slightly

### Phase 3: **Validation** (Final)
8. Run another 1,000 game simulation
9. Target metrics:
   - Lives lost: 2-3 per game
   - Ticks/life: 200-350
   - Too Easy: <15%
   - Poison deaths: 15-30%

---

## Conclusion

The game currently provides a **pleasant but unchallenging experience**. While the 72.9% balanced distribution is good, players survive far too long (27× longer than target) and accumulate lives instead of losing them.

**Priority Actions**:
1. ⚠️ **CRITICAL**: Increase enemy density and speed
2. ⚠️ **CRITICAL**: Make poison mushrooms actually dangerous
3. ⚠️ **IMPORTANT**: Reduce life generosity
4. ✅ **OPTIONAL**: Add speed scaling for progression

**Expected Outcome**: Game will feel like classic arcade Centipede - challenging, fair, and requiring actual skill to survive.

---

*Report generated by automated balance testing harness*
*Test code: `/Users/michaellavery/github/centipede/test_balance.go`*
