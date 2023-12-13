// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package apptoken

import (
	"context"
	"testing"

	"github.com/hashicorp/boundary/internal/db"
	"github.com/hashicorp/boundary/internal/errors"
	"github.com/hashicorp/boundary/internal/iam"
	"github.com/hashicorp/boundary/internal/kms"
	"github.com/stretchr/testify/require"
)

func TestRepository_DeleteAppToken(t *testing.T) {
	testCtx := context.Background()
	testConn, _ := db.TestSetup(t, "postgres")
	testRw := db.New(testConn)
	testWrapper := db.TestWrapper(t)
	testKms := kms.TestKms(t, testConn, testWrapper)
	testIamRepo := iam.TestRepo(t, testConn, testWrapper)
	testOrg, _ := iam.TestScopes(t, testIamRepo)
	testRepo, err := NewRepository(testCtx, testRw, testRw, testKms, testIamRepo)
	testUser := iam.TestUser(t, testIamRepo, testOrg.GetPublicId())
	require.NoError(t, err)

	testUserHistoryId, err := testRepo.ResolveUserHistoryId(testCtx, testUser.GetPublicId())
	require.NoError(t, err)

	tests := []struct {
		name            string
		reader          db.Reader
		writer          db.Writer
		kms             kms.GetWrapperer
		publicId        string
		createdBy       string
		grantsStr       string
		wantRowsDeleted int
		wantErrMatch    *errors.Template
		wantErrContains string
	}{
		{
			name:            "valid",
			reader:          testRw,
			writer:          testRw,
			kms:             testKms,
			publicId:        testOrg.PublicId,
			createdBy:       testUserHistoryId,
			grantsStr:       "id=*;type=*;actions=*",
			wantRowsDeleted: 1,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			at, atg := TestAppToken(t, testConn, tc.publicId, tc.createdBy, tc.grantsStr)
		})
	}
}
