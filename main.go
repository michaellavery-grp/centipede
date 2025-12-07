// Centipede - A terminal-based centipede game using Bubble Tea
// Created by Claude Code (Anthropic)
package main

import (
	"fmt"
	"math/rand"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Game entity positions
type Position struct {
	X, Y int
}

// Player with improved gun character
type Player struct {
	pos Position
}

// Bullet with improved rendering
type Bullet struct {
	pos    Position
	active bool
}

func (b *Bullet) Update() {
	if b.active {
		b.pos.Y--
		if b.pos.Y < 0 {
			b.active = false
		}
	}
}

// Centipede Segment with head tracking
type Segment struct {
	pos       Position
	direction int  // 1 = right, -1 = left
	isHead    bool // Track head segment for special rendering
}

// Mushroom obstacle
type Mushroom struct {
	pos    Position
	health int // 0-4 hits to destroy
}

// Game state
type Game struct {
	width     int
	height    int
	player    Player
	segments  []Segment
	bullets   []Bullet
	mushrooms []Mushroom
	score     int
	level     int
	gameOver  bool
	won       bool
}

func NewGame(width, height int) *Game {
	g := &Game{
		width:  width,
		height: height,
		player: Player{pos: Position{X: width / 2, Y: height - 2}},
		level:  1,
	}

	// Create initial centipede at top with head
	g.spawnCentipede(10)

	// Create random mushrooms
	g.spawnMushrooms(15)

	return g
}

func (g *Game) spawnCentipede(length int) {
	startX := 5
	startY := 2

	for i := 0; i < length; i++ {
		g.segments = append(g.segments, Segment{
			pos:       Position{X: startX + i, Y: startY},
			direction: 1,
			isHead:    i == 0, // First segment is the head
		})
	}
}

func (g *Game) spawnMushrooms(count int) {
	for i := 0; i < count; i++ {
		x := rand.Intn(g.width-2) + 1
		y := rand.Intn(g.height-5) + 2 // Avoid player area
		g.mushrooms = append(g.mushrooms, Mushroom{
			pos:    Position{X: x, Y: y},
			health: 4,
		})
	}
}

func (g *Game) Update() {
	if g.gameOver || g.won {
		return
	}

	// Update bullets
	for i := range g.bullets {
		g.bullets[i].Update()
	}

	// Update centipede segments with improved falling behavior
	for i := range g.segments {
		seg := &g.segments[i]
		seg.pos.X += seg.direction

		// Hit edge - drop down and reverse
		if seg.pos.X <= 0 || seg.pos.X >= g.width-1 {
			seg.pos.Y++
			seg.direction *= -1
		}

		// Check if hit mushroom - drop down and reverse
		for _, mush := range g.mushrooms {
			if seg.pos.X == mush.pos.X && seg.pos.Y == mush.pos.Y {
				seg.pos.Y++
				seg.direction *= -1
				break
			}
		}

		// Check if reached player area - game over
		if seg.pos.Y >= g.height-3 {
			g.gameOver = true
		}
	}

	// Check bullet collisions
	for i := range g.bullets {
		if !g.bullets[i].active {
			continue
		}

		// Bullet vs Centipede
		for j := range g.segments {
			if g.bullets[i].pos.X == g.segments[j].pos.X &&
				g.bullets[i].pos.Y == g.segments[j].pos.Y {
				g.bullets[i].active = false

				// Extra points for head
				if g.segments[j].isHead {
					g.score += 100
				} else {
					g.score += 10
				}

				// Remove segment
				g.segments = append(g.segments[:j], g.segments[j+1:]...)

				// Update head if we removed first segment
				if len(g.segments) > 0 && j == 0 {
					g.segments[0].isHead = true
				}
				break
			}
		}

		// Bullet vs Mushroom
		for j := range g.mushrooms {
			if g.bullets[i].pos.X == g.mushrooms[j].pos.X &&
				g.bullets[i].pos.Y == g.mushrooms[j].pos.Y {
				g.bullets[i].active = false
				g.mushrooms[j].health--
				g.score += 1

				// Remove mushroom if destroyed
				if g.mushrooms[j].health <= 0 {
					g.mushrooms = append(g.mushrooms[:j], g.mushrooms[j+1:]...)
					g.score += 4
				}
				break
			}
		}
	}

	// Check win condition
	if len(g.segments) == 0 {
		g.won = true
	}
}

func (g *Game) MovePlayer(dx int) {
	newX := g.player.pos.X + dx
	if newX > 0 && newX < g.width-1 {
		// Check mushroom collision
		canMove := true
		for _, mush := range g.mushrooms {
			if newX == mush.pos.X && g.player.pos.Y == mush.pos.Y {
				canMove = false
				break
			}
		}
		if canMove {
			g.player.pos.X = newX
		}
	}
}

func (g *Game) MovePlayerY(dy int) {
	newY := g.player.pos.Y + dy
	// Allow movement in bottom quarter of screen
	if newY >= g.height-6 && newY < g.height-1 {
		// Check mushroom collision
		canMove := true
		for _, mush := range g.mushrooms {
			if g.player.pos.X == mush.pos.X && newY == mush.pos.Y {
				canMove = false
				break
			}
		}
		if canMove {
			g.player.pos.Y = newY
		}
	}
}

func (g *Game) Shoot() {
	// Only allow 2 bullets on screen
	activeBullets := 0
	for _, b := range g.bullets {
		if b.active {
			activeBullets++
		}
	}

	if activeBullets < 2 {
		g.bullets = append(g.bullets, Bullet{
			pos:    Position{X: g.player.pos.X, Y: g.player.pos.Y - 1},
			active: true,
		})
	}
}

func (g *Game) GetBoard() [][]rune {
	board := make([][]rune, g.height)
	for i := range board {
		board[i] = make([]rune, g.width)
		for j := range board[i] {
			board[i][j] = ' '
		}
	}

	// Draw player gun character (improved)
	board[g.player.pos.Y][g.player.pos.X] = 'A'

	// Draw mushrooms with different characters based on health
	for _, mush := range g.mushrooms {
		if mush.pos.Y >= 0 && mush.pos.Y < g.height &&
			mush.pos.X >= 0 && mush.pos.X < g.width {
			switch mush.health {
			case 4:
				board[mush.pos.Y][mush.pos.X] = 'M'
			case 3:
				board[mush.pos.Y][mush.pos.X] = 'm'
			case 2:
				board[mush.pos.Y][mush.pos.X] = '*'
			case 1:
				board[mush.pos.Y][mush.pos.X] = '.'
			}
		}
	}

	// Draw centipede segments with head differentiation
	for _, seg := range g.segments {
		if seg.pos.Y >= 0 && seg.pos.Y < g.height &&
			seg.pos.X >= 0 && seg.pos.X < g.width {
			if seg.isHead {
				board[seg.pos.Y][seg.pos.X] = '@' // Head
			} else {
				board[seg.pos.Y][seg.pos.X] = 'O' // Body
			}
		}
	}

	// Draw bullets (improved)
	for _, bullet := range g.bullets {
		if bullet.active && bullet.pos.Y >= 0 && bullet.pos.Y < g.height {
			board[bullet.pos.Y][bullet.pos.X] = '|'
		}
	}

	return board
}

// Bubble Tea Model
type tickMsg time.Time

type model struct {
	game   *Game
	paused bool
	width  int
	height int
}

// Styles
var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("205")).
			MarginBottom(1)

	playerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("10"))

	centipedeHeadStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("13"))

	centipedeBodyStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("93"))

	mushroomStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("2"))

	bulletStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("11"))

	borderStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("62"))

	statsStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("86")).
			Bold(true)

	gameOverStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("196")).
			MarginTop(1).
			MarginBottom(1)

	winStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("10")).
			MarginTop(1).
			MarginBottom(1)
)

