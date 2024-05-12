INSERT INTO muscle_group (id, name)
VALUES ('49eb115c-f18a-4f39-b750-270b7ccb1eef', 'Abductors');

INSERT INTO exercise (id, name, demonstration_giphy_id, exercise_type_id)
VALUES ('faa2b566-0d73-4c28-acc0-878b110d3309', 'Side Leg Raises', 'jgMpJCqtCij8XGsvIe', '8ffe7196-4e3d-4439-ae19-3159ad5387bd');
INSERT INTO exercise_muscle_group (exercise_id, muscle_group_id)
VALUES ('faa2b566-0d73-4c28-acc0-878b110d3309', '49eb115c-f18a-4f39-b750-270b7ccb1eef');
INSERT INTO exercise_measurement_type (exercise_id, measurement_type_id)
VALUES ('faa2b566-0d73-4c28-acc0-878b110d3309', 'a4345a46-44fa-4c9a-beee-ab53f48c8a33');


INSERT INTO exercise (id, name, demonstration_giphy_id, exercise_type_id)
VALUES ('aa893498-2b47-460a-a4ca-4475ffec2685', 'Extended-Range, Side-Lying Hip Abduction', 'RtBE6pTJU3zm9qng6e', '8ffe7196-4e3d-4439-ae19-3159ad5387bd');
INSERT INTO exercise_muscle_group (exercise_id, muscle_group_id)
VALUES ('aa893498-2b47-460a-a4ca-4475ffec2685', '49eb115c-f18a-4f39-b750-270b7ccb1eef');
INSERT INTO exercise_measurement_type (exercise_id, measurement_type_id)
VALUES ('aa893498-2b47-460a-a4ca-4475ffec2685', 'a4345a46-44fa-4c9a-beee-ab53f48c8a33');