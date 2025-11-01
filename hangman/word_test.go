package hangman

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestWord_Indices(t *testing.T) {
	tests := []struct {
		name     string
		word     Word
		expected map[string][]int
	}{
		{
			name: "simple word",
			word: Word{Text: "hello"},
			expected: map[string][]int{
				"h": {0},
				"e": {1},
				"l": {2, 3},
				"o": {4},
			},
		},
		{
			name: "word with duplicates",
			word: Word{Text: "banana"},
			expected: map[string][]int{
				"b": {0},
				"a": {1, 3, 5},
				"n": {2, 4},
			},
		},
		{
			name: "mixed case",
			word: Word{Text: "Hello"},
			expected: map[string][]int{
				"h": {0},
				"e": {1},
				"l": {2, 3},
				"o": {4},
			},
		},
		{
			name: "single character",
			word: Word{Text: "a"},
			expected: map[string][]int{
				"a": {0},
			},
		},
		{
			name: "word with space",
			word: Word{Text: "hi world"},
			expected: map[string][]int{
				"h": {0},
				"i": {1},
				" ": {2},
				"w": {3},
				"o": {4},
				"r": {5},
				"l": {6},
				"d": {7},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.word.Indices()
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("Word.Indices() = %v; want %v", result, tt.expected)
			}
		})
	}
}

func TestWord_PreAnswer(t *testing.T) {
	tests := []struct {
		name     string
		word     Word
		expected []string
	}{
		{
			name:     "all letters",
			word:     Word{Text: "hello"},
			expected: []string{"_", "_", "_", "_", "_"},
		},
		{
			name:     "with space",
			word:     Word{Text: "hello world"},
			expected: []string{"_", "_", "_", "_", "_", " ", "_", "_", "_", "_", "_"},
		},
		{
			name:     "with apostrophe",
			word:     Word{Text: "it's"},
			expected: []string{"_", "_", "'", "_"},
		},
		{
			name:     "single character",
			word:     Word{Text: "a"},
			expected: []string{"_"},
		},
		{
			name:     "with hyphen",
			word:     Word{Text: "co-op"},
			expected: []string{"_", "_", "-", "_", "_"},
		},
		{
			name:     "empty word",
			word:     Word{Text: ""},
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.word.PreAnswer()
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("Word.PreAnswer() = %v; want %v", result, tt.expected)
			}
		})
	}
}

func TestWord_AlphabetLength(t *testing.T) {
	tests := []struct {
		name     string
		word     Word
		expected int
	}{
		{
			name:     "all letters",
			word:     Word{Text: "hello"},
			expected: 5,
		},
		{
			name:     "with space",
			word:     Word{Text: "hello world"},
			expected: 10,
		},
		{
			name:     "with punctuation",
			word:     Word{Text: "it's cool"},
			expected: 7,
		},
		{
			name:     "no letters",
			word:     Word{Text: "123 !@#"},
			expected: 0,
		},
		{
			name:     "single character",
			word:     Word{Text: "a"},
			expected: 1,
		},
		{
			name:     "empty word",
			word:     Word{Text: ""},
			expected: 0,
		},
		{
			name:     "mixed alphanumeric",
			word:     Word{Text: "abc123"},
			expected: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.word.AlphabetLength()
			if result != tt.expected {
				t.Errorf("Word.AlphabetLength() = %d; want %d", result, tt.expected)
			}
		})
	}
}

func TestWordLoader_LoadFile(t *testing.T) {
	tests := []struct {
		name          string
		filePath      string
		wantCategory  string
		wantWordCount int
		wantErr       bool
	}{
		{
			name:          "valid file",
			filePath:      "testdata/fruits.txt",
			wantCategory:  "Fruits",
			wantWordCount: 3,
			wantErr:       false,
		},
		{
			name:     "file not found",
			filePath: "testdata/nonexistent.txt",
			wantErr:  true,
		},
		{
			name:     "empty file",
			filePath: "testdata/empty.txt",
			wantErr:  true,
		},
		{
			name:     "invalid format - no comma",
			filePath: "testdata/invalid.txt",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loader := NewWordLoader()
			category, words, err := loader.LoadFile(tt.filePath)

			if (err != nil) != tt.wantErr {
				t.Errorf("WordLoader.LoadFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if category != tt.wantCategory {
					t.Errorf("WordLoader.LoadFile() category = %v, want %v", category, tt.wantCategory)
				}
				if len(words) != tt.wantWordCount {
					t.Errorf("WordLoader.LoadFile() word count = %v, want %v", len(words), tt.wantWordCount)
				}
				if len(words) > 0 {
					if words[0].Text == "" || words[0].Hint == "" {
						t.Errorf("WordLoader.LoadFile() word has empty text or hint")
					}
				}
			}
		})
	}
}

