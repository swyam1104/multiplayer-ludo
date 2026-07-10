package game

import (
	"errors"
	"math/rand"
	"time"

	"github.com/multiplayer-ludo/internal/models"
)

var (
	ErrGameNotStarted = errors.New("game has not started")
	ErrGameOver       = errors.New("game is already over")
	ErrNotYourTurn    = errors.New("not your turn")
	ErrAlreadyRolled  = errors.New("dice already rolled this turn")
	ErrMustRollFirst  = errors.New("must roll dice first")
	ErrInvalidToken   = errors.New("invalid token")
	ErrInvalidMove    = errors.New("invalid move for this token")
)

const (
	TotalStepsToHome = 57
)

type TokenState string

const (
	TokenStateYard        TokenState = "yard"
	TokenStateTrack       TokenState = "track"
	TokenStateHomeStretch TokenState = "home_stretch"
	TokenStateHome        TokenState = "home"
)

type Token struct {
	ID         int               `json:"id"`
	Color      models.TokenColor `json:"color"`
	State      TokenState        `json:"state"`
	StepsMoved int               `json:"steps_moved"`
}

type Player struct {
	UserID string            `json:"user_id"`
	Color  models.TokenColor `json:"color"`
	Tokens []*Token          `json:"tokens"`
}

type GameState string

const (
	StateWaiting  GameState = "waiting"
	StatePlaying  GameState = "playing"
	StateFinished GameState = "finished"
)

type Engine struct {
	State            GameState
	Players          []*Player
	TurnOrder        []models.TokenColor
	CurrentTurnIdx   int
	CurrentTurnColor models.TokenColor
	DiceValue        int
	ConsecutiveSixes int
	Winner           models.TokenColor
	rand             *rand.Rand
}

