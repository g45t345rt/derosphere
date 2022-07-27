package utils

import (
	_ "embed"
	"encoding/hex"
	"fmt"
	"log"

	"github.com/deroproject/derohe/cryptography/crypto"
	"github.com/deroproject/derohe/rpc"
)

//go:embed g45_nft_public.bas
var G45_NFT_PUBLIC string

//go:embed g45_nft_private.bas
var G45_NFT_PRIVATE string

//go:embed g45_nft_collection.bas
var G45_NFT_COLLECTION string

type G45NFTCollection struct {
	Token    string
	Frozen   bool
	Owner    string
	NFTCount uint64
}

func (nft *G45NFTCollection) Print() {
	fmt.Println("Asset Token: ", nft.Token)
	fmt.Println("Frozen: ", nft.Frozen)
	fmt.Println("Owner: ", nft.Owner)
}

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

func ParseG45NFTCollection(token string, result *rpc.GetSC_Result) (*G45NFTCollection, error) {
	values := result.VariableStringKeys
	nftCollection := &G45NFTCollection{}

	if result.Code != G45_NFT_COLLECTION {
		return nil, fmt.Errorf("not a valid G45-NFT-Collection")
	}

	nftCollection.Token = token
	nftCollection.Frozen = values["frozen"].(float64) != 0
	nftCollection.NFTCount = values["nftCount"].(uint64)

	p := new(crypto.Point)
	key, err := hex.DecodeString(values["owner"].(string))
	if err != nil {
		return nil, err
	}

	err = p.DecodeCompressed(key)
	if err != nil {
		return nil, err
	}

	nftCollection.Owner = rpc.NewAddressFromKeys(p).String()
	return nftCollection, nil
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
