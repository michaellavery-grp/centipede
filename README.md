# 🐛 Centipede

A terminal-based clone of the classic arcade game Centipede, built with [Bubble Tea](https://github.com/charmbracelet/bubbletea) and Go.

![Centipede Demo](https://img.shields.io/badge/status-playable-green)
![Go Version](https://img.shields.io/badge/go-1.18+-blue)
![License](https://img.shields.io/badge/license-MIT-blue)

## 🎮 Features

- **Classic Centipede Gameplay**: Shoot the descending centipede segments as they zigzag down the screen
- **Mushroom Obstacles**: Destroyable mushrooms (4 hits) that affect centipede movement
- **Smart Falling Mechanics**: Centipedes drop down and reverse direction when hitting edges or mushrooms
- **Player Movement**: Full directional control in the bottom quarter of the screen
- **Bullet System**: Fire up to 2 bullets at once
- **Scoring System**:
  - Body segment: 10 points
  - Head segment: 100 points
  - Mushroom hit: 1 point
  - Mushroom destroyed: 5 points total
- **Color-Coded Display**: Using Lip Gloss for terminal styling
  - Green player ('A')
  - Magenta centipede head ('@')
  - Purple centipede body ('O')
  - Green mushrooms ('M' → 'm' → '*' → '.')
  - Yellow bullets ('|')
- **Game States**: Win/lose conditions with restart capability
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
| `←` / `→` or `A` / `D` | Move left/right |
| `↑` / `↓` or `W` / `S` | Move up/down (in player area) |
| `Space` | Shoot |
| `P` | Pause/Unpause |
| `R` | Restart (after game over/win) |
| `Q` or `Ctrl+C` | Quit |

## 🎯 How to Play

1. **Objective**: Destroy all centipede segments before they reach the bottom
2. **Strategy**:
   - Aim for the head (@) for bonus points (100 vs 10)
   - Destroy mushrooms to create clear lanes
   - Use vertical movement to dodge and position
   - Watch centipede drop patterns when hitting obstacles
3. **Game Over**: Centipede reaches the player area (bottom 3 rows)
4. **Victory**: Destroy all 10 segments

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
- Collision detection: Position-based (X, Y matching)

### Code Structure

```go
main.go
├── Position struct          // X, Y coordinates
├── Player struct           // Player gun position
├── Bullet struct           // Active bullets with position
├── Segment struct          // Centipede segments (head/body)
├── Mushroom struct         // Obstacles with health (1-4)
├── Game struct             // Main game state
├── model struct            // Bubble Tea model
├── Update() methods        // Game logic
├── View() method           // Terminal rendering
└── main()                  // Entry point
```

## 🎨 Visual Elements

```
┌──────────────────────────────────────────────────┐
│                                                  │
│       @OOOOOOOOO                                 │  ← Centipede (@ = head, O = body)
│          M      m    *         .                 │  ← Mushrooms (M=full, m=damaged, *=weak, .=critical)
│                                                  │
│                     |                            │  ← Bullet
│                                                  │
│                                                  │
│                     A                            │  ← Player gun
└──────────────────────────────────────────────────┘

Score: 120  |  Level: 1  |  Segments: 7  |  Mushrooms: 12
```

## 🚀 Future Enhancements

Potential additions for future versions:

- [ ] Multiple levels with increasing difficulty
- [ ] Additional enemies (Spider, Flea, Scorpion)
- [ ] Lives system with respawn
- [ ] High score tracking
- [ ] Sound effects (terminal bell)
- [ ] Configuration file (TOML)
- [ ] Centipede segment splitting when hit mid-body
- [ ] Speed increases as segments are destroyed
- [ ] Mushroom regeneration between levels

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