func initialModel() model {
	return model{
		game: NewGame(50, 28),
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(
		tickCmd(),
		tea.EnterAltScreen,
	)
}

func tickCmd() tea.Cmd {
	return tea.Tick(time.Millisecond*80, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "left", "a":
			m.game.MovePlayer(-1)
		case "right", "d":
			m.game.MovePlayer(1)
		case "up", "w":
			m.game.MovePlayerY(-1)
		case "down", "s":
			m.game.MovePlayerY(1)
		case "space":
			m.game.Shoot()
		case "p":
			m.paused = !m.paused
		case "r":
			// Restart game
			if m.game.gameOver || m.game.won {
				m.game = NewGame(50, 28)
			}
		}

	case tickMsg:
		if !m.paused {
			m.game.Update()
		}
		return m, tickCmd()
	}

	return m, nil
}

func (m model) View() string {
	board := m.game.GetBoard()

	// Title
	title := titleStyle.Render("🐛 CENTIPEDE 🐛")

	// Build game board with colors
	var boardStr string
	boardStr += "┌" + lipgloss.NewStyle().Foreground(lipgloss.Color("62")).Render(
		lipgloss.PlaceHorizontal(len(board[0]), lipgloss.Center, "")) + "┐\n"

	for _, row := range board {
		boardStr += "│"
		for _, cell := range row {
			char := string(cell)
			switch cell {
			case 'A': // Player
				char = playerStyle.Render(char)
			case '@': // Centipede head
				char = centipedeHeadStyle.Render(char)
			case 'O': // Centipede body
				char = centipedeBodyStyle.Render(char)
			case 'M', 'm', '*', '.': // Mushrooms
				char = mushroomStyle.Render(char)
			case '|': // Bullets
				char = bulletStyle.Render(char)
			}
			boardStr += char
		}
		boardStr += "│\n"
	}

	boardStr += "└" + lipgloss.NewStyle().Foreground(lipgloss.Color("62")).Render(
		lipgloss.PlaceHorizontal(len(board[0]), lipgloss.Center, "")) + "┘"

	// Stats
	stats := statsStyle.Render(fmt.Sprintf(
		"Score: %d  |  Level: %d  |  Segments: %d  |  Mushrooms: %d",
		m.game.score, m.game.level, len(m.game.segments), len(m.game.mushrooms)))

	// Controls
	controls := lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render(
		"[←→ or A/D] Move  [↑↓ or W/S] Up/Down  [Space] Shoot  [P] Pause  [Q] Quit")

	// Status messages
	status := ""
	if m.paused {
		status = lipgloss.NewStyle().
			Foreground(lipgloss.Color("11")).
			Bold(true).
			Render("⏸  PAUSED")
	}
	if m.game.gameOver {
		status = gameOverStyle.Render("💥 GAME OVER! Press [R] to restart")
	}
	if m.game.won {
		status = winStyle.Render("🎉 YOU WIN! Press [R] to play again")
	}

	// Combine everything
	return lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		boardStr,
		"",
		stats,
		controls,
		status,
	)
}

func main() {
	rand.Seed(time.Now().UnixNano())

	p := tea.NewProgram(
		initialModel(),
		tea.WithAltScreen(),
	)

	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(1)
	}
}
