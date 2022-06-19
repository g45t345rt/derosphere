package utils

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"os"

	"github.com/g45t345rt/derosphere/config"
)

func GetCommitCounts() map[string]uint64 {
	content, err := ioutil.ReadFile(config.DATA_FOLDER + "/commit_counts.json")
	var counts = make(map[string]uint64)

	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return counts
		}

		log.Fatal(err)
	}

	err = json.Unmarshal(content, &counts)
	if err != nil {
		log.Fatal(err)
	}

	return counts
}

func SetCommitCount(name string, count uint64) {
	counts := GetCommitCounts()
	counts[name] = count
	countsString, err := json.Marshal(counts)
	if err != nil {
		log.Fatal(err)
	}

	err = ioutil.WriteFile(config.DATA_FOLDER+"/commit_counts.json", countsString, os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}
}
