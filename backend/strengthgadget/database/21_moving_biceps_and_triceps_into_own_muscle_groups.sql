UPDATE muscle_group
SET name = 'Biceps'
WHERE id = '5c20da1c-186d-4619-a9cf-07bbff89ab8c';

INSERT INTO muscle_group (id, name)
VALUES ('3ab57998-1a41-4aec-84b7-9704f78ea864', 'Triceps');

UPDATE exercise_muscle_group
SET muscle_group_id = '3ab57998-1a41-4aec-84b7-9704f78ea864'
WHERE exercise_id IN ('3dafcddd-ed0f-406f-9a3d-4438dc24172d',
'4e4a58fd-d80d-482f-b6c6-156612047f23',
'897ac686-0357-4f65-a585-6ea95e21c24d',
'8e4576dc-795a-40f1-bf0b-e3a08673f9e9') AND muscle_group_id = '5c20da1c-186d-4619-a9cf-07bbff89ab8c';

INSERT INTO exercise (id, name, demonstration_giphy_id, exercise_type_id)
VALUES ('800bbf8d-7fcd-4ccd-8604-2f8c26243039', 'Sumo Squat with Barbell', 'a7Y2DhvZX4D3EdluiZ', '6bdb3624-bed1-41a9-bf8c-7b1066411446');
INSERT INTO exercise_muscle_group (exercise_id, muscle_group_id)
VALUES ('800bbf8d-7fcd-4ccd-8604-2f8c26243039', '38f352d6-f290-443a-84f7-baf38f54b5b2');
INSERT INTO exercise_muscle_group (exercise_id, muscle_group_id)
VALUES ('800bbf8d-7fcd-4ccd-8604-2f8c26243039', '305a9027-d72c-4e60-9ab3-c07d854f76c5');
INSERT INTO exercise_muscle_group (exercise_id, muscle_group_id)
VALUES ('800bbf8d-7fcd-4ccd-8604-2f8c26243039', 'b2b1aa14-b7eb-4d98-a038-c73a0cb343f2');
INSERT INTO exercise_measurement_type (exercise_id, measurement_type_id)
VALUES ('800bbf8d-7fcd-4ccd-8604-2f8c26243039', '0c7a1eb7-1b61-4bcd-b136-8778e99e5b06');