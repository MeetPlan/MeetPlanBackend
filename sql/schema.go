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
	role         VARCHAR(50)   NOT NULL
);
`
