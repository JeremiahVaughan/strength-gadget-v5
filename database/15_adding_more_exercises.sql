INSERT INTO exercise (id, name, demonstration_giphy_id, exercise_type_id)
VALUES ('5f19abeb-9c35-4dd5-a95a-ca0923d8b693', 'Plank with Hip Twist', 'F9JNNNn1YVGe3cACYZ', '8ffe7196-4e3d-4439-ae19-3159ad5387bd');
INSERT INTO exercise_muscle_group (exercise_id, muscle_group_id)
VALUES ('5f19abeb-9c35-4dd5-a95a-ca0923d8b693', '4fbc0e2d-41d2-49e7-b78a-d0f30e0b8950');


INSERT INTO exercise (id, name, demonstration_giphy_id, exercise_type_id)
VALUES ('b3303e30-9a51-4e44-942e-dd33bc64cb8f', 'Plank to Forearm Plank', 'mvmrk7kZ8nORvgloyt', '8ffe7196-4e3d-4439-ae19-3159ad5387bd');
INSERT INTO exercise_muscle_group (exercise_id, muscle_group_id)
VALUES ('b3303e30-9a51-4e44-942e-dd33bc64cb8f', '4fbc0e2d-41d2-49e7-b78a-d0f30e0b8950');


INSERT INTO exercise (id, name, demonstration_giphy_id, exercise_type_id)
VALUES ('f21a3b80-b69e-4f34-a903-2196baf9b825', 'Standing Ab Wheel Rollouts', 'PECk6hDHVod2x4PBe0', '8ffe7196-4e3d-4439-ae19-3159ad5387bd');
INSERT INTO exercise_muscle_group (exercise_id, muscle_group_id)
VALUES ('f21a3b80-b69e-4f34-a903-2196baf9b825', '4fbc0e2d-41d2-49e7-b78a-d0f30e0b8950');



INSERT INTO measurement_type (id, name, abbreviated_name)
VALUES ('a4345a46-44fa-4c9a-beee-ab53f48c8a33', 'repetition', 'rep');

alter table exercise_measurement_type
    rename column exercise_type_id to measurement_type_id;



INSERT INTO exercise_measurement_type (exercise_id, measurement_type_id)
VALUES ('029e64e2-81d5-4ff5-be23-df0cd9b18823', 'a4345a46-44fa-4c9a-beee-ab53f48c8a33');

INSERT INTO exercise_measurement_type (exercise_id, measurement_type_id)
VALUES ('0c1c7312-3c82-464e-bf9a-04c007713f33', 'a4345a46-44fa-4c9a-beee-ab53f48c8a33');

INSERT INTO exercise_measurement_type (exercise_id, measurement_type_id)
VALUES ('0dc37e9e-1fa1-47f4-b1ee-2cc411c0fb2d', 'ca0a001b-eddc-4b90-8fd1-b06e811819f5');

INSERT INTO exercise_measurement_type (exercise_id, measurement_type_id)
VALUES ('15cb4343-1dd6-42c9-8211-10780e8a11a9', 'a4345a46-44fa-4c9a-beee-ab53f48c8a33');

INSERT INTO exercise_measurement_type (exercise_id, measurement_type_id)
VALUES ('21b50e5f-d4ae-4a4f-8026-93b6cdbcba12', '0c7a1eb7-1b61-4bcd-b136-8778e99e5b06');

INSERT INTO exercise_measurement_type (exercise_id, measurement_type_id)
VALUES ('259bbe35-ad68-4158-b520-e21596d190ed', 'ca0a001b-eddc-4b90-8fd1-b06e811819f5');

INSERT INTO exercise_measurement_type (exercise_id, measurement_type_id)
VALUES ('2b82618c-368a-4585-81f7-9c9a6d14d257', 'c94e6ede-2da4-42fc-92e3-0b1ff2b2fcb3');

INSERT INTO exercise_measurement_type (exercise_id, measurement_type_id)
VALUES ('3dafcddd-ed0f-406f-9a3d-4438dc24172d', '0c7a1eb7-1b61-4bcd-b136-8778e99e5b06');

INSERT INTO exercise_measurement_type (exercise_id, measurement_type_id)
VALUES ('3f958ede-410a-4b49-9f8f-beb58ba0a79d', 'ca0a001b-eddc-4b90-8fd1-b06e811819f5');

INSERT INTO exercise_measurement_type (exercise_id, measurement_type_id)
VALUES ('49b90ba8-0c4c-42c3-8032-d289fa91238e', 'a4345a46-44fa-4c9a-beee-ab53f48c8a33');

INSERT INTO exercise_measurement_type (exercise_id, measurement_type_id)
VALUES ('4d0b3cf5-cbf3-467d-9610-9168794a4915', '0c7a1eb7-1b61-4bcd-b136-8778e99e5b06');

