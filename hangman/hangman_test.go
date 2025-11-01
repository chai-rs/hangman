package hangman

import (
	"testing"
)

func TestNewHangmanGame(t *testing.T) {
	validWord := &Word{
		Text: "hello",
		Hint: "A greeting",
	}

	tests := []struct {
		name       string
		word       *Word
		maxGuesses int
		wantErr    bool
		checkFunc  func(*testing.T, *HangmanGame)
	}{
		{
			name:       "valid word and maxGuesses",
			word:       validWord,
			maxGuesses: 3,
			wantErr:    false,
			checkFunc: func(t *testing.T, game *HangmanGame) {
				if game.hint != "A greeting" {
					t.Errorf("Expected hint 'A greeting', got '%s'", game.hint)
				}
				if game.remaining != 8 {
					t.Errorf("Expected remaining 8 (5 letters + 3 guesses), got %d", game.remaining)
				}
				if game.score != 0 {
					t.Errorf("Expected initial score 0, got %d", game.score)
				}
				if game.streak != 0 {
					t.Errorf("Expected initial streak 0, got %d", game.streak)
				}
				if game.alphabetLength != 5 {
					t.Errorf("Expected alphabet length 5, got %d", game.alphabetLength)
				}
			},
		},
		{
			name:       "nil word",
			word:       nil,
			maxGuesses: 3,
			wantErr:    true,
		},
		{
			name:       "empty word text",
			word:       &Word{Text: "", Hint: "Empty"},
			maxGuesses: 3,
			wantErr:    true,
		},
		{
			name:       "negative maxGuesses",
			word:       validWord,
			maxGuesses: -1,
			wantErr:    true,
		},
		{
			name:       "zero maxGuesses",
			word:       validWord,
			maxGuesses: 0,
			wantErr:    false,
			checkFunc: func(t *testing.T, game *HangmanGame) {
				if game.remaining != 5 {
					t.Errorf("Expected remaining 5 (5 letters + 0 guesses), got %d", game.remaining)
				}
			},
		},
		{
			name:       "word with spaces",
			word:       &Word{Text: "hello world", Hint: "Greeting phrase"},
			maxGuesses: 3,
			wantErr:    false,
			checkFunc: func(t *testing.T, game *HangmanGame) {
				if game.alphabetLength != 10 {
					t.Errorf("Expected alphabet length 10, got %d", game.alphabetLength)
				}
				if game.remaining != 13 {
					t.Errorf("Expected remaining 13 (10 letters + 3 guesses), got %d", game.remaining)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			game, err := NewHangmanGame(tt.word, tt.maxGuesses)

			if (err != nil) != tt.wantErr {
				t.Errorf("NewHangmanGame() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.checkFunc != nil {
				tt.checkFunc(t, game)
			}
		})
	}
}

func TestHangmanGame_processGuess(t *testing.T) {
	word := &Word{
		Text: "hello",
		Hint: "A greeting",
	}

	tests := []struct {
		name           string
		letter         string
		setupFunc      func(*HangmanGame)
		checkFunc      func(*testing.T, *HangmanGame)
	}{
		{
			name:   "correct guess - first time",
			letter: "h",
			setupFunc: func(g *HangmanGame) {
			},
			checkFunc: func(t *testing.T, g *HangmanGame) {
				if g.score != 10 {
					t.Errorf("Expected score 10, got %d", g.score)
				}
				if g.streak != 1 {
					t.Errorf("Expected streak 1, got %d", g.streak)
				}
				if g.fill != 1 {
					t.Errorf("Expected fill 1, got %d", g.fill)
				}
				if g.answer[0] != "h" {
					t.Errorf("Expected answer[0] = 'h', got '%s'", g.answer[0])
				}
				if !g.guesses["h"] {
					t.Error("Expected 'h' to be marked as guessed")
				}
			},
		},
		{
			name:   "correct guess - multiple occurrences",
			letter: "l",
			setupFunc: func(g *HangmanGame) {
			},
			checkFunc: func(t *testing.T, g *HangmanGame) {
				if g.fill != 2 {
					t.Errorf("Expected fill 2 (two 'l's), got %d", g.fill)
				}
				if g.answer[2] != "l" || g.answer[3] != "l" {
					t.Errorf("Expected answer[2] and answer[3] to be 'l'")
				}
			},
		},
		{
			name:   "incorrect guess",
			letter: "x",
			setupFunc: func(g *HangmanGame) {
			},
			checkFunc: func(t *testing.T, g *HangmanGame) {
				if g.remaining != 7 {
					t.Errorf("Expected remaining 7, got %d", g.remaining)
				}
				if g.streak != 0 {
					t.Errorf("Expected streak 0, got %d", g.streak)
				}
				if len(g.incorrect) != 1 || g.incorrect[0] != "x" {
					t.Errorf("Expected incorrect = ['x'], got %v", g.incorrect)
				}
			},
		},
		{
			name:   "duplicate guess",
			letter: "h",
			setupFunc: func(g *HangmanGame) {
				g.guesses["h"] = true
				g.answer[0] = "h"
			},
			checkFunc: func(t *testing.T, g *HangmanGame) {
				if g.remaining != 7 {
					t.Errorf("Expected remaining 7 (decreased for duplicate), got %d", g.remaining)
				}
				if g.streak != 0 {
					t.Errorf("Expected streak 0 (reset), got %d", g.streak)
				}
			},
		},
		{
			name:   "streak bonus",
			letter: "e",
			setupFunc: func(g *HangmanGame) {
				g.guesses["h"] = true
				g.answer[0] = "h"
				g.fill = 1
				g.streak = 1
				g.score = 10
			},
			checkFunc: func(t *testing.T, g *HangmanGame) {
				if g.streak != 2 {
					t.Errorf("Expected streak 2, got %d", g.streak)
				}
				if g.score != 30 {
					t.Errorf("Expected score 30 (10 + 10*2), got %d", g.score)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			game, err := NewHangmanGame(word, 3)
			if err != nil {
				t.Fatalf("Failed to create game: %v", err)
			}

			if tt.setupFunc != nil {
				tt.setupFunc(game)
			}

			game.processGuess(tt.letter)

			if tt.checkFunc != nil {
				tt.checkFunc(t, game)
			}
		})
	}
}

func TestHangmanGame_isWin(t *testing.T) {
	word := &Word{
		Text: "hello",
		Hint: "A greeting",
	}

	tests := []struct {
		name      string
		setupFunc func(*HangmanGame)
		wantWin   bool
	}{
		{
			name: "not won - no guesses",
			setupFunc: func(g *HangmanGame) {
			},
			wantWin: false,
		},
		{
			name: "not won - partial guesses",
			setupFunc: func(g *HangmanGame) {
				g.fill = 3
				g.alphabetLength = 5
			},
			wantWin: false,
		},
		{
			name: "won - all letters guessed",
			setupFunc: func(g *HangmanGame) {
				g.fill = 5
				g.alphabetLength = 5
			},
			wantWin: true,
		},
		{
			name: "edge case - zero alphabet length",
			setupFunc: func(g *HangmanGame) {
				g.fill = 0
				g.alphabetLength = 0
			},
			wantWin: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			game, err := NewHangmanGame(word, 3)
			if err != nil {
				t.Fatalf("Failed to create game: %v", err)
			}

			if tt.setupFunc != nil {
				tt.setupFunc(game)
			}

			result := game.isWin()
			if result != tt.wantWin {
				t.Errorf("isWin() = %v, want %v", result, tt.wantWin)
			}
		})
	}
}

func TestHangmanGame_FullGameScenario(t *testing.T) {
	word := &Word{
		Text: "cat",
		Hint: "A pet",
	}

	game, err := NewHangmanGame(word, 2)
	if err != nil {
		t.Fatalf("Failed to create game: %v", err)
	}

	if game.isWin() {
		t.Error("Game should not be won initially")
	}

	game.processGuess("c")
	if game.score != 10 {
		t.Errorf("Expected score 10 after first correct guess, got %d", game.score)
	}
	if game.isWin() {
		t.Error("Game should not be won after one letter")
	}

	game.processGuess("a")
	if game.score != 30 {
		t.Errorf("Expected score 30 (10 + 20), got %d", game.score)
	}

	game.processGuess("x")
	if game.streak != 0 {
		t.Errorf("Expected streak reset to 0, got %d", game.streak)
	}
	if game.remaining != 4 {
		t.Errorf("Expected remaining 4, got %d", game.remaining)
	}

	game.processGuess("t")
	if !game.isWin() {
		t.Error("Game should be won after all letters guessed")
	}
	if game.fill != 3 {
		t.Errorf("Expected fill 3, got %d", game.fill)
	}
}