func NewEngine() *Engine {
	return &Engine{
		State:   StateWaiting,
		Players: make([]*Player, 0),
		rand:    rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// Deterministic engine for testing
func NewEngineWithSeed(seed int64) *Engine {
	return &Engine{
		State:   StateWaiting,
		Players: make([]*Player, 0),
		rand:    rand.New(rand.NewSource(seed)),
	}
}

func (e *Engine) AddPlayer(userID string, color models.TokenColor) error {
	if e.State != StateWaiting {
		return errors.New("cannot add player, game already started")
	}
	if len(e.Players) >= 4 {
		return errors.New("game is full")
	}

	p := &Player{
		UserID: userID,
		Color:  color,
		Tokens: make([]*Token, 4),
	}
	for i := 0; i < 4; i++ {
		p.Tokens[i] = &Token{
			ID:         i,
			Color:      color,
			State:      TokenStateYard,
			StepsMoved: 0,
		}
	}
	e.Players = append(e.Players, p)
	e.TurnOrder = append(e.TurnOrder, color)
	return nil
}

func (e *Engine) StartGame() error {
	if len(e.Players) < 2 {
		return errors.New("need at least 2 players to start")
	}
	e.State = StatePlaying
	e.CurrentTurnIdx = 0
	e.CurrentTurnColor = e.TurnOrder[0]
	return nil
}

func (e *Engine) RollDice(userID string) (int, error) {
	if e.State != StatePlaying {
		return 0, ErrGameNotStarted
	}
	currentPlayer := e.GetPlayerByColor(e.CurrentTurnColor)
	if currentPlayer.UserID != userID {
		return 0, ErrNotYourTurn
	}
	if e.DiceValue != 0 {
		return 0, ErrAlreadyRolled
	}

	val := e.rand.Intn(6) + 1
	e.DiceValue = val

	if val == 6 {
		e.ConsecutiveSixes++
	} else {
		e.ConsecutiveSixes = 0
	}

	// 3 consecutive 6s forfeits the turn
	if e.ConsecutiveSixes == 3 {
		e.DiceValue = 0
		e.ConsecutiveSixes = 0
		e.nextTurn()
		return val, nil // Forfeited, handled transparently
	}

	// If no valid moves are possible, automatically skip to next turn.
	if len(e.GetValidMoves(e.CurrentTurnColor, e.DiceValue)) == 0 {
		e.DiceValue = 0
		if val != 6 {
			e.nextTurn()
		}
	}

	return val, nil
}

func (e *Engine) MoveToken(userID string, tokenID int) error {
	if e.State != StatePlaying {
		return ErrGameNotStarted
	}
	currentPlayer := e.GetPlayerByColor(e.CurrentTurnColor)
	if currentPlayer.UserID != userID {
		return ErrNotYourTurn
	}
	if e.DiceValue == 0 {
		return ErrMustRollFirst
	}
	if tokenID < 0 || tokenID > 3 {
		return ErrInvalidToken
	}

	token := currentPlayer.Tokens[tokenID]
	valid := false
	validMoves := e.GetValidMoves(e.CurrentTurnColor, e.DiceValue)
	for _, id := range validMoves {
		if id == tokenID {
			valid = true
			break
		}
	}
	if !valid {
		return ErrInvalidMove
	}

	// Apply move
	if token.State == TokenStateYard {
		token.State = TokenStateTrack
		token.StepsMoved = 1
	} else {
		token.StepsMoved += e.DiceValue
		if token.StepsMoved > 51 && token.StepsMoved < TotalStepsToHome {
			token.State = TokenStateHomeStretch
		} else if token.StepsMoved == TotalStepsToHome {
			token.State = TokenStateHome
		}
	}

	// Check captures
	captureOccurred := e.checkCaptures(token)

	// Check win condition
	if e.checkWin(currentPlayer) {
		e.State = StateFinished
		e.Winner = e.CurrentTurnColor
		return nil
	}

	valRolled := e.DiceValue
	e.DiceValue = 0

	// Extra turn logic: rolled a 6 OR captured an opponent OR token reached home
	// MVP simplifies extra turn to just rolling a 6 for now to reduce complexity,
	// but let's include capture and home for completeness if possible.
	extraTurn := valRolled == 6 || captureOccurred || token.State == TokenStateHome

	if !extraTurn {
		e.nextTurn()
	}

	return nil
}

func (e *Engine) GetValidMoves(color models.TokenColor, diceVal int) []int {
	player := e.GetPlayerByColor(color)
	var valid []int
	for _, t := range player.Tokens {
		if t.State == TokenStateHome {
			continue
		}
		if t.State == TokenStateYard {
			if diceVal == 6 {
				valid = append(valid, t.ID)
			}
			continue
		}
		if t.StepsMoved+diceVal <= TotalStepsToHome {
			valid = append(valid, t.ID)
		}
	}
	return valid
}

func (e *Engine) nextTurn() {
	e.CurrentTurnIdx = (e.CurrentTurnIdx + 1) % len(e.TurnOrder)
	e.CurrentTurnColor = e.TurnOrder[e.CurrentTurnIdx]
	e.DiceValue = 0
	e.ConsecutiveSixes = 0
}

func (e *Engine) GetPlayerByColor(color models.TokenColor) *Player {
	for _, p := range e.Players {
		if p.Color == color {
			return p
		}
	}
	return nil
}

func (e *Engine) checkWin(p *Player) bool {
	for _, t := range p.Tokens {
		if t.State != TokenStateHome {
			return false
		}
	}
	return true
}

// Convert logical steps (1-51) to absolute track cell (0-51)
func (e *Engine) getAbsoluteTrackPosition(color models.TokenColor, steps int) int {
	var offset int
	switch color {
	case models.ColorRed:
		offset = 0
	case models.ColorGreen:
		offset = 13
	case models.ColorYellow:
		offset = 26
	case models.ColorBlue:
		offset = 39
	}
	return (offset + steps - 1) % 52
}

func isSafeCell(pos int) bool {
	safeCells := []int{0, 8, 13, 21, 26, 34, 39, 47}
	for _, sc := range safeCells {
		if pos == sc {
			return true
		}
	}
	return false
}

func (e *Engine) checkCaptures(movedToken *Token) bool {
	if movedToken.State != TokenStateTrack {
		return false
	}

	movedAbsPos := e.getAbsoluteTrackPosition(movedToken.Color, movedToken.StepsMoved)

	if isSafeCell(movedAbsPos) {
		return false
	}

	captured := false
	for _, p := range e.Players {
		if p.Color == movedToken.Color {
			continue
		}
		for _, t := range p.Tokens {
			if t.State == TokenStateTrack {
				absPos := e.getAbsoluteTrackPosition(t.Color, t.StepsMoved)
				if absPos == movedAbsPos {
					// Capture!
					t.State = TokenStateYard
					t.StepsMoved = 0
					captured = true
				}
			}
		}
	}
	return captured
}
