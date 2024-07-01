create table public.athlete
(
    id              bigserial                          not null
        primary key,
    email           text                               not null
        constraint email__index
            unique,
    password_hash   text                               not null,
    current_routine smallint default 0                 not null
);


create table public.access_attempt
(
    id             bigserial                      not null
        constraint access_attempt_pk
            primary key,
    time           bigint                         not null,
    access_granted boolean                        not null,
    type           smallint                       not null,
    user_id        bigint                         not null
        constraint access_attempt_user_id_fk
            references public.athlete
);


create index access_attempt_type_index
    on public.access_attempt (type);

create index access_attempt_access_granted_index
    on public.access_attempt (access_granted);

create index access_attempt_time_index
    on public.access_attempt (time);

create table public.verification_code
(
    id       bigserial not null
        constraint verification_code_pk
            primary key,
    code    text                           not null,
    user_id bigint                         not null
        constraint verification_code_user_id_fk
            references public.athlete,
    expires bigint                         not null
);

create index verification_code_user_id_index
    on public.verification_code (user_id);

create index verification_code_expires_index
    on public.verification_code (expires);

create table public.last_completed_measurement
(
    user_id     bigint not null,
    exercise_id smallint not null,
    measurement smallint not null,
    constraint last_completed_measurement_pk
        primary key (user_id)
);

