// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package cache

import (
	"context"
	"database/sql"
	"encoding/json"
	stdErrors "errors"
	"fmt"
	"time"

	"github.com/hashicorp/boundary/api/authtokens"
	"github.com/hashicorp/boundary/api/targets"
	"github.com/hashicorp/boundary/internal/db"
	"github.com/hashicorp/boundary/internal/errors"
	"github.com/hashicorp/boundary/internal/types/resource"
	"github.com/hashicorp/boundary/internal/util"
	"github.com/hashicorp/mql"
)

const (
	personaLimit          = 50
	personaStalenessLimit = 36 * time.Hour
)

// tokenLookupFn takes a token name and returns the token
type tokenLookupFn func(keyring string, tokenName string) *authtokens.AuthToken

type Repository struct {
	rw            *db.Db
	tokenLookupFn tokenLookupFn
}

func NewRepository(ctx context.Context, s *Store, tFn tokenLookupFn, opt ...Option) (*Repository, error) {
	const op = "cache.NewRepository"
	switch {
	case util.IsNil(s):
		return nil, errors.New(ctx, errors.InvalidParameter, op, "missing store")
	case util.IsNil(tFn):
		return nil, errors.New(ctx, errors.InvalidParameter, op, "missing token lookup function")
	}

	return &Repository{rw: db.New(s.conn), tokenLookupFn: tFn}, nil
}

func (r *Repository) SaveError(ctx context.Context, resourceType string, err error) error {
	const op = "cache.(Repository).StoreError"
	switch {
	case resourceType == "":
		return errors.New(ctx, errors.InvalidParameter, op, "resource type is empty")
	case err == nil:
		return errors.New(ctx, errors.InvalidParameter, op, "error is nil")
	}
	apiErr := &ApiError{
		ResourceType: resourceType,
		Error:        err.Error(),
	}
	onConflict := db.OnConflict{
		Target: db.Columns{"token_name", "resource_type"},
		Action: db.SetColumns([]string{"error", "create_time"}),
	}
	if err := r.rw.Create(ctx, apiErr, db.WithOnConflict(&onConflict)); err != nil {
		return errors.Wrap(ctx, err, op)
	}
	return nil
}

