create table muscle_group
(
    id   uuid default gen_random_uuid() not null
        constraint muscle_group_pk
            primary key,
    name text                           not null
);

-- 3 muscle group groups a day makes for 72 hours of rest for each muscle group. This assumes exercises won't overlap muscle groups
-- from other groups, so this may not work
INSERT INTO muscle_group (id, name)
VALUES ('4fbc0e2d-41d2-49e7-b78a-d0f30e0b8950', 'Core');
INSERT INTO muscle_group (id, name)
VALUES ('27e80e2c-86c8-4701-80f2-afc1d05484d1', 'Calves');
INSERT INTO muscle_group (id, name)
VALUES ('74710ced-e25a-4e31-ba4f-19c2c64849b0', 'Forearms and Grip Strength');
INSERT INTO muscle_group (id, name)
VALUES ('5ec1ca63-0513-4bb8-aca7-7cca468eb98d', 'Cardio');

INSERT INTO muscle_group (id, name)
VALUES ('305a9027-d72c-4e60-9ab3-c07d854f76c5', 'Glutes');
INSERT INTO muscle_group (id, name)
VALUES ('38f352d6-f290-443a-84f7-baf38f54b5b2', 'Quadriceps');
INSERT INTO muscle_group (id, name)
VALUES ('86cf648e-4aa0-45eb-beee-7380b1a1e00f', 'Hamstrings');
INSERT INTO muscle_group (id, name)
VALUES ('874b7e61-b489-4e8d-8dae-f53c914b5fc7', 'Hip Flexors');

INSERT INTO muscle_group (id, name)
VALUES ('da0fd6de-41c6-462a-9fb6-3f661b85bdb0', 'Back');
INSERT INTO muscle_group (id, name)
VALUES ('5a2a9865-74d9-4092-9f7e-011bb6931f6a', 'Chest');
INSERT INTO muscle_group (id, name)
VALUES ('9042cec4-faab-4920-999f-5fb7335bb8f1', 'Shoulders');
INSERT INTO muscle_group (id, name)
VALUES ('5c20da1c-186d-4619-a9cf-07bbff89ab8c', 'Biceps and Triceps');


create table exercise
(
    id                     uuid default gen_random_uuid() not null
        constraint exercise_pk
            primary key,
    name                   text                           not null,
    demonstration_giphy_id text                           not null
);

-- Core
INSERT INTO exercise (id, name, demonstration_giphy_id) VALUES ('029e64e2-81d5-4ff5-be23-df0cd9b18823', 'Leg Raises', '55atXlETBRZm0h9NrO');
INSERT INTO exercise (id, name, demonstration_giphy_id)
VALUES ('0dc37e9e-1fa1-47f4-b1ee-2cc411c0fb2d', 'Bird Dog', 'RpwQmzE45R3NTBaszK');
INSERT INTO exercise (id, name, demonstration_giphy_id) VALUES ('259bbe35-ad68-4158-b520-e21596d190ed', 'Russian Twists', '8mRpGumgGIswSG2dD5');
INSERT INTO exercise (id, name, demonstration_giphy_id)
VALUES ('3f958ede-410a-4b49-9f8f-beb58ba0a79d', 'Dead Bug', 'XjvQ1Tfu8y1xYK23vi');
INSERT INTO exercise (id, name, demonstration_giphy_id) VALUES ('63d7154c-c486-422a-847d-df970bf0b57f', 'Bicycle Crunches', 'IkFw3Mnwi6g7tO5AN1');
INSERT INTO exercise (id, name, demonstration_giphy_id)
VALUES ('6d56462f-aef6-4cce-a5ac-b6e6bc860a4e', 'Plank', 'HxBrA4zpPtUb3quip7');
INSERT INTO exercise (id, name, demonstration_giphy_id) VALUES ('6f908487-b829-4487-9610-fc8d885b0cb7', 'High Plank Knee-to-Elbow', 'JFdWzEYkbK7tyf1A1y');
INSERT INTO exercise (id, name, demonstration_giphy_id)
VALUES ('b1387dc3-1105-4701-9f7e-3cb34132dd3f', 'Mountain Climbers', '4NojW5eV2t2yY4R0JA');
INSERT INTO exercise (id, name, demonstration_giphy_id)
VALUES ('b673529b-7e20-49b1-9e92-e4edad081573', 'Side Plank', 'YeYfAFgpamPVi29TNI');
INSERT INTO exercise (id, name, demonstration_giphy_id) VALUES ('f5029db9-24d4-4850-a5c2-e5606a359e4d', 'Superman', 'xxQIlMGciMFUsDA6DX');

-- Calves
-- INSERT INTO exercise (id, name, demonstration_giphy_id) VALUES ('0ea0f5f3-85d1-43de-9da4-2dbe66a6818d', 'Single Leg Calf Raises', '');
-- INSERT INTO exercise (id, name, demonstration_giphy_id) VALUES ('39ca5b58-2b5e-42c2-a4f8-c604b886f0fc', 'Box Jumps', '');
-- INSERT INTO exercise (id, name, demonstration_giphy_id) VALUES ('6c01d2fb-eeb9-497d-b2bb-e01fc7855348', 'Running or Walking on Tiptoes', '');
-- INSERT INTO exercise (id, name, demonstration_giphy_id) VALUES ('7fdaa7fb-028b-49e5-9b27-9bbf5adeafa0', 'Hill or Stair Climbing', '');
-- INSERT INTO exercise (id, name, demonstration_giphy_id) VALUES ('94b90ec5-d0ae-4991-93ef-213666ba043c', 'Downward Dog', '');
-- INSERT INTO exercise (id, name, demonstration_giphy_id) VALUES ('9d4372f9-ced3-461e-93dd-76a2107d5a9a', 'Calf Raises', '');
-- INSERT INTO exercise (id, name, demonstration_giphy_id) VALUES ('a5bc9d2d-5b2b-4d50-8fac-c8f99a56cdf4', 'Jumping Rope', '');


