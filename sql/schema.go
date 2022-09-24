package sql

const schema string = `
CREATE OR REPLACE FUNCTION update_changetimestamp_column()
    RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TABLE IF NOT EXISTS users (
	id                       UUID           PRIMARY KEY     DEFAULT gen_random_uuid(),
	email                    VARCHAR(250)   NOT NULL,
	pass                     VARCHAR(250)   NOT NULL,
	name                     VARCHAR(250)   NOT NULL,
	role                     VARCHAR(50)    NOT NULL,
	birth_certificate_number VARCHAR(200),
	birthday                 VARCHAR(200),
	country_of_birth         VARCHAR(200),
	city_of_birth            VARCHAR(200),
	login_token              VARCHAR(400),
	users                    VARCHAR(200)   DEFAULT('[]'),
	is_passing               BOOLEAN,
	
	created_at               TIMESTAMP      NOT NULL DEFAULT now(),
	updated_at               TIMESTAMP      NOT NULL DEFAULT now()
);
CREATE TABLE IF NOT EXISTS classes (
	id                       UUID           PRIMARY KEY     DEFAULT gen_random_uuid(),
	students                 JSON           DEFAULT('[]'),
	name                     VARCHAR(100)   NOT NULL,
	class_year               VARCHAR(20)    DEFAULT(''),
	last_school_date         INTEGER,
	teacher                  UUID,
	sok                      INTEGER,
	eok                      INTEGER,
	
	created_at               TIMESTAMP      NOT NULL DEFAULT now(),
	updated_at               TIMESTAMP      NOT NULL DEFAULT now(),
                                   
    CONSTRAINT FK_ClassesTeacher FOREIGN KEY (teacher) REFERENCES users(id)
);
CREATE TABLE IF NOT EXISTS subject (
	id                      UUID           PRIMARY KEY     DEFAULT gen_random_uuid(),
	teacher_id              UUID,
	name                    VARCHAR(200),
	long_name               VARCHAR(200),
	inherits_class          BOOLEAN,
	realization             FLOAT,
	location                VARCHAR(100)    NOT NULL,
	class_id                UUID,
	students                JSON            DEFAULT('[]'),
	selected_hours          FLOAT           DEFAULT(1.0),
	color                   VARCHAR(10),
	is_graded               BOOLEAN,
	
	created_at              TIMESTAMP      NOT NULL DEFAULT now(),
	updated_at              TIMESTAMP      NOT NULL DEFAULT now(),

	CONSTRAINT FK_SubjectTeacher FOREIGN KEY (teacher_id) REFERENCES users(id),
	CONSTRAINT FK_SubjectClass   FOREIGN KEY (class_id)   REFERENCES classes(id)
);
CREATE TABLE IF NOT EXISTS testing (
	id                       UUID           PRIMARY KEY     DEFAULT gen_random_uuid(),
	user_id                  UUID           NOT NULL,
	date                     VARCHAR(250)   NOT NULL,
	teacher_id               UUID           NOT NULL,
	class_id                 UUID           NOT NULL,
	result                   VARCHAR(250)   NOT NULL,
	
	created_at               TIMESTAMP      NOT NULL DEFAULT now(),
	updated_at               TIMESTAMP      NOT NULL DEFAULT now(),
	
	CONSTRAINT FK_TestingUser    FOREIGN KEY (user_id) REFERENCES users(id),
	CONSTRAINT FK_TestingTeacher FOREIGN KEY (teacher_id) REFERENCES users(id),
	CONSTRAINT FK_TestingClass   FOREIGN KEY (class_id) REFERENCES classes(id)
);
CREATE TABLE IF NOT EXISTS meetings (
	id                      UUID            PRIMARY KEY     DEFAULT gen_random_uuid(),
	meeting_name            VARCHAR(200)    NOT NULL,
	url                     VARCHAR(300)    NOT NULL,
	details                 VARCHAR(1000)   NOT NULL,
	teacher_id              UUID            NOT NULL,
	subject_id              UUID            NOT NULL,
	hour                    INTEGER         NOT NULL,
	date                    VARCHAR(20)     NOT NULL,
	location                VARCHAR(100)    NOT NULL,
	is_mandatory            BOOLEAN         NOT NULL,
	is_grading              BOOLEAN         NOT NULL,
	is_written_assessment   BOOLEAN,
	is_test                 BOOLEAN         NOT NULL,
	is_substitution         BOOLEAN         NOT NULL,
	is_beta                 BOOLEAN         NOT NULL,
	
	created_at              TIMESTAMP      NOT NULL DEFAULT now(),
	updated_at              TIMESTAMP      NOT NULL DEFAULT now(),

	CONSTRAINT FK_MeetingTeacher FOREIGN KEY (teacher_id) REFERENCES users(id),
	CONSTRAINT FK_MeetingSubject FOREIGN KEY (subject_id) REFERENCES subject(id)
);
CREATE TABLE IF NOT EXISTS absence (
	id                      UUID           PRIMARY KEY     DEFAULT gen_random_uuid(),
	user_id                 UUID,
	meeting_id              UUID,
	teacher_id              UUID,
	absence_type            VARCHAR(200),
	is_excused              BOOLEAN,
	
	created_at              TIMESTAMP      NOT NULL DEFAULT now(),
	updated_at              TIMESTAMP      NOT NULL DEFAULT now(),

	CONSTRAINT FK_AbsenceUser    FOREIGN KEY (user_id) REFERENCES users(id),
	CONSTRAINT FK_AbsenceMeeting FOREIGN KEY (meeting_id) REFERENCES meetings(id),
	CONSTRAINT FK_AbsenceTeacher FOREIGN KEY (teacher_id) REFERENCES users(id)
);
CREATE TABLE IF NOT EXISTS grades (
	id                      UUID           PRIMARY KEY     DEFAULT gen_random_uuid(),
	user_id                 UUID,
	teacher_id              UUID,
	subject_id              UUID,
	date                    VARCHAR(200),
	is_written              BOOLEAN,
	grade                   INTEGER,
	period                  INTEGER,
	is_final                BOOLEAN,
	can_patch               BOOLEAN,
	description             VARCHAR(200),
	
	created_at              TIMESTAMP      NOT NULL DEFAULT now(),
	updated_at              TIMESTAMP      NOT NULL DEFAULT now(),

	CONSTRAINT FK_GradesUser    FOREIGN KEY (user_id)    REFERENCES users(id),
	CONSTRAINT FK_GradesTeacher FOREIGN KEY (teacher_id) REFERENCES users(id),
	CONSTRAINT FK_GradesSubject FOREIGN KEY (subject_id) REFERENCES subject(id)
);
CREATE TABLE IF NOT EXISTS homework (
	id                      UUID           PRIMARY KEY     DEFAULT gen_random_uuid(),
	teacher_id              UUID,
	subject_id              UUID,
	name                    VARCHAR(200),
	description             VARCHAR(1000),
	from_date               VARCHAR(200),
	to_date                 VARCHAR(200),
	
	created_at              TIMESTAMP      NOT NULL DEFAULT now(),
	updated_at              TIMESTAMP      NOT NULL DEFAULT now(),

    CONSTRAINT FK_HomeworkSubject FOREIGN KEY (subject_id) REFERENCES subject(id),
    CONSTRAINT FK_HomeworkTeacher FOREIGN KEY (teacher_id) REFERENCES users(id)
);
CREATE TABLE IF NOT EXISTS student_homework (
	id                      UUID           PRIMARY KEY     DEFAULT gen_random_uuid(),
	user_id                 UUID,
	homework_id             UUID,
	status                  VARCHAR(200),
	
	created_at              TIMESTAMP      NOT NULL DEFAULT now(),
	updated_at              TIMESTAMP      NOT NULL DEFAULT now(),

	CONSTRAINT FK_StudentHomeworkUser     FOREIGN KEY (user_id)     REFERENCES users(id),
	CONSTRAINT FK_StudentHomeworkHomework FOREIGN KEY (homework_id) REFERENCES homework(id)
);
CREATE TABLE IF NOT EXISTS communication (
	id                      UUID            PRIMARY KEY     DEFAULT gen_random_uuid(),
	people                  JSON            DEFAULT('[]'),
	title                   VARCHAR(200),
	date_created            VARCHAR(200),
	
	created_at              TIMESTAMP      NOT NULL DEFAULT now(),
	updated_at              TIMESTAMP      NOT NULL DEFAULT now()
);
CREATE TABLE IF NOT EXISTS message (
	id                      UUID           PRIMARY KEY     DEFAULT gen_random_uuid(),
	communication_id        UUID,
	user_id                 UUID,
	body                    VARCHAR(3000),
	seen                    JSON,
	date_created            VARCHAR(200),
	
	created_at              TIMESTAMP      NOT NULL DEFAULT now(),
	updated_at              TIMESTAMP      NOT NULL DEFAULT now(),
    
    CONSTRAINT FK_MessageCommunication FOREIGN KEY (communication_id) REFERENCES communication(id),
    CONSTRAINT FK_MessageUser          FOREIGN KEY (user_id)          REFERENCES users(id)
);
CREATE TABLE IF NOT EXISTS meals (
	id                      UUID           PRIMARY KEY     DEFAULT gen_random_uuid(),
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
	block_orders            BOOLEAN,
	
	created_at              TIMESTAMP      NOT NULL DEFAULT now(),
	updated_at              TIMESTAMP      NOT NULL DEFAULT now()
);
CREATE TABLE IF NOT EXISTS notifications (
	id                      UUID           PRIMARY KEY     DEFAULT gen_random_uuid(),
	notification            VARCHAR(3000),
	
	created_at              TIMESTAMP      NOT NULL DEFAULT now(),
	updated_at              TIMESTAMP      NOT NULL DEFAULT now()
);
CREATE TABLE IF NOT EXISTS improvements (
	id                      UUID           PRIMARY KEY     DEFAULT gen_random_uuid(),
	message                 VARCHAR(3000),
	student_id              UUID,
    meeting_id              UUID,
	teacher_id              UUID,
	
	created_at              TIMESTAMP      NOT NULL DEFAULT now(),
	updated_at              TIMESTAMP      NOT NULL DEFAULT now(),
    
	CONSTRAINT FK_ImprovementsStudent FOREIGN KEY (student_id) REFERENCES users(id),
	CONSTRAINT FK_ImprovementsMeeting FOREIGN KEY (meeting_id) REFERENCES meetings(id),
	CONSTRAINT FK_ImprovementsTeacher FOREIGN KEY (teacher_id) REFERENCES users(id)
);

-- Documents are special and don't use UUID type due to small space on some documents
CREATE TABLE IF NOT EXISTS documents (
    id                      VARCHAR(50)     PRIMARY KEY,
    exported_by             UUID            NOT NULL,
    document_type           INTEGER         NOT NULL,
    is_signed               BOOLEAN         NOT NULL,
    
    created_at              TIMESTAMP      NOT NULL DEFAULT now(),
	updated_at              TIMESTAMP      NOT NULL DEFAULT now(),
    
    CONSTRAINT FK_DocumentsExporter FOREIGN KEY (exported_by) REFERENCES users(id)
);

CREATE OR REPLACE TRIGGER update_users_updated_at BEFORE UPDATE ON users FOR EACH ROW EXECUTE PROCEDURE update_changetimestamp_column();
CREATE OR REPLACE TRIGGER update_classes_updated_at BEFORE UPDATE ON classes FOR EACH ROW EXECUTE PROCEDURE update_changetimestamp_column();
CREATE OR REPLACE TRIGGER update_testing_updated_at BEFORE UPDATE ON testing FOR EACH ROW EXECUTE PROCEDURE update_changetimestamp_column();
CREATE OR REPLACE TRIGGER update_meetings_updated_at BEFORE UPDATE ON meetings FOR EACH ROW EXECUTE PROCEDURE update_changetimestamp_column();
CREATE OR REPLACE TRIGGER update_absence_updated_at BEFORE UPDATE ON absence FOR EACH ROW EXECUTE PROCEDURE update_changetimestamp_column();
CREATE OR REPLACE TRIGGER update_grades_updated_at BEFORE UPDATE ON grades FOR EACH ROW EXECUTE PROCEDURE update_changetimestamp_column();
CREATE OR REPLACE TRIGGER update_subject_updated_at BEFORE UPDATE ON subject FOR EACH ROW EXECUTE PROCEDURE update_changetimestamp_column();
CREATE OR REPLACE TRIGGER update_student_homework_updated_at BEFORE UPDATE ON student_homework FOR EACH ROW EXECUTE PROCEDURE update_changetimestamp_column();
CREATE OR REPLACE TRIGGER update_homework_updated_at BEFORE UPDATE ON homework FOR EACH ROW EXECUTE PROCEDURE update_changetimestamp_column();
CREATE OR REPLACE TRIGGER update_communication_updated_at BEFORE UPDATE ON communication FOR EACH ROW EXECUTE PROCEDURE update_changetimestamp_column();
CREATE OR REPLACE TRIGGER update_message_updated_at BEFORE UPDATE ON message FOR EACH ROW EXECUTE PROCEDURE update_changetimestamp_column();
CREATE OR REPLACE TRIGGER update_meals_updated_at BEFORE UPDATE ON meals FOR EACH ROW EXECUTE PROCEDURE update_changetimestamp_column();
CREATE OR REPLACE TRIGGER update_notifications_updated_at BEFORE UPDATE ON notifications FOR EACH ROW EXECUTE PROCEDURE update_changetimestamp_column();
CREATE OR REPLACE TRIGGER update_improvements_updated_at BEFORE UPDATE ON improvements FOR EACH ROW EXECUTE PROCEDURE update_changetimestamp_column();
CREATE OR REPLACE TRIGGER update_documents_updated_at BEFORE UPDATE ON documents FOR EACH ROW EXECUTE PROCEDURE update_changetimestamp_column();

`
