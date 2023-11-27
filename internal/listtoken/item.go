// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package listtoken

import (
	"github.com/hashicorp/boundary/internal/db/timestamp"
	"github.com/hashicorp/boundary/internal/types/resource"
)

// Item represents a generic resource with a public ID,
// create time, update time and resource type.
type Item struct {
	publicId     string
	createTime   *timestamp.Timestamp
	updateTime   *timestamp.Timestamp
	resourceType resource.Type
}

// GetPublicId gets the public ID of the item.
func (p *Item) GetPublicId() string {
	return p.publicId
}

// GetCreateTime gets the create time of the item.
func (p *Item) GetCreateTime() *timestamp.Timestamp {
	return p.createTime
}

// GetUpdateTime gets the update time of the item.
func (p *Item) GetUpdateTime() *timestamp.Timestamp {
	return p.updateTime
}

// GetResourceType gets the resource type of the item.
func (p *Item) GetResourceType() resource.Type {
	return p.resourceType
}
