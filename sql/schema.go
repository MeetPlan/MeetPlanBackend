package sql

const schema string = `
CREATE TABLE IF NOT EXISTS testing (
	id           INTEGER                    PRIMARY KEY,
	user_id      INTEGER                    NOT NULL,
    date         VARCHAR(250)               NOT NULL,
	teacher_id   INTEGER                    NOT NULL,
	class_id     INTEGER                    NOT NULL,
	result       VARCHAR(250)               NOT NULL
);
CREATE TABLE IF NOT EXISTS users (
    id                       INTEGER        PRIMARY KEY,
    email                    VARCHAR(250)   NOT NULL,
    pass                     VARCHAR(250)   NOT NULL,
	name                     VARCHAR(250)   NOT NULL,
	role                     VARCHAR(50)    NOT NULL,
    birth_certificate_number VARCHAR(200),
    birthday                 VARCHAR(200),
    country_of_birth         VARCHAR(200),
    city_of_birth            VARCHAR(200),
    users                    VARCHAR(200)   DEFAULT('[]')
);
CREATE TABLE IF NOT EXISTS classes (
	id                       INTEGER        PRIMARY KEY,
	students                 JSON           DEFAULT('[]'),
	name                     VARCHAR(100)   NOT NULL,
    class_year               VARCHAR(20)    DEFAULT(''),
	last_school_date         INTEGER,
	teacher                  INTEGER,
    sok                      INTEGER,
    eok                      INTEGER
);
CREATE TABLE IF NOT EXISTS meetings (
	id                      INTEGER         PRIMARY KEY,
	meeting_name            VARCHAR(200)    NOT NULL,
	url                     VARCHAR(300)    NOT NULL,
	details                 VARCHAR(1000)   NOT NULL,
	teacher_id              INTEGER         NOT NULL,
	subject_id              INTEGER         NOT NULL,
	hour                    INTEGER         NOT NULL,
	date                    INTEGER         NOT NULL,
	is_mandatory            BOOLEAN         NOT NULL,
	is_grading              BOOLEAN         NOT NULL,
	is_written_assessment   BOOLEAN,
	is_test                 BOOLEAN         NOT NULL,
	is_substitution         BOOLEAN         NOT NULL
);
CREATE TABLE IF NOT EXISTS absence (
	id                      INTEGER         PRIMARY KEY,
	user_id                 INTEGER,
	meeting_id              INTEGER,
	teacher_id              INTEGER,
	absence_type            VARCHAR(200),
	is_excused              BOOLEAN
);
CREATE TABLE IF NOT EXISTS grades (
	id                      INTEGER         PRIMARY KEY,
	user_id                 INTEGER,
	teacher_id              INTEGER,
	subject_id              INTEGER,
	date                    VARCHAR(200),
	is_written              BOOLEAN,
	grade                   INTEGER,
	period                  INTEGER,
	is_final                BOOLEAN,
	description             VARCHAR(200)
);
CREATE TABLE IF NOT EXISTS subject (
	id                      INTEGER         PRIMARY KEY,
	teacher_id              INTEGER,
	name                    VARCHAR(200),
    long_name               VARCHAR(200),
	inherits_class          BOOLEAN,
	class_id                INTEGER         DEFAULT(-1),
	students                JSON            DEFAULT('[]')
);
CREATE TABLE IF NOT EXISTS student_homework (
	id                      INTEGER,
	user_id                 INTEGER,
	homework_id             INTEGER,
	status                  VARCHAR(200)
);
CREATE TABLE IF NOT EXISTS homework (
	id                      INTEGER         PRIMARY KEY,
	teacher_id              INTEGER,
	subject_id              INTEGER,
	name                    VARCHAR(200),
	description             VARCHAR(1000),
	from_date               VARCHAR(200),
	to_date                 VARCHAR(200)
);
CREATE TABLE IF NOT EXISTS communication (
	id                      INTEGER         PRIMARY KEY,
	people                  JSON            DEFAULT('[]'),
	title                   VARCHAR(200),
	date_created            VARCHAR(200)
);
CREATE TABLE IF NOT EXISTS message (
	id                      INTEGER         PRIMARY KEY,
	communication_id        INTEGER,
	user_id                 INTEGER,
	body                    VARCHAR(3000),
	seen                    JSON,
	date_created            VARCHAR(200)
);
CREATE TABLE IF NOT EXISTS meals (
	id                      INTEGER         PRIMARY KEY,
	meals                   VARCHAR(3000),
	date                    VARCHAR(200),
	meal_title              VARCHAR(3000),
	price                   FLOAT,
	orders                  JSON,
	is_limited              BOOLEAN,
	order_limit             INTEGER,
	is_vegan                BOOLEAN,
	is_vegetarian           BOOLEAN,
	is_lactose_free         BOOLEAN,
	block_orders            BOOLEAN
);
`
