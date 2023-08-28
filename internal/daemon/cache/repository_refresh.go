package cache

import (
	"context"
	stderrors "errors"
	"fmt"

	"github.com/hashicorp/boundary/api"
	"github.com/hashicorp/boundary/api/targets"
	"github.com/hashicorp/boundary/internal/errors"
	"github.com/hashicorp/boundary/internal/observability/event"
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

	// TODO: Only refresh once per user instead of once per token
	var retErr error
	for _, st := range storedTokens {
		at := r.tokenLookupFn(st.KeyringType, st.TokenName)
		if at == nil {
			if err := r.deleteStoredToken(ctx, st); err != nil {
				retErr = stderrors.Join(retErr, err)
			}
			continue
		}
		u := &User{UserId: st.UserId, BoundaryAddr: st.BoundaryAddr}
		t, err := opts.withTargetRetrievalFunc(ctx, st.BoundaryAddr, at.Token)
		if err != nil {
			retErr = stderrors.Join(retErr, errors.Wrap(ctx, err, op, errors.WithMsg("for user %v", u)))
			continue
		}
		event.WriteSysEvent(ctx, op, fmt.Sprintf("updating %d targets for user %v", len(t), u))
		if err := r.refreshTargets(ctx, u, t); err != nil {
			retErr = stderrors.Join(retErr, errors.Wrap(ctx, err, op, errors.WithMsg("for user %v", u)))
			continue
		}
	}
	return retErr
}
