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

//go:embed g45_c.bas
var G45_BC_CODE string

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

func (asset *G45_FAT) Parse(scId string, result *rpc.GetSC_Result) error {
	values := result.VariableStringKeys

	code := strings.ReplaceAll(strings.ReplaceAll(result.Code, "\r", ""), "\n", "")
	g45_fat_public_code := strings.ReplaceAll(strings.ReplaceAll(G45_FAT_PUBLIC_CODE, "\r", ""), "\n", "")
	g45_fat_private_code := strings.ReplaceAll(strings.ReplaceAll(G45_FAT_PRIVATE_CODE, "\r", ""), "\n", "")

	switch code {
	case g45_fat_public_code:
		asset.Private = false
	case g45_fat_private_code:
		asset.Private = true
	default:
		return fmt.Errorf("not a valid G45-FAT")
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

/** G45-BC **/

type G45_BC struct {
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

func (asset *G45_BC) Print() {
	fmt.Println("SCID: ", asset.SCID)
	fmt.Println("Frozen Assets: ", asset.FrozenAssets)
	fmt.Println("Frozen Metadata: ", asset.FrozenMetadata)
	fmt.Println("Metadata Format: ", asset.MetadataFormat)
	fmt.Println("Metadata: ", asset.Metadata)
	fmt.Println("Owner: ", asset.Owner)
	fmt.Println("Original Owner: ", asset.OriginalOwner)
	fmt.Println("Timestamp: ", asset.Timestamp)
}

func (asset *G45_BC) JsonMetadata() (map[string]interface{}, error) {
	return formatMetadata(asset.MetadataFormat, asset.Metadata)
}

func (collection *G45_BC) Parse(scId string, result *rpc.GetSC_Result) error {
	values := result.VariableStringKeys
	code := strings.ReplaceAll(strings.ReplaceAll(result.Code, "\r", ""), "\n", "")
	g45_c_code := strings.ReplaceAll(strings.ReplaceAll(G45_BC_CODE, "\r", ""), "\n", "")
	if code != g45_c_code {
		return fmt.Errorf("not a valid G45-C")
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

/** G45-C **/

type G45_C struct {
	SCID           string
	FrozenAssets   bool
	FrozenMetadata bool
	Owner          string
	OriginalOwner  string
	AssetCount     uint64
	MetadataFormat string
	Metadata       string
	Assets         map[string]uint64
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
	fmt.Println("Asset Count: ", asset.AssetCount)
	fmt.Println("Timestamp: ", asset.Timestamp)
}

func (asset *G45_C) JsonMetadata() (map[string]interface{}, error) {
	return formatMetadata(asset.MetadataFormat, asset.Metadata)
}

func (collection *G45_C) Parse(scId string, result *rpc.GetSC_Result) error {
	values := result.VariableStringKeys
	code := strings.ReplaceAll(strings.ReplaceAll(result.Code, "\r", ""), "\n", "")
	g45_dc_code := strings.ReplaceAll(strings.ReplaceAll(G45_C_CODE, "\r", ""), "\n", "")
	if code != g45_dc_code {
		return fmt.Errorf("not a valid G45-DC")
	}

	collection.SCID = scId
	collection.FrozenAssets = values["frozenAssets"].(float64) != 0
	collection.FrozenMetadata = values["frozenMetadata"].(float64) != 0
	collection.AssetCount = uint64(values["assetCount"].(float64))
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

	assetKey, err := regexp.Compile(`asset_(.+)`)
	if err != nil {
		return err
	}

	collection.Assets = make(map[string]uint64)
	for key, value := range values {
		if assetKey.Match([]byte(key)) {
			assetSCID := assetKey.ReplaceAllString(key, "$1")
			collection.Assets[assetSCID] = uint64(value.(float64))
		}
	}

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

func (asset *G45_AT) Parse(scId string, result *rpc.GetSC_Result) error {
	values := result.VariableStringKeys
	code := strings.ReplaceAll(strings.ReplaceAll(result.Code, "\r", ""), "\n", "")
	g45_at_public_code := strings.ReplaceAll(strings.ReplaceAll(G45_AT_PUBLIC_CODE, "\r", ""), "\n", "")
	g45_at_private_code := strings.ReplaceAll(strings.ReplaceAll(G45_AT_PRIVATE_CODE, "\r", ""), "\n", "")

	switch code {
	case g45_at_public_code:
		asset.Private = false
	case g45_at_private_code:
		asset.Private = true
	default:
		return fmt.Errorf("not a valid G45-AT")
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

func (asset *G45_NFT) Parse(scId string, result *rpc.GetSC_Result) error {
	values := result.VariableStringKeys

	code := strings.ReplaceAll(strings.ReplaceAll(result.Code, "\r", ""), "\n", "")
	g45_nft_public_code := strings.ReplaceAll(strings.ReplaceAll(G45_NFT_PUBLIC_CODE, "\r", ""), "\n", "")
	g45_nft_private_code := strings.ReplaceAll(strings.ReplaceAll(G45_NFT_PRIVATE_CODE, "\r", ""), "\n", "")

	switch code {
	case g45_nft_public_code:
		asset.Private = false
	case g45_nft_private_code:
		asset.Private = true
	default:
		return fmt.Errorf("not a valid G45-NFT")
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