create table exercise_muscle_group
(
    exercise_id     uuid not null
        constraint exercise_muscle_group_exercise_id_fk
            references exercise,
    muscle_group_id uuid not null
        constraint exercise_muscle_group_muscle_group_id_fk
            references muscle_group,
    constraint exercise_muscle_group_pk
        primary key (exercise_id, muscle_group_id)
);

-- Core
INSERT INTO exercise_muscle_group (exercise_id, muscle_group_id)
VALUES ('029e64e2-81d5-4ff5-be23-df0cd9b18823', '4fbc0e2d-41d2-49e7-b78a-d0f30e0b8950');
INSERT INTO exercise_muscle_group (exercise_id, muscle_group_id)
VALUES ('0dc37e9e-1fa1-47f4-b1ee-2cc411c0fb2d', '4fbc0e2d-41d2-49e7-b78a-d0f30e0b8950');
INSERT INTO exercise_muscle_group (exercise_id, muscle_group_id)
VALUES ('259bbe35-ad68-4158-b520-e21596d190ed', '4fbc0e2d-41d2-49e7-b78a-d0f30e0b8950');
INSERT INTO exercise_muscle_group (exercise_id, muscle_group_id)
VALUES ('3f958ede-410a-4b49-9f8f-beb58ba0a79d', '4fbc0e2d-41d2-49e7-b78a-d0f30e0b8950');
INSERT INTO exercise_muscle_group (exercise_id, muscle_group_id)
VALUES ('63d7154c-c486-422a-847d-df970bf0b57f', '4fbc0e2d-41d2-49e7-b78a-d0f30e0b8950');
INSERT INTO exercise_muscle_group (exercise_id, muscle_group_id)
VALUES ('6d56462f-aef6-4cce-a5ac-b6e6bc860a4e', '4fbc0e2d-41d2-49e7-b78a-d0f30e0b8950');
INSERT INTO exercise_muscle_group (exercise_id, muscle_group_id)
VALUES ('6f908487-b829-4487-9610-fc8d885b0cb7', '4fbc0e2d-41d2-49e7-b78a-d0f30e0b8950');
INSERT INTO exercise_muscle_group (exercise_id, muscle_group_id)
VALUES ('b1387dc3-1105-4701-9f7e-3cb34132dd3f', '4fbc0e2d-41d2-49e7-b78a-d0f30e0b8950');
INSERT INTO exercise_muscle_group (exercise_id, muscle_group_id)
VALUES ('b673529b-7e20-49b1-9e92-e4edad081573', '4fbc0e2d-41d2-49e7-b78a-d0f30e0b8950');
INSERT INTO exercise_muscle_group (exercise_id, muscle_group_id)
VALUES ('f5029db9-24d4-4850-a5c2-e5606a359e4d', '4fbc0e2d-41d2-49e7-b78a-d0f30e0b8950');
--
-- -- Calves
-- INSERT INTO exercise_muscle_group (exercise_id, muscle_group_id)
-- VALUES ('0ea0f5f3-85d1-43de-9da4-2dbe66a6818d', '27e80e2c-86c8-4701-80f2-afc1d05484d1');
-- INSERT INTO exercise_muscle_group (exercise_id, muscle_group_id)
-- VALUES ('39ca5b58-2b5e-42c2-a4f8-c604b886f0fc', '27e80e2c-86c8-4701-80f2-afc1d05484d1');
-- INSERT INTO exercise_muscle_group (exercise_id, muscle_group_id)
-- VALUES ('6c01d2fb-eeb9-497d-b2bb-e01fc7855348', '27e80e2c-86c8-4701-80f2-afc1d05484d1');
-- INSERT INTO exercise_muscle_group (exercise_id, muscle_group_id)
-- VALUES ('7fdaa7fb-028b-49e5-9b27-9bbf5adeafa0', '27e80e2c-86c8-4701-80f2-afc1d05484d1');
-- INSERT INTO exercise_muscle_group (exercise_id, muscle_group_id)
-- VALUES ('94b90ec5-d0ae-4991-93ef-213666ba043c', '27e80e2c-86c8-4701-80f2-afc1d05484d1');
-- INSERT INTO exercise_muscle_group (exercise_id, muscle_group_id)
-- VALUES ('9d4372f9-ced3-461e-93dd-76a2107d5a9a', '27e80e2c-86c8-4701-80f2-afc1d05484d1');
-- INSERT INTO exercise_muscle_group (exercise_id, muscle_group_id)
-- VALUES ('a5bc9d2d-5b2b-4d50-8fac-c8f99a56cdf4', '27e80e2c-86c8-4701-80f2-afc1d05484d1');
