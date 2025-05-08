create table public.last_completed_measurement
(
    user_id     bigint   not null,
    exercise_id smallint not null,
    measurement smallint not null,
    constraint last_completed_measurement_pk
        primary key (user_id, exercise_id)
);

