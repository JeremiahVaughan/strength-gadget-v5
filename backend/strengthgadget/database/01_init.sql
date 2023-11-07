CREATE TABLE public."user" (
  id UUID NOT NULL DEFAULT gen_random_uuid(),
  email STRING NOT NULL,
  password_hash STRING NOT NULL,
  CONSTRAINT user_pkey PRIMARY KEY (id ASC),
  UNIQUE INDEX email__index (email ASC)
);


CREATE TABLE public.access_attempt_type (
  id UUID NOT NULL DEFAULT gen_random_uuid(),
  name STRING NOT NULL,
  CONSTRAINT access_attempt_pk PRIMARY KEY (id ASC)
);

INSERT INTO public.access_attempt_type (id, name)
VALUES
('14cb4661-74e5-49e8-8532-ebe93d1e806a', 'PASSWORD_RESET'),
('288e1dae-5865-4707-b242-ce818ee8145f', 'LOGIN'),
('ca33f4f1-e2ba-49e1-8222-be982a57c231', 'VERIFICATION');



CREATE TABLE public.access_attempt (
  id UUID NOT NULL DEFAULT gen_random_uuid(),
  "time" INT8 NOT NULL,
  access_granted BOOL NOT NULL,
  type UUID NOT NULL,
  user_id UUID NOT NULL,
  CONSTRAINT access_attempt_pk PRIMARY KEY (id ASC),
  CONSTRAINT access_attempt_access_attempt_type_id_fk FOREIGN KEY (type) REFERENCES public.access_attempt_type(id),
  CONSTRAINT access_attempt_user_id_fk FOREIGN KEY (user_id) REFERENCES public."user"(id),
  INDEX access_attempt_access_granted_index (access_granted ASC),
  INDEX access_attempt_time_index ("time" ASC),
  INDEX access_attempt_type_index (type ASC)
);


CREATE TABLE public.verification_code (
  id UUID NOT NULL DEFAULT gen_random_uuid(),
  code STRING NOT NULL,
  user_id UUID NOT NULL,
  expires INT8 NOT NULL,
  CONSTRAINT verification_code_pk PRIMARY KEY (id ASC),
  CONSTRAINT verification_code_user_id_fk FOREIGN KEY (user_id) REFERENCES public."user"(id),
  INDEX verification_code_expires_index (expires ASC)
);
