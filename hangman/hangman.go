package hangman

import (
	"errors"
	"fmt"
	"strings"

	"github.com/manifoldco/promptui"
	"github.com/sirupsen/logrus"
)

const (
	DefaultDataDir           = "data"
	DefaultAdditionalGuesses = 3
	PointsPerCorrectGuess    = 10
)

type GameState string

const (
	GameStatePending = GameState("pending")
	GameStatePlaying = GameState("playing")
	GameStateWin     = GameState("win")
	GameStateLose    = GameState("lose")
	GameStateQuit    = GameState("quit")
)

type Hangman struct {
	WordLoader *WordLoader

	gameState            GameState
	additionalMaxGuesses int
}

// NewHangman creates a new Hangman game instance and loads words from the default data directory.
// Returns an error if the word data cannot be loaded.
func NewHangman() (*Hangman, error) {
	wordLoader := NewWordLoader()
	if err := wordLoader.Load(DefaultDataDir); err != nil {
		return nil, err
	}

	return &Hangman{
		WordLoader:           wordLoader,
		gameState:            GameStatePending,
		additionalMaxGuesses: DefaultAdditionalGuesses,
	}, nil
}

// Start runs the main game loop, managing state transitions between pending, playing, win, lose, and quit states.
// The loop continues until the user quits the game.
func (h *Hangman) Start() {
	var game *HangmanGame
	for {
		switch h.gameState {
		case GameStatePending:
			var err error
			game, h.gameState, err = h.createGame()
			if err != nil {
				logrus.WithError(err).Error("failed to create game")
				h.gameState = GameStateQuit
			}
		case GameStatePlaying:
			if game != nil {
				h.gameState = game.Play()
			} else {
				logrus.Error("game is nil in playing state")
				h.gameState = GameStateQuit
			}
		case GameStateWin:
			logrus.Info("🎉 You win!")
			h.gameState = GameStatePending
		case GameStateLose:
			logrus.Info("😢 You lose!")
			h.gameState = GameStatePending
		case GameStateQuit:
			logrus.Info("👋 Quit...")
			return
		}
	}
}

// createGame displays a category selection menu and creates a new game instance.
// Returns the created game, the next game state, and any error that occurred.
// If the user selects quit or an error occurs, returns GameStateQuit.
func (h *Hangman) createGame() (*HangmanGame, GameState, error) {
	categories := h.WordLoader.Categories()
	items := make([]string, 0)
	for _, category := range categories {
		items = append(items, fmt.Sprintf("📂 %s", category))
	}
	items = append(items, "❌ Quit")

	prompt := promptui.Select{
		Label: "Hangman Menu - Select Category",
		Items: items,
	}

	idx, _, err := prompt.Run()
	if err != nil {
		return nil, GameStateQuit, fmt.Errorf("prompt failed: %w", err)
	}

	if idx == len(items)-1 {
		return nil, GameStateQuit, nil
	}

	word, err := h.WordLoader.RandomWord(categories[idx])
	if err != nil {
		return nil, GameStateQuit, fmt.Errorf("failed to get random word: %w", err)
	}

	game, err := NewHangmanGame(word, h.additionalMaxGuesses)
	if err != nil {
		return nil, GameStateQuit, fmt.Errorf("failed to create game: %w", err)
	}

	return game, GameStatePlaying, nil
}

type HangmanGame struct {
	hint           string
	wordIndices    map[string][]int
	guesses        map[string]bool
	answer         []string
	incorrect      []string
	alphabetLength int
	fill           int
	remaining      int
	score          int
	streak         int
}

// NewHangmanGame creates a new game instance for a specific word.
// The remaining attempts are calculated as the word's alphabet length plus additional guesses.
// Returns an error if the word is nil, empty, or if maxGuesses is negative.
func NewHangmanGame(word *Word, maxGuesses int) (*HangmanGame, error) {
	if word == nil {
		return nil, errors.New("word cannot be nil")
	}
	if len(word.Text) == 0 {
		return nil, errors.New("word text cannot be empty")
	}
	if maxGuesses < 0 {
		return nil, errors.New("maxGuesses cannot be negative")
	}

	return &HangmanGame{
		hint:           word.Hint,
		wordIndices:    word.Indices(),
		answer:         word.PreAnswer(),
		alphabetLength: word.AlphabetLength(),
		remaining:      word.AlphabetLength() + maxGuesses,
		incorrect:      make([]string, 0),
		guesses:        make(map[string]bool),
		score:          0,
		streak:         0,
	}, nil
}

// Play runs the main game loop for a single round of Hangman.
// Displays the hint, collects user input, processes guesses, and checks for win/loss conditions.
// Returns the next game state: GameStateWin, GameStateLose, or GameStateQuit.
func (g *HangmanGame) Play() GameState {
	logrus.Info("Hint: ", g.hint)
	for g.remaining > 0 {
		g.displayAnswer()
		letter, err := g.input()
		if err != nil {
			logrus.WithError(err).Error("failed to input the guess letter")
			return GameStateQuit
		}

		g.processGuess(letter)
		if g.isWin() {
			g.displayAnswer() // Show final answer
			return GameStateWin
		}
	}

	g.displayAnswer() // Show final answer on loss
	return GameStateLose
}

// displayAnswer shows the current game state including the answer with guessed letters,
// the current score, remaining attempts, and incorrect guesses.
func (g *HangmanGame) displayAnswer() {
	builder := new(strings.Builder)
	builder.WriteString(strings.Join(g.answer, " "))
	fmt.Fprintf(builder, "\tscore: %d,", g.score)
	fmt.Fprintf(builder, "\tremaining: %d", g.remaining)
	fmt.Fprintf(builder, "\tincorrect: %s", strings.Join(g.incorrect, ","))
	logrus.Info(builder.String())
}

// input prompts the user to enter a single alphabetic character guess.
// Validates that the input is exactly one alphabetic character.
// Returns the lowercase letter and any error that occurred during input.
func (g *HangmanGame) input() (string, error) {
	prompt := promptui.Prompt{
		Label: ">",
		Validate: func(s string) error {
			if len(s) != 1 {
				return errors.New("you must input single character")
			}

			if !IsAlphabet(s) {
				return errors.New("you must input alphabetic character")
			}

			return nil
		},
	}

	in, err := prompt.Run()
	if err != nil {
		return "", fmt.Errorf("prompt failed: %w", err)
	}
	return strings.ToLower(in), nil
}

// processGuess processes a letter guess and updates the game state accordingly.
// If the letter was already guessed, decrements remaining and resets streak.
// If the letter is incorrect, adds it to incorrect list and resets streak.
// If the letter is correct, fills in the answer positions and increases score based on streak.
func (g *HangmanGame) processGuess(letter string) {
	if g.guesses[letter] {
		g.streak = 0
		g.remaining--
		logrus.Warn("already guessed")
		return
	}

	locs, ok := g.wordIndices[letter]
	if !ok {
		g.remaining--
		g.streak = 0
		g.guesses[letter] = true
		g.incorrect = append(g.incorrect, letter)
		return
	}

	for _, loc := range locs {
		g.fill++
		g.answer[loc] = letter
	}

	g.streak++
	g.score += PointsPerCorrectGuess * g.streak
	g.guesses[letter] = true
}

// isWin checks if the player has won by comparing filled letters to the total alphabet length.
// Returns true if all alphabetic characters in the word have been guessed.
func (g *HangmanGame) isWin() bool {
	return g.fill == g.alphabetLength
}
