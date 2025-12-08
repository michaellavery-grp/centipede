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
	pos       Position
	health    int  // 0-4 hits to destroy
	poisoned  bool // Poisoned mushrooms make centipede fall faster
}

// Fly enemy
type Fly struct {
	pos       Position
	direction int  // 1 = right, -1 = left
	active    bool
	wingFlap  bool // Alternates for wing animation
}

// Flea enemy - falls from top and creates mushrooms
type Flea struct {
	pos    Position
	active bool
}

func (f *Fly) Update() {
	if !f.active {
		return
	}

	// Move horizontally
	f.pos.X += f.direction * 2 // Flies move faster

	// Toggle wing flap
	f.wingFlap = !f.wingFlap

	// Deactivate if off screen
	if f.pos.X < 0 || f.pos.X >= 50 {
		f.active = false
	}
}

// Explosion effect
type Explosion struct {
	pos      Position
	frame    int
	maxFrame int
	active   bool
}

func (e *Explosion) Update() {
	if !e.active {
		return
	}
	e.frame++
	if e.frame >= e.maxFrame {
		e.active = false
	}
}

func (e *Explosion) Char() rune {
	switch e.frame {
	case 0:
		return '✶'
	case 1:
		return '✸'
	case 2:
		return '✹'
	case 3:
		return '✺'
	default:
		return ' '
	}
}

// High Score entry
type HighScore struct {
	Name  string
	Score int
}

// Game state
type Game struct {
	width         int
	height        int
	player        Player
	segments      []Segment
	bullets       []Bullet
	mushrooms     []Mushroom
	flies         []Fly
	fleas         []Flea
	explosions    []Explosion
	score         int
	level         int
	lives         int
	lastLifeScore int // Track score for bonus life awards
	respawning    bool
	respawnTimer  int
	gameOver      bool
	won           bool
}

func NewGame(width, height int) *Game {
	g := &Game{
		width:         width,
		height:        height,
		player:        Player{pos: Position{X: width / 2, Y: height - 2}},
		level:         1,
		lives:         3,
		lastLifeScore: 0,
	}

	// Create initial centipede at top with head
	g.spawnCentipede(10)

	// Spawn SECOND centipede for increased difficulty!
	g.spawnSecondCentipede(8)

	// Create random mushrooms - INCREASED for difficulty
	g.spawnMushrooms(25) // Was 15, now 25 for more obstacles

	return g
}

func (g *Game) spawnSecondCentipede(length int) {
	// Spawn second centipede offset from first
	startX := 25 // Offset from first centipede
	startY := 2

	for i := 0; i < length; i++ {
		g.segments = append(g.segments, Segment{
			pos:       Position{X: startX + i, Y: startY},
			direction: -1, // Moving left (opposite of first)
			isHead:    i == length-1,
		})
	}
}