INSERT INTO exercise_measurement_type (exercise_id, measurement_type_id)
VALUES ('50deb36b-bae1-4b96-8025-dd481bc07c47', 'a4345a46-44fa-4c9a-beee-ab53f48c8a33');

INSERT INTO exercise_measurement_type (exercise_id, measurement_type_id)
VALUES ('5e07b51d-7dac-41fa-b7df-51d7d8086171', '0c7a1eb7-1b61-4bcd-b136-8778e99e5b06');

INSERT INTO exercise_measurement_type (exercise_id, measurement_type_id)
VALUES ('63d7154c-c486-422a-847d-df970bf0b57f', 'ca0a001b-eddc-4b90-8fd1-b06e811819f5');

INSERT INTO exercise_measurement_type (exercise_id, measurement_type_id)
VALUES ('6d56462f-aef6-4cce-a5ac-b6e6bc860a4e', 'ca0a001b-eddc-4b90-8fd1-b06e811819f5');

INSERT INTO exercise_measurement_type (exercise_id, measurement_type_id)
VALUES ('6f908487-b829-4487-9610-fc8d885b0cb7', 'ca0a001b-eddc-4b90-8fd1-b06e811819f5');

INSERT INTO exercise_measurement_type (exercise_id, measurement_type_id)
VALUES ('70a931af-c145-40e9-b539-71332468078c', '0c7a1eb7-1b61-4bcd-b136-8778e99e5b06');

INSERT INTO exercise_measurement_type (exercise_id, measurement_type_id)
VALUES ('75aff446-c084-4ff2-9374-c3157e645e74', 'a4345a46-44fa-4c9a-beee-ab53f48c8a33');

INSERT INTO exercise_measurement_type (exercise_id, measurement_type_id)
VALUES ('7ec3823c-3eb1-42c4-ac6b-c255761d3bf7', 'ca0a001b-eddc-4b90-8fd1-b06e811819f5');

INSERT INTO exercise_measurement_type (exercise_id, measurement_type_id)
VALUES ('7f0b3b71-1f3e-41fc-b522-0819db869317', 'ca0a001b-eddc-4b90-8fd1-b06e811819f5');

INSERT INTO exercise_measurement_type (exercise_id, measurement_type_id)
VALUES ('85c5fe0a-da11-4ba0-81b6-2d862953198a', '0c7a1eb7-1b61-4bcd-b136-8778e99e5b06');

INSERT INTO exercise_measurement_type (exercise_id, measurement_type_id)
VALUES ('878cdd10-e11f-4925-bd2e-d0909d616b1b', '0c7a1eb7-1b61-4bcd-b136-8778e99e5b06');

INSERT INTO exercise_measurement_type (exercise_id, measurement_type_id)
VALUES ('897ac686-0357-4f65-a585-6ea95e21c24d', 'a4345a46-44fa-4c9a-beee-ab53f48c8a33');

INSERT INTO exercise_measurement_type (exercise_id, measurement_type_id)
VALUES ('8e3b8279-fe1e-4bb1-9ce4-bf2f6b53a74a', 'a4345a46-44fa-4c9a-beee-ab53f48c8a33');

INSERT INTO exercise_measurement_type (exercise_id, measurement_type_id)
VALUES ('92f6a301-bfdb-4807-9c50-a3727feda152', 'c94e6ede-2da4-42fc-92e3-0b1ff2b2fcb3');

INSERT INTO exercise_measurement_type (exercise_id, measurement_type_id)
VALUES ('a2247e91-577b-4e20-a573-a0d3a36e1d56', 'a4345a46-44fa-4c9a-beee-ab53f48c8a33');

INSERT INTO exercise_measurement_type (exercise_id, measurement_type_id)
VALUES ('a3fbf86c-6e7b-4899-8636-8f20171dfe95', 'a4345a46-44fa-4c9a-beee-ab53f48c8a33');

INSERT INTO exercise_measurement_type (exercise_id, measurement_type_id)
VALUES ('a6104691-5392-4601-89fe-83062a98ebbc', 'ca0a001b-eddc-4b90-8fd1-b06e811819f5');

INSERT INTO exercise_measurement_type (exercise_id, measurement_type_id)
VALUES ('b1387dc3-1105-4701-9f7e-3cb34132dd3f', 'ca0a001b-eddc-4b90-8fd1-b06e811819f5');

INSERT INTO exercise_measurement_type (exercise_id, measurement_type_id)
VALUES ('b673529b-7e20-49b1-9e92-e4edad081573', 'ca0a001b-eddc-4b90-8fd1-b06e811819f5');

