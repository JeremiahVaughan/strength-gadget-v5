drop table exercise_measurement_type;


create table last_completed_measurement
(
    user_id     uuid not null,
    exercise_id uuid not null
        constraint lcm_exercise_measurement_type_exercise_id_fk
            references exercise (id),
    measurement integer,
    constraint last_completed_measurement_pk
        primary key (user_id, exercise_id)
);