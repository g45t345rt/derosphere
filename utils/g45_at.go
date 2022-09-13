package utils

import (
	_ "embed"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/deroproject/derohe/cryptography/crypto"
	"github.com/deroproject/derohe/rpc"
	"github.com/g45t345rt/derosphere/rpc_client"
)

//go:embed g45_at_public.bas
var G45_AT_PUBLIC_CODE string

//go:embed g45_at_private.bas
var G45_AT_PRIVATE_CODE string

//go:embed g45_c.bas
var G45_C_CODE string

//go:embed g45_nft_public.bas
var G45_NFT_PUBLIC_CODE string

//go:embed g45_nft_private.bas
var G45_NFT_PRIVATE_CODE string

type G45_C struct {
	SCID             string
	FrozenCollection bool
	FrozenMetadata   bool
	Owner            string
	OriginalOwner    string
	AssetCount       uint64
	MetadataFormat   string
	Metadata         string
	Assets           map[string]uint64
	Timestamp        uint64
}

func (a *G45_C) Print() {
	fmt.Println("SCID: ", a.SCID)
	fmt.Println("Frozen Collection: ", a.FrozenCollection)
	fmt.Println("Frozen Metadata: ", a.FrozenMetadata)
	fmt.Println("Metadata Format: ", a.MetadataFormat)
	fmt.Println("Metadata: ", a.Metadata)
	fmt.Println("Owner: ", a.Owner)
	fmt.Println("Original Owner: ", a.OriginalOwner)
	fmt.Println("Timestamp: ", a.Timestamp)
}

func (a *G45_C) JsonMetadata() (map[string]interface{}, error) {
	var metadata map[string]interface{}
	if a.MetadataFormat == "json" {
		err := json.Unmarshal([]byte(a.Metadata), &metadata)
		if err != nil {
			return nil, err
		}
	} else {
		return nil, errors.New("metadata format is not JSON")
	}
	return metadata, nil
}

type G45_AT struct {
	SCID             string
	Private          bool
	Minter           string
	OriginalMinter   string
	FrozenMetadata   bool
	FrozenMint       bool
	FrozenCollection bool
	MetadataFormat   string
	Metadata         string
	TotalSupply      uint64
	Decimals         uint64
	Collection       string
	Owners           map[string]uint64
	Timestamp        uint64
}

func (a *G45_AT) Print() {
	fmt.Println("SCID: ", a.SCID)
	fmt.Println("Private: ", a.Private)
	fmt.Println("Minter: ", a.Minter)
	fmt.Println("Original Minter: ", a.OriginalMinter)
	fmt.Println("Timestamp: ", a.Timestamp)
	fmt.Println("Collection SCID: ", a.Collection)
	fmt.Println("Frozen Metadata: ", a.FrozenMetadata)
	fmt.Println("Frozen Mint: ", a.FrozenMint)
	fmt.Println("Frozen Collection: ", a.FrozenCollection)
	fmt.Println("Metadata Format: ", a.MetadataFormat)
	fmt.Println("Metadata: ", a.Metadata)
	fmt.Println("Total Supply: ", a.TotalSupply)
	fmt.Println("Decimals: ", a.Decimals)
}

func (a *G45_AT) JsonMetadata() (map[string]interface{}, error) {
	var metadata map[string]interface{}
	if a.MetadataFormat == "json" {
		err := json.Unmarshal([]byte(a.Metadata), &metadata)
		if err != nil {
			return nil, err
		}
	} else {
		return nil, errors.New("metadata format is not JSON")
	}
	return metadata, nil
}

type G45_NFT struct {
	SCID           string
	Private        bool
	Minter         string
	MetadataFormat string
	Metadata       string
	Collection     string
	Owner          string
	Timestamp      uint64
}

func (a *G45_NFT) Print() {
	fmt.Println("SCID: ", a.SCID)
	fmt.Println("Private: ", a.Private)
	fmt.Println("Minter: ", a.Minter)
	fmt.Println("Timestamp: ", a.Timestamp)
	fmt.Println("Collection SCID: ", a.Collection)
	fmt.Println("Metadata Format: ", a.MetadataFormat)
	fmt.Println("Metadata: ", a.Metadata)
	fmt.Println("Owner: ", a.Owner)
}

