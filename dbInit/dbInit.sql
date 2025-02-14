create schema if not exists public;

alter schema public owner to pg_database_owner;

SET search_path TO public;

create table if not exists users(
    username text primary key,
    password bytea not null,
    balance int not null check(balance >= 0)
);

create index username_index on public.users(username);

create table if not exists merch (
    title text primary key,
    price int not null
);

create table if not exists inventory (
    username text references public.users(username),
    item text references public.merch(title),
    amount int
);

create index user_inventory_index on public.inventory(username);

create table if not exists history(
    sender text references public.users(username),
    reciever text references public.users(username),
    amount int
);

create index sender_index on public.history(sender);
create index reciever_index on public.history(sender);

create or replace procedure add_item(
    in in_username text,
    in in_item text
)
language plpgsql
as $$
begin
    if exists(select 1 from public.inventory where username = in_username and item = in_item) then 
		update public.inventory set amount = amount + 1 where username = in_username and item = in_item;
	else 
		insert into public.inventory values(in_username, in_item, 1);
	end if;
end; 
$$;

insert into public.merch values('t-shirt', 80);
insert into public.merch values('cup', 20);
insert into public.merch values('book', 50);
insert into public.merch values('pen', 10);
insert into public.merch values('powerbank', 200);
insert into public.merch values('hoody', 300);
insert into public.merch values('umbrella', 200);
insert into public.merch values('socks', 10);
insert into public.merch values('wallet', 50);
insert into public.merch values('pink-hoody', 500);

insert into public.users values('1', '', 0);
insert into public.users values('2', '', 0);
insert into public.users values('3', '', 0);
insert into public.users values('4', '', 0);
insert into public.users values('5', '', 0);
insert into public.users values('6', '', 0);
insert into public.users values('7', '', 0);
insert into public.users values('8', '', 0);
insert into public.users values('9', '', 0);
insert into public.users values('10', '', 0);