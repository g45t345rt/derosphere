package utils

import (
	"encoding/hex"
	"fmt"
	"log"

	"github.com/deroproject/derohe/cryptography/crypto"
	"github.com/deroproject/derohe/rpc"
)

var G45_NFT_PUBLIC = `
Function Initialize() Uint64
10 STORE("minter", SIGNER())
20 STORE("type", "G45-NFT")
30 STORE("init", 0)
40 RETURN 0
End Function

Function InitStore(collection String, supply Uint64, metadata String, frozenMetadata Uint64, frozenSupply Uint64) Uint64
10 IF LOAD("minter") == SIGNER() THEN GOTO 30
20 RETURN 1
30 IF LOAD("init") == 0 THEN GOTO 50
40 RETURN 1
50 IF supply > 0 THEN GOTO 70
60 RETURN 1
70 IF frozenMetadata <= 1  THEN GOTO 90
80 RETURN 1
90 IF frozenSupply <= 1  THEN GOTO 110
100 RETURN 1
110 SEND_ASSET_TO_ADDRESS(LOAD("minter"), supply, SCID())
120 STORE("collection", collection)
130 STORE("metadata", metadata)
140 STORE("supply", supply)
150 STORE("frozenMetadata", frozenMetadata)
160 STORE("frozenSupply", frozenSupply)
170 STORE("init", 1)
180 RETURN 0
End Function

Function SetMetadata(metadata String) Uint64
10 IF LOAD("minter") == SIGNER() THEN GOTO 30
20 RETURN 1
30 IF LOAD("frozenMetadata") == 0 THEN GOTO 50
40 RETURN 1
50 STORE("metadata", metadata)
60 RETURN 0
End Function

Function AddSupply(supply Uint64) Uint64
10 IF LOAD("minter") == SIGNER() THEN GOTO 30
20 RETURN 1
30 IF LOAD("frozenSupply") == 0 THEN GOTO 50
40 RETURN 1
50 STORE("supply", LOAD("supply") + supply)
60 SEND_ASSET_TO_ADDRESS(LOAD("minter"), supply, SCID())
70 RETURN 0
End Function

Function FreezeMetadata() Uint64
10 IF LOAD("minter") == SIGNER() THEN GOTO 30
20 RETURN 1
30 STORE("frozenMetadata", 1)
40 RETURN 0
End Function

Function FreezeSupply() Uint64
10 IF LOAD("minter") == SIGNER() THEN GOTO 30
20 RETURN 1
30 STORE("frozenSupply", 1)
40 RETURN 0
End Function
`

var G45_NFT_PRIVATE = `
Function InitializePrivate() Uint64
10 STORE("minter", SIGNER())
20 STORE("type", "G45-NFT")
30 STORE("init", 0)
40 RETURN 0
End Function

Function InitStore(collection String, supply Uint64, metadata String, frozenMetadata Uint64, frozenSupply Uint64) Uint64
10 IF LOAD("minter") == SIGNER() THEN GOTO 30
20 RETURN 1
30 IF LOAD("init") == 0 THEN GOTO 50
40 RETURN 1
50 IF supply > 0 THEN GOTO 70
60 RETURN 1
70 IF frozenMetadata <= 1  THEN GOTO 90
80 RETURN 1
90 IF frozenSupply <= 1  THEN GOTO 110
100 RETURN 1
110 SEND_ASSET_TO_ADDRESS(LOAD("minter"), supply, SCID())
120 STORE("collection", collection)
130 STORE("metadata", metadata)
140 STORE("supply", supply)
150 STORE("frozenMetadata", frozenMetadata)
160 STORE("frozenSupply", frozenSupply)
170 STORE("init", 1)
180 RETURN 0
End Function

Function SetMetadata(metadata String) Uint64
10 IF LOAD("minter") == SIGNER() THEN GOTO 30
20 RETURN 1
30 IF LOAD("frozenMetadata") == 0 THEN GOTO 50
40 RETURN 1
50 STORE("metadata", metadata)
60 RETURN 0
End Function

Function AddSupply(supply Uint64) Uint64
10 IF LOAD("minter") == SIGNER() THEN GOTO 30
20 RETURN 1
30 IF LOAD("frozenSupply") == 0 THEN GOTO 50
40 RETURN 1
50 STORE("supply", LOAD("supply") + supply)
60 SEND_ASSET_TO_ADDRESS(LOAD("minter"), supply, SCID())
70 RETURN 0
End Function

Function FreezeMetadata() Uint64
10 IF LOAD("minter") == SIGNER() THEN GOTO 30
20 RETURN 1
30 STORE("frozenMetadata", 1)
40 RETURN 0
End Function

Function FreezeSupply() Uint64
10 IF LOAD("minter") == SIGNER() THEN GOTO 30
20 RETURN 1
30 STORE("frozenSupply", 1)
40 RETURN 0
End Function
`

var G45_NFT_COLLECTION = `
Function Initialize() Uint64
10 STORE("owner", SIGNER())
20 STORE("type", "G45-NFT-COLLECTION")
30 STORE("nft_count", 0)
40 RETURN 0
End Function

Function Add(nft String) Uint64
10 DIM ctr as Uint64
20 IF LOAD("owner") == SIGNER() THEN GOTO 40
30 RETURN 1
40 IF EXISTS("nft_" + nft) == 0 THEN GOTO 60
50 RETURN 1
60 LET ctr = LOAD("nft_count")
70 STORE("nft_" + ctr, nft)
80 STORE("nft_" + nft, "")
90 STORE("nft_count", ctr + 1)
100 RETURN 0
End Function
`

type G45NFT struct {
	Id             string
	Init           bool
	Minter         string
	FrozenMetadata bool
	FrozenSupply   bool
	Metadata       string
	Supply         uint64
	Collection     string
}

func (nft *G45NFT) Print() {
	fmt.Println("Asset ID: ", nft.Id)
	fmt.Println("Collection ID: ", nft.Collection)
	fmt.Println("Init: ", nft.Init)
	fmt.Println("Minter: ", nft.Minter)
	fmt.Println("Frozen Metadata: ", nft.FrozenMetadata)
	fmt.Println("Frozen Supply: ", nft.FrozenSupply)
	fmt.Println("Metadata: ", nft.Metadata)
	fmt.Println("Supply: ", nft.Supply)
}

func decodeString(value string) string {
	bytes, err := hex.DecodeString(value)
	if err != nil {
		log.Fatal(err)
	}

	return string(bytes)
}

func ParseG45NFT(id string, result *rpc.GetSC_Result) (*G45NFT, error) {
	values := result.VariableStringKeys
	nft := &G45NFT{}

	switch result.Code {
	case G45_NFT_PUBLIC:
	case G45_NFT_PRIVATE:
	case G45_NFT_COLLECTION:
	default:
		return nil, fmt.Errorf("not a valid G45-NFT")
	}

	nft.Id = id
	nft.Init = values["init"].(float64) != 0
	nft.Collection = decodeString(values["collection"].(string))
	nft.FrozenMetadata = values["frozenMetadata"].(float64) != 0
	nft.FrozenSupply = values["frozenSupply"].(float64) != 0
	nft.Metadata = decodeString(values["metadata"].(string))

	p := new(crypto.Point)
	key, err := hex.DecodeString(values["minter"].(string))
	if err != nil {
		return nil, err
	}

	err = p.DecodeCompressed(key)
	if err != nil {
		return nil, err
	}

	nft.Minter = rpc.NewAddressFromKeys(p).String()
	nft.Supply = uint64(values["supply"].(float64))

	return nft, nil
}
