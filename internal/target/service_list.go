// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package target

import (
	"context"
	"runtime/trace"
	"time"

	"github.com/hashicorp/boundary/internal/errors"
	"github.com/hashicorp/boundary/internal/pagination"
)

// List lists up to page size targets, filtering out entries that
// do not pass the filter item function. It will automatically request
// more targets from the database, at page size chunks, to fill the page.
// It returns a new list token used to continue pagination or refresh items.
// Targets are ordered by create time descending (most recently created first).
func List(
	ctx context.Context,
	grantsHash []byte,
	pageSize int,
	filterItemFn pagination.ListFilterFunc[Target],
	repo *Repository,
) (*pagination.ListResponse[Target], error) {
	defer trace.StartRegion(ctx, "targets.List").End()
	const op = "target.List"

	if len(grantsHash) == 0 {
		return nil, errors.New(ctx, errors.InvalidParameter, op, "missing grants hash")
	}
	if pageSize < 1 {
		return nil, errors.New(ctx, errors.InvalidParameter, op, "page size must be at least 1")
	}
	if filterItemFn == nil {
		return nil, errors.New(ctx, errors.InvalidParameter, op, "missing filter item callback")
	}
	if repo == nil {
		return nil, errors.New(ctx, errors.InvalidParameter, op, "missing repo")
	}

	listItemsFn := func(ctx context.Context, lastPageItem Target, limit int) ([]Target, time.Time, error) {
		defer trace.StartRegion(ctx, "List.listItemsFn").End()
		opts := []Option{
			WithLimit(limit),
		}
		if lastPageItem != nil {
			opts = append(opts, WithStartPageAfterItem(lastPageItem))
		}
		return repo.listTargets(ctx, opts...)
	}

	return pagination.List(ctx, grantsHash, pageSize, filterItemFn, listItemsFn, repo.estimatedCount)
}
