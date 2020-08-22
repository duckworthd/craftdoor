package model

import (
	"context"

	"github.com/jmoiron/sqlx"
)

// MemberModel accesses the db
type MemberModel struct {
	db *sqlx.DB
}

// NewMemberModel returns a new model
func NewMemberModel(db *sqlx.DB) *MemberModel {
	return &MemberModel{db: db}
}

// Member represents a single row
type Member struct {
	// Member ID. Auto incremented.
	ID int64 `json:"id" db:"id"`

	// Member's name.
	Name string `json:"name" db:"name"`
}

// MemberInfo contains all details about a member.
type MemberInfo struct {
	// Member's basic information.
	Member Member `json:"member"`

	// Keys associated with this member.
	Keys []Key `json:"keys"`
}

// Create creates a new entry in the table
func (m *MemberModel) Create(ctx context.Context, t *Member) error {
	res, err := m.db.NamedExecContext(ctx, queryCreateMember, t)
	if err != nil {
		return err
	}
	t.ID, err = res.LastInsertId()
	return err
}

// List returns all entries from the table
func (m *MemberModel) List(ctx context.Context) ([]Member, error) {
	res := []Member{}
	err := m.db.SelectContext(ctx, &res, queryListMembers)
	if err != nil {
		return nil, err
	}
	return res, nil
}

// Get a single row by id.
func (m *MemberModel) Get(ctx context.Context, id int64) (*MemberInfo, error) {
	var err error

	// Get member's core details.
	member := Member{}
	err = m.db.GetContext(ctx, &member, queryGetMember, id)
	if err != nil {
		return nil, err
	}

	// Get keys associated with this member.
	//
	// TODO(duckworthd): Move this query to KeyModel.
	keys := []Key{}
	err = m.db.SelectContext(ctx, &keys, queryKeysByMemberID, id)
	if err != nil {
		return nil, err
	}

	res := &MemberInfo{
		Member: member,
		Keys:   keys,
	}
	return res, nil
}

// Update updates a single entry in the table
func (m *MemberModel) Update(ctx context.Context, t Member) error {
	_, err := m.db.NamedExecContext(ctx, queryUpdateMember, t)
	return err
}

// Delete deletes a single entry from the table
func (m *MemberModel) Delete(ctx context.Context, id int64) error {
	_, err := m.db.ExecContext(ctx, queryDeleteMember, id)
	return err
}

const (
	queryCreateMember = `
INSERT INTO "member"
("name")
VALUES
(:name)`
	queryListMembers = `
SELECT "id"
	, "name"
FROM "member"
ORDER BY "id"`
	queryGetMember = `
SELECT "id"
	, "name"
FROM "member"
WHERE id = ?`
	queryKeysByMemberID = `
SELECT "id"
	, "uuid"
	, "member_id"
FROM "key"
WHERE member_id = ?`
	queryUpdateMember = `
UPDATE "member"
SET "name" = :name
WHERE "id" = :id`
	queryDeleteMember = `
DELETE FROM "member"
WHERE id = ?`
)
