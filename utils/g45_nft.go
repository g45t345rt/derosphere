package utils

import (
	_ "embed"
	"encoding/hex"
	"fmt"
	"log"
	"strconv"

	"github.com/deroproject/derohe/cryptography/crypto"
	"github.com/deroproject/derohe/rpc"
	"github.com/g45t345rt/derosphere/rpc_client"
)

//go:embed g45_nft_public.bas
var G45_NFT_PUBLIC string

//go:embed g45_nft_private.bas
var G45_NFT_PRIVATE string

//go:embed g45_nft_collection.bas
var G45_NFT_COLLECTION string

type G45NFTCollection struct {
	Token         string
	Frozen        bool
	Owner         string
	OriginalOwner string
	NFTCount      uint64
}

func (nft *G45NFTCollection) Print() {
	fmt.Println("Asset Token: ", nft.Token)
	fmt.Println("Frozen: ", nft.Frozen)
	fmt.Println("Owner: ", nft.Owner)
	fmt.Println("Original Owner: ", nft.OriginalOwner)
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
	fmt.Println("Init: ", nft.Init)
	fmt.Println("Private: ", nft.Private)
	fmt.Println("Minter: ", nft.Minter)
	if nft.Init {
		fmt.Println("Collection Token: ", nft.Collection)
		fmt.Println("Frozen Metadata: ", nft.FrozenMetadata)
		fmt.Println("Frozen Supply: ", nft.FrozenSupply)
		fmt.Println("Metadata: ", nft.Metadata)
		fmt.Println("Supply: ", nft.Supply)
	}
}

func decodeString(value string) string {
	bytes, err := hex.DecodeString(value)
	if err != nil {
		log.Fatal(err)
	}

	return string(bytes)
}

func decodeAddress(value string) (string, error) {
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

func GetG45NftCollection(scid string, daemon *rpc_client.Daemon) (*G45NFTCollection, error) {
	result, err := daemon.GetSC(&rpc.GetSC_Params{
		SCID:       scid,
		Code:       true,
		Variables:  false,
		KeysString: []string{"owner", "originalOwner"}, // {"frozen", "nftCount", "owner", "originalOwner"} can't do that frozen is missing don't know why so I use KeysBytes
		KeysBytes:  [][]byte{[]byte("frozen"), []byte("nftCount")},
	})

	if err != nil {
		return nil, err
	}

	nftCollection := &G45NFTCollection{}

	if result.Code != G45_NFT_COLLECTION {
		return nil, fmt.Errorf("not a valid G45-NFT-Collection")
	}

	nftCollection.Token = scid
	nftCollection.Frozen, _ = strconv.ParseBool(result.ValuesBytes[0])
	nftCollection.NFTCount, _ = strconv.ParseUint(result.ValuesBytes[1], 10, 64)

	owner, err := decodeAddress(result.ValuesString[0])
	if err != nil {
		return nil, err
	}

	originalOwner, err := decodeAddress(result.ValuesString[1])
	if err != nil {
		return nil, err
	}

	nftCollection.Owner = owner
	nftCollection.OriginalOwner = originalOwner

	return nftCollection, nil
}

func GetG45NFT(scid string, daemon *rpc_client.Daemon) (*G45NFT, error) {
	result, err := daemon.GetSC(&rpc.GetSC_Params{
		SCID:      scid,
		Code:      true,
		Variables: true,
	})

	if err != nil {
		return nil, err
	}

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

	nft.Token = scid
	nft.Init = values["init"].(float64) != 0
	if nft.Init {
		nft.Collection = decodeString(values["collection"].(string))
		nft.FrozenMetadata = values["frozenMetadata"].(float64) != 0
		nft.FrozenSupply = values["frozenSupply"].(float64) != 0
		nft.Metadata = decodeString(values["metadata"].(string))
		nft.Supply = uint64(values["supply"].(float64))
	}

	minter, err := decodeAddress(values["minter"].(string))
	if err != nil {
		return nil, err
	}

	nft.Minter = minter
	return nft, nil
}
