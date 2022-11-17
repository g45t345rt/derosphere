package utils

import (
	_ "embed"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/deroproject/derohe/rpc"
)

//go:embed g45_at_public.bas
var G45_AT_PUBLIC_CODE string

//go:embed g45_at_private.bas
var G45_AT_PRIVATE_CODE string

//go:embed g45_fat_public.bas
var G45_FAT_PUBLIC_CODE string

//go:embed g45_fat_private.bas
var G45_FAT_PRIVATE_CODE string

//go:embed g45_c.bas
var G45_C_CODE string

//go:embed g45_nft_public.bas
var G45_NFT_PUBLIC_CODE string

//go:embed g45_nft_private.bas
var G45_NFT_PRIVATE_CODE string

func formatMetadata(format string, value string) (map[string]interface{}, error) {
	var metadata map[string]interface{}
	if format == "json" {
		err := json.Unmarshal([]byte(value), &metadata)
		if err != nil {
			return metadata, err
		}
	} else {
		return metadata, errors.New("metadata format is not JSON")
	}
	return metadata, nil
}

func trimCode(code string) string {
	return strings.ReplaceAll(strings.ReplaceAll(code, "\r", ""), "\n", "")
}

/** G45-FAT **/

type G45_FAT struct {
	SCID             string
	Private          bool
	Minter           string
	FrozenMetadata   bool
	FrozenCollection bool
	MetadataFormat   string
	Metadata         string
	MaxSupply        uint64
	TotalSupply      uint64
	Decimals         uint64
	Collection       string
	Owners           map[string]uint64
	Timestamp        uint64
}

func (asset *G45_FAT) Print() {
	fmt.Println("SCID: ", asset.SCID)
	fmt.Println("Private: ", asset.Private)
	fmt.Println("Minter: ", asset.Minter)
	fmt.Println("Timestamp: ", asset.Timestamp)
	fmt.Println("Collection SCID: ", asset.Collection)
	fmt.Println("Frozen Metadata: ", asset.FrozenMetadata)
	fmt.Println("Frozen Collection: ", asset.FrozenCollection)
	fmt.Println("Metadata Format: ", asset.MetadataFormat)
	fmt.Println("Metadata: ", asset.Metadata)
	fmt.Println("Max Supply: ", asset.MaxSupply)
	fmt.Println("Total Supply: ", asset.TotalSupply)
	fmt.Println("Decimals: ", asset.Decimals)
}

func (asset *G45_FAT) JsonMetadata() (map[string]interface{}, error) {
	return formatMetadata(asset.MetadataFormat, asset.Metadata)
}

func (asset *G45_FAT) Validate(code string) (bool, error) {
	switch trimCode(code) {
	case trimCode(G45_FAT_PUBLIC_CODE):
		asset.Private = false
	case trimCode(G45_FAT_PRIVATE_CODE):
		asset.Private = true
	default:
		return false, fmt.Errorf("not a valid G45-FAT")
	}

	return true, nil
}

func (asset *G45_FAT) Parse(scId string, result *rpc.GetSC_Result) error {
	values := result.VariableStringKeys

	_, err := asset.Validate(result.Code)
	if err != nil {
		return err
	}

	asset.SCID = scId
	asset.Timestamp = uint64(values["timestamp"].(float64))
	collection, err := DecodeString(values["collection"].(string))
	if err != nil {
		return err
	}
	asset.Collection = collection

	asset.FrozenMetadata = values["frozenMetadata"].(float64) != 0
	asset.FrozenCollection = values["frozenCollection"].(float64) != 0

	metadataFormat, err := DecodeString(values["metadataFormat"].(string))
	if err != nil {
		return err
	}
	asset.MetadataFormat = metadataFormat

	metadata, err := DecodeString(values["metadata"].(string))
	if err != nil {
		return err
	}
	asset.Metadata = metadata

	asset.MaxSupply = uint64(values["maxSupply"].(float64))
	asset.TotalSupply = uint64(values["totalSupply"].(float64))
	asset.Decimals = uint64(values["decimals"].(float64))

	minter, err := DecodeAddress(values["minter"].(string))
	if err != nil {
		return err
	}

	asset.Minter = minter

	ownerKey, err := regexp.Compile(`owner_(.+)`)
	if err != nil {
		return err
	}

	asset.Owners = make(map[string]uint64)
	for key, value := range result.VariableStringKeys {
		if ownerKey.Match([]byte(key)) {
			owner := ownerKey.ReplaceAllString(key, "$1")
			asset.Owners[owner] = uint64(value.(float64))
		}
	}

	return nil
}

