// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package cache

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/boundary/api/authtokens"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testAuthTokenLookup(k, t string) *authtokens.AuthToken {
	return &authtokens.AuthToken{
		Id:           fmt.Sprintf("at_%s", t),
		Token:        fmt.Sprintf("at_%s_%s", t, k),
		UserId:       fmt.Sprintf("u_%s", t),
		AuthMethodId: fmt.Sprintf("ampw_%s", t),
		AccountId:    fmt.Sprintf("acctpw_%s", t),
	}
}

func TestRepository_CleanupOrphanedUsers(t *testing.T) {
	ctx := context.Background()
	s, err := Open(ctx)
	require.NoError(t, err)
	t.Cleanup(func() {
		s.conn.Close(ctx)
	})
	r, err := NewRepository(ctx, s, testAuthTokenLookup)
	require.NoError(t, err)

	userWithToken := &User{UserId: "userid", BoundaryAddr: "address"}
	require.NoError(t, r.rw.Create(ctx, userWithToken))
	tokenForUser := &storedToken{
		KeyringType:  "keyring",
		TokenName:    "token",
		BoundaryAddr: userWithToken.BoundaryAddr,
		UserId:       userWithToken.UserId,
		AuthTokenId:  "at_someid",
	}
	require.NoError(t, r.rw.Create(ctx, tokenForUser))

	orphanedUser := &User{UserId: "unmatched", BoundaryAddr: "unmatched"}
	require.NoError(t, r.rw.Create(ctx, orphanedUser))

	require.NoError(t, r.cleanupOrphanedUsers(ctx))

	got, err := r.listUsers(ctx)
	require.NoError(t, err)
	assert.ElementsMatch(t, got, []*User{userWithToken})
}

