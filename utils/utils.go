package utils

import (
	"errors"
	"log"
	"os"
	"strings"

	"github.com/urfave/cli/v2"
)

func AppAuthors(app *cli.App) string {
	var authorNames []string
	for _, author := range app.Authors {
		authorNames = append(authorNames, author.Name)
	}
	return strings.Join(authorNames, ", ")
}

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
