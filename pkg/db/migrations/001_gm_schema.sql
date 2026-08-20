if object_id('gm_users', 'U') is null
begin
  create table gm_users (
    username nvarchar(64) not null primary key,
    password_hash nvarchar(255) not null,
    is_active bit not null default 1,
    created_at datetime2 not null default sysutcdatetime()
  );
end;
if object_id('gm_action_requests', 'U') is null
begin
  create table gm_action_requests (
    id bigint identity(1,1) not null primary key,
    idempotency_key nvarchar(128) not null unique,
    gm_username nvarchar(64) not null,
    action_name nvarchar(64) not null,
    target_name nvarchar(128) not null,
    response_json nvarchar(max) null,
    created_at datetime2 not null default sysutcdatetime()
  );
end;
if object_id('gm_user_roles', 'U') is null
begin
  create table gm_user_roles (
    username nvarchar(64) not null,
    role_name nvarchar(32) not null,
    primary key (username, role_name),
    constraint fk_gm_user_roles_user foreign key (username) references gm_users(username)
  );
end;
if object_id('gm_audit_logs', 'U') is null
begin
  create table gm_audit_logs (
    id bigint identity(1,1) not null primary key,
    request_id nvarchar(64) not null,
    gm_username nvarchar(64) not null,
    action_name nvarchar(64) not null,
    target_name nvarchar(128) null,
    reason nvarchar(500) null,
    outcome nvarchar(32) not null,
    details nvarchar(max) null,
    created_at datetime2 not null default sysutcdatetime()
  );
  create index ix_gm_audit_logs_created_at on gm_audit_logs(created_at desc);
end;
