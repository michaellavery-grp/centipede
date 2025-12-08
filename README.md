# 🐛 Centipede

A terminal-based clone of the classic arcade game Centipede, built with [Bubble Tea](https://github.com/charmbracelet/bubbletea) and Go.

![Centipede Demo](https://img.shields.io/badge/status-playable-green)
![Go Version](https://img.shields.io/badge/go-1.18+-blue)
![License](https://img.shields.io/badge/license-MIT-blue)

## 🎮 Features

- **Splash Screen**: ASCII art title with green worm, spider, flea, and fly characters
- **High Score System**: Top 10 high scores saved to `highscores.txt` with name entry
- **Flashing Messages**: Animated "Press any key to continue" on splash screen
- **Classic Centipede Gameplay**: Shoot the descending centipede segments as they zigzag down the screen
- **Lives System**: Start with 3 lives (♥♥♥), earn bonus life every 10,000 points!
- **Player Collision**: Lose a life when centipede touches you - respawn with 2.4 second invincibility
- **Progressive Difficulty**: Each level spawns longer centipedes (10 + level×2 segments)
- **Correct Head Position**: Head segment (@) is at the FRONT of the centipede (direction of movement)
- **Mushroom Obstacles**: Destroyable mushrooms (4 hits) that affect centipede movement
- **Mushroom Regeneration**: All mushrooms restore to full health on level complete or player death
- **Poison Mushrooms**: Flies that hit mushrooms create poison mushrooms (X) - creates deadly 3-char zigzag chute!
- **Poison Mushroom Chute**: Centipedes hitting poison mushrooms drop in tight zigzag pattern straight down
- **Smart Falling Mechanics**: Centipedes drop down and reverse direction when hitting edges or mushrooms
- **Player Movement**: Full directional control in the bottom quarter of the screen
- **Unlimited Rapid Fire**: Hold spacebar to fire bullets continuously (10 per second!)
- **Fly Enemy**: Animated flies cross the screen with flickering wing trails (✺~.)
- **Explosion Effects**: Animated explosions (✶✸✹✺) when hitting enemies
- **Scoring System**:
  - Body segment: 10 points
  - Head segment: 100 points
  - Fly enemy: 200 points
  - Mushroom hit: 1 point
  - Mushroom destroyed: 5 points total
- **Color-Coded Display**: Using Lip Gloss for terminal styling
  - Green player ('A')
  - Magenta centipede head ('@')
  - Purple centipede body ('O')
  - Green mushrooms ('M' → 'm' → '*' → '.')
  - Magenta poison mushrooms ('X')
  - Yellow bullets ('|')
- **Game States**: Continuous play with progressive levels
- **Improved Game Over**: Player controls freeze when game ends, 'R' to restart works properly
- **Pause Function**: Freeze the action with 'P'

## 📦 Installation

### Prerequisites

- Go 1.18 or higher

### Building from Source

```bash
# Clone the repository
git clone https://github.com/michaellavery-grp/centipede.git
cd centipede

# Install dependencies
go mod download

# Build the game
go build -o centipede main.go

# Run it!
./centipede
```

### Quick Run

```bash
go run main.go
```

## 🕹️ Controls

| Key | Action |
|-----|--------|
| `Any Key` | Start game (from splash screen) |
| `←` / `→` or `A` / `D` | Move left/right |
| `↑` / `↓` or `W` / `S` | Move up/down (in player area) |
| `Space` | UNLIMITED RAPID FIRE! (Hold = 10/sec) |
| `P` | Pause/Unpause |
| `R` | Restart (after game over/win) |
| `Q` or `Ctrl+C` | Quit |
| `Letters` | Enter name (high score screen) |
| `Enter` | Submit name (high score screen) |

## 🎯 How to Play

1. **Start**: Press any key on the splash screen to begin with 3 lives (♥♥♥)
2. **Objective**: Destroy all centipede segments before they touch you!
3. **Strategy**:
   - Aim for the head (@) for bonus points (100 vs 10) - head is at the FRONT!
   - Hold spacebar for UNLIMITED rapid fire (10 bullets/second!)
   - Shoot flies (✺) for 200 points - watch for their flickering wings
   - **AVOID poison mushrooms (X)** - they create deadly 3-char zigzag chutes!
   - Flies that hit mushrooms create poison mushrooms - prevent this!
   - Poison mushrooms force centipedes into tight zigzag descent
   - Destroy mushrooms to create clear lanes
   - Use vertical movement to dodge and position
   - Watch centipede drop patterns when hitting obstacles
   - Explosions appear when you hit enemies!
4. **Lives System**:
   - Start with 3 lives
   - Lose a life when centipede touches you
   - Respawn with 2.4 second invincibility
   - Earn bonus life every 10,000 points
   - Mushrooms regenerate to full health on death or level complete
   - Game over when all lives lost
5. **Progressive Levels**: Destroy all segments to spawn a longer, harder centipede!
6. **High Score**: Enter your name if you make the top 10!

## 🏗️ Technical Details

### Architecture

The game follows the **Elm Architecture** pattern used by Bubble Tea:

- **Model**: Game state (player, centipede segments, bullets, mushrooms, score)
- **Update**: Message handler for input and game logic
- **View**: Renderer that builds the terminal display

### Technologies

- **[Bubble Tea](https://github.com/charmbracelet/bubbletea)**: TUI framework
- **[Lip Gloss](https://github.com/charmbracelet/lipgloss)**: Terminal styling and colors
- **Go**: Programming language

### Game Loop

- Tick rate: 80ms (~12.5 FPS)
- Centipede movement: 1 cell per tick
- Bullet speed: 1 cell upward per tick
- Rapid fire rate: 100ms (10 bullets/second when holding space)
- **UNLIMITED BULLETS**: No limit on active bullets!
- Fly movement: 2 cells per tick (faster than centipede)
- Fly spawn rate: 2% chance per tick
- Wing animation: Alternates each tick (~. pattern)
- Explosion animation: 4 frames (✶→✸→✹→✺)
- Collision detection: Position-based (X, Y matching)
- Splash screen: Flashing text at tick rate
- High scores: Saved to `highscores.txt` (CSV format: Name,Score)

### Code Structure

```go
main.go
├── Position struct          // X, Y coordinates
├── Player struct           // Player gun position
├── Bullet struct           // Active bullets with position
├── Segment struct          // Centipede segments (head/body)
├── Mushroom struct         // Obstacles with health (1-4)
├── Fly struct              // Animated enemy with wing flap
├── Explosion struct        // 4-frame explosion animation
├── HighScore struct        // Name and score
├── Game struct             // Main game state
├── model struct            // Bubble Tea model with game states
├── loadHighScores()        // Read from highscores.txt
├── saveHighScore()         // Write to highscores.txt
├── Update() methods        // Game logic + rapid fire
├── View() method           // Terminal rendering
├── renderSplash()          // Splash screen with ASCII art
├── renderNameEntry()       // High score name entry
└── main()                  // Entry point
```

## 🎨 Visual Elements

```
┌──────────────────────────────────────────────────┐
│                                                  │
│       OOOOOOOO@                                  │  ← Centipede (@ = head at FRONT, O = body)
│          M      m    *    X    .                 │  ← Mushrooms (M=full, m=damaged, *=weak, .=critical, X=POISON!)
│                           ✺~.                    │  ← Fly with wing trail
│                     |  ✹                         │  ← Bullet & Explosion
│                     |                            │
│                     |                            │
│                     A                            │  ← Player gun
└──────────────────────────────────────────────────┘

Score: 320  |  Lives: ♥♥♥  |  Bullets: 15  |  Segments: 7  |  Flies: 1  |  Level: 2
```

## 🚀 Future Enhancements

Potential additions for future versions:

- [x] Multiple levels with increasing difficulty (DONE! v4.0)
- [x] Correct head positioning (DONE! v4.0)
- [x] Poison mushrooms (DONE! v4.0)
- [x] Poison mushroom zigzag chute (DONE! v5.0)
- [x] Improved game over controls (DONE! v4.0)
- [x] Lives system with respawn (DONE! v5.0)
- [x] Bonus lives every 10k points (DONE! v5.0)
- [x] Player collision detection (DONE! v5.0)
- [x] Mushroom regeneration on death/level (DONE! v5.0)
- [ ] Additional enemies (Spider, Flea, Scorpion) - Flea/Spider on splash
- [x] High score tracking (DONE!)
- [x] Splash screen with ASCII art (DONE!)
- [x] Unlimited rapid fire bullets (DONE!)
- [x] Fly enemy with animation (DONE!)
- [x] Explosion effects (DONE!)
- [ ] Sound effects (terminal bell)
- [ ] Configuration file (TOML)
- [ ] Centipede segment splitting when hit mid-body
- [ ] Speed increases as segments are destroyed
- [ ] Online leaderboard

## 🐛 Known Issues

- None currently! Report issues on GitHub.

## 📝 Credits

- **Game Development**: Claude Code (Anthropic) - 2025
- **Original Centipede**: Atari (1981) - Dona Bailey & Ed Logg
- **Frameworks**:
  - [Charm](https://charm.sh/) - Bubble Tea & Lip Gloss
- **Inspired by**: [Tetrigo](https://github.com/Broderick-Westrope/tetrigo) - Bubble Tea Tetris implementation

## 📄 License

MIT License - See LICENSE file for details

## 🤝 Contributing

Contributions welcome! Feel free to:

- Report bugs
- Suggest features
- Submit pull requests
- Improve documentation

## 🎓 Learning Resource

This game was built following the Bubble Tea framework patterns as a learning exercise in:

- Terminal UI development in Go
- The Elm Architecture pattern
- Game loop implementation
- Collision detection
- State management

Perfect for developers learning Bubble Tea or terminal game development!

---

**Enjoy the game!** 🐛🎮

*Built with ❤️ using Go and Bubble Tea*
