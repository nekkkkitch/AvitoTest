create schema if not exists public;

alter schema public owner to pg_database_owner;

SET search_path TO public;

create table users if not exists(
    username text primary key,
    password bytea[] not null,
    balance int not null check(balance >= 0)
);

create index username_index on public.users(username);

create table merch if not exists(
    title text primary key,
    price int not null
);

create table inventory if not exists(
    username text references public.users(username),
    item text references public.merch(title),
    amount int
);

create index user_inventory_index on public.inventory(username);

create table history if not exists(
    sender text references public.users(username),
    reciever text references public.users(username),
    amount int
);

create index sender_index on public.history(sender);
create index reciever_index on public.history(sender);
