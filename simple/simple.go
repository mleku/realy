package simple

import (
	"realy.lol/kinds"
	"realy.lol/tag"
	"realy.lol/tags"
	"realy.lol/timestamp"
)

type Filter struct {
	Kinds   *kinds.T     `json:"kinds,omitempty"`
	Authors *tag.T       `json:"authors,omitempty"`
	Tags    *tags.T      `json:"-,omitempty"`
	Since   *timestamp.T `json:"since,omitempty"`
	Until   *timestamp.T `json:"until,omitempty"`
}

type Fulltext struct {
	Filter
	Search []byte
}