func TestRepository_StorkedToken_ListUsers(t *testing.T) {
	ctx := context.Background()

	cases := []struct {
		name          string
		addedTokens   []*storedToken
		deletedTokens []*storedToken
		want          []*User
	}{
		{
			name: "single token",
			addedTokens: []*storedToken{
				{
					BoundaryAddr: "address",
					TokenName:    "token",
					KeyringType:  "keyring",
					AuthTokenId:  "at_token",
				},
			},
			want: []*User{
				{
					UserId:       "u_token",
					BoundaryAddr: "address",
				},
			},
		},
		{
			name: "deleted token",
			addedTokens: []*storedToken{
				{
					BoundaryAddr: "address",
					TokenName:    "token",
					KeyringType:  "keyring",
					AuthTokenId:  "at_token",
				},
			},
			deletedTokens: []*storedToken{
				{
					BoundaryAddr: "address",
					TokenName:    "token",
					KeyringType:  "keyring",
					AuthTokenId:  "at_token",
				},
			},
			want: []*User{},
		},
		{
			name: "same keyring and token name different address",
			addedTokens: []*storedToken{
				{
					BoundaryAddr: "address",
					TokenName:    "token",
					KeyringType:  "keyring",
					AuthTokenId:  "at_token",
				},
				{
					BoundaryAddr: "address2",
					TokenName:    "token",
					KeyringType:  "keyring",
					AuthTokenId:  "at_token",
				},
			},
			want: []*User{
				{
					UserId:       "u_token",
					BoundaryAddr: "address2",
				},
			},
		},
		{
			name: "different keyring and token name",
			addedTokens: []*storedToken{
				{
					BoundaryAddr: "address",
					TokenName:    "token",
					KeyringType:  "keyring",
					AuthTokenId:  "at_token",
				},
				{
					BoundaryAddr: "address",
					TokenName:    "token2",
					KeyringType:  "keyring2",
					AuthTokenId:  "at_token2",
				},
			},
			want: []*User{
				{
					UserId:       "u_token",
					BoundaryAddr: "address",
				},
				{
					UserId:       "u_token2",
					BoundaryAddr: "address",
				},
			},
		},
		{
			name: "different keyring and token name one deleted",
			addedTokens: []*storedToken{
				{
					BoundaryAddr: "address",
					TokenName:    "token",
					KeyringType:  "keyring",
					AuthTokenId:  "at_token",
				},
				{
					BoundaryAddr: "address",
					TokenName:    "token2",
					KeyringType:  "keyring2",
					AuthTokenId:  "at_token2",
				},
			},
			deletedTokens: []*storedToken{
				{
					BoundaryAddr: "address",
					TokenName:    "token",
					KeyringType:  "keyring",
					AuthTokenId:  "at_token",
				},
			},
			want: []*User{
				{
					UserId:       "u_token2",
					BoundaryAddr: "address",
				},
			},
		},
		{
			name: "different address and token",
			addedTokens: []*storedToken{
				{
					BoundaryAddr: "address",
					TokenName:    "token",
					KeyringType:  "keyring",
					AuthTokenId:  "at_token",
				},
				{
					BoundaryAddr: "address2",
					TokenName:    "token2",
					KeyringType:  "keyring2",
					AuthTokenId:  "at_token2",
				},
			},
			want: []*User{
				{
					UserId:       "u_token",
					BoundaryAddr: "address",
				},
				{
					UserId:       "u_token2",
					BoundaryAddr: "address2",
				},
			},
		},
		{
			name: "different address and token name one deleted",
			addedTokens: []*storedToken{
				{
					BoundaryAddr: "address",
					TokenName:    "token",
					KeyringType:  "keyring",
					AuthTokenId:  "at_token",
				},
				{
					BoundaryAddr: "address2",
					TokenName:    "token2",
					KeyringType:  "keyring2",
					AuthTokenId:  "at_token2",
				},
			},
			deletedTokens: []*storedToken{
				{
					BoundaryAddr: "address",
					TokenName:    "token",
					KeyringType:  "keyring",
					AuthTokenId:  "at_token",
				},
			},
			want: []*User{
				{
					UserId:       "u_token2",
					BoundaryAddr: "address2",
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s, err := Open(ctx)
			require.NoError(t, err)
			t.Cleanup(func() {
				s.conn.Close(ctx)
			})
			r, err := NewRepository(ctx, s, testAuthTokenLookup)
			require.NoError(t, err)
			for _, tok := range tc.addedTokens {
				require.NoError(t, r.AddStoredToken(ctx, tok.BoundaryAddr, tok.TokenName, tok.KeyringType, tok.AuthTokenId))
			}
			for _, tok := range tc.deletedTokens {
				require.NoError(t, r.deleteStoredToken(ctx, tok))
			}
			got, err := r.listUsers(ctx)
			require.NoError(t, err)
			assert.ElementsMatch(t, got, tc.want)
		})
	}
}

func TestRepository_AddStoredToken_EvictsOverLimit(t *testing.T) {
	ctx := context.Background()
	s, err := Open(ctx)
	require.NoError(t, err)

	r, err := NewRepository(ctx, s, testAuthTokenLookup)
	require.NoError(t, err)

	addr := "address"
	keyringType := "keyring"
	tokenName := "token"
	at := testAuthTokenLookup(keyringType, tokenName)

	assert.NoError(t, r.AddStoredToken(ctx, addr, tokenName, keyringType, at.Id))
	assert.NoError(t, r.AddStoredToken(ctx, addr, tokenName, keyringType, at.Id))
	for i := 0; i < personaLimit; i++ {
		kr := fmt.Sprintf("%s%d", keyringType, i)
		tn := fmt.Sprintf("%s%d", tokenName, i)
		at := testAuthTokenLookup(kr, tn)
		assert.NoError(t, r.AddStoredToken(ctx, addr, tn, kr, at.Id))
	}
	// Lookup the first persona added. It should have been evicted for being
	// used the least recently.
	gotAtId, err := r.LookupStoredAuthTokenId(ctx, addr, tokenName, keyringType)
	assert.NoError(t, err)
	assert.Empty(t, gotAtId)

	gotAtId, err = r.LookupStoredAuthTokenId(ctx, addr, tokenName+"0", keyringType+"0")
	assert.NoError(t, err)
	assert.NotEmpty(t, gotAtId)
}

func TestRepository_AddStoredToken_AddingExistingUpdatesLastAccessedTime(t *testing.T) {
	ctx := context.Background()
	s, err := Open(ctx, WithDebug(true))
	require.NoError(t, err)

	r, err := NewRepository(ctx, s, testAuthTokenLookup)
	require.NoError(t, err)

	addr := "address"
	tokenname := "token"
	keyringType := "keyring"
	at := testAuthTokenLookup(keyringType, tokenname)
	assert.NoError(t, r.AddStoredToken(ctx, addr, tokenname, keyringType, at.Id))

	us, err := r.listUsers(ctx)
	require.NoError(t, err)
	assert.Len(t, us, 1)

	// same token info, different address means different user
	addr2 := "address2"
	assert.NoError(t, r.AddStoredToken(ctx, addr2, tokenname, keyringType, at.Id))

	us, err = r.listUsers(ctx)
	require.NoError(t, err)
	assert.Len(t, us, 1)

	time.Sleep(10 * time.Millisecond)
	assert.NoError(t, r.AddStoredToken(ctx, addr, tokenname, keyringType, at.Id))

	us, err = r.listUsers(ctx)
	require.NoError(t, err)
	assert.Len(t, us, 1)

	gotP1, err := r.lookupStoredAuthToken(ctx, addr, tokenname, keyringType)
	require.NoError(t, err)
	require.NotNil(t, gotP1)

	gotP2, err := r.lookupStoredAuthToken(ctx, addr2, tokenname, keyringType)
	require.NoError(t, err)
	require.NotNil(t, gotP2)

	assert.Greater(t, gotP1.LastAccessedTime, gotP2.LastAccessedTime)
}

func TestRepository_ListStoredTokens(t *testing.T) {
	ctx := context.Background()
	s, err := Open(ctx)
	require.NoError(t, err)

	r, err := NewRepository(ctx, s, testAuthTokenLookup)
	require.NoError(t, err)

	t.Run("no token", func(t *testing.T) {
		gotP, err := r.listStoredTokens(ctx)
		assert.NoError(t, err)
		assert.Empty(t, gotP)
	})

	personaCount := 15
	addr := "address"
	keyringType := "keyring"
	tokenName := "token"
	at := testAuthTokenLookup(keyringType, tokenName)

	for i := 0; i < personaCount; i++ {
		thisAddr := fmt.Sprintf("%s%d", addr, i)
		require.NoError(t, r.AddStoredToken(ctx, thisAddr, tokenName, keyringType, at.Id))
	}

	t.Run("many tokens", func(t *testing.T) {
		gotP, err := r.listStoredTokens(ctx)
		assert.NoError(t, err)
		assert.Len(t, gotP, personaCount)
	})
}

func TestRepository_DeletePersona(t *testing.T) {
	ctx := context.Background()
	s, err := Open(ctx)
	require.NoError(t, err)

	r, err := NewRepository(ctx, s, testAuthTokenLookup)
	require.NoError(t, err)

	t.Run("delete non existing", func(t *testing.T) {
		assert.ErrorContains(t, r.deleteStoredToken(ctx, &storedToken{BoundaryAddr: "unknown", KeyringType: "Unknown", TokenName: "Unknown"}), "not found")
	})

	t.Run("delete existing", func(t *testing.T) {
		addr := "address"
		keyringType := "keyring"
		tokenName := "token"
		at := testAuthTokenLookup(keyringType, tokenName)
		assert.NoError(t, r.AddStoredToken(ctx, addr, tokenName, keyringType, at.Id))
		p, err := r.lookupStoredAuthToken(ctx, addr, tokenName, keyringType)
		require.NoError(t, err)
		require.NotNil(t, p)

		assert.NoError(t, r.deleteStoredToken(ctx, p))

		got, err := r.LookupStoredAuthTokenId(ctx, addr, tokenName, keyringType)
		require.NoError(t, err)
		require.Empty(t, got)
	})
}

func TestRepository_LookupStoredAuthTokenId(t *testing.T) {
	ctx := context.Background()
	s, err := Open(ctx)
	require.NoError(t, err)

	r, err := NewRepository(ctx, s, testAuthTokenLookup)
	require.NoError(t, err)

	t.Run("empty address", func(t *testing.T) {
		p, err := r.LookupStoredAuthTokenId(ctx, "", "token", "keyring")
		assert.ErrorContains(t, err, "address is empty")
		assert.Empty(t, p)
	})
	t.Run("empty token name", func(t *testing.T) {
		p, err := r.LookupStoredAuthTokenId(ctx, "address", "", "keyring")
		assert.ErrorContains(t, err, "token name is empty")
		assert.Empty(t, p)
	})
	t.Run("empty keyring type", func(t *testing.T) {
		p, err := r.LookupStoredAuthTokenId(ctx, "address", "token", "")
		assert.ErrorContains(t, err, "keyring type is empty")
		assert.Empty(t, p)
	})
	t.Run("not found", func(t *testing.T) {
		p, err := r.LookupStoredAuthTokenId(ctx, "address", "token", "keyring")
		assert.NoError(t, err)
		assert.Empty(t, p)
	})
	t.Run("found", func(t *testing.T) {
		addr := "address"
		keyringType := "keyring"
		tokenName := "token"
		at := testAuthTokenLookup(keyringType, tokenName)

		assert.NoError(t, r.AddStoredToken(ctx, addr, tokenName, keyringType, at.Id))
		p, err := r.lookupStoredAuthToken(ctx, addr, tokenName, keyringType)
		assert.NoError(t, err)
		assert.NotEmpty(t, p)
	})
}

func TestRepository_RemoveStalePersonas(t *testing.T) {
	ctx := context.Background()
	s, err := Open(ctx)
	require.NoError(t, err)

	r, err := NewRepository(ctx, s, testAuthTokenLookup)
	require.NoError(t, err)

	staleTime := time.Now().Add(-(personaStalenessLimit + 1*time.Hour))
	oldNotStaleTime := time.Now().Add(-(personaStalenessLimit - 1*time.Hour))

	addr := "address"
	keyringType := "keyring"
	tokenName := "token"
	at := testAuthTokenLookup(keyringType, tokenName)
	for i := 0; i < personaLimit; i++ {
		bAddr := fmt.Sprintf("%s%d", addr, i)
		assert.NoError(t, r.AddStoredToken(ctx, bAddr, tokenName, keyringType, at.Id))
		st := &storedToken{
			KeyringType:  keyringType,
			TokenName:    tokenName,
			BoundaryAddr: bAddr,
			UserId:       at.UserId,
			AuthTokenId:  at.Id,
		}
		switch i % 3 {
		case 0:
			st.LastAccessedTime = staleTime
			_, err := r.rw.Update(ctx, st, []string{"LastAccessedTime"}, nil)
			require.NoError(t, err)
		case 1:
			st.LastAccessedTime = oldNotStaleTime
			_, err := r.rw.Update(ctx, st, []string{"LastAccessedTime"}, nil)
			require.NoError(t, err)
		}
	}

	assert.NoError(t, r.removeStaleStoredTokens(ctx))
	lp, err := r.listStoredTokens(ctx)
	assert.NoError(t, err)
	assert.Len(t, lp, personaLimit*2/3)
}
