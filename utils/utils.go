package utils

import (
	"database/sql"
	"encoding/hex"
	"errors"
	"log"
	"os"
	"strings"

	"github.com/deroproject/derohe/cryptography/crypto"
	"github.com/deroproject/derohe/rpc"
	"github.com/g45t345rt/derosphere/config"
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

func OpenDB(env string) *sql.DB {
	filePath := config.DATA_FOLDER + "/" + env + ".db"
	CreateFoldersIfNotExists(config.DATA_FOLDER)
	db, err := sql.Open("sqlite3", filePath)
	if err != nil {
		log.Fatal(err)
	}

	return db
}

// https://stackoverflow.com/questions/40266633/golang-insert-null-into-sql-instead-of-empty-string
func NewNullString(s string) sql.NullString {
	if len(s) == 0 {
		return sql.NullString{}
	}
	return sql.NullString{
		String: s,
		Valid:  true,
	}
}

func DecodeString(value string) string {
	bytes, err := hex.DecodeString(value)
	if err != nil {
		log.Fatal(err)
	}

	return string(bytes)
}

func DecodeAddress(value string) (string, error) {
	p := new(crypto.Point)
	key, err := hex.DecodeString(value)
	if err != nil {
		return "", err
	}

	err = p.DecodeCompressed(key)
	if err != nil {
		return "", err
	}

	return rpc.NewAddressFromKeys(p).String(), nil
}
