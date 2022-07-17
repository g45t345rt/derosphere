package utils

import (
	"encoding/hex"
	"fmt"
	"log"

	"github.com/deroproject/derohe/cryptography/crypto"
	"github.com/deroproject/derohe/rpc"
)

var G45_NFT_PUBLIC = `
Function storeTX()
10 STORE("txid_" + HEX(TXID()), 1)
20 RETURN
End Function

Function Initialize() Uint64
10 IF EXISTS("owner") == 0 THEN GOTO 30
20 RETURN 1
30 STORE("minter", SIGNER())
40 STORE("type", "G45-NFT")
50 STORE("init", 0)
60 RETURN 0
End Function

Function InitStore(collection String, supply Uint64, metadata String, freezeMetadata Uint64, freezeSupply Uint64) Uint64
10 IF LOAD("minter") == SIGNER() THEN GOTO 30
20 RETURN 1
30 IF LOAD("init") == 0 THEN GOTO 50
40 RETURN 1
50 IF supply > 0 THEN GOTO 70
60 RETURN 1
70 IF freezeMetadata <= 1  THEN GOTO 90
80 RETURN 1
90 IF freezeSupply <= 1  THEN GOTO 110
100 RETURN 1
110 SEND_ASSET_TO_ADDRESS(LOAD("minter"), supply, SCID())
120 STORE("collection", collection)
130 STORE("metadata", metadata)
140 STORE("supply", supply)
150 STORE("frozenMetadata", freezeMetadata)
160 STORE("frozenSupply", freezeSupply)
170 STORE("init", 1)
180 storeTX()
190 RETURN 0
End Function

Function SetMetadata(metadata String) Uint64
10 IF LOAD("minter") == SIGNER() THEN GOTO 30
20 RETURN 1
30 IF LOAD("frozenMetadata") == 0 THEN GOTO 50
40 RETURN 1
50 STORE("metadata", metadata)
60 storeTX()
70 RETURN 0
End Function

Function AddSupply(supply Uint64) Uint64
10 IF LOAD("minter") == SIGNER() THEN GOTO 30
20 RETURN 1
30 IF LOAD("frozenSupply") == 0 THEN GOTO 50
40 RETURN 1
50 STORE("supply", LOAD("supply") + supply)
60 SEND_ASSET_TO_ADDRESS(LOAD("minter"), supply, SCID())
70 storeTX()
80 RETURN 0
End Function

Function FreezeMetadata() Uint64
10 IF LOAD("minter") == SIGNER() THEN GOTO 30
20 RETURN 1
30 STORE("frozenMetadata", 1)
40 storeTX()
50 RETURN 0
End Function

Function FreezeSupply() Uint64
10 IF LOAD("minter") == SIGNER() THEN GOTO 30
20 RETURN 1
30 STORE("frozenSupply", 1)
40 storeTX()
50 RETURN 0
End Function
`

var G45_NFT_PRIVATE = `
Function storeTX()
10 STORE("txid_" + HEX(TXID()), 1)
20 RETURN
End Function

Function InitializePrivate() Uint64
10 IF EXISTS("owner") == 0 THEN GOTO 30
20 RETURN 1
30 STORE("minter", SIGNER())
40 STORE("type", "G45-NFT")
50 STORE("init", 0)
60 RETURN 0
End Function

Function InitStore(collection String, supply Uint64, metadata String, freezeMetadata Uint64, freezeSupply Uint64) Uint64
10 IF LOAD("minter") == SIGNER() THEN GOTO 30
20 RETURN 1
30 IF LOAD("init") == 0 THEN GOTO 50
40 RETURN 1
50 IF supply > 0 THEN GOTO 70
60 RETURN 1
70 IF freezeMetadata <= 1  THEN GOTO 90
80 RETURN 1
90 IF freezeSupply <= 1  THEN GOTO 110
100 RETURN 1
110 SEND_ASSET_TO_ADDRESS(LOAD("minter"), supply, SCID())
120 STORE("collection", collection)
130 STORE("metadata", metadata)
140 STORE("supply", supply)
150 STORE("frozenMetadata", freezeMetadata)
160 STORE("frozenSupply", freezeSupply)
170 STORE("init", 1)
180 storeTX()
190 RETURN 0
End Function

Function SetMetadata(metadata String) Uint64
10 IF LOAD("minter") == SIGNER() THEN GOTO 30
20 RETURN 1
30 IF LOAD("frozenMetadata") == 0 THEN GOTO 50
40 RETURN 1
50 STORE("metadata", metadata)
60 storeTX()
70 RETURN 0
End Function

Function AddSupply(supply Uint64) Uint64
10 IF LOAD("minter") == SIGNER() THEN GOTO 30
20 RETURN 1
30 IF LOAD("frozenSupply") == 0 THEN GOTO 50
40 RETURN 1
50 STORE("supply", LOAD("supply") + supply)
60 SEND_ASSET_TO_ADDRESS(LOAD("minter"), supply, SCID())
70 storeTX()
80 RETURN 0
End Function

Function FreezeMetadata() Uint64
10 IF LOAD("minter") == SIGNER() THEN GOTO 30
20 RETURN 1
30 STORE("frozenMetadata", 1)
40 storeTX()
50 RETURN 0
End Function

Function FreezeSupply() Uint64
10 IF LOAD("minter") == SIGNER() THEN GOTO 30
20 RETURN 1
30 STORE("frozenSupply", 1)
40 storeTX()
50 RETURN 0
End Function
`

