package lib

import "github.com/google/uuid"

// State contains information about the current tag in front of the reader.
type State struct {
	// Unique identifier for the state of this system.
	UUID uuid.UUID `json:"uuid"`

	// If True, TagInfo will be non-null.
	IsTagAvailable bool `json:"is_tag_available"`

	// Information about the current tag in front of the reader.
	TagInfo *TagInfo `json:"tag_info"`
}

// TagInfo contains details about the tag in front of the reader.
type TagInfo struct {
	// ID of the tag.
	ID string `json:"id"`

	// Additional data stored in the tag's data blocks.
	Data string `json:"data"`
}