func (a *G45_NFT) JsonMetadata() (map[string]interface{}, error) {
	var metadata map[string]interface{}
	if a.MetadataFormat == "json" {
		err := json.Unmarshal([]byte(a.Metadata), &metadata)
		if err != nil {
			return nil, err
		}
	} else {
		return nil, errors.New("metadata format is not JSON")
	}
	return metadata, nil
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

func GetG45_C(scid string, daemon *rpc_client.Daemon) (*G45_C, error) {
	result, err := daemon.GetSC(&rpc.GetSC_Params{
		SCID:      scid,
		Code:      true,
		Variables: true,
	})

	if err != nil {
		return nil, err
	}

	collection := &G45_C{}

	values := result.VariableStringKeys
	code := strings.ReplaceAll(strings.ReplaceAll(result.Code, "\r", ""), "\n", "")
	g45_atc_code := strings.ReplaceAll(strings.ReplaceAll(G45_C_CODE, "\r", ""), "\n", "")
	if code != g45_atc_code {
		return nil, fmt.Errorf("not a valid G45-ATC")
	}

	collection.SCID = scid
	collection.FrozenCollection = values["frozenCollection"].(float64) != 0
	collection.FrozenMetadata = values["frozenMetadata"].(float64) != 0
	collection.AssetCount = uint64(values["assetCount"].(float64))
	collection.MetadataFormat = decodeString(values["metadataFormat"].(string))
	collection.Metadata = decodeString(values["metadata"].(string))
	collection.Timestamp = uint64(values["timestamp"].(float64))

	owner, err := decodeAddress(values["owner"].(string))
	if err != nil {
		return nil, err
	}

	originalOwner, err := decodeAddress(values["originalOwner"].(string))
	if err != nil {
		return nil, err
	}

	collection.Owner = owner
	collection.OriginalOwner = originalOwner

	assetKey, _ := regexp.Compile(`asset_(.+)`)
	collection.Assets = make(map[string]uint64)
	for key, value := range result.VariableStringKeys {
		if assetKey.Match([]byte(key)) {
			assetSCID := assetKey.ReplaceAllString(key, "$1")
			collection.Assets[assetSCID] = uint64(value.(float64))
		}
	}

	return collection, nil
}

func GetG45_AT(scid string, daemon *rpc_client.Daemon) (*G45_AT, error) {
	result, err := daemon.GetSC(&rpc.GetSC_Params{
		SCID:      scid,
		Code:      true,
		Variables: true,
	})

	if err != nil {
		return nil, err
	}

	values := result.VariableStringKeys
	asset := &G45_AT{}

	code := strings.ReplaceAll(strings.ReplaceAll(result.Code, "\r", ""), "\n", "")
	g45_at_public_code := strings.ReplaceAll(strings.ReplaceAll(G45_AT_PUBLIC_CODE, "\r", ""), "\n", "")
	g45_at_private_code := strings.ReplaceAll(strings.ReplaceAll(G45_AT_PRIVATE_CODE, "\r", ""), "\n", "")

	switch code {
	case g45_at_public_code:
		asset.Private = false
	case g45_at_private_code:
		asset.Private = true
	default:
		return nil, fmt.Errorf("not a valid G45-AT")
	}

	asset.SCID = scid
	asset.Timestamp = uint64(values["timestamp"].(float64))
	asset.Collection = decodeString(values["collection"].(string))
	asset.FrozenMetadata = values["frozenMetadata"].(float64) != 0
	asset.FrozenMint = values["frozenMint"].(float64) != 0
	asset.FrozenCollection = values["frozenCollection"].(float64) != 0
	asset.MetadataFormat = decodeString(values["metadataFormat"].(string))
	asset.Metadata = decodeString(values["metadata"].(string))
	asset.TotalSupply = uint64(values["totalSupply"].(float64))
	asset.Decimals = uint64(values["decimals"].(float64))

	minter, err := decodeAddress(values["minter"].(string))
	if err != nil {
		return nil, err
	}

	asset.Minter = minter

	originalMinter, err := decodeAddress(values["originalMinter"].(string))
	if err != nil {
		return nil, err
	}

	asset.OriginalMinter = originalMinter

	ownerKey, _ := regexp.Compile(`owner_(.+)`)
	asset.Owners = make(map[string]uint64)
	for key, value := range result.VariableStringKeys {
		if ownerKey.Match([]byte(key)) {
			owner := ownerKey.ReplaceAllString(key, "$1")
			asset.Owners[owner] = uint64(value.(float64))
		}
	}

	return asset, nil
}

func GetG45_NFT(scid string, daemon *rpc_client.Daemon) (*G45_NFT, error) {
	result, err := daemon.GetSC(&rpc.GetSC_Params{
		SCID:      scid,
		Code:      true,
		Variables: true,
	})

	if err != nil {
		return nil, err
	}

	values := result.VariableStringKeys
	asset := &G45_NFT{}

	code := strings.ReplaceAll(strings.ReplaceAll(result.Code, "\r", ""), "\n", "")
	g45_nft_public_code := strings.ReplaceAll(strings.ReplaceAll(G45_NFT_PUBLIC_CODE, "\r", ""), "\n", "")
	g45_nft_private_code := strings.ReplaceAll(strings.ReplaceAll(G45_NFT_PRIVATE_CODE, "\r", ""), "\n", "")

	switch code {
	case g45_nft_public_code:
		asset.Private = false
	case g45_nft_private_code:
		asset.Private = true
	default:
		return nil, fmt.Errorf("not a valid G45-NFT")
	}

	asset.SCID = scid
	asset.Timestamp = uint64(values["timestamp"].(float64))
	asset.Collection = decodeString(values["collection"].(string))
	asset.MetadataFormat = decodeString(values["metadataFormat"].(string))
	asset.Metadata = decodeString(values["metadata"].(string))
	asset.Owner = decodeString(values["owner"].(string))

	minter, err := decodeAddress(values["minter"].(string))
	if err != nil {
		return nil, err
	}

	asset.Minter = minter

	return asset, nil
}
