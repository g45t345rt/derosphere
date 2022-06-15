package rpc_client

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"encoding/hex"

	"github.com/deroproject/derohe/rpc"
)

type Commit struct {
	Action string
	Key    string
	Value  string
}

func (d *Daemon) GetSCCommitCount(scid string) uint64 {
	result, err := d.GetSC(&rpc.GetSC_Params{
		SCID:       scid,
		Variables:  false,
		Code:       false,
		KeysString: []string{"commit_count"},
	})

	if err != nil {
		log.Fatal(err)
	}

	commitCount, err := strconv.ParseUint(result.ValuesString[0], 10, 64)
	if err != nil {
		log.Fatal(err)
	}

	return commitCount
}

func (d *Daemon) GetSCCommits(scid string, start uint64, end uint64) []Commit {
	commitKeys := []string{}
	for i := start; i < end; i++ {
		commitKeys = append(commitKeys, fmt.Sprintf("commit_%d", i))
	}

	result, err := d.GetSC(&rpc.GetSC_Params{
		SCID:       scid,
		Variables:  false,
		Code:       false,
		KeysString: commitKeys,
	})

	if err != nil {
		log.Fatal(err)
	}

	commits := []Commit{}
	for _, value := range result.ValuesString {
		valuestring, err := hex.DecodeString(value)
		if err != nil {
			log.Fatal(err)
		}

		values := strings.Split(string(valuestring), "::")
		commits = append(commits, Commit{
			Action: values[0],
			Key:    values[1],
			Value:  values[2],
		})
	}

	return commits
}