func (r *Repository) refreshTargets(ctx context.Context, u *User, targets []*targets.Target) error {
	const op = "cache.(Repository).refreshTargets"
	switch {
	case util.IsNil(u):
		return errors.New(ctx, errors.InvalidParameter, op, "user is missing")
	case u.UserId == "":
		return errors.New(ctx, errors.InvalidParameter, op, "user id is missing")
	case u.BoundaryAddr == "":
		return errors.New(ctx, errors.InvalidParameter, op, "boundary address is missing")
	}

	found := u.clone()
	if err := r.rw.LookupById(ctx, found); err != nil {
		// if this user isn't known, error out.l
		return errors.Wrap(ctx, err, op, errors.WithMsg("looking up user"))
	}

	_, err := r.rw.DoTx(ctx, db.StdRetryCnt, db.ExpBackoff{}, func(r db.Reader, w db.Writer) error {
		// TODO: Instead of deleting everything, use refresh tokens and apply the delta
		if _, err := w.Exec(ctx, "delete from cache_target where boundary_addr = @boundary_addr and boundary_user_id = @boundary_user_id",
			[]any{sql.Named("boundary_addr", u.BoundaryAddr), sql.Named("boundary_user_id", u.UserId)}); err != nil {
			return err
		}

		for _, t := range targets {
			item, err := json.Marshal(t)
			if err != nil {
				return err
			}
			newTarget := Target{
				BoundaryAddr:   u.BoundaryAddr,
				BoundaryUserId: u.UserId,
				Id:             t.Id,
				Name:           t.Name,
				Description:    t.Description,
				Address:        t.Address,
				Item:           string(item),
			}
			onConflict := db.OnConflict{
				Target: db.Columns{"boundary_addr", "boundary_user_id", "id"},
				Action: db.UpdateAll(true),
			}
			if err := w.Create(ctx, newTarget, db.WithOnConflict(&onConflict)); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		if saveErr := r.SaveError(ctx, resource.Target.String(), err); saveErr != nil {
			return stdErrors.Join(err, errors.Wrap(ctx, saveErr, op))
		}
		return errors.Wrap(ctx, err, op)
	}
	return nil
}

func (r *Repository) ListTargets(ctx context.Context, boundaryAddr, tokenName, keyringType string) ([]*targets.Target, error) {
	const op = "cache.(Repository).ListTargets"
	switch {
	case tokenName == "":
		return nil, errors.New(ctx, errors.InvalidParameter, op, "token name is missing")
	case keyringType == "":
		return nil, errors.New(ctx, errors.InvalidParameter, op, "keyring type is missing")
	case boundaryAddr == "":
		return nil, errors.New(ctx, errors.InvalidParameter, op, "boundary address is missing")
	}
	st, err := r.lookupStoredAuthToken(ctx, boundaryAddr, tokenName, keyringType)
	if err != nil {
		return nil, errors.Wrap(ctx, err, op)
	}
	if st == nil {
		return nil, errors.New(ctx, errors.NotFound, op, "stored auth token not found")
	}
	ret, err := r.searchTargets(ctx, st, "true", nil)
	if err != nil {
		return nil, errors.Wrap(ctx, err, op)
	}
	return ret, nil
}

func (r *Repository) QueryTargets(ctx context.Context, boundaryAddr, tokenName, keyringType, query string) ([]*targets.Target, error) {
	const op = "cache.(Repository).QueryTargets"
	switch {
	case tokenName == "":
		return nil, errors.New(ctx, errors.InvalidParameter, op, "token name is missing")
	case keyringType == "":
		return nil, errors.New(ctx, errors.InvalidParameter, op, "keyring type is missing")
	case boundaryAddr == "":
		return nil, errors.New(ctx, errors.InvalidParameter, op, "boundary address is missing")
	case query == "":
		return nil, errors.New(ctx, errors.InvalidParameter, op, "query is missing")
	}

	st, err := r.lookupStoredAuthToken(ctx, boundaryAddr, tokenName, keyringType)
	if err != nil {
		return nil, errors.Wrap(ctx, err, op)
	}
	if st == nil {
		return nil, errors.New(ctx, errors.NotFound, op, "stored auth token not found")
	}

	w, err := mql.Parse(query, Target{}, mql.WithIgnoredFields("BoundaryAddr", "PersonaUserId", "Item"))
	if err != nil {
		return nil, errors.Wrap(ctx, err, op, errors.WithCode(errors.InvalidParameter))
	}
	ret, err := r.searchTargets(ctx, st, w.Condition, w.Args)
	if err != nil {
		return nil, errors.Wrap(ctx, err, op)
	}
	return ret, nil
}

func (r *Repository) searchTargets(ctx context.Context, st *storedToken, condition string, searchArgs []any) ([]*targets.Target, error) {
	const op = "cache.(Repository).searchTargets"
	switch {
	case st == nil:
		return nil, errors.New(ctx, errors.InvalidParameter, op, "persona is missing")
	case st.UserId == "":
		return nil, errors.New(ctx, errors.InvalidParameter, op, "user id is missing")
	case st.BoundaryAddr == "":
		return nil, errors.New(ctx, errors.InvalidParameter, op, "boundary address is missing")
	case condition == "":
		return nil, errors.New(ctx, errors.InvalidParameter, op, "condition is missing")
	}

	condition = fmt.Sprintf("%s and (boundary_addr = ? and boundary_user_id = ?)", condition)
	args := append(searchArgs, st.BoundaryAddr, st.UserId)
	var cachedTargets []*Target
	if err := r.rw.SearchWhere(ctx, &cachedTargets, condition, args); err != nil {
		return nil, errors.Wrap(ctx, err, op)
	}

	retTargets := make([]*targets.Target, 0, len(cachedTargets))
	for _, cachedTar := range cachedTargets {
		var tar targets.Target
		if err := json.Unmarshal([]byte(cachedTar.Item), &tar); err != nil {
			return nil, errors.Wrap(ctx, err, op)
		}
		retTargets = append(retTargets, &tar)
	}
	return retTargets, nil
}

type Target struct {
	BoundaryAddr   string `gorm:"primaryKey"`
	BoundaryUserId string `gorm:"primaryKey"`
	Id             string `gorm:"primaryKey"`
	Name           string
	Description    string
	Address        string
	Item           string
}

func (*Target) TableName() string {
	return "cache_target"
}

type ApiError struct {
	TokenName    string `gorm:"primaryKey"`
	ResourceType string `gorm:"primaryKey"`
	Error        string
	CreateTime   time.Time
}

func (*ApiError) TableName() string {
	return "cache_api_error"
}
