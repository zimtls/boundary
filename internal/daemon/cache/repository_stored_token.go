package cache

import (
	"context"
	"database/sql"
	stderrors "errors"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/boundary/internal/db"
	"github.com/hashicorp/boundary/internal/errors"
)

// listUsers returns a list of users known by the cache.
func (r *Repository) listUsers(ctx context.Context) ([]*User, error) {
	const op = "cache.(Repository).listUsers"
	var ret []*User
	if err := r.rw.SearchWhere(ctx, &ret, "true", nil); err != nil {
		return nil, errors.Wrap(ctx, err, op)
	}
	return ret, nil
}

// AddStoredToken adds a stored token to the repository.  If the token in the
// keyring doesn't match the id provided an error is returned.  If the number of
// stored tokens now exceed a limit, the stored token retrieved least recently is deleted.
func (r *Repository) AddStoredToken(ctx context.Context, bAddr, tokenName, keyringType, authTokId string) error {
	const op = "cache.(Repository).AddStoredToken"
	switch {
	case tokenName == "":
		return errors.New(ctx, errors.InvalidParameter, op, "token name is empty")
	case keyringType == "":
		return errors.New(ctx, errors.InvalidParameter, op, "keyring type is empty")
	case bAddr == "":
		return errors.New(ctx, errors.InvalidParameter, op, "boundary address is empty")
	}

	at := r.tokenLookupFn(keyringType, tokenName)
	if at == nil {
		return errors.New(ctx, errors.InvalidParameter, op, "unable to find token in the keyring specified")
	}
	if authTokId != at.Id {
		return errors.New(ctx, errors.InvalidParameter, op, "provided auth token id doesn't match the one stored")
	}
	// Even though the auth token is already stored, we still call create so
	// the last accessed timestamps can get updated since calling this method
	// indicates that the token was used and is still valid.
	_, err := r.rw.DoTx(ctx, db.StdRetryCnt, db.ExpBackoff{}, func(reader db.Reader, writer db.Writer) error {
		// A stored token must reference a user in storage so ensure one exists
		// before creating the stored token
		{
			u := &User{
				BoundaryAddr: bAddr,
				UserId:       at.UserId,
			}
			onConflict := &db.OnConflict{
				Target: db.Columns{"boundary_addr", "user_id"},
				Action: db.UpdateAll(true),
			}
			if err := writer.Create(ctx, u, db.WithOnConflict(onConflict)); err != nil {
				return errors.Wrap(ctx, err, op)
			}
		}

		{
			st := &storedToken{
				KeyringType:      keyringType,
				TokenName:        tokenName,
				BoundaryAddr:     bAddr,
				UserId:           at.UserId,
				AuthTokenId:      at.Id,
				LastAccessedTime: time.Now(),
			}
			onConflict := &db.OnConflict{
				Target: db.Columns{"keyring_type", "token_name"},
				Action: db.SetColumns([]string{"auth_token_id", "boundary_addr", "user_id", "last_accessed_time"}),
			}
			if err := writer.Create(ctx, st, db.WithOnConflict(onConflict)); err != nil {
				return errors.Wrap(ctx, err, op)
			}
		}

		var tokens []*storedToken
		if err := reader.SearchWhere(ctx, &tokens, "", []any{}, db.WithLimit(-1)); err != nil {
			return errors.Wrap(ctx, err, op)
		}
		if len(tokens) <= personaLimit {
			return nil
		}

		var oldestToken *storedToken
		for _, p := range tokens {
			if oldestToken == nil || oldestToken.LastAccessedTime.After(p.LastAccessedTime) {
				oldestToken = p
			}
		}
		if oldestToken != nil {
			if _, err := writer.Delete(ctx, oldestToken); err != nil {
				return errors.Wrap(ctx, err, op)
			}
		}
		return nil
	})
	if err != nil {
		return err
	}

	// Since a token may have been deleted, or an address changed,
	// clean all orphaned users.
	if err := r.cleanupOrphanedUsers(ctx); err != nil {
		return errors.Wrap(ctx, err, op)
	}

	return nil
}

// cleanupOrphanedUsers deletes users from the cache that no longer have a
// known token stored in the cache.
func (r *Repository) cleanupOrphanedUsers(ctx context.Context) error {
	const op = "cache.(Repository).cleanupOrphanedUsers"
	st, err := r.listStoredTokens(ctx)
	if err != nil {
		return errors.Wrap(ctx, err, op)
	}
	var userQueryParts []string
	var userParams []any
	for _, t := range st {
		userQueryParts = append(userQueryParts, "(user_id = ? and boundary_addr = ?)")
		userParams = append(userParams, t.UserId, t.BoundaryAddr)
	}
	whereClause := strings.Join(userQueryParts, " or ")
	if len(userParams) == 0 {
		// There are no more tokens, so delete all users.
		whereClause = "false"
	}
	_, err = r.rw.Exec(ctx, fmt.Sprintf("delete from cache_user where not (%s)", whereClause), userParams)
	if err != nil {
		return errors.Wrap(ctx, err, op)
	}
	return nil
}

