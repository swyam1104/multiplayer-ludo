package game

import (
	"testing"

	"github.com/multiplayer-ludo/internal/models"
)

func TestEngine_Initialization(t *testing.T) {
	eng := NewEngine()
	if eng.State != StateWaiting {
		t.Errorf("expected state waiting, got %v", eng.State)
	}

	eng.AddPlayer("user1", models.ColorRed)
	eng.AddPlayer("user2", models.ColorBlue)

	if len(eng.Players) != 2 {
		t.Errorf("expected 2 players, got %d", len(eng.Players))
	}

	err := eng.StartGame()
	if err != nil {
		t.Fatalf("failed to start game: %v", err)
	}
	if eng.State != StatePlaying {
		t.Errorf("expected state playing, got %v", eng.State)
	}
}

func TestEngine_TurnOrderAndDice(t *testing.T) {
	eng := NewEngineWithSeed(42)
	eng.AddPlayer("user1", models.ColorRed)
	eng.AddPlayer("user2", models.ColorBlue)
	eng.StartGame()

	// Initial state
	if eng.CurrentTurnColor != models.ColorRed {
		t.Errorf("expected red's turn, got %v", eng.CurrentTurnColor)
	}

	// Try rolling for wrong user
	_, err := eng.RollDice("user2")
	if err != ErrNotYourTurn {
		t.Errorf("expected ErrNotYourTurn, got %v", err)
	}

	// Roll for correct user
	val, err := eng.RollDice("user1")
	if err != nil {
		t.Fatalf("expected successful roll, got err: %v", err)
	}

	if val < 1 || val > 6 {
		t.Errorf("expected dice value between 1 and 6, got %d", val)
	}

	// Try rolling again
	_, err = eng.RollDice("user1")
	if err != ErrAlreadyRolled {
		t.Errorf("expected ErrAlreadyRolled, got %v", err)
	}
}

func TestEngine_Movement(t *testing.T) {
	eng := NewEngineWithSeed(42) // Ensure deterministic if needed, though we will manipulate dice manually here
	eng.AddPlayer("user1", models.ColorRed)
	eng.AddPlayer("user2", models.ColorBlue)
	eng.StartGame()

	// Inject dice value 6 to get out of yard
	eng.DiceValue = 6
	err := eng.MoveToken("user1", 0)
	if err != nil {
		t.Fatalf("expected successful move, got %v", err)
	}

	redPlayer := eng.getPlayerByColor(models.ColorRed)
	if redPlayer.Tokens[0].State != TokenStateTrack || redPlayer.Tokens[0].StepsMoved != 1 {
		t.Errorf("expected token on track at step 1, got state %v, step %d", redPlayer.Tokens[0].State, redPlayer.Tokens[0].StepsMoved)
	}
}

func TestEngine_Capture(t *testing.T) {
	eng := NewEngine()
	eng.AddPlayer("user1", models.ColorRed)
	eng.AddPlayer("user2", models.ColorBlue) // Blue offset is 39
	eng.StartGame()

	// Set up capture scenario
	p1 := eng.getPlayerByColor(models.ColorRed)
	p2 := eng.getPlayerByColor(models.ColorBlue)

	p1.Tokens[0].State = TokenStateTrack
	p1.Tokens[0].StepsMoved = 41 // Absolute position 40 (since red offset is 0, pos = 41 - 1 = 40)

	p2.Tokens[0].State = TokenStateTrack
	p2.Tokens[0].StepsMoved = 2 // Absolute position 40 (since blue offset is 39, pos = 39 + 2 - 1 = 40)

	// It's red's turn, red is at 41, blue is at 2.
	// Red is at step 41 (absolute 40). Blue is at step 2 (absolute 40).
	// If Red rolls a 1, red moves to step 42 (absolute 41).
	// Wait, we want red to land on 40.
	// Red is at step 40 (absolute 39). If red rolls 1, red goes to 41 (absolute 40).
	p1.Tokens[0].StepsMoved = 40
	
	eng.DiceValue = 1
	err := eng.MoveToken("user1", 0)
	if err != nil {
		t.Fatalf("move failed: %v", err)
	}

	if p2.Tokens[0].State != TokenStateYard || p2.Tokens[0].StepsMoved != 0 {
		t.Errorf("expected blue token to be captured and in yard, got state %v steps %d", p2.Tokens[0].State, p2.Tokens[0].StepsMoved)
	}

	if p1.Tokens[0].StepsMoved != 41 {
		t.Errorf("expected red token to be at step 41, got %d", p1.Tokens[0].StepsMoved)
	}
}
