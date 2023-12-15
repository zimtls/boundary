-- Copyright (c) HashiCorp, Inc.
-- SPDX-License-Identifier: BUSL-1.1

begin;

  -- Add new indexes for the create and update time queries.
  create index session_create_time_list_idx
            on session (create_time desc,
                        public_id,
                        project_id,
                        user_id,
                        termination_reason);
  create index session_update_time_list_idx
            on session (update_time desc,
                        public_id,
                        project_id,
                        user_id,
                        termination_reason);

commit;