// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package password

import (
	"context"
	"time"

	"github.com/hashicorp/boundary/internal/auth/password/store"
	"github.com/hashicorp/boundary/internal/db/timestamp"
	"github.com/hashicorp/boundary/internal/errors"
	"github.com/hashicorp/boundary/internal/oplog"
	"github.com/hashicorp/boundary/internal/types/resource"
	"google.golang.org/protobuf/proto"
)

// An Account contains a user name. It is owned by an auth method.
type Account struct {
	*store.Account
	tableName string

	// CredentialId is included when Authenticate or ChangePassword is
	// called. A new CredentialId is generated when a password is changed.
	CredentialId string `gorm:"->"`
}

func allocAccount() *Account {
	return &Account{
		Account: &store.Account{},
	}
}

// NewAccount creates a new in memory Account. LoginName, name, and
// description are the only valid options. All other options are ignored.
func NewAccount(ctx context.Context, authMethodId string, opt ...Option) (*Account, error) {
	const op = "password.NewAccount"
	// NOTE(mgaffney): The scopeId in the embedded *store.Account is
	// populated by a trigger in the database.
	if authMethodId == "" {
		return nil, errors.New(ctx, errors.InvalidParameter, op, "missing auth method id")
	}

	opts := GetOpts(opt...)
	a := &Account{
		Account: &store.Account{
			AuthMethodId: authMethodId,
			LoginName:    opts.WithLoginName,
			Name:         opts.withName,
			Description:  opts.withDescription,
		},
	}
	return a, nil
}

func (a *Account) clone() *Account {
	cp := proto.Clone(a.Account)
	return &Account{
		Account: cp.(*store.Account),
	}
}

// TableName returns the table name.
func (a *Account) TableName() string {
	if a.tableName != "" {
		return a.tableName
	}
	return "auth_password_account"
}

// SetTableName sets the table name.
func (a *Account) SetTableName(n string) {
	a.tableName = n
}

// GetResourceType returns the resource type of the Account
func (a *Account) GetResourceType() resource.Type {
	return resource.Account
}

// GetEmail returns the email, which will always be empty as this type doesn't
// currently support email
func (a *Account) GetEmail() string {
	return ""
}

// GetSubject returns the subject, which will always be empty as this type
// doesn't currently support subject
func (a *Account) GetSubject() string {
	return ""
}

func (a *Account) oplog(op oplog.OpType) oplog.Metadata {
	metadata := oplog.Metadata{
		"resource-public-id": []string{a.GetPublicId()},
		"resource-type":      []string{"password account"},
		"op-type":            []string{op.String()},
	}
	if a.AuthMethodId != "" {
		metadata["auth-method-id"] = []string{a.AuthMethodId}
	}
	return metadata
}

// sortAccount is used to encapsulate the information
// needed for sorting and filtering for pagination.
type sortAccount struct {
	publicId   string
	updateTime time.Time
}

type deletedAccount struct {
	PublicId   string `gorm:"primary_key"`
	DeleteTime *timestamp.Timestamp
}

// TableName returns the tablename to override the default gorm table name
func (s *deletedAccount) TableName() string {
	return "auth_account_deleted"
}
