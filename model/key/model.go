package key

import (
	"context"

	"github.com/jmoiron/sqlx"
)

// Model accesses the key table
type Model struct {
	db *sqlx.DB
}

// New returns a new model
func New(db *sqlx.DB) *Model {
	return &Model{db: db}
}

// Key represents a single row
type Key struct {
	// Integer ID. Auto incremented.
	ID int64 `json:"id" db:"id"`

	// UUID stored on the key.
	UUID string `json:"uuid" db:"uuid"`

	// ID of member associated with this key.
	MemberID *int64 `json:"member_id" db:"member_id"`
}

// List returns all entries from the table
func (m *Model) List(ctx context.Context) ([]Key, error) {
	res := []Key{}
	err := m.db.SelectContext(ctx, &res, queryList)
	if err != nil {
		return nil, err
	}
	return res, nil
}

// Create inserts a new row into the table
func (m *Model) Create(ctx context.Context, k *Key) error {
	res, err := m.db.NamedExecContext(ctx, queryCreate, k)
	if err != nil {
		return err
	}
	k.ID, err = res.LastInsertId()
	return err
}

// Update updates a single row's fields.
func (m *Model) Update(ctx context.Context, k *Key) error {
	res, err := m.db.NamedExecContext(ctx, queryUpdate, k)
	if err != nil {
		return err
	}
	k.ID, err = res.LastInsertId()
	return err
}

// Delete deletes a single row from the table
func (m *Model) Delete(ctx context.Context, id int64) error {
	_, err := m.db.ExecContext(ctx, queryDelete, id)
	return err
}

// IsAccessAllowed returns whether the key has access.
func (m *Model) IsAccessAllowed(ctx context.Context, keyID string) (bool, error) {
	var res bool
	err := m.db.GetContext(ctx, &res, accessAllowed, keyID)
	return res, err
}

const (
	queryCreate = `
INSERT INTO "main"."key"
( uuid,  member_id)
VALUES
(:uuid, :member_id)`
	queryList = `
SELECT "id"
    , "uuid"
	, "member_id"
FROM "key"
ORDER BY "id"`
	queryUpdate = `
UPDATE "key"
SET   "uuid"      = :uuid
	, "member_id" = :member_id
WHERE "id" = :id`
	queryDelete = `
DELETE FROM "key"
WHERE id = ?`
	accessAllowed = `
SELECT COUNT(*) > 0
FROM key
WHERE
	key.uuid = ?`
)
