UPDATE muscle_group
SET name = 'Rectus Abdominis'
WHERE id = '4fbc0e2d-41d2-49e7-b78a-d0f30e0b8950';

INSERT INTO muscle_group (id, name)
VALUES ('00b4144c-9eaa-495a-a1c4-4a30b668f4fa', 'Obliques');

INSERT INTO muscle_group (id, name)
VALUES ('6ed59cfe-6f23-4c6e-a0c6-63c1af3f113e', 'Transverse Abdominis');

INSERT INTO muscle_group (id, name)
VALUES ('9869a55c-d3c9-4687-b32e-1079ce689420', 'Erector Spinae');

INSERT INTO muscle_group (id, name)
VALUES ('19060fbc-9588-44d0-8455-41a5e6ca1ff1', 'Multifidus');

INSERT INTO muscle_group (id, name)
VALUES ('285c2ae7-cd80-4f36-bf9c-ef3348c5edaa', 'Quadratus Lumborum');



DELETE
FROM exercise_muscle_group emg
WHERE muscle_group_id = '4fbc0e2d-41d2-49e7-b78a-d0f30e0b8950';




INSERT INTO exercise_muscle_group (exercise_id, muscle_group_id)
VALUES ('029e64e2-81d5-4ff5-be23-df0cd9b18823', '4fbc0e2d-41d2-49e7-b78a-d0f30e0b8950');
INSERT INTO exercise_muscle_group (exercise_id, muscle_group_id)
VALUES ('029e64e2-81d5-4ff5-be23-df0cd9b18823', '6ed59cfe-6f23-4c6e-a0c6-63c1af3f113e');




INSERT INTO exercise_muscle_group (exercise_id, muscle_group_id)
VALUES ('0dc37e9e-1fa1-47f4-b1ee-2cc411c0fb2d', '6ed59cfe-6f23-4c6e-a0c6-63c1af3f113e');
INSERT INTO exercise_muscle_group (exercise_id, muscle_group_id)
VALUES ('0dc37e9e-1fa1-47f4-b1ee-2cc411c0fb2d', '9869a55c-d3c9-4687-b32e-1079ce689420');




INSERT INTO exercise_muscle_group (exercise_id, muscle_group_id)
VALUES ('259bbe35-ad68-4158-b520-e21596d190ed', '00b4144c-9eaa-495a-a1c4-4a30b668f4fa');
INSERT INTO exercise_muscle_group (exercise_id, muscle_group_id)
VALUES ('259bbe35-ad68-4158-b520-e21596d190ed', '6ed59cfe-6f23-4c6e-a0c6-63c1af3f113e');



INSERT INTO exercise_muscle_group (exercise_id, muscle_group_id)
VALUES ('3f958ede-410a-4b49-9f8f-beb58ba0a79d', '4fbc0e2d-41d2-49e7-b78a-d0f30e0b8950');
INSERT INTO exercise_muscle_group (exercise_id, muscle_group_id)
VALUES ('3f958ede-410a-4b49-9f8f-beb58ba0a79d', '6ed59cfe-6f23-4c6e-a0c6-63c1af3f113e');



INSERT INTO exercise_muscle_group (exercise_id, muscle_group_id)
VALUES ('5f19abeb-9c35-4dd5-a95a-ca0923d8b693', '4fbc0e2d-41d2-49e7-b78a-d0f30e0b8950');
INSERT INTO exercise_muscle_group (exercise_id, muscle_group_id)
VALUES ('5f19abeb-9c35-4dd5-a95a-ca0923d8b693', '00b4144c-9eaa-495a-a1c4-4a30b668f4fa');
INSERT INTO exercise_muscle_group (exercise_id, muscle_group_id)
VALUES ('5f19abeb-9c35-4dd5-a95a-ca0923d8b693', '6ed59cfe-6f23-4c6e-a0c6-63c1af3f113e');
INSERT INTO exercise_muscle_group (exercise_id, muscle_group_id)
VALUES ('5f19abeb-9c35-4dd5-a95a-ca0923d8b693', '285c2ae7-cd80-4f36-bf9c-ef3348c5edaa');



INSERT INTO exercise_muscle_group (exercise_id, muscle_group_id)
VALUES ('63d7154c-c486-422a-847d-df970bf0b57f', '4fbc0e2d-41d2-49e7-b78a-d0f30e0b8950');
INSERT INTO exercise_muscle_group (exercise_id, muscle_group_id)
VALUES ('63d7154c-c486-422a-847d-df970bf0b57f', '00b4144c-9eaa-495a-a1c4-4a30b668f4fa');
INSERT INTO exercise_muscle_group (exercise_id, muscle_group_id)
VALUES ('63d7154c-c486-422a-847d-df970bf0b57f', '6ed59cfe-6f23-4c6e-a0c6-63c1af3f113e');



INSERT INTO exercise_muscle_group (exercise_id, muscle_group_id)
VALUES ('6d56462f-aef6-4cce-a5ac-b6e6bc860a4e', '4fbc0e2d-41d2-49e7-b78a-d0f30e0b8950');
INSERT INTO exercise_muscle_group (exercise_id, muscle_group_id)
VALUES ('6d56462f-aef6-4cce-a5ac-b6e6bc860a4e', '6ed59cfe-6f23-4c6e-a0c6-63c1af3f113e');



