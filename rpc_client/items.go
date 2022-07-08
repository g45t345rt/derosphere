package rpc_client

import (
	"fmt"
	"log"
	"strconv"

	"encoding/hex"

	"github.com/deroproject/derohe/rpc"
)

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

func (d *Daemon) GetSCKeyValues(scid string, prefixKey string, start uint64, end uint64, columns []string) map[string]string {
	keys := []string{}
	for i := start; i < end; i++ {
		if len(columns) == 0 {
			keys = append(keys, fmt.Sprintf("%s%d", prefixKey, i))
		} else {
			for _, column := range columns {
				keys = append(keys, fmt.Sprintf("%s%d%s", prefixKey, i, column))
			}
		}
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

	keyValues := make(map[string]string)
	for index, value := range result.ValuesString {
		key := keys[index]
		valuestring, err := hex.DecodeString(value)
		if err != nil {
			log.Fatal(err)
		}

		keyValues[key] = string(valuestring)
	}

	return keyValues
}
