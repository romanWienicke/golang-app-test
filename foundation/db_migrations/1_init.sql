-- +goose Up

create table users (
    id serial primary key,
    name varchar(100) not null,
    email varchar(100) unique not null,
    created_at timestamp with time zone default current_timestamp
);

-- +goose Down
drop table if exists users;