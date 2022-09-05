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

//go:embed g45_at_public.bas
var G45_AT_PUBLIC_CODE string

//go:embed g45_at_private.bas
var G45_AT_PRIVATE_CODE string

//go:embed g45_atc.bas
var G45_ATC_CODE string

type G45_ATC struct {
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

func (a *G45_ATC) Print() {
	fmt.Println("SCID: ", a.SCID)
	fmt.Println("Frozen Collection: ", a.FrozenCollection)
	fmt.Println("Frozen Metadata: ", a.FrozenMetadata)
	fmt.Println("Metadata Format: ", a.MetadataFormat)
	fmt.Println("Metadata: ", a.Metadata)
	fmt.Println("Owner: ", a.Owner)
	fmt.Println("Original Owner: ", a.OriginalOwner)
	fmt.Println("Timestamp: ", a.Timestamp)
}

type G45_AT struct {
	SCID             string
	Private          bool
	Minter           string
	OriginalMinter   string
	FrozenMetadata   bool
	FrozenSupply     bool
	FrozenCollection bool
	MetadataFormat   string
	Metadata         string
	Supply           uint64
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
	fmt.Println("Frozen Supply: ", a.FrozenSupply)
	fmt.Println("Frozen Collection: ", a.FrozenCollection)
	fmt.Println("Metadata Format: ", a.MetadataFormat)
	fmt.Println("Metadata: ", a.Metadata)
	fmt.Println("Supply: ", a.Supply)
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

func GetG45_ATC(scid string, daemon *rpc_client.Daemon) (*G45_ATC, error) {
	result, err := daemon.GetSC(&rpc.GetSC_Params{
		SCID:      scid,
		Code:      true,
		Variables: true,
	})

	if err != nil {
		return nil, err
	}

	collection := &G45_ATC{}

	values := result.VariableStringKeys
	code := strings.ReplaceAll(strings.ReplaceAll(result.Code, "\r", ""), "\n", "")
	g45_atc_code := strings.ReplaceAll(strings.ReplaceAll(G45_ATC_CODE, "\r", ""), "\n", "")
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
	asset.FrozenSupply = values["frozenSupply"].(float64) != 0
	asset.FrozenCollection = values["frozenCollection"].(float64) != 0
	asset.MetadataFormat = decodeString(values["metadataFormat"].(string))
	asset.Metadata = decodeString(values["metadata"].(string))
	asset.Supply = uint64(values["supply"].(float64))

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
