begin;

create table if not exists cache_user (
	boundary_addr text not null,
	user_id text not null,
	last_accessed_time timestamp not null default (strftime('%Y-%m-%d %H:%M:%f','now')),
    primary key (boundary_addr, user_id)
);

create table if not exists cache_stored_token (
  keyring_type text not null,
  token_name text not null,
  boundary_addr text not null,
  auth_token_id text not null,
  user_id text not null,
  last_accessed_time timestamp not null default (strftime('%Y-%m-%d %H:%M:%f','now')),
  foreign key (boundary_addr, user_id)
	references cache_user(boundary_addr, user_id)
	on delete cascade,
  primary key (keyring_type, token_name)
);

create table if not exists cache_target (
  boundary_addr text not null,
  boundary_user_id text not null,
  id text not null,
  name text,
  description text,
  address text,
  item text,
  foreign key (boundary_addr, boundary_user_id)
	references cache_user(boundary_addr, user_id)
	on delete cascade,
  primary key (boundary_addr, boundary_user_id, id)
);

create table if not exists cache_api_error (
	token_name text not null,
	resource_type text not null,
	error text not null,
	create_time timestamp not null default current_timestamp,
	primary key (token_name, resource_type)
);

commit;