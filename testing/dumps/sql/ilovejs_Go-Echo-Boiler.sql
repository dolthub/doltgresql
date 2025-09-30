-- Downloaded from: https://github.com/ilovejs/Go-Echo-Boiler/blob/f50ee58be50f7d14f2831e6ed07728175835d31a/tables_schema.sql
/*
-- Note: No mysqldump or pg_dump equivalent for Microsoft SQL Server,
-- so generated tests must be supplemented by tables_schema.sql with CREATE TABLE ... queries
-- Don't run in test environment


USE [master]
DROP DATABASE test2;
create database test2;
ALTER AUTHORIZATION ON DATABASE::test2 TO tester;
use test2;
*/

create table basic_trades
(
    id         int identity,
    name       varchar(200)       not null,
    is_active  bit      default 1 not null,
    is_deleted bit      default 0 not null,
    created    datetime,
    updated    datetime default getdate(),
    primary key (id)
)
go

-- create unique index basic_trades_name_uindex
--     on basic_trades (name)
-- go

alter table trade_categories
    add constraint DF__basic_tra__is_ac__48CFD27E default 1 for is_active
go

alter table trade_categories
    add constraint DF__basic_tra__is_de__49C3F6B7 default 0 for is_deleted
go

alter table trade_categories
    add constraint DF__basic_tra__updat__4AB81AF0 default getdate() for updated
go

create table roles
(
    id             int identity,
    user_role_type varchar(50) not null,
    created        datetime    not null,
    updated        datetime default getdate(),
    primary key (id)
)
go

alter table roles
    add constraint DF__roles__updated__37A5467C default getdate() for updated
go

create table users
(
    id              int identity,
    user_role_id    int                not null,
    username        varchar(200)       not null,
    password        varchar(256)       not null,
    email           varchar(256)       not null,
    password_token  varchar(200),
    token_expires   datetime,
    deletion_reason varchar(200),
    is_active       bit      default 1 not null,
    is_deleted      bit      default 0 not null,
    created         datetime,
    updated         datetime default getdate(),
    primary key (id),
--     constraint users_roles_id_fk
--         foreign key (user_role_id)
--             references roles
)
go

create table profiles
(
    id         int identity,
    user_id    int                not null,
    image_url  varchar(200),
    mobile     varchar(50),
    company    varchar(200),
    first_name varchar(200),
    last_name  varchar(200),
    is_active  bit      default 1 not null,
    is_deleted bit      default 0 not null,
    created    datetime,
    updated    datetime default getdate(),
    primary key (id),
--     constraint profiles_users_id_fk
--         foreign key (user_id)
--             references users
)
go

alter table profiles
    add constraint DF__profiles__is_act__3F466844 default 1 for is_active
go

alter table profiles
    add constraint DF__profiles__is_del__403A8C7D default 0 for is_deleted
go

alter table profiles
    add constraint DF__profiles__update__412EB0B6 default getdate() for updated
go

create table projects
(
    id                     int identity,
    manager_id             int                not null,
    creator_id             int                not null,
    name                   varchar(200),
    total_item_breakdown   float,
    contractor_total_claim float,
    serial_no              varchar(50),
    address                varchar(200),
    total_contract_value   float,
    quantity_surveyor      varchar(200),
    notes                  varchar(200),
    is_active              bit      default 1 not null,
    is_deleted             bit      default 0 not null,
    created                datetime,
    updated                datetime default getdate(),
    primary key (id)
--     constraint fk_project_creator_id
--         foreign key (creator_id)
--             references users,
--     constraint fk_project_manager_id
--         foreign key (manager_id)
--             references users
)
go

alter table projects
    add constraint DF__projects__is_act__440B1D61 default 1 for is_active
go

alter table projects
    add constraint DF__projects__is_del__44FF419A default 0 for is_deleted
go

alter table projects
    add constraint DF__projects__update__45F365D3 default getdate() for updated
go

create table trades
(
    id             int identity,
    basic_trade_id int                not null,
    surveyor_id    int                not null,
    project_id     int                not null,
    floor_level    varchar(20)        not null,
    work_desc      varchar(200),
    item_breakdown float,
    tempcheck      bit,
    is_active      bit      default 1 not null,
    is_deleted     bit      default 0 not null,
    created        datetime,
    updated        datetime default getdate(),
    primary key (id),
--     constraint trades_basic_trades_id_fk
--         foreign key (basic_trade_id)
--             references basic_trades,
--     constraint trades_projects_id_fk
--         foreign key (project_id)
--             references projects,
--     constraint trades_users_id_fk
--         foreign key (surveyor_id)
--             references users
)
go

create table claims
(
    id                 int identity,
    trade_id           int                not null,
    user_id            int                not null,
    basic_trade_id     int                not null,
    total_amount       float,
    claimed_amount     float,
    previous_claimed   float,
    amount_due         float,
    cost_to_completed  float,
    claim_number       varchar(200),
    claim_period       varchar(50),
    action_claim       bit,
    old_claimed_amount float,
    claim_percentage   float,
    is_active          bit      default 1 not null,
    is_deleted         bit      default 0 not null,
    created            datetime,
    updated            datetime default getdate(),
    primary key (id),
--     constraint claims_basic_trades_id_fk
--         foreign key (basic_trade_id)
--             references basic_trades,
--     constraint claims_trades_id_fk
--         foreign key (trade_id)
--             references trades,
--     constraint claims_users_id_fk
--         foreign key (user_id)
--             references users
)
go

create table claim_histories
(
    id               int identity,
    trade_id         int                not null,
    claim_id         int                not null,
    profile_id       int                not null,
    previous_claimed float,
    is_active        bit      default 1 not null,
    is_deleted       bit      default 0 not null,
    created          datetime,
    updated          datetime default getdate(),
    primary key (id),
--     constraint claim_histories_claims_id_fk
--         foreign key (claim_id)
--             references claims,
--     constraint claim_histories_trades_id_fk
--         foreign key (trade_id)
--             references trades
)
go

alter table claim_histories
    add constraint DF__claim_his__is_ac__571DF1D5 default 1 for is_active
go

alter table claim_histories
    add constraint DF__claim_his__is_de__5812160E default 0 for is_deleted
go

alter table claim_histories
    add constraint DF__claim_his__updat__59063A47 default getdate() for updated
go

alter table claims
    add constraint DF__claims__is_activ__52593CB8 default 1 for is_active
go

alter table claims
    add constraint DF__claims__is_delet__534D60F1 default 0 for is_deleted
go

alter table claims
    add constraint DF__claims__updated__5441852A default getdate() for updated
go

alter table trades
    add constraint DF__trades__is_activ__4D94879B default 1 for editable
go

alter table trades
    add constraint DF__trades__is_delet__4E88ABD4 default 0 for is_deleted
go

alter table trades
    add constraint DF__trades__updated__4F7CD00D default getdate() for updated
go

-- create unique index users_email_uindex
--     on users (email)
-- go

alter table users
    add constraint DF__users__is_active__3A81B327 default 1 for is_active
go

alter table users
    add constraint DF__users__is_delete__3B75D760 default 0 for is_deleted
go

alter table users
    add constraint DF__users__updated__3C69FB99 default getdate() for updated
go



