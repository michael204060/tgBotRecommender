create table if not exists dialogs(
    index serial primary key,
    content varchar(500),
    sender int not null
);