/** G45-C **/

type G45_C struct {
	SCID           string
	FrozenAssets   bool
	FrozenMetadata bool
	Owner          string
	OriginalOwner  string
	Collection     string
	MetadataFormat string
	Metadata       string
	Assets         map[string]uint64
	AssetCount     uint64
	Timestamp      uint64
}

func (asset *G45_C) Print() {
	fmt.Println("SCID: ", asset.SCID)
	fmt.Println("Frozen Assets: ", asset.FrozenAssets)
	fmt.Println("Frozen Metadata: ", asset.FrozenMetadata)
	fmt.Println("Metadata Format: ", asset.MetadataFormat)
	fmt.Println("Metadata: ", asset.Metadata)
	fmt.Println("Owner: ", asset.Owner)
	fmt.Println("Original Owner: ", asset.OriginalOwner)
	fmt.Println("Timestamp: ", asset.Timestamp)
}

func (asset *G45_C) JsonMetadata() (map[string]interface{}, error) {
	return formatMetadata(asset.MetadataFormat, asset.Metadata)
}

func (asset *G45_C) Validate(code string) (bool, error) {
	if trimCode(G45_C_CODE) != trimCode(code) {
		return false, fmt.Errorf("not a valid G45-C")
	}

	return true, nil
}

func (collection *G45_C) Parse(scId string, result *rpc.GetSC_Result) error {
	values := result.VariableStringKeys

	_, err := collection.Validate(result.Code)
	if err != nil {
		return err
	}

	collection.SCID = scId
	collection.FrozenAssets = values["frozenAssets"].(float64) != 0
	collection.FrozenMetadata = values["frozenMetadata"].(float64) != 0

	metadataFormat, err := DecodeString(values["metadataFormat"].(string))
	if err != nil {
		return err
	}
	collection.MetadataFormat = metadataFormat

	metadata, err := DecodeString(values["metadata"].(string))
	if err != nil {
		return err
	}
	collection.Metadata = metadata

	collection.Timestamp = uint64(values["timestamp"].(float64))

	owner, err := DecodeAddress(values["owner"].(string))
	if err != nil {
		return err
	}

	originalOwner, err := DecodeAddress(values["originalOwner"].(string))
	if err != nil {
		return err
	}

	collection.Owner = owner
	collection.OriginalOwner = originalOwner

	assetKey, err := regexp.Compile(`assets_(.+)`)
	if err != nil {
		return err
	}

	collection.Assets = make(map[string]uint64)
	for sKey, sValue := range values {
		if assetKey.Match([]byte(sKey)) {
			valueBytes, err := hex.DecodeString(sValue.(string))
			if err != nil {
				return err
			}

			var assets map[string]uint64
			err = json.Unmarshal(valueBytes, &assets)
			if err != nil {
				return err
			}

			for key, value := range assets {
				collection.Assets[key] = value
			}
		}
	}

	collection.AssetCount = uint64(len(collection.Assets))
	return nil
}

/** G45-AT **/

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
	MaxSupply        uint64
	TotalSupply      uint64
	Decimals         uint64
	Collection       string
	Owners           map[string]uint64
	Timestamp        uint64
}

func (asset *G45_AT) Print() {
	fmt.Println("SCID: ", asset.SCID)
	fmt.Println("Private: ", asset.Private)
	fmt.Println("Minter: ", asset.Minter)
	fmt.Println("Original Minter: ", asset.OriginalMinter)
	fmt.Println("Timestamp: ", asset.Timestamp)
	fmt.Println("Collection SCID: ", asset.Collection)
	fmt.Println("Frozen Metadata: ", asset.FrozenMetadata)
	fmt.Println("Frozen Mint: ", asset.FrozenMint)
	fmt.Println("Frozen Collection: ", asset.FrozenCollection)
	fmt.Println("Metadata Format: ", asset.MetadataFormat)
	fmt.Println("Metadata: ", asset.Metadata)
	fmt.Println("Max Supply: ", asset.MaxSupply)
	fmt.Println("Total Supply: ", asset.TotalSupply)
	fmt.Println("Decimals: ", asset.Decimals)
}

