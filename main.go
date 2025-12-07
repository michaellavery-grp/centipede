// Centipede - A terminal-based centipede game using Bubble Tea
// Created by Claude Code (Anthropic)
package main

import (
	"fmt"
	"math/rand"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const highScoreFile = "highscores.txt"

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

// High Score entry
type HighScore struct {
	Name  string
	Score int
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

// High Score Management
func loadHighScores() []HighScore {
	scores := []HighScore{}
	data, err := os.ReadFile(highScoreFile)
	if err != nil {
		return scores // Return empty if file doesn't exist
	}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.Split(line, ",")
		if len(parts) == 2 {
			score, err := strconv.Atoi(parts[1])
			if err == nil {
				scores = append(scores, HighScore{Name: parts[0], Score: score})
			}
		}
	}

	// Sort by score descending
	sort.Slice(scores, func(i, j int) bool {
		return scores[i].Score > scores[j].Score
	})

	return scores
}

func saveHighScore(name string, score int) error {
	scores := loadHighScores()
	scores = append(scores, HighScore{Name: name, Score: score})

	// Sort by score descending
	sort.Slice(scores, func(i, j int) bool {
		return scores[i].Score > scores[j].Score
	})

	// Keep top 10
	if len(scores) > 10 {
		scores = scores[:10]
	}

	// Write to file
	var lines []string
	for _, s := range scores {
		lines = append(lines, fmt.Sprintf("%s,%d", s.Name, s.Score))
	}

	return os.WriteFile(highScoreFile, []byte(strings.Join(lines, "\n")), 0644)
}

// Bubble Tea Model
type tickMsg time.Time
type shootMsg time.Time

type gameState int

const (
	splashScreen gameState = iota
	playingGame
	gameOverScreen
)

type model struct {
	game            *Game
	paused          bool
	width           int
	height          int
	state           gameState
	flashOn         bool
	spacePressed    bool
	lastShot        time.Time
	highScores      []HighScore
	playerName      string
	enteringName    bool
	scoreSaved      bool
}

// Styles
var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("205")).
			MarginBottom(1)

	splashTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("10"))

	flashStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("11"))

	highScoreStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("14")).
			Bold(true)

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
		game:       NewGame(50, 28),
		state:      splashScreen,
		highScores: loadHighScores(),
		lastShot:   time.Now(),
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(
		tickCmd(),
		shootTickCmd(),
		tea.EnterAltScreen,
	)
}

func tickCmd() tea.Cmd {
	return tea.Tick(time.Millisecond*80, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func shootTickCmd() tea.Cmd {
	return tea.Tick(time.Millisecond*500, func(t time.Time) tea.Msg {
		return shootMsg(t)
	})
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
		// Handle splash screen
		if m.state == splashScreen {
			m.state = playingGame
			return m, nil
		}

		// Handle name entry
		if m.enteringName {
			switch msg.String() {
			case "enter":
				if m.playerName != "" {
					saveHighScore(m.playerName, m.game.score)
					m.highScores = loadHighScores()
					m.scoreSaved = true
					m.enteringName = false
				}
			case "backspace":
				if len(m.playerName) > 0 {
					m.playerName = m.playerName[:len(m.playerName)-1]
				}
			default:
				if len(msg.String()) == 1 && len(m.playerName) < 10 {
					m.playerName += msg.String()
				}
			}
			return m, nil
		}

		// Handle game controls
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "left", "a":
			if m.state == playingGame {
				m.game.MovePlayer(-1)
			}
		case "right", "d":
			if m.state == playingGame {
				m.game.MovePlayer(1)
			}
		case "up", "w":
			if m.state == playingGame {
				m.game.MovePlayerY(-1)
			}
		case "down", "s":
			if m.state == playingGame {
				m.game.MovePlayerY(1)
			}
		case " ": // Spacebar
			if m.state == playingGame {
				m.spacePressed = true
				m.game.Shoot()
			}
		case "p":
			if m.state == playingGame {
				m.paused = !m.paused
			}
		case "r":
			// Restart game
			if m.game.gameOver || m.game.won {
				m.game = NewGame(50, 28)
				m.state = playingGame
				m.enteringName = false
				m.scoreSaved = false
				m.playerName = ""
			}
		}

	case shootMsg:
		// Rapid fire when holding space (2 shots per second)
		if m.spacePressed && m.state == playingGame && !m.paused {
			now := time.Now()
			if now.Sub(m.lastShot) >= time.Millisecond*500 {
				m.game.Shoot()
				m.lastShot = now
			}
		}
		return m, shootTickCmd()

	case tickMsg:
		// Flash "Press any key" message
		m.flashOn = !m.flashOn

		if m.state == playingGame && !m.paused {
			m.game.Update()

			// Check if game ended and score is high enough
			if (m.game.gameOver || m.game.won) && !m.scoreSaved && !m.enteringName {
				scores := m.highScores
				if len(scores) < 10 || m.game.score > scores[len(scores)-1].Score {
					m.enteringName = true
				}
			}
		}
		return m, tickCmd()
	}

	// Reset space key when released (key up events)
	if _, ok := msg.(tea.KeyMsg); ok {
		m.spacePressed = false
	}

	return m, nil
}