func TestWordLoader_Load(t *testing.T) {
	tests := []struct {
		name     string
		dirPath  string
		wantErr  bool
		wantCats int
	}{
		{
			name:     "valid directory",
			dirPath:  "testdata/valid_data",
			wantErr:  false,
			wantCats: 2,
		},
		{
			name:     "non-existent directory",
			dirPath:  "testdata/nonexistent",
			wantErr:  true,
			wantCats: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loader := NewWordLoader()
			err := loader.Load(tt.dirPath)

			if (err != nil) != tt.wantErr {
				t.Errorf("WordLoader.Load() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				categories := loader.Categories()
				if len(categories) < tt.wantCats {
					t.Errorf("WordLoader.Load() loaded %v categories, want at least %v", len(categories), tt.wantCats)
				}
			}
		})
	}
}

func TestWordLoader_GetWords(t *testing.T) {
	loader := NewWordLoader()
	loader.categoryWords["TestCategory"] = []Word{
		{Text: "Apple", Hint: "A fruit"},
		{Text: "Banana", Hint: "Another fruit"},
	}
	loader.categories = []string{"TestCategory"}

	tests := []struct {
		name     string
		category string
		wantErr  bool
		wantLen  int
	}{
		{
			name:     "valid category",
			category: "TestCategory",
			wantErr:  false,
			wantLen:  2,
		},
		{
			name:     "invalid category",
			category: "NonExistent",
			wantErr:  true,
			wantLen:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			words, err := loader.GetWords(tt.category)

			if (err != nil) != tt.wantErr {
				t.Errorf("WordLoader.GetWords() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && len(words) != tt.wantLen {
				t.Errorf("WordLoader.GetWords() returned %v words, want %v", len(words), tt.wantLen)
			}
		})
	}
}

func TestWordLoader_Categories(t *testing.T) {
	loader := NewWordLoader()

	if len(loader.Categories()) != 0 {
		t.Errorf("New WordLoader should have 0 categories, got %v", len(loader.Categories()))
	}

	loader.categories = []string{"Cat1", "Cat2", "Cat3"}
	categories := loader.Categories()

	if len(categories) != 3 {
		t.Errorf("WordLoader.Categories() returned %v categories, want 3", len(categories))
	}

	expected := []string{"Cat1", "Cat2", "Cat3"}
	if !reflect.DeepEqual(categories, expected) {
		t.Errorf("WordLoader.Categories() = %v, want %v", categories, expected)
	}
}

func TestWordLoader_RandomWord(t *testing.T) {
	loader := NewWordLoader()
	loader.categoryWords["TestCategory"] = []Word{
		{Text: "Apple", Hint: "A fruit"},
		{Text: "Banana", Hint: "Another fruit"},
		{Text: "Orange", Hint: "Citrus fruit"},
	}

	tests := []struct {
		name     string
		category string
		wantErr  bool
	}{
		{
			name:     "valid category",
			category: "TestCategory",
			wantErr:  false,
		},
		{
			name:     "invalid category",
			category: "NonExistent",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			word, err := loader.RandomWord(tt.category)

			if (err != nil) != tt.wantErr {
				t.Errorf("WordLoader.RandomWord() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if word == nil {
					t.Error("WordLoader.RandomWord() returned nil word")
					return
				}
				if word.Text == "" || word.Hint == "" {
					t.Error("WordLoader.RandomWord() returned word with empty text or hint")
				}
			}
		})
	}
}

func TestWordLoader_Integration(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}

	testDataPath := filepath.Join(cwd, "testdata", "valid_data")

	loader := NewWordLoader()
	err = loader.Load(testDataPath)
	if err != nil {
		t.Fatalf("Failed to load test data: %v", err)
	}

	categories := loader.Categories()
	if len(categories) == 0 {
		t.Error("Expected at least one category to be loaded")
	}

	if len(categories) > 0 {
		words, err := loader.GetWords(categories[0])
		if err != nil {
			t.Errorf("Failed to get words from category %s: %v", categories[0], err)
		}
		if len(words) == 0 {
			t.Errorf("Expected words in category %s", categories[0])
		}

		randomWord, err := loader.RandomWord(categories[0])
		if err != nil {
			t.Errorf("Failed to get random word from category %s: %v", categories[0], err)
		}
		if randomWord == nil {
			t.Errorf("Random word should not be nil")
		}
	}
}
