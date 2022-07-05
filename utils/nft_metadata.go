package utils

type NFTMetadataCollection struct {
	Name        string        `json:"name"`
	Description string        `json:"description"`
	Collection  []NFTMetadata `json:"collection"`
}

type NFTMetadata struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	ExternalUrl string                 `json:"external_url"`
	Image       string                 `json:"image"`
	Attributes  []NFTMetadataAttribute `json:"attributes"`
}

type NFTMetadataAttribute struct {
	TraitType string `json:"trait_type"`
	Value     string `json:"value"`
}
