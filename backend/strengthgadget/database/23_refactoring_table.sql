alter table exercise
    add measurement_type_id uuid;

alter table exercise
    add constraint exercise_measurement_type_id_fk
        foreign key (measurement_type_id) references measurement_type;
