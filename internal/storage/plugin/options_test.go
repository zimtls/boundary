// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package plugin

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/structpb"
)

func Test_GetOpts(t *testing.T) {
	t.Parallel()
	t.Run("WithChunkSize", func(t *testing.T) {
		opts := getOpts(WithChunkSize(1024))
		testOpts := getDefaultOptions()
		testOpts.withChunkSize = 1024
		assert.Equal(t, opts, testOpts)
	})
	t.Run("WithAttributes", func(t *testing.T) {
		opts := getOpts(WithAttributes(&structpb.Struct{Fields: map[string]*structpb.Value{"foo": structpb.NewStringValue("bar")}}))
		testOpts := getDefaultOptions()
		testOpts.withAttributes = &structpb.Struct{Fields: map[string]*structpb.Value{"foo": structpb.NewStringValue("bar")}}
		assert.Equal(t, opts, testOpts)
	})
	t.Run("WithName", func(t *testing.T) {
		opts := getOpts(WithName("test"))
		testOpts := getDefaultOptions()
		testOpts.withName = "test"
		assert.Equal(t, opts, testOpts)
	})
	t.Run("WithDescription", func(t *testing.T) {
		opts := getOpts(WithDescription("test desc"))
		testOpts := getDefaultOptions()
		testOpts.withDescription = "test desc"
		assert.Equal(t, opts, testOpts)
	})
	t.Run("WithSecrets", func(t *testing.T) {
		opts := getOpts(WithSecrets(&structpb.Struct{Fields: map[string]*structpb.Value{"foo": structpb.NewStringValue("bar")}}))
		testOpts := getDefaultOptions()
		testOpts.withSecrets = &structpb.Struct{Fields: map[string]*structpb.Value{"foo": structpb.NewStringValue("bar")}}
		assert.Equal(t, opts, testOpts)
	})
	t.Run("WithBucketPrefix", func(t *testing.T) {
		opts := getOpts(WithBucketPrefix("test prefix"))
		testOpts := getDefaultOptions()
		testOpts.withBucketPrefix = "test prefix"
		assert.Equal(t, opts, testOpts)
	})
	t.Run("WithWorkerFilter", func(t *testing.T) {
		opts := getOpts(WithWorkerFilter("test filter"))
		testOpts := getDefaultOptions()
		testOpts.withWorkerFilter = "test filter"
		assert.Equal(t, opts, testOpts)
	})
	t.Run("WithLimit", func(t *testing.T) {
		opts := getOpts(WithLimit(12345))
		testOpts := getDefaultOptions()
		testOpts.withLimit = 12345
		assert.Equal(t, opts, testOpts)
	})
}
