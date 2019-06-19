DROP TABLE IF EXISTS students;

CREATE TABLE students(
    ID SERIAL PRIMARY KEY,
    FIRSTNAME CHAR(70) NOT NULL,
    LASTNAME CHAR(70) NOT NULL,
    DEPARTMENT CHAR(70) NOT NULL,
    GPA REAL NOT NULL
);

INSERT INTO students (FIRSTNAME,LASTNAME,DEPARTMENT,GPA) VALUES ('Mark', 'Johnson', 'Software Engineering', 3.3), ('Sofi', 'Hawley-Weld', 'Art', 4.0), ('Dmitrii', 'Shostak', 'Software Engineering', 4.0);
