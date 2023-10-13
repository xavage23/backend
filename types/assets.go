package types

import "time"

type AssetMetadata struct {
	Exists       bool       `json:"exists" description:"Whether the asset exists or not"`
	Path         string     `json:"path,omitempty" description:"The path to the asset based on $cdnUrl"`
	DefaultPath  string     `json:"default_path" description:"The path to the default asset based on $cdnUrl. May be empty if there is no default asset"`
	Type         string     `json:"type,omitempty" description:"Asset type (banner, icon etc.)"`
	Size         int64      `json:"size,omitempty" description:"The size of the asset in bytes, if it exists"`
	LastModified *time.Time `json:"last_modified,omitempty" description:"The last modified date of the asset, if it exists"`
	Errors       []string   `json:"errors,omitempty" description:"Any errors that occurred while trying to get the asset"`
}
