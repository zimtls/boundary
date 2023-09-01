package cache

import (
	"context"
	stderrors "errors"
	"fmt"

	"github.com/hashicorp/boundary/api"
	"github.com/hashicorp/boundary/api/targets"
	"github.com/hashicorp/boundary/internal/errors"
	"github.com/hashicorp/boundary/internal/observability/event"
	"golang.org/x/exp/slices"
)

type targetRetrievalFunc func(ctx context.Context, keyringstring, tokenName string) ([]*targets.Target, error)

func defaultTargetFunc(ctx context.Context, addr string, token string) ([]*targets.Target, error) {
	const op = "cache.defaultTargetFunc"
	client, err := api.NewClient(&api.Config{
		Addr:  addr,
		Token: token,
	})
	if err != nil {
		return nil, errors.Wrap(ctx, err, op)
	}
	tarClient := targets.NewClient(client)
	l, err := tarClient.List(ctx, "global", targets.WithRecursive(true))
	if err != nil {
		return nil, errors.Wrap(ctx, err, op)
	}
	return l.Items, nil
}

func (r *Repository) Refresh(ctx context.Context, opt ...Option) error {
	const op = "cache.(Repository).Refresh"
	if err := r.removeStaleStoredTokens(ctx); err != nil {
		return errors.Wrap(ctx, err, op)
	}

	opts, err := getOpts(opt...)
	if err != nil {
		return errors.Wrap(ctx, err, op)
	}
	if opts.withTargetRetrievalFunc == nil {
		opts.withTargetRetrievalFunc = defaultTargetFunc
	}

	storedTokens, err := r.listStoredTokens(ctx)
	if err != nil {
		return errors.Wrap(ctx, err, op)
	}

	// build the users from the retrieved stored tokens
	users := make(map[User][]*storedToken)
	for _, st := range storedTokens {
		key := User{UserId: st.UserId, BoundaryAddr: st.BoundaryAddr}
		users[key] = append(users[key], st)
	}

	var retErr error
	for u, tokens := range users {
		// sort the tokens so the most recently accessed is first
		slices.SortFunc(tokens, func(a, b *storedToken) int {
			return a.LastAccessedTime.Compare(b.LastAccessedTime)
		})
		slices.Reverse(tokens)

		var lookupErr error
		var retrievedTargets []*targets.Target
		for _, t := range tokens {
			lookupErr = nil

			at := r.tokenLookupFn(t.KeyringType, t.TokenName)
			if at == nil {
				if err := r.deleteStoredToken(ctx, t); err != nil {
					lookupErr = err
				}
				continue
			}

			tars, err := opts.withTargetRetrievalFunc(ctx, u.BoundaryAddr, at.Token)
			if err != nil {
				lookupErr = errors.Wrap(ctx, err, op, errors.WithMsg("for user %v using stored token %#v", u, t))
				continue
			}
			retrievedTargets = tars
		}
		retErr = stderrors.Join(retErr, lookupErr)

		event.WriteSysEvent(ctx, op, fmt.Sprintf("updating %d targets for user %v", len(retrievedTargets), u))
		if err := r.refreshTargets(ctx, &u, retrievedTargets); err != nil {
			retErr = stderrors.Join(retErr, errors.Wrap(ctx, err, op, errors.WithMsg("for user %v", u)))
			continue
		}
	}
	return retErr
}