INSERT INTO exercise_muscle_group (exercise_id, muscle_group_id)
VALUES ('6f908487-b829-4487-9610-fc8d885b0cb7', '4fbc0e2d-41d2-49e7-b78a-d0f30e0b8950');
INSERT INTO exercise_muscle_group (exercise_id, muscle_group_id)
VALUES ('6f908487-b829-4487-9610-fc8d885b0cb7', '00b4144c-9eaa-495a-a1c4-4a30b668f4fa');
INSERT INTO exercise_muscle_group (exercise_id, muscle_group_id)
VALUES ('6f908487-b829-4487-9610-fc8d885b0cb7', '6ed59cfe-6f23-4c6e-a0c6-63c1af3f113e');



INSERT INTO exercise_muscle_group (exercise_id, muscle_group_id)
VALUES ('7f0b3b71-1f3e-41fc-b522-0819db869317', '4fbc0e2d-41d2-49e7-b78a-d0f30e0b8950');
INSERT INTO exercise_muscle_group (exercise_id, muscle_group_id)
VALUES ('7f0b3b71-1f3e-41fc-b522-0819db869317', '6ed59cfe-6f23-4c6e-a0c6-63c1af3f113e');



INSERT INTO exercise_muscle_group (exercise_id, muscle_group_id)
VALUES ('85c5fe0a-da11-4ba0-81b6-2d862953198a', '285c2ae7-cd80-4f36-bf9c-ef3348c5edaa');
INSERT INTO exercise_muscle_group (exercise_id, muscle_group_id)
VALUES ('85c5fe0a-da11-4ba0-81b6-2d862953198a', '00b4144c-9eaa-495a-a1c4-4a30b668f4fa');



INSERT INTO exercise_muscle_group (exercise_id, muscle_group_id)
VALUES ('b1387dc3-1105-4701-9f7e-3cb34132dd3f', '4fbc0e2d-41d2-49e7-b78a-d0f30e0b8950');
INSERT INTO exercise_muscle_group (exercise_id, muscle_group_id)
VALUES ('b1387dc3-1105-4701-9f7e-3cb34132dd3f', '6ed59cfe-6f23-4c6e-a0c6-63c1af3f113e');



INSERT INTO exercise_muscle_group (exercise_id, muscle_group_id)
VALUES ('b3303e30-9a51-4e44-942e-dd33bc64cb8f', '4fbc0e2d-41d2-49e7-b78a-d0f30e0b8950');
INSERT INTO exercise_muscle_group (exercise_id, muscle_group_id)
VALUES ('b3303e30-9a51-4e44-942e-dd33bc64cb8f', '6ed59cfe-6f23-4c6e-a0c6-63c1af3f113e');



INSERT INTO exercise_muscle_group (exercise_id, muscle_group_id)
VALUES ('b673529b-7e20-49b1-9e92-e4edad081573', '00b4144c-9eaa-495a-a1c4-4a30b668f4fa');
INSERT INTO exercise_muscle_group (exercise_id, muscle_group_id)
VALUES ('b673529b-7e20-49b1-9e92-e4edad081573', '285c2ae7-cd80-4f36-bf9c-ef3348c5edaa');



INSERT INTO exercise_muscle_group (exercise_id, muscle_group_id)
VALUES ('bd53ab36-8d90-4db5-a486-4a78b9d01032', '9869a55c-d3c9-4687-b32e-1079ce689420');


INSERT INTO exercise_muscle_group (exercise_id, muscle_group_id)
VALUES ('d5bf762b-a279-4e93-b37a-376e83c949b8', '4fbc0e2d-41d2-49e7-b78a-d0f30e0b8950');
INSERT INTO exercise_muscle_group (exercise_id, muscle_group_id)
VALUES ('d5bf762b-a279-4e93-b37a-376e83c949b8', '6ed59cfe-6f23-4c6e-a0c6-63c1af3f113e');



INSERT INTO exercise_muscle_group (exercise_id, muscle_group_id)
VALUES ('f21a3b80-b69e-4f34-a903-2196baf9b825', '4fbc0e2d-41d2-49e7-b78a-d0f30e0b8950');
INSERT INTO exercise_muscle_group (exercise_id, muscle_group_id)
VALUES ('f21a3b80-b69e-4f34-a903-2196baf9b825', '6ed59cfe-6f23-4c6e-a0c6-63c1af3f113e');
INSERT INTO exercise_muscle_group (exercise_id, muscle_group_id)
VALUES ('f21a3b80-b69e-4f34-a903-2196baf9b825', '9869a55c-d3c9-4687-b32e-1079ce689420');
INSERT INTO exercise_muscle_group (exercise_id, muscle_group_id)
VALUES ('f21a3b80-b69e-4f34-a903-2196baf9b825', 'da0fd6de-41c6-462a-9fb6-3f661b85bdb0');



INSERT INTO exercise_muscle_group (exercise_id, muscle_group_id)
VALUES ('f5029db9-24d4-4850-a5c2-e5606a359e4d', '9869a55c-d3c9-4687-b32e-1079ce689420');
INSERT INTO exercise_muscle_group (exercise_id, muscle_group_id)
VALUES ('f5029db9-24d4-4850-a5c2-e5606a359e4d', '19060fbc-9588-44d0-8455-41a5e6ca1ff1');