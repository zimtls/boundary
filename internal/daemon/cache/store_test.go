// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package cache

import (
	"context"
	"testing"
	"time"

	"github.com/hashicorp/boundary/internal/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStoredToken(t *testing.T) {
	ctx := context.Background()
	s, err := Open(ctx)
	require.NoError(t, err)

	rw := db.New(s.conn)

	u := User{
		BoundaryAddr: "boundary",
		UserId:       "u_123456",
	}
	require.NoError(t, rw.Create(ctx, &u))

	st := storedToken{
		BoundaryAddr: u.BoundaryAddr,
		UserId:       u.UserId,
		KeyringType:  "keyring",
		TokenName:    "default",
		AuthTokenId:  "at_1234567890",
	}
	before := time.Now().Truncate(1 * time.Millisecond)
	require.NoError(t, rw.Create(ctx, &st))

	require.NoError(t, rw.LookupById(ctx, &st))
	assert.GreaterOrEqual(t, st.LastAccessedTime, before)

	st.AuthTokenId = "at_0987654321"
	n, err := rw.Update(ctx, &st, []string{"AuthTokenId"}, nil)
	assert.NoError(t, err)
	assert.Equal(t, 1, n)

	n, err = rw.Delete(ctx, &st)
	assert.NoError(t, err)
	assert.Equal(t, 1, n)
}

func TestUser(t *testing.T) {
	ctx := context.Background()
	s, err := Open(ctx)
	require.NoError(t, err)

	rw := db.New(s.conn)

	u := User{
		BoundaryAddr: "boundary",
		UserId:       "u_123456",
	}
	require.NoError(t, rw.Create(ctx, &u))

	st := storedToken{
		BoundaryAddr: u.BoundaryAddr,
		UserId:       u.UserId,
		KeyringType:  "keyring",
		TokenName:    "default",
		AuthTokenId:  "at_1234567890",
	}
	require.NoError(t, rw.Create(ctx, &st))

	require.NoError(t, rw.LookupById(ctx, &st))
	assert.NotNil(t, st)

	// Deleting the user deletes the tokens
	n, err := rw.Delete(ctx, &u)
	assert.NoError(t, err)
	assert.Equal(t, 1, n)

	require.ErrorContains(t, rw.LookupById(ctx, &st), "not found")
}

func TestTarget(t *testing.T) {
	ctx := context.Background()
	s, err := Open(ctx)
	require.NoError(t, err)

	rw := db.New(s.conn)

	u := &User{
		BoundaryAddr: "boundary",
		UserId:       "u_1234567890",
	}
	require.NoError(t, rw.Create(ctx, u))

	t.Run("target without persona user id", func(t *testing.T) {
		unknownTarget := &Target{
			BoundaryAddr: "some unknown addr",
			Id:           "tssh_1234567890",
			Name:         "target",
			Description:  "target desc",
			Address:      "some address",
			Item:         "{id:'tssh_1234567890'}",
		}
		require.ErrorContains(t, rw.Create(ctx, unknownTarget), "FOREIGN KEY constraint")
	})

	t.Run("target actions", func(t *testing.T) {
		target := &Target{
			BoundaryAddr:   u.BoundaryAddr,
			BoundaryUserId: u.UserId,
			Id:             "tssh_1234567890",
			Name:           "target",
			Description:    "target desc",
			Address:        "some address",
			Item:           "{id:'tssh_1234567890'}",
		}

		require.NoError(t, rw.Create(ctx, target))

		require.NoError(t, rw.LookupById(ctx, target))

		target.Address = "new address"
		n, err := rw.Update(ctx, target, []string{"address"}, nil)
		assert.NoError(t, err)
		assert.Equal(t, 1, n)

		n, err = rw.Delete(ctx, target)
		assert.NoError(t, err)
		assert.Equal(t, 1, n)
	})

	target := &Target{
		BoundaryAddr:   u.BoundaryAddr,
		BoundaryUserId: u.UserId,
		Id:             "tssh_1234567890",
		Name:           "target",
		Description:    "target desc",
		Address:        "some address",
		Item:           "{id:'tssh_1234567890'}",
	}
	require.NoError(t, rw.Create(ctx, target))

	t.Run("lookup a target", func(t *testing.T) {
		lookTar := &Target{
			BoundaryAddr:   target.BoundaryAddr,
			BoundaryUserId: u.UserId,
			Id:             target.Id,
		}
		assert.NoError(t, rw.LookupById(ctx, lookTar))
		assert.NotNil(t, lookTar)
	})

	t.Run("deleting the user deletes the target", func(t *testing.T) {
		n, err := rw.Delete(ctx, u)
		require.NoError(t, err)
		require.Equal(t, 1, n)

		lookTar := &Target{
			BoundaryAddr:   target.BoundaryAddr,
			BoundaryUserId: target.BoundaryUserId,
			Id:             target.Id,
		}
		assert.ErrorContains(t, rw.LookupById(ctx, lookTar), "not found")
	})
}
