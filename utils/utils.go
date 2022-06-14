package utils

import (
	"errors"
	"log"
	"os"
)

func CreateFoldersIfNotExists(folder string) {
	_, err := os.Stat(folder)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			err := os.MkdirAll(folder, os.ModePerm)
			if err != nil {
				log.Fatal(err)
			}
		} else {
			log.Fatal(err)
		}
	}
}
