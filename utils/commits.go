package utils

import (
	"fmt"
	"strconv"

	"github.com/g45t345rt/derosphere/rpc_client"
	"github.com/tidwall/buntdb"
)

func GetCommitAt(db *buntdb.DB) (uint64, error) {
	var start uint64 = 0
	err := db.View(func(tx *buntdb.Tx) error {
		commitAt, err := tx.Get("commit_at")
		if err != nil {
			return err
		}

		start, err = strconv.ParseUint(commitAt, 10, 64)
		if err != nil {
			return err
		}
		return nil
	})

	if err != nil && err != buntdb.ErrNotFound {
		return 0, err
	}

	return start, nil
}

func SyncCommits(db *buntdb.DB, daemon *rpc_client.Daemon, scid string) error {
	commitCount := daemon.GetSCCommitCount(scid)
	commitAt, err := GetCommitAt(db)
	if err != nil {
		return err
	}

	chunk := uint64(1000)

	var i uint64
	for i = commitAt; i < commitCount; i += chunk {
		var commits []rpc_client.Commit
		end := i + chunk
		if end > commitCount {
			commitAt = commitCount
			commits = daemon.GetSCCommits(scid, i, commitCount)
		} else {
			commitAt = end
			commits = daemon.GetSCCommits(scid, i, commitAt)
		}

		// fmt.Print(commits)

		err := db.Update(func(tx *buntdb.Tx) error {
			for _, commit := range commits {
				if commit.Action == "S" {
					tx.Set(commit.Key, commit.Value, nil)
				}

				if commit.Action == "D" {
					tx.Delete(commit.Key)
				}
			}

			tx.Set("commit_at", fmt.Sprint(commitAt), nil)
			return nil
		})

		if err != nil {
			return err
		}
	}

	return nil
}
