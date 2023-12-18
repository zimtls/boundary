-- Copyright (c) HashiCorp, Inc.
-- SPDX-License-Identifier: BUSL-1.1

begin;
  select plan(2);

  select has_index('session', 'session_create_time_list_idx', array['create_time', 'public_id', 'project_id', 'user_id', 'termination_reason']);
  select has_index('session', 'session_update_time_list_idx', array['update_time', 'public_id', 'project_id', 'user_id', 'termination_reason']);

  select * from finish();

rollback;