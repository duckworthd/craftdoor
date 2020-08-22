package model

import (
	"github.com/jmoiron/sqlx"
)

// Model holds all models
type Model struct {
	KeyModel    *KeyModel
	MemberModel *MemberModel
}

// New returns all models initialized
func New(db *sqlx.DB) Model {
	return Model{
		KeyModel:    NewKeyModel(db),
		MemberModel: NewMemberModel(db),
	}
}
