// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package listtoken

import (
	"github.com/hashicorp/boundary/internal/db/timestamp"
	"github.com/hashicorp/boundary/internal/types/resource"
)

// Item defines a subset of a boundary.Resource that can
// be used as an input to a DB operation for the purposes
// of pagination and sorting.
type Item interface {
	GetPublicId() string
	GetCreateTime() *timestamp.Timestamp
	GetUpdateTime() *timestamp.Timestamp
	GetResourceType() resource.Type
}

// Item represents a generic resource with a public ID,
// create time, update time and resource type.
type item struct {
	publicId     string
	createTime   *timestamp.Timestamp
	updateTime   *timestamp.Timestamp
	resourceType resource.Type
}

// GetPublicId gets the public ID of the item.
func (p *item) GetPublicId() string {
	return p.publicId
}

// GetCreateTime gets the create time of the item.
func (p *item) GetCreateTime() *timestamp.Timestamp {
	return p.createTime
}

// GetUpdateTime gets the update time of the item.
func (p *item) GetUpdateTime() *timestamp.Timestamp {
	return p.updateTime
}

// GetResourceType gets the resource type of the item.
func (p *item) GetResourceType() resource.Type {
	return p.resourceType
}
