// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package apptoken

import (
	"context"

	"github.com/hashicorp/boundary/internal/db"
	"github.com/hashicorp/boundary/internal/errors"
	"github.com/hashicorp/boundary/internal/kms"
	"github.com/hashicorp/go-dbw"
)

// DeleteAppToken will delete the app token from the repository. It is
// idempotent so if the app token was not found, return 0 (no rows affected)
// and nil. No options are currently supported.
func (r *Repository) DeleteAppToken(ctx context.Context, publicId string, _ ...Option) (int, error) {
	const op = "apptoken.(Repository).DeleteAppToken"
	if publicId == "" {
		return db.NoRowsAffected, errors.New(ctx, errors.InvalidPublicId, op, "missing public id")
	}
	at := AllocAppToken()
	at.PublicId = publicId
	if err := r.reader.LookupByPublicId(ctx, at); err != nil {
		if errors.Is(err, dbw.ErrRecordNotFound) {
			return db.NoRowsAffected, nil
		}
	}

	oplogWrapper, err := r.kms.GetWrapper(ctx, at.ScopeId, kms.KeyPurposeOplog)
	if err != nil {
		return db.NoRowsAffected, errors.Wrap(ctx, err, op, errors.WithMsg("unable to get oplog wrapper"))
	}
}