var G45_NFT_COLLECTION = `
Function storeCommitString(action String, key String, value String)
10 DIM commit_count as Uint64
20 LET commit_count = MAPGET("commit_count")
30 STORE("commit_" + commit_count, action + "::" + key + "::" + value)
40 MAPSTORE("commit_count", commit_count + 1)
50 RETURN
End Function

Function storeCommitInt(action String, key String, value Uint64)
10 DIM commit_count as Uint64
20 LET commit_count = MAPGET("commit_count")
30 STORE("commit_" + commit_count, action + "::" + key + "::" + value)
40 MAPSTORE("commit_count", commit_count + 1)
50 RETURN
End Function

Function initCommit()
10 STORE("commit_count", 0)
20 RETURN
End Function

Function beginCommit()
10 MAPSTORE("commit_count", LOAD("commit_count"))
20 RETURN
End Function

Function endCommit()
10 STORE("commit_count", MAPGET("commit_count"))
20 RETURN
End Function

Function storeStateString(key String, value String)
10 STORE("state_" + key, value)
20 storeCommitString("S", "state_" + key, value)
30 RETURN
End Function

Function storeStateInt(key String, value Uint64)
10 STORE("state_" + key, value)
20 storeCommitInt("S", "state_" + key, value)
30 RETURN
End Function

Function deleteState(key String)
10 DELETE("state_" + key)
20 storeCommitInt("D", "state_" + key, 0)
30 RETURN
End Function

Function loadStateString(key String) String
10 RETURN LOAD("state_" + key)
End Function

Function loadStateInt(key String) Uint64
10 RETURN LOAD("state_" + key)
End Function

Function stateExists(key String) Uint64
10 RETURN EXISTS("state_" + key)
End Function

Function storeTX()
10 STORE("txid_" + HEX(TXID()), 1)
20 RETURN
End Function

Function Initialize() Uint64
10 IF EXISTS("owner") == 0 THEN GOTO 30
20 RETURN 1
30 STORE("owner", SIGNER())
40 STORE("type", "G45-NFT-COLLECTION")
50 STORE("lock", 0)
60 initCommit()
70 RETURN 0
End Function

Function Lock() Uint64
10 IF LOAD("owner") == SIGNER() THEN GOTO 30
20 RETURN 1
30 STORE("lock", 1)
40 storeTX()
50 RETURN 0
End Function

Function Set(nft String, index Uint64) Uint64
10 IF LOAD("owner") == SIGNER() THEN GOTO 30
20 RETURN 1
30 IF LOAD("lock") == 0 THEN GOTO 50
40 RETURN 1
50 beginCommit()
60 storeStateInt("nft_" + nft, index)
70 endCommit()
80 storeTX()
90 RETURN 0
End Function

Function Del(nft String) Uint64
10 IF LOAD("owner") == SIGNER() THEN GOTO 30
20 RETURN 1
30 IF LOAD("lock") == 0 THEN GOTO 50
40 RETURN 1
50 IF stateExists("nft_" + nft) == 1 THEN GOTO 70
60 RETURN 1
70 beginCommit()
80 deleteState("nft_" + nft)
90 endCommit()
100 storeTX()
110 RETURN 0
End Function
`

type G45NFT struct {
	Token          string
	Init           bool
	Private        bool
	Minter         string
	FrozenMetadata bool
	FrozenSupply   bool
	Metadata       string
	Supply         uint64
	Collection     string
}

func (nft *G45NFT) Print() {
	fmt.Println("Asset Token: ", nft.Token)
	fmt.Println("Collection Token: ", nft.Collection)
	fmt.Println("Init: ", nft.Init)
	fmt.Println("Private: ", nft.Private)
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

func ParseG45NFT(token string, result *rpc.GetSC_Result) (*G45NFT, error) {
	values := result.VariableStringKeys
	nft := &G45NFT{}

	switch result.Code {
	case G45_NFT_PUBLIC:
		nft.Private = false
	case G45_NFT_PRIVATE:
		nft.Private = true
	default:
		return nil, fmt.Errorf("not a valid G45-NFT")
	}

	nft.Token = token
	nft.Init = values["init"].(float64) != 0
	if nft.Init {
		nft.Collection = decodeString(values["collection"].(string))
		nft.FrozenMetadata = values["frozenMetadata"].(float64) != 0
		nft.FrozenSupply = values["frozenSupply"].(float64) != 0
		nft.Metadata = decodeString(values["metadata"].(string))
		nft.Supply = uint64(values["supply"].(float64))
	}

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

	return nft, nil
}
