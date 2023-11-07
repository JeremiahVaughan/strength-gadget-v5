create table exercise_type
(
    id   uuid not null
        constraint exercise_type_pk
            primary key,
    name text not null
);

alter table exercise
    add exercise_type_id uuid;

alter table exercise
    add constraint exercise_exercise_type_id_fk
        foreign key (exercise_type_id) references exercise_type;

INSERT INTO exercise_type (id, name)
VALUES ('982d0b18-a67c-401a-95f2-ddb702ba80b5', 'cardio');

INSERT INTO exercise_type (id, name)
VALUES ('8ffe7196-4e3d-4439-ae19-3159ad5387bd', 'calisthenics');

INSERT INTO exercise_type (id, name)
VALUES ('6bdb3624-bed1-41a9-bf8c-7b1066411446', 'weightlifting');