func (asset *G45_AT) JsonMetadata() (map[string]interface{}, error) {
	return formatMetadata(asset.MetadataFormat, asset.Metadata)
}

func (asset *G45_AT) Validate(code string) (bool, error) {
	switch trimCode(code) {
	case trimCode(G45_AT_PUBLIC_CODE):
		asset.Private = false
	case trimCode(G45_AT_PRIVATE_CODE):
		asset.Private = true
	default:
		return false, fmt.Errorf("not a valid G45-AT")
	}

	return true, nil
}

func (asset *G45_AT) Parse(scId string, result *rpc.GetSC_Result) error {
	values := result.VariableStringKeys

	_, err := asset.Validate(result.Code)
	if err != nil {
		return err
	}

	asset.SCID = scId
	asset.Timestamp = uint64(values["timestamp"].(float64))

	collection, err := DecodeString(values["collection"].(string))
	if err != nil {
		return err
	}
	asset.Collection = collection

	asset.FrozenMetadata = values["frozenMetadata"].(float64) != 0
	asset.FrozenMint = values["frozenMint"].(float64) != 0
	asset.FrozenCollection = values["frozenCollection"].(float64) != 0

	metadataFormat, err := DecodeString(values["metadataFormat"].(string))
	if err != nil {
		return err
	}
	asset.MetadataFormat = metadataFormat

	metadata, err := DecodeString(values["metadata"].(string))
	if err != nil {
		return err
	}
	asset.Metadata = metadata

	asset.MaxSupply = uint64(values["maxSupply"].(float64))
	asset.TotalSupply = uint64(values["totalSupply"].(float64))
	asset.Decimals = uint64(values["decimals"].(float64))

	minter, err := DecodeAddress(values["minter"].(string))
	if err != nil {
		return err
	}

	asset.Minter = minter

	originalMinter, err := DecodeAddress(values["originalMinter"].(string))
	if err != nil {
		return err
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
	return nil
}

/** G45-NFT **/

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

func (asset *G45_NFT) Print() {
	fmt.Println("SCID: ", asset.SCID)
	fmt.Println("Private: ", asset.Private)
	fmt.Println("Minter: ", asset.Minter)
	fmt.Println("Timestamp: ", asset.Timestamp)
	fmt.Println("Collection SCID: ", asset.Collection)
	fmt.Println("Metadata Format: ", asset.MetadataFormat)
	fmt.Println("Metadata: ", asset.Metadata)
	fmt.Println("Owner: ", asset.Owner)
}

func (asset *G45_NFT) JsonMetadata() (map[string]interface{}, error) {
	return formatMetadata(asset.MetadataFormat, asset.Metadata)
}

func (asset *G45_NFT) Validate(code string) (bool, error) {
	switch trimCode(code) {
	case trimCode(G45_NFT_PUBLIC_CODE):
		asset.Private = false
	case trimCode(G45_NFT_PRIVATE_CODE):
		asset.Private = true
	default:
		return false, fmt.Errorf("not a valid G45-NFT")
	}

	return true, nil
}

func (asset *G45_NFT) Parse(scId string, result *rpc.GetSC_Result) error {
	values := result.VariableStringKeys

	_, err := asset.Validate(result.Code)
	if err != nil {
		return err
	}

	asset.SCID = scId
	asset.Timestamp = uint64(values["timestamp"].(float64))

	collection, err := DecodeString(values["collection"].(string))
	if err != nil {
		return err
	}
	asset.Collection = collection

	metadataFormat, err := DecodeString(values["metadataFormat"].(string))
	if err != nil {
		return err
	}
	asset.MetadataFormat = metadataFormat

	metadata, err := DecodeString(values["metadata"].(string))
	if err != nil {
		return err
	}
	asset.Metadata = metadata

	owner, err := DecodeString(values["owner"].(string))
	if err != nil {
		return err
	}
	asset.Owner = owner

	minter, err := DecodeAddress(values["minter"].(string))
	if err != nil {
		return err
	}

	asset.Minter = minter
	return nil
}
