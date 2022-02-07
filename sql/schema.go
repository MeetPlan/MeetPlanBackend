package sql

const schema string = `
CREATE TABLE IF NOT EXISTS testing (
	id           INTEGER       PRIMARY KEY,
	user_id      INTEGER       NOT NULL,
    date         VARCHAR(250)  NOT NULL,
	teacher_id   INTEGER       NOT NULL,
	class_id     INTEGER       NOT NULL,
	result       VARCHAR(250)  NOT NULL
);
CREATE TABLE IF NOT EXISTS users (
    id           INTEGER       PRIMARY KEY,
    email        VARCHAR(250)  NOT NULL,
    pass         VARCHAR(250)  NOT NULL,
	name         VARCHAR(250)  NOT NULL,
	role         VARCHAR(50)   NOT NULL
);
CREATE TABLE IF NOT EXISTS classes (
	id           INTEGER       PRIMARY KEY,
	students     JSON          DEFAULT('[]'),
	name         VARCHAR(100)  NOT NULL,
	teacher      INTEGER
);
CREATE TABLE IF NOT EXISTS meetings (
	id                      INTEGER       PRIMARY KEY,
	meeting_name            VARCHAR(200)  NOT NULL,
	url                     VARCHAR(300)  NOT NULL,
	details                 VARCHAR(1000) NOT NULL,
	teacher_id              INTEGER       NOT NULL,
	class_id                INTEGER       NOT NULL,
	hour                    INTEGER       NOT NULL,
	date                    INTEGER       NOT NULL,
	is_mandatory            BOOLEAN       NOT NULL,
	is_grading              BOOLEAN       NOT NULL,
	is_written_assessment   BOOLEAN,
	is_test                 BOOLEAN       NOT NULL
);
`
