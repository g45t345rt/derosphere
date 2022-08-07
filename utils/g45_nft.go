package utils

import (
	_ "embed"
	"encoding/hex"
	"fmt"
	"log"
	"regexp"
	"strings"

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
	Token            string
	FrozenCollection bool
	FrozenMetadata   bool
	Owner            string
	OriginalOwner    string
	NFTCount         uint64
	Metadata         string
	NFTs             map[string]uint64
}

func (nft *G45NFTCollection) Print() {
	fmt.Println("Asset Token: ", nft.Token)
	fmt.Println("Frozen Collection: ", nft.FrozenCollection)
	fmt.Println("Frozen Metadata: ", nft.FrozenMetadata)
	fmt.Println("Metadata: ", nft.Metadata)
	fmt.Println("Owner: ", nft.Owner)
	fmt.Println("Original Owner: ", nft.OriginalOwner)
}

type G45NFT struct {
	Token            string
	Init             bool
	Private          bool
	Minter           string
	OriginalMinter   string
	FrozenMetadata   bool
	FrozenSupply     bool
	FrozenCollection bool
	Metadata         string
	Supply           uint64
	Collection       string
	Owners           map[string]uint64
}

func (nft *G45NFT) Print() {
	fmt.Println("Asset Token: ", nft.Token)
	fmt.Println("Init: ", nft.Init)
	fmt.Println("Private: ", nft.Private)
	fmt.Println("Minter: ", nft.Minter)
	fmt.Println("Original Minter: ", nft.OriginalMinter)
	if nft.Init {
		fmt.Println("Collection SCID: ", nft.Collection)
		fmt.Println("Frozen Metadata: ", nft.FrozenMetadata)
		fmt.Println("Frozen Supply: ", nft.FrozenSupply)
		fmt.Println("Frozen Collection: ", nft.FrozenCollection)
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
		SCID:      scid,
		Code:      true,
		Variables: true,
	})

	if err != nil {
		return nil, err
	}

	nftCollection := &G45NFTCollection{}

	values := result.VariableStringKeys
	code := strings.ReplaceAll(strings.ReplaceAll(result.Code, "\r", ""), "\n", "")
	g45_nft_collection := strings.ReplaceAll(strings.ReplaceAll(G45_NFT_COLLECTION, "\r", ""), "\n", "")
	if code != g45_nft_collection {
		return nil, fmt.Errorf("not a valid G45-NFT-Collection")
	}

	nftCollection.Token = scid
	nftCollection.FrozenCollection = values["frozenCollection"].(float64) != 0
	nftCollection.FrozenMetadata = values["frozenMetadata"].(float64) != 0
	nftCollection.NFTCount = uint64(values["nftCount"].(float64))
	nftCollection.Metadata = decodeString(values["metadata"].(string))

	owner, err := decodeAddress(values["owner"].(string))
	if err != nil {
		return nil, err
	}

	originalOwner, err := decodeAddress(values["originalOwner"].(string))
	if err != nil {
		return nil, err
	}

	nftCollection.Owner = owner
	nftCollection.OriginalOwner = originalOwner

	nftKey, _ := regexp.Compile(`nft_(.+)`)
	nftCollection.NFTs = make(map[string]uint64)
	for key, value := range result.VariableStringKeys {
		if nftKey.Match([]byte(key)) {
			nftId := nftKey.ReplaceAllString(key, "$1")
			nftCollection.NFTs[nftId] = uint64(value.(float64))
		}
	}

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

	code := strings.ReplaceAll(strings.ReplaceAll(result.Code, "\r", ""), "\n", "")
	g45_nft_public := strings.ReplaceAll(strings.ReplaceAll(G45_NFT_PUBLIC, "\r", ""), "\n", "")
	g45_nft_private := strings.ReplaceAll(strings.ReplaceAll(G45_NFT_PRIVATE, "\r", ""), "\n", "")

	switch code {
	case g45_nft_public:
		nft.Private = false
	case g45_nft_private:
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
		nft.FrozenCollection = values["frozenCollection"].(float64) != 0
		nft.Metadata = decodeString(values["metadata"].(string))
		nft.Supply = uint64(values["supply"].(float64))
	}

	minter, err := decodeAddress(values["minter"].(string))
	if err != nil {
		return nil, err
	}

	nft.Minter = minter

	originalMinter, err := decodeAddress(values["originalMinter"].(string))
	if err != nil {
		return nil, err
	}

	nft.OriginalMinter = originalMinter

	ownerKey, _ := regexp.Compile(`owner_(.+)`)
	nft.Owners = make(map[string]uint64)
	for key, value := range result.VariableStringKeys {
		if ownerKey.Match([]byte(key)) {
			owner := ownerKey.ReplaceAllString(key, "$1")
			nft.Owners[owner] = uint64(value.(float64))
		}
	}

	return nft, nil
}
