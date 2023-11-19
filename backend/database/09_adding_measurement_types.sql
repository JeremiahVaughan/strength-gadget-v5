create table measurement_type
(
    id   uuid    not null
        constraint measurement_type_pk
            primary key,
    name text not null,
    abbreviated_name text null
);

create table exercise_measurement_type
(
    exercise_id      uuid not null,
    exercise_type_id uuid not null,
    constraint exercise_measurement_type_pk
        primary key (exercise_id, exercise_type_id)
);

create index exercise_measurement_type_exercise_id_index
    on exercise_measurement_type (exercise_id);

create table set_type
(
    id   uuid not null
        constraint set_type_pk
            primary key,
    name text not null
);