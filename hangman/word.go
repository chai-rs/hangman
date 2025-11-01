package hangman

import (
	"bufio"
	"errors"
	"io/fs"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
)

// Word represents a word with its hint.
type Word struct {
	Text string
	Hint string
}

// Indices returns a map of letters to their positions in the word.
func (w *Word) Indices() map[string][]int {
	indices := make(map[string][]int)
	for i, ch := range w.Text {
		letter := strings.ToLower(string(ch))
		indices[letter] = append(indices[letter], i)
	}
	return indices
}

// PreAnswer generates the initial answer slice with underscores.
func (w *Word) PreAnswer() []string {
	answer := make([]string, len(w.Text))
	for i := 0; i < len(w.Text); i++ {
		if IsAlphabet(string(w.Text[i])) {
			answer[i] = "_"
		} else {
			answer[i] = string(w.Text[i])
		}
	}
	return answer
}

func (w *Word) AlphabetLength() int {
	count := 0
	for _, ch := range w.Text {
		if IsAlphabet(string(ch)) {
			count++
		}
	}
	return count
}

// WordLoader is responsible for loading words from files.
type WordLoader struct {
	categories    []string
	categoryWords map[string][]Word
}

// NewWordLoader creates a new instance of WordLoader.
func NewWordLoader() *WordLoader {
	return &WordLoader{
		categories:    []string{},
		categoryWords: make(map[string][]Word),
	}
}

// Load loads words from the specified directory path.
func (l *WordLoader) Load(path string) error {
	return filepath.Walk(path, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		category, words, err := l.LoadFile(path)
		if err != nil {
			return err
		}

		l.categoryWords[category] = append(l.categoryWords[category], words...)
		l.categories = append(l.categories, category)
		return nil
	})
}

// LoadFile is a stub function for loading words from a file. which will returned the category, words slice and error.
func (l *WordLoader) LoadFile(path string) (string, []Word, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	category := ""
	words := []Word{}

	for scanner.Scan() {
		line := scanner.Text()
		if category == "" {
			category = line
		} else {
			texts := strings.Split(line, ",")

			if len(texts) != 2 {
				return "", nil, errors.New("invalid word format in file, expected 'word,hint'")
			}

			words = append(words, Word{Text: strings.TrimSpace(texts[0]), Hint: strings.TrimSpace(texts[1])})
		}
	}

	if err := scanner.Err(); err != nil {
		return "", nil, err
	}

	if category == "" || len(words) == 0 {
		return "", nil, errors.New("file is missing category or words")
	}

	return category, words, nil
}

// GetWords retrieves words for a given category.
func (l *WordLoader) GetWords(category string) ([]Word, error) {
	words, ok := l.categoryWords[category]
	if !ok {
		return nil, errors.New("category not found")
	}

	if len(words) == 0 {
		return nil, errors.New("no words available in this category")
	}

	return words, nil
}

// Categories returns the list of available categories.
func (l *WordLoader) Categories() []string {
	return l.categories
}

// RandomWord retrieves a random word from the specified category.
func (l *WordLoader) RandomWord(category string) (*Word, error) {
	words, err := l.GetWords(category)
	if err != nil {
		return nil, err
	}

	index := rand.Intn(len(words))
	return &words[index], nil
}