INSERT INTO exercise_measurement_type (exercise_id, measurement_type_id)
VALUES ('bd53ab36-8d90-4db5-a486-4a78b9d01032', 'ca0a001b-eddc-4b90-8fd1-b06e811819f5');

INSERT INTO exercise_measurement_type (exercise_id, measurement_type_id)
VALUES ('c2b62577-48d4-41ef-8a09-dd76a7eae1fe', '0c7a1eb7-1b61-4bcd-b136-8778e99e5b06');

INSERT INTO exercise_measurement_type (exercise_id, measurement_type_id)
VALUES ('cd7d6cf1-c6be-4ff1-a6e3-1e2787cabe41', 'a4345a46-44fa-4c9a-beee-ab53f48c8a33');

INSERT INTO exercise_measurement_type (exercise_id, measurement_type_id)
VALUES ('d5472f88-fafb-48d8-bfdb-7f8b6c68ea70', '0c7a1eb7-1b61-4bcd-b136-8778e99e5b06');

INSERT INTO exercise_measurement_type (exercise_id, measurement_type_id)
VALUES ('d5bf762b-a279-4e93-b37a-376e83c949b8', 'ca0a001b-eddc-4b90-8fd1-b06e811819f5');

INSERT INTO exercise_measurement_type (exercise_id, measurement_type_id)
VALUES ('db1e81f0-f47e-4713-8c27-99822cb651c4', '0c7a1eb7-1b61-4bcd-b136-8778e99e5b06');

INSERT INTO exercise_measurement_type (exercise_id, measurement_type_id)
VALUES ('ed1b3fc5-e11b-4cfc-819a-2c37a409c132', '0c7a1eb7-1b61-4bcd-b136-8778e99e5b06');

INSERT INTO exercise_measurement_type (exercise_id, measurement_type_id)
VALUES ('f35115f5-0bc4-409c-930e-be71adfc1835', '0c7a1eb7-1b61-4bcd-b136-8778e99e5b06');

INSERT INTO exercise_measurement_type (exercise_id, measurement_type_id)
VALUES ('f5029db9-24d4-4850-a5c2-e5606a359e4d', 'ca0a001b-eddc-4b90-8fd1-b06e811819f5');

INSERT INTO exercise_measurement_type (exercise_id, measurement_type_id)
VALUES ('f75d9a21-191c-4d0d-881a-4ccf72f16e0b', '0c7a1eb7-1b61-4bcd-b136-8778e99e5b06');

INSERT INTO exercise_measurement_type (exercise_id, measurement_type_id)
VALUES ('fbc1b600-e475-4d17-99f2-cd70356d53bd', '0c7a1eb7-1b61-4bcd-b136-8778e99e5b06');

INSERT INTO exercise_measurement_type (exercise_id, measurement_type_id)
VALUES ('5c53ddb0-6650-45ef-a097-c66c4d56962c', '0c7a1eb7-1b61-4bcd-b136-8778e99e5b06');

INSERT INTO exercise_measurement_type (exercise_id, measurement_type_id)
VALUES ('5f19abeb-9c35-4dd5-a95a-ca0923d8b693', 'ca0a001b-eddc-4b90-8fd1-b06e811819f5');

INSERT INTO exercise_measurement_type (exercise_id, measurement_type_id)
VALUES ('83088bee-ebdb-4cf6-83ba-ecb77767fe30', '0c7a1eb7-1b61-4bcd-b136-8778e99e5b06');

INSERT INTO exercise_measurement_type (exercise_id, measurement_type_id)
VALUES ('9495a0bc-1439-49c9-aa87-9368d8a3ec15', 'ca0a001b-eddc-4b90-8fd1-b06e811819f5');

INSERT INTO exercise_measurement_type (exercise_id, measurement_type_id)
VALUES ('b3303e30-9a51-4e44-942e-dd33bc64cb8f', 'ca0a001b-eddc-4b90-8fd1-b06e811819f5');

INSERT INTO exercise_measurement_type (exercise_id, measurement_type_id)
VALUES ('c73b4cdc-0885-4505-adf3-563c8bc8269a', 'ca0a001b-eddc-4b90-8fd1-b06e811819f5');

INSERT INTO exercise_measurement_type (exercise_id, measurement_type_id)
VALUES ('f21a3b80-b69e-4f34-a903-2196baf9b825', 'a4345a46-44fa-4c9a-beee-ab53f48c8a33');

alter table exercise_measurement_type
    add constraint exercise_measurement_type_exercise_id_fk
        foreign key (exercise_id) references exercise;

alter table exercise_measurement_type
    add constraint exercise_measurement_type_measurement_type_id_fk
        foreign key (measurement_type_id) references measurement_type;