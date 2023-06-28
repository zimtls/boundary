// Code generated by "make api"; DO NOT EDIT.
// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sessionrecordings

import (
	"fmt"

	"github.com/mitchellh/mapstructure"
)

type UsernamePasswordCredentialAttributes struct {
	Username     string `json:"username,omitempty"`
	PasswordHmac string `json:"password_hmac,omitempty"`
}

func AttributesMapToUsernamePasswordCredentialAttributes(in map[string]interface{}) (*UsernamePasswordCredentialAttributes, error) {
	if in == nil {
		return nil, fmt.Errorf("nil input map")
	}
	var out UsernamePasswordCredentialAttributes
	dec, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		Result:  &out,
		TagName: "json",
	})
	if err != nil {
		return nil, fmt.Errorf("error creating mapstructure decoder: %w", err)
	}
	if err := dec.Decode(in); err != nil {
		return nil, fmt.Errorf("error decoding: %w", err)
	}
	return &out, nil
}

func (pt *Credential) GetUsernamePasswordCredentialAttributes() (*UsernamePasswordCredentialAttributes, error) {
	if pt.Type != "username_password" {
		return nil, fmt.Errorf("asked to fetch %s-type attributes but credential is of type %s", "username_password", pt.Type)
	}
	return AttributesMapToUsernamePasswordCredentialAttributes(pt.Attributes)
}
