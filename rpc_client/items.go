package rpc_client

import (
	"fmt"
	"log"
	"strconv"

	"encoding/hex"

	"github.com/deroproject/derohe/rpc"
)

type Item struct {
	Key   string
	Value string
}

func (d *Daemon) GetSCItemCount(scid string, key string) uint64 {
	result, err := d.GetSC(&rpc.GetSC_Params{
		SCID:       scid,
		Variables:  false,
		Code:       false,
		KeysString: []string{key},
	})

	if err != nil {
		log.Fatal(err)
	}

	count, err := strconv.ParseUint(result.ValuesString[0], 10, 64)
	if err != nil {
		log.Fatal(err)
	}

	return count
}

func (d *Daemon) GetSCItems(scid string, prefixKey string, start uint64, end uint64) []Item {
	keys := []string{}
	for i := start; i < end; i++ {
		keys = append(keys, fmt.Sprintf("%s%d", prefixKey, i))
	}

	result, err := d.GetSC(&rpc.GetSC_Params{
		SCID:       scid,
		Variables:  false,
		Code:       false,
		KeysString: keys,
	})

	if err != nil {
		log.Fatal(err)
	}

	items := []Item{}
	for index, value := range result.ValuesString {
		key := keys[index]
		valuestring, err := hex.DecodeString(value)
		if err != nil {
			log.Fatal(err)
		}

		items = append(items, Item{
			Key:   key,
			Value: string(valuestring),
		})
	}

	return items
}
