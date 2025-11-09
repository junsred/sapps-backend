create database face;

create table users
(
    id                      text not null
        constraint users_pk
            primary key,
    firebase_token          text
        constraint users_pk_2
            unique,
    last_login              timestamp,
    last_online             timestamp,
    registered_at           timestamp default now(),
    session                 uuid      default gen_random_uuid(),
    premium_id              text
        constraint users_pk_3
            unique,
    last_token              text,
    device_id               text,
    firebase_id             text
        constraint users_pk_5
            unique,
    coin                    integer,
    coin_reset_date         timestamp,
    language                text,
    debug                   boolean,
    ip_address              text,
    country                 text,
    special_offer_deadline  timestamp,
    register_not            boolean,
    build_number            integer,
    store                   text,
    notification_permission boolean,
    timezone                text,
    device_info             jsonb
);


create table premium_data
(
    id           text not null
        constraint premium_data_pk
            primary key,
    premium_type text,
    created_date timestamp default now(),
    expire_date  timestamp
);


create table costs
(
    id           text      default gen_random_uuid() not null
        constraint costs_pk
            primary key,
    user_id      text,
    created_date timestamp default now(),
    reason       text,
    price        double precision,
    task_id      text,
    ip_address   text
);

create index costs_user_id_index
    on costs (user_id);

create index costs_task_id_index
    on costs (task_id);

create table revenuecat_logs
(
    revenuecat_event_id         text not null
        constraint pk_revenuecat_logs
            primary key,
    app_user_id                 text,
    original_app_user_id        text,
    product_id                  text,
    price                       double precision,
    currency                    text,
    price_in_purchased_currency double precision,
    takehome_percentage         double precision,
    purchased_at_ms             bigint,
    expiration_at_ms            bigint,
    store                       text,
    environment                 text,
    transaction_id              text,
    original_transaction_id     text,
    other_data                  json,
    event_type                  text,
    current_user_info           json,
    event_timestamp_ms          bigint generated always as (((other_data ->> 'event_timestamp_ms'::text))::bigint) stored,
    created_at                  timestamp default now(),
    user_id                     text
);
