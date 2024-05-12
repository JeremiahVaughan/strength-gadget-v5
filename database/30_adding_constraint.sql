UPDATE public.exercise
SET measurement_type_id = '0c7a1eb7-1b61-4bcd-b136-8778e99e5b06'
WHERE id = 'b84acd17-670d-4449-a249-4bba40f00c8c'::uuid;

UPDATE public.exercise
SET measurement_type_id = '0c7a1eb7-1b61-4bcd-b136-8778e99e5b06'
WHERE id = '8bd3e0c6-7037-4efd-8527-f422c0bcfc21'::uuid;


alter table exercise
    alter column measurement_type_id set not null;
alter table exercise
    alter column exercise_type_id set not null;
alter table exercise
    alter column demonstration_giphy_id set not null;
alter table exercise
    alter column name set not null;