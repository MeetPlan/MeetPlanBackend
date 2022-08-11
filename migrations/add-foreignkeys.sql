ALTER TABLE users ADD PRIMARY KEY (id);
ALTER TABLE testing ADD PRIMARY KEY (id);
ALTER TABLE classes ADD PRIMARY KEY (id);
ALTER TABLE meetings ADD PRIMARY KEY (id);
ALTER TABLE absence ADD PRIMARY KEY (id);
ALTER TABLE grades ADD PRIMARY KEY (id);
ALTER TABLE subject ADD PRIMARY KEY (id);
ALTER TABLE student_homework ADD PRIMARY KEY (id);
ALTER TABLE homework ADD PRIMARY KEY (id);
ALTER TABLE communication ADD PRIMARY KEY (id);
ALTER TABLE message ADD PRIMARY KEY (id);
ALTER TABLE meals ADD PRIMARY KEY (id);
ALTER TABLE notifications ADD PRIMARY KEY (id);
ALTER TABLE improvements ADD PRIMARY KEY (id);
ALTER TABLE documents ADD PRIMARY KEY (id);

ALTER TABLE testing ADD FOREIGN KEY (user_id) REFERENCES users(id);
ALTER TABLE testing ADD FOREIGN KEY (teacher_id) REFERENCES users(id);
ALTER TABLE testing ADD FOREIGN KEY (class_id) REFERENCES classes(id);

ALTER TABLE meetings ADD FOREIGN KEY (teacher_id) REFERENCES users(id);
ALTER TABLE meetings ADD FOREIGN KEY (subject_id) REFERENCES subject(id);

ALTER TABLE absence ADD FOREIGN KEY (user_id) REFERENCES users(id);
ALTER TABLE absence ADD FOREIGN KEY (meeting_id) REFERENCES meetings(id);
ALTER TABLE absence ADD FOREIGN KEY (teacher_id) REFERENCES users(id);

ALTER TABLE grades ADD FOREIGN KEY (user_id) REFERENCES users(id);
ALTER TABLE grades ADD FOREIGN KEY (teacher_id) REFERENCES users(id);
ALTER TABLE grades ADD FOREIGN KEY (subject_id) REFERENCES subject(id);

ALTER TABLE subject ADD FOREIGN KEY (teacher_id) REFERENCES users(id);
ALTER TABLE subject ADD FOREIGN KEY (class_id) REFERENCES classes(id);

ALTER TABLE student_homework ADD FOREIGN KEY (user_id) REFERENCES users(id);
ALTER TABLE student_homework ADD FOREIGN KEY (homework_id) REFERENCES homework(id);

ALTER TABLE homework ADD FOREIGN KEY (subject_id) REFERENCES subject(id);
ALTER TABLE homework ADD FOREIGN KEY (teacher_id) REFERENCES users(id);

ALTER TABLE message ADD FOREIGN KEY (communication_id) REFERENCES communication(id);
ALTER TABLE message ADD FOREIGN KEY (user_id) REFERENCES users(id);

ALTER TABLE improvements ADD FOREIGN KEY (student_id) REFERENCES users(id);
ALTER TABLE improvements ADD FOREIGN KEY (meeting_id) REFERENCES meetings(id);
ALTER TABLE improvements ADD FOREIGN KEY (teacher_id) REFERENCES users(id);

ALTER TABLE documents ADD FOREIGN KEY (exported_by) REFERENCES users(id);