// LookupStoredAuthTokenId returns the auth token id in the cache if one exists.
// Accepts withUpdateLastAccessedTime option.
func (r *Repository) LookupStoredAuthTokenId(ctx context.Context, addr, tokenName, keyringType string, opt ...Option) (string, error) {
	const op = "cache.(Repository).LookupStoredAuthTokenId"
	switch {
	case addr == "":
		return "", errors.New(ctx, errors.InvalidParameter, op, "address is empty")
	case keyringType == "":
		return "", errors.New(ctx, errors.InvalidParameter, op, "keyring type is empty")
	case tokenName == "":
		return "", errors.New(ctx, errors.InvalidParameter, op, "token name is empty")
	}
	p, err := r.lookupStoredAuthToken(ctx, addr, tokenName, keyringType, opt...)
	if err != nil {
		return "", errors.Wrap(ctx, err, op)
	}
	if p == nil {
		return "", nil
	}

	return p.AuthTokenId, nil
}

func (r *Repository) lookupStoredAuthToken(ctx context.Context, addr, tokenName, keyringType string, opt ...Option) (*storedToken, error) {
	const op = "cache.(Repository).lookupStoredAuthToken"
	switch {
	case addr == "":
		return nil, errors.New(ctx, errors.InvalidParameter, op, "address is empty")
	case keyringType == "":
		return nil, errors.New(ctx, errors.InvalidParameter, op, "keyring type is empty")
	case tokenName == "":
		return nil, errors.New(ctx, errors.InvalidParameter, op, "token name is empty")
	}
	opts, err := getOpts(opt...)
	if err != nil {
		return nil, errors.Wrap(ctx, err, op)
	}

	p := &storedToken{
		KeyringType: keyringType,
		TokenName:   tokenName,
	}
	if err := r.rw.LookupById(ctx, p); err != nil {
		if errors.IsNotFoundError(err) {
			return nil, nil
		}
		return nil, errors.Wrap(ctx, err, op)
	}
	if p.BoundaryAddr != addr {
		// If we found a stored token that doesn't have the provided address it
		// is not the correct one, so the return should indicate the looked up
		// stored token could not be found.
		return nil, nil
	}

	if opts.withUpdateLastAccessedTime {
		updatedP := &storedToken{
			BoundaryAddr:     p.BoundaryAddr,
			TokenName:        p.TokenName,
			KeyringType:      p.KeyringType,
			LastAccessedTime: time.Now(),
		}
		if _, err := r.rw.Update(ctx, updatedP, []string{"LastAccessedTime"}, nil); err != nil {
			return nil, errors.Wrap(ctx, err, op)
		}
	}
	return p, nil
}

// deleteStoredToken deletes a stored token
func (r *Repository) deleteStoredToken(ctx context.Context, s *storedToken) (retErr error) {
	const op = "cache.(Repository).deleteStoredToken"
	switch {
	case s == nil:
		return errors.New(ctx, errors.InvalidParameter, op, "missing persona")
	case s.TokenName == "":
		return errors.New(ctx, errors.InvalidParameter, op, "missing token name")
	case s.KeyringType == "":
		return errors.New(ctx, errors.InvalidParameter, op, "missing keyring type")
	}

	n, err := r.rw.Delete(ctx, s)
	if err != nil {
		return errors.Wrap(ctx, err, op)
	}

	defer func() {
		if err := r.cleanupOrphanedUsers(ctx); err != nil {
			retErr = stderrors.Join(retErr, errors.Wrap(ctx, err, op))
		}
	}()

	switch n {
	case 1:
		return nil
	case 0:
		return errors.New(ctx, errors.RecordNotFound, op, "stored token not found when attempting deletion")
	default:
		return errors.New(ctx, errors.MultipleRecords, op, "multiple stored tokens deleted when one was requested")
	}
}

// removeStaleStoredTokens removes all personas which are older than the staleness
func (r *Repository) removeStaleStoredTokens(ctx context.Context, opt ...Option) error {
	const op = "cache.(Repository).removeStaleStoredTokens"
	if _, err := r.rw.Exec(ctx, "delete from cache_stored_token where last_accessed_time < @last_accessed_time",
		[]any{sql.Named("last_accessed_time", time.Now().Add(-personaStalenessLimit))}); err != nil {
		return errors.Wrap(ctx, err, op)
	}
	if err := r.cleanupOrphanedUsers(ctx); err != nil {
		return errors.Wrap(ctx, err, op)
	}
	return nil
}

// listStoredTokens returns all known stored tokens in the cache
func (r *Repository) listStoredTokens(ctx context.Context) ([]*storedToken, error) {
	const op = "cache.(Repository).listStoredTokens"
	var ret []*storedToken
	if err := r.rw.SearchWhere(ctx, &ret, "true", nil); err != nil {
		return nil, errors.Wrap(ctx, err, op)
	}
	return ret, nil
}

type storedToken struct {
	KeyringType      string `gorm:"primaryKey"`
	TokenName        string `gorm:"primaryKey"`
	BoundaryAddr     string
	UserId           string
	AuthTokenId      string
	LastAccessedTime time.Time `gorm:"default:(strftime('%Y-%m-%d %H:%M:%f','now'))"`
}

func (*storedToken) TableName() string {
	return "cache_stored_token"
}

type User struct {
	BoundaryAddr string `gorm:"primaryKey"`
	UserId       string `gorm:"primaryKey"`
}

func (*User) TableName() string {
	return "cache_user"
}

func (u *User) clone() *User {
	return &User{
		BoundaryAddr: u.BoundaryAddr,
		UserId:       u.UserId,
	}
}
