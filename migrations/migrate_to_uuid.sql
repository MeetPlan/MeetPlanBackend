-- noinspection SqlWithoutWhereForFile

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

DROP TABLE improvements CASCADE;
DROP TABLE classes CASCADE;
DROP TABLE testing CASCADE;
DROP TABLE meetings CASCADE;
DROP TABLE absence CASCADE;
DROP TABLE grades CASCADE;
DROP TABLE subject CASCADE;
DROP TABLE student_homework CASCADE;
DROP TABLE homework CASCADE;
DROP TABLE communication CASCADE;
DROP TABLE message CASCADE;
DROP TABLE meals CASCADE;
DROP TABLE documents CASCADE;
DROP TABLE notifications CASCADE;

ALTER TABLE users ALTER COLUMN id SET DATA TYPE UUID USING (uuid_generate_v4());
UPDATE users SET id=gen_random_uuid();
UPDATE users SET users='[]';
