package main

import (
	"hangman/hangman"

	"github.com/sirupsen/logrus"
)

func main() {
	hangman, err := hangman.NewHangman()
	if err != nil {
		logrus.Fatal(err)
	}

	hangman.Start()
}