func (m model) View() string {
	if m.state == splashScreen {
		return m.renderSplash()
	}

	if m.enteringName {
		return m.renderNameEntry()
	}

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

func (m model) renderSplash() string {
	centipede := splashTitleStyle.Render(`
   _____ ______ _   _ _______ _____ _____  ______ _____  ______
  / ____|  ____| \ | |__   __|_   _|  __ \|  ____|  __ \|  ____|
 | |    | |__  |  \| |  | |    | | | |__) | |__  | |  | | |__
 | |    |  __| | . \ |  | |    | | |  ___/|  __| | |  | |  __|
 | |____| |____| |\  |  | |   _| |_| |    | |____| |__| | |____
  \_____|______|_| \_|  |_|  |_____|_|    |______|_____/|______|
`)

	worm := lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Render(`
        ╔═══════════════════════════════════════╗
        ║    @OOOOOOOOOOOOOO    Green Worm     ║
        ║                                       ║
        ║    ╱╲  ╱╲  ╱╲                        ║
        ║   ╱  ╲╱  ╲╱  ╲       Spider          ║
        ║  ╱    ╲    ╲  ╲                      ║
        ║                                       ║
        ║    ┃                 Flea             ║
        ║    ●                                  ║
        ║    ┃                                  ║
        ║                                       ║
        ║    ✺   Fly                            ║
        ╚═══════════════════════════════════════╝
`)

	// High scores
	highScoreTitle := highScoreStyle.Render("\n═══ HIGH SCORES ═══\n")
	var scoreLines []string
	for i, score := range m.highScores {
		if i >= 10 {
			break
		}
		scoreLines = append(scoreLines,
			lipgloss.NewStyle().Foreground(lipgloss.Color("14")).Render(
				fmt.Sprintf("%2d. %-10s  %6d", i+1, score.Name, score.Score)))
	}
	highScoreList := strings.Join(scoreLines, "\n")

	// Flashing "Press any key"
	pressKey := ""
	if m.flashOn {
		pressKey = flashStyle.Render("\n\n>>> PRESS ANY KEY TO CONTINUE <<<")
	} else {
		pressKey = "\n\n                                  "
	}

	return lipgloss.JoinVertical(
		lipgloss.Center,
		centipede,
		worm,
		highScoreTitle,
		highScoreList,
		pressKey,
	)
}

func (m model) renderNameEntry() string {
	title := gameOverStyle.Render("NEW HIGH SCORE!")
	scoreText := statsStyle.Render(fmt.Sprintf("Your Score: %d", m.game.score))
	prompt := lipgloss.NewStyle().Foreground(lipgloss.Color("11")).Render(
		"Enter your name (max 10 chars):")
	nameDisplay := lipgloss.NewStyle().
		Foreground(lipgloss.Color("10")).
		Bold(true).
		Render(m.playerName + "_")
	instruction := lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render(
		"Press [Enter] to save")

	return lipgloss.JoinVertical(
		lipgloss.Center,
		"",
		"",
		"",
		title,
		"",
		scoreText,
		"",
		prompt,
		nameDisplay,
		"",
		instruction,
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
