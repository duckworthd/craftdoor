package model

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
)

// KeyModel accesses the key table.
type KeyModel struct {
	db *sqlx.DB
}

// NewKeyModel returns a new model.
func NewKeyModel(db *sqlx.DB) *KeyModel {
	return &KeyModel{db: db}
}

// Key represents a single row.
type Key struct {
	// Integer ID. Auto incremented.
	ID int64 `json:"id" db:"id"`

	// UUID stored on the key.
	UUID string `json:"uuid" db:"uuid"`

	// ID of member associated with this key.
	MemberID *int64 `json:"member_id" db:"member_id"`
}

// KeyInfo contains all details about a key.
type KeyInfo struct {
	Key    Key     `json:"key"`
	Member *Member `json:"member"`
}

// List returns all entries from the table
func (m *KeyModel) List(ctx context.Context) ([]Key, error) {
	res := []Key{}
	err := m.db.SelectContext(ctx, &res, queryListKeys)
	if err != nil {
		return nil, err
	}
	return res, nil
}

// Get a single row by id.
func (m *KeyModel) Get(ctx context.Context, id int64) (*KeyInfo, error) {
	key := Key{}
	err := m.db.GetContext(ctx, &key, queryGetKey, id)
	if err != nil {
		return nil, err
	}

	// Get Member associated with this key. There may be 0 or 1.
	//
	// TODO(duckworthd): Delegate this logic to MemberModel.
	members := []Member{}
	err = m.db.SelectContext(ctx, &members, queryGetMemberByID, key.MemberID)
	if err != nil {
		return nil, err
	}

	// Extract one and only member if possible.
	var member *Member
	switch n := len(members); n {
	case 0:
		member = nil
		break
	case 1:
		member = &members[0]
		break
	default:
		return nil, fmt.Errorf("Found %d > 1 members matching key=%s", n, key.UUID)
	}

	res := KeyInfo{
		Key:    key,
		Member: member,
	}
	return &res, nil
}

// Create inserts a new row into the table
func (m *KeyModel) Create(ctx context.Context, k *Key) error {
	res, err := m.db.NamedExecContext(ctx, queryCreateKey, k)
	if err != nil {
		return err
	}
	k.ID, err = res.LastInsertId()
	return err
}

// Update updates a single row's fields.
func (m *KeyModel) Update(ctx context.Context, k *Key) error {
	res, err := m.db.NamedExecContext(ctx, queryUpdateKey, k)
	if err != nil {
		return err
	}
	k.ID, err = res.LastInsertId()
	return err
}

// Delete deletes a single row from the table
func (m *KeyModel) Delete(ctx context.Context, id int64) error {
	_, err := m.db.ExecContext(ctx, queryDeleteKey, id)
	return err
}

// IsAccessAllowed returns whether the key has access.
func (m *KeyModel) IsAccessAllowed(ctx context.Context, keyID string) (bool, error) {
	var res bool
	err := m.db.GetContext(ctx, &res, accessAllowed, keyID)
	return res, err
}

const (
	queryCreateKey = `
INSERT INTO "main"."key"
( uuid,  member_id)
VALUES
(:uuid, :member_id)`
	queryListKeys = `
SELECT "id"
    , "uuid"
	, "member_id"
FROM "key"
ORDER BY "id"`
	queryGetKey = `
SELECT "id"
    , "uuid"
	, "member_id"
FROM "key"
WHERE id = ?`
	queryGetMemberByID = `
SELECT "id"
	, "name"
FROM "member"
WHERE id = ?`
	queryUpdateKey = `
UPDATE "key"
SET   "uuid"      = :uuid
	, "member_id" = :member_id
WHERE "id" = :id`
	queryDeleteKey = `
DELETE FROM "key"
WHERE id = ?`
	accessAllowed = `
SELECT COUNT(*) > 0
FROM key
JOIN  member
	ON (key.member_id = member.id)
WHERE
	key.uuid = ?`
)
