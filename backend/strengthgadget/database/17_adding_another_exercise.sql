UPDATE public.exercise
SET name = 'Strait Leg Deadlift with Dumbbells'
WHERE id = '5e07b51d-7dac-41fa-b7df-51d7d8086171';

INSERT INTO exercise (id, name, demonstration_giphy_id, exercise_type_id)
VALUES ('58cdd50c-6ce4-401e-a53d-b328c86d3f68', 'Strait Leg Deadlift with Barbell', 'oYK8O344YusZHZKW7S', '6bdb3624-bed1-41a9-bf8c-7b1066411446');
INSERT INTO exercise_muscle_group (exercise_id, muscle_group_id)
VALUES ('58cdd50c-6ce4-401e-a53d-b328c86d3f68', '86cf648e-4aa0-45eb-beee-7380b1a1e00f');
INSERT INTO exercise_measurement_type (exercise_id, measurement_type_id)
VALUES ('58cdd50c-6ce4-401e-a53d-b328c86d3f68', '0c7a1eb7-1b61-4bcd-b136-8778e99e5b06');