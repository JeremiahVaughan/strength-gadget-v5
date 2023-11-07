INSERT INTO public.exercise (id, name, demonstration_giphy_id) VALUES ('cd7d6cf1-c6be-4ff1-a6e3-1e2787cabe41', 'Squats', 'wxNwwnoYUyxxWqvySO');
INSERT INTO public.exercise (id, name, demonstration_giphy_id) VALUES ('a3fbf86c-6e7b-4899-8636-8f20171dfe95', 'Lunges', 'N9hhNLh26xarKxizrY');
INSERT INTO public.exercise (id, name, demonstration_giphy_id) VALUES ('8e3b8279-fe1e-4bb1-9ce4-bf2f6b53a74a', 'Jump Squats', 'I6YSGpRMYGNwzxAo0d');
INSERT INTO public.exercise (id, name, demonstration_giphy_id) VALUES ('7ec3823c-3eb1-42c4-ac6b-c255761d3bf7', 'Wall Sit', 'kAsOw4LRzKvKZPh6fc');

UPDATE public.exercise SET demonstration_giphy_id = 'kTdej6DP88WuWkggCp' WHERE id = '6d56462f-aef6-4cce-a5ac-b6e6bc860a4e';


INSERT INTO public.exercise_muscle_group (exercise_id, muscle_group_id)
VALUES ('cd7d6cf1-c6be-4ff1-a6e3-1e2787cabe41', '305a9027-d72c-4e60-9ab3-c07d854f76c5');
INSERT INTO public.exercise_muscle_group (exercise_id, muscle_group_id)
VALUES ('cd7d6cf1-c6be-4ff1-a6e3-1e2787cabe41', '38f352d6-f290-443a-84f7-baf38f54b5b2');


INSERT INTO public.exercise_muscle_group (exercise_id, muscle_group_id)
VALUES ('a3fbf86c-6e7b-4899-8636-8f20171dfe95', '305a9027-d72c-4e60-9ab3-c07d854f76c5');
INSERT INTO public.exercise_muscle_group (exercise_id, muscle_group_id)
VALUES ('a3fbf86c-6e7b-4899-8636-8f20171dfe95', '38f352d6-f290-443a-84f7-baf38f54b5b2');


INSERT INTO public.exercise_muscle_group (exercise_id, muscle_group_id)
VALUES ('8e3b8279-fe1e-4bb1-9ce4-bf2f6b53a74a', '305a9027-d72c-4e60-9ab3-c07d854f76c5');
INSERT INTO public.exercise_muscle_group (exercise_id, muscle_group_id)
VALUES ('8e3b8279-fe1e-4bb1-9ce4-bf2f6b53a74a', '38f352d6-f290-443a-84f7-baf38f54b5b2');


INSERT INTO public.exercise_muscle_group (exercise_id, muscle_group_id)
VALUES ('7ec3823c-3eb1-42c4-ac6b-c255761d3bf7', '305a9027-d72c-4e60-9ab3-c07d854f76c5');
INSERT INTO public.exercise_muscle_group (exercise_id, muscle_group_id)
VALUES ('7ec3823c-3eb1-42c4-ac6b-c255761d3bf7', '38f352d6-f290-443a-84f7-baf38f54b5b2');