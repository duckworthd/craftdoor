package model

import (
	"github.com/jmoiron/sqlx"
	"github.com/pakohan/craftdoor/model/key"
	"github.com/pakohan/craftdoor/model/member"
)

// Model holds all models
type Model struct {
	KeyModel    *key.Model
	MemberModel *member.Model
}

// New returns all models initialized
func New(db *sqlx.DB) Model {
	return Model{
		KeyModel:    key.New(db),
		MemberModel: member.New(db),
	}
}