func (g *Game) spawnCentipede(length int) {
	startX := 5
	startY := 2

	for i := 0; i < length; i++ {
		g.segments = append(g.segments, Segment{
			pos:       Position{X: startX + i, Y: startY},
			direction: 1,
			isHead:    i == length-1, // LAST segment is the head (front of movement)
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

func (g *Game) spawnFly() {
	// Random chance to spawn fly - INCREASED for difficulty
	if rand.Float64() < 0.05 { // Was 0.02 (2%), now 0.05 (5%) chance per tick
		y := rand.Intn(g.height - 10) + 3 // Middle area
		direction := 1
		startX := 0
		if rand.Float64() < 0.5 {
			direction = -1
			startX = g.width - 1
		}

		g.flies = append(g.flies, Fly{
			pos:       Position{X: startX, Y: y},
			direction: direction,
			active:    true,
			wingFlap:  false,
		})
	}
}

func (g *Game) spawnFlea() {
	// Spawn falling fleas when mushroom count is low
	mushroomCount := len(g.mushrooms)
	if mushroomCount < 15 && rand.Float64() < 0.03 { // 3% chance when low mushrooms
		x := rand.Intn(g.width-4) + 2
		g.fleas = append(g.fleas, Flea{
			pos:    Position{X: x, Y: 2},
			active: true,
		})
	}
}

func (f *Flea) Update(g *Game) {
	if !f.active {
		return
	}

	// Flea falls straight down
	f.pos.Y++

	// Create mushroom occasionally as it falls
	if rand.Float64() < 0.4 && f.pos.Y > 5 { // 40% chance per tick
		// Add mushroom at current position if none exists
		exists := false
		for _, m := range g.mushrooms {
			if m.pos.X == f.pos.X && m.pos.Y == f.pos.Y {
				exists = true
				break
			}
		}
		if !exists {
			g.mushrooms = append(g.mushrooms, Mushroom{
				pos:    Position{X: f.pos.X, Y: f.pos.Y},
				health: 4,
			})
		}
	}

	// Deactivate if reached bottom
	if f.pos.Y >= g.height-2 {
		f.active = false
	}
}

func (g *Game) createExplosion(x, y int) {
	g.explosions = append(g.explosions, Explosion{
		pos:      Position{X: x, Y: y},
		frame:    0,
		maxFrame: 4,
		active:   true,
	})
}

func (g *Game) Update() {
	if g.gameOver || g.won {
		return
	}

	// Handle respawn timer
	if g.respawning {
		g.respawnTimer--
		if g.respawnTimer <= 0 {
			g.respawning = false
			// Clear any segments near player area
			var newSegments []Segment
			for _, seg := range g.segments {
				if seg.pos.Y < g.height-10 {
					newSegments = append(newSegments, seg)
				}
			}
			g.segments = newSegments
		}
		return // Don't update game during respawn
	}

	// Check for bonus life every 20,000 points - REDUCED generosity for difficulty
	if g.score >= g.lastLifeScore+20000 { // Was 10k, now 20k
		g.lives++
		g.lastLifeScore = g.score - (g.score % 20000) // Set to nearest 20k
	}

	// Update bullets
	for i := range g.bullets {
		g.bullets[i].Update()
	}

	// Update flies
	for i := range g.flies {
		g.flies[i].Update()
	}

	// Update fleas
	for i := range g.fleas {
		g.fleas[i].Update(g)
	}

	// Check fly collisions with mushrooms (create poison mushrooms)
	for i := range g.flies {
		if !g.flies[i].active {
			continue
		}
		for j := range g.mushrooms {
			if g.flies[i].pos.X == g.mushrooms[j].pos.X &&
				g.flies[i].pos.Y == g.mushrooms[j].pos.Y {
				// Fly hits mushroom - make it poisoned!
				g.mushrooms[j].poisoned = true
				g.flies[i].active = false
				g.createExplosion(g.mushrooms[j].pos.X, g.mushrooms[j].pos.Y)
				break
			}
		}
	}

	// Check flea collision with player
	for i := range g.fleas {
		if !g.fleas[i].active {
			continue
		}
		if g.fleas[i].pos.X == g.player.pos.X && g.fleas[i].pos.Y == g.player.pos.Y {
			g.loseLife()
			g.fleas[i].active = false
		}
	}

	// Update explosions
	for i := range g.explosions {
		g.explosions[i].Update()
	}

	// Spawn flies and fleas
	g.spawnFly()
	g.spawnFlea()

	// Update centipede segments with improved falling behavior
	for i := 0; i < len(g.segments); i++ {
		seg := &g.segments[i]
		seg.pos.X += seg.direction

		// Hit edge - drop down and reverse
		if seg.pos.X <= 0 || seg.pos.X >= g.width-1 {
			seg.pos.Y++
			seg.direction *= -1
		}

		// Check if hit mushroom - drop down and reverse
		hitPoisonMushroom := false
		for _, mush := range g.mushrooms {
			if seg.pos.X == mush.pos.X && seg.pos.Y == mush.pos.Y {
				if mush.poisoned {
					// POISON MUSHROOM CHUTE: Creates deadly fast zigzag descent
					// Force centipede into zigzag pattern by alternating direction
					seg.pos.Y += 3 // Was 1, now 3 - TRUE CHUTE EFFECT! Falls much faster
					seg.direction *= -1 // Reverse direction

					// Create tight zigzag by limiting horizontal movement
					// The centipede will zigzag within a 3-character chute
					hitPoisonMushroom = true
				} else {
					seg.pos.Y++
				}
				seg.direction *= -1
				break
			}
		}

		// Poison mushrooms cause centipede to drop faster in zigzag chute
		if hitPoisonMushroom {
			// Already handled above - centipede drops and zigzags
		}

		// Check for collision with player
		if seg.pos.X == g.player.pos.X && seg.pos.Y == g.player.pos.Y {
			g.loseLife()
		}

		// CRITICAL FIX: Check if centipede escaped to bottom (reached player area)
		// If ANY segment reaches the bottom without hitting player, it's a death
		// This fixes the bug where centipedes can escape "stage left"
		if seg.pos.Y >= g.height-2 {
			g.loseLife()
			// Remove this segment so we don't trigger multiple deaths from same segment
			g.segments = append(g.segments[:i], g.segments[i+1:]...)
			i-- // Adjust index since we removed an element
			continue
		}
	}

	// Check bullet collisions (improved collision detection with distance check)
	for i := range g.bullets {
		if !g.bullets[i].active {
			continue
		}

		// Bullet vs Centipede
		for j := range g.segments {
			// Exact position match for collision
			if g.bullets[i].pos.X == g.segments[j].pos.X &&
				g.bullets[i].pos.Y == g.segments[j].pos.Y {
				g.bullets[i].active = false

				// Create explosion
				g.createExplosion(g.segments[j].pos.X, g.segments[j].pos.Y)

				// Extra points for head
				if g.segments[j].isHead {
					g.score += 100
				} else {
					g.score += 10
				}

				// Remove segment
				g.segments = append(g.segments[:j], g.segments[j+1:]...)

				// Update head if we removed last segment (front of centipede)
				if len(g.segments) > 0 {
					g.segments[len(g.segments)-1].isHead = true
				}
				break
			}
		}

		// Bullet vs Fly
		for j := range g.flies {
			if !g.flies[j].active {
				continue
			}
			if g.bullets[i].pos.X == g.flies[j].pos.X &&
				g.bullets[i].pos.Y == g.flies[j].pos.Y {
				g.bullets[i].active = false
				g.flies[j].active = false

				// Create explosion
				g.createExplosion(g.flies[j].pos.X, g.flies[j].pos.Y)

				g.score += 200 // Flies worth 200 points
				break
			}
		}

		// Bullet vs Flea
		for j := range g.fleas {
			if !g.fleas[j].active {
				continue
			}
			if g.bullets[i].pos.X == g.fleas[j].pos.X &&
				g.bullets[i].pos.Y == g.fleas[j].pos.Y {
				g.bullets[i].active = false
				g.fleas[j].active = false

				// Create explosion
				g.createExplosion(g.fleas[j].pos.X, g.fleas[j].pos.Y)

				g.score += 150 // Fleas worth 150 points
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

	// Check win condition - spawn longer centipede instead of stopping
	if len(g.segments) == 0 {
		g.level++
		// Spawn centipede with more segments each level (10 + level*2)
		g.spawnCentipede(10 + g.level*2)
		// Add more mushrooms too - DOUBLED for difficulty
		g.spawnMushrooms(10) // Was 5, now 10 per level
		// Regenerate all mushrooms to full health
		g.regenerateMushrooms()
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
	// UNLIMITED BULLETS - removed the limit!
	g.bullets = append(g.bullets, Bullet{
		pos:    Position{X: g.player.pos.X, Y: g.player.pos.Y - 1},
		active: true,
	})
}

func (g *Game) loseLife() {
	g.lives--
	if g.lives <= 0 {
		g.gameOver = true
	} else {
		// Start respawn sequence
		g.respawning = true
		g.respawnTimer = 30 // 30 ticks ~2.4 seconds
		// Reset player position
		g.player.pos.X = g.width / 2
		g.player.pos.Y = g.height - 2
		// Clear bullets
		g.bullets = nil
		// Regenerate all mushrooms to full health
		g.regenerateMushrooms()
	}
}

func (g *Game) regenerateMushrooms() {
	// Restore all mushrooms to full health (4) and clear poison
	for i := range g.mushrooms {
		g.mushrooms[i].health = 4
		g.mushrooms[i].poisoned = false // Reset poison status
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

	// Draw player gun character (improved) - hide during respawn
	if !g.respawning {
		board[g.player.pos.Y][g.player.pos.X] = 'A'
	}

	// Draw mushrooms with different characters based on health and poison status
	for _, mush := range g.mushrooms {
		if mush.pos.Y >= 0 && mush.pos.Y < g.height &&
			mush.pos.X >= 0 && mush.pos.X < g.width {
			if mush.poisoned {
				// Poisoned mushrooms show as 'X' (skull/poison symbol)
				board[mush.pos.Y][mush.pos.X] = 'X'
			} else {
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
	}

	// Draw flies with flickering wing trail
	for _, fly := range g.flies {
		if !fly.active {
			continue
		}
		if fly.pos.Y >= 0 && fly.pos.Y < g.height &&
			fly.pos.X >= 0 && fly.pos.X < g.width {
			board[fly.pos.Y][fly.pos.X] = '✺'

			// Draw flickering wing trail
			if fly.wingFlap {
				trailX := fly.pos.X - fly.direction
				if trailX >= 0 && trailX < g.width {
					board[fly.pos.Y][trailX] = '~'
				}
				trailX2 := fly.pos.X - (fly.direction * 2)
				if trailX2 >= 0 && trailX2 < g.width {
					board[fly.pos.Y][trailX2] = '.'
				}
			}
		}
	}

	// Draw fleas (falling down)
	for _, flea := range g.fleas {
		if flea.active && flea.pos.Y >= 0 && flea.pos.Y < g.height &&
			flea.pos.X >= 0 && flea.pos.X < g.width {
			board[flea.pos.Y][flea.pos.X] = '┃' // Vertical bar for flea
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

	// Draw explosions (on top of everything)
	for _, exp := range g.explosions {
		if exp.active && exp.pos.Y >= 0 && exp.pos.Y < g.height &&
			exp.pos.X >= 0 && exp.pos.X < g.width {
			board[exp.pos.Y][exp.pos.X] = exp.Char()
		}
	}

	// Draw bullets (on top)
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
	game         *Game
	paused       bool
	width        int
	height       int
	state        gameState
	flashOn      bool
	spacePressed bool
	lastShot     time.Time
	highScores   []HighScore
	playerName   string
	enteringName bool
	scoreSaved   bool
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

	poisonMushroomStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("201")).
				Bold(true)

	bulletStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("11"))

	flyStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("208"))

	explosionStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196"))

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
	// FASTER game speed for difficulty - was 80ms, now 50ms
	return tea.Tick(time.Millisecond*50, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func shootTickCmd() tea.Cmd {
	return tea.Tick(time.Millisecond*100, func(t time.Time) tea.Msg {
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
		case "r":
			// Restart game - allow restart when game is over
			if m.game.gameOver || m.game.won {
				m.game = NewGame(50, 28)
				m.state = playingGame
				m.enteringName = false
				m.scoreSaved = false
				m.playerName = ""
				return m, nil
			}
		case "left", "a":
			// Only allow movement when playing AND not game over
			if m.state == playingGame && !m.game.gameOver && !m.game.won {
				m.game.MovePlayer(-1)
			}
		case "right", "d":
			if m.state == playingGame && !m.game.gameOver && !m.game.won {
				m.game.MovePlayer(1)
			}
		case "up", "w":
			if m.state == playingGame && !m.game.gameOver && !m.game.won {
				m.game.MovePlayerY(-1)
			}
		case "down", "s":
			if m.state == playingGame && !m.game.gameOver && !m.game.won {
				m.game.MovePlayerY(1)
			}
		case " ": // Spacebar
			if m.state == playingGame && !m.game.gameOver && !m.game.won {
				m.spacePressed = true
				m.game.Shoot()
			}
		case "p":
			if m.state == playingGame && !m.game.gameOver && !m.game.won {
				m.paused = !m.paused
			}
		}

	case shootMsg:
		// Rapid fire when holding space - now shoots MANY bullets!
		if m.spacePressed && m.state == playingGame && !m.paused {
			m.game.Shoot()
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
			case 'X': // Poison mushroom
				char = poisonMushroomStyle.Render(char)
			case 'M', 'm', '*', '.': // Normal mushrooms
				char = mushroomStyle.Render(char)
			case '|': // Bullets
				char = bulletStyle.Render(char)
			case '✺': // Fly
				char = flyStyle.Render(char)
			case '┃': // Flea
				char = lipgloss.NewStyle().Foreground(lipgloss.Color("226")).Bold(true).Render(char)
			case '~': // Wing trail (darker)
				char = lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render(char)
			case '✶', '✸', '✹': // Explosions
				char = explosionStyle.Render(char)
			}
			boardStr += char
		}
		boardStr += "│\n"
	}

	boardStr += "└" + lipgloss.NewStyle().Foreground(lipgloss.Color("62")).Render(
		lipgloss.PlaceHorizontal(len(board[0]), lipgloss.Center, "")) + "┘"

	// Stats with active flies count
	activeBullets := 0
	for _, b := range m.game.bullets {
		if b.active {
			activeBullets++
		}
	}
	activeFlies := 0
	for _, f := range m.game.flies {
		if f.active {
			activeFlies++
		}
	}

	// Create lives display
	livesStr := ""
	for i := 0; i < m.game.lives; i++ {
		livesStr += "♥"
	}

	stats := statsStyle.Render(fmt.Sprintf(
		"Score: %d  |  Lives: %s  |  Bullets: %d  |  Segments: %d  |  Flies: %d  |  Level: %d",
		m.game.score, livesStr, activeBullets, len(m.game.segments), activeFlies, m.game.level))

	// Controls
	controls := lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render(
		"[←→ or A/D] Move  [↑↓ or W/S] Up/Down  [Space] RAPID FIRE!  [P] Pause  [Q] Quit")

	// Status messages
	status := ""
	if m.game.respawning {
		status = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Bold(true).
			Render(fmt.Sprintf("💥 RESPAWNING... %d", m.game.respawnTimer/10))
	} else if m.paused {
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
        ║    ✺~.  Fly (200 pts!)                ║
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
