-- Skema awal SIAKAD Mini. Dipakai oleh Tahap 2; urutan file menentukan
-- urutan eksekusi migration.

CREATE TABLE users (
    id          BIGSERIAL    PRIMARY KEY,
    email       TEXT         NOT NULL UNIQUE,
    password    TEXT         NOT NULL,
    role        TEXT         NOT NULL CHECK (role IN ('admin', 'mahasiswa')),
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT now()
);

CREATE TABLE students (
    id             BIGSERIAL     PRIMARY KEY,
    user_id        BIGINT        NOT NULL UNIQUE REFERENCES users(id) ON DELETE RESTRICT,
    nim            TEXT          NOT NULL UNIQUE CHECK (nim ~ '^[0-9]{12}$'),
    nama           TEXT          NOT NULL,
    prodi          TEXT          NOT NULL,
    angkatan       INT             NOT NULL CHECK (angkatan BETWEEN 1000 AND 9999),
    ipk_terakhir   NUMERIC(3,2)  CHECK (ipk_terakhir IS NULL OR (ipk_terakhir >= 0 AND ipk_terakhir <= 4)),
    deleted_at     TIMESTAMPTZ,
    created_at     TIMESTAMPTZ   NOT NULL DEFAULT now()
);

CREATE INDEX students_prodi_angkatan_active_idx
    ON students (prodi, angkatan)
    WHERE deleted_at IS NULL;

CREATE TABLE courses (
    id        BIGSERIAL   PRIMARY KEY,
    kode_mk   TEXT        NOT NULL UNIQUE,
    nama_mk   TEXT        NOT NULL,
    sks       INT         NOT NULL CHECK (sks > 0),
    semester  INT         NOT NULL CHECK (semester > 0),
    kuota     INT         NOT NULL CHECK (kuota >= 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE enrollments (
    id              BIGSERIAL   PRIMARY KEY,
    student_id      BIGINT      NOT NULL REFERENCES students(id) ON DELETE RESTRICT,
    course_id       BIGINT      NOT NULL REFERENCES courses(id) ON DELETE RESTRICT,
    tahun_akademik  TEXT        NOT NULL
        CHECK (tahun_akademik ~ '^[0-9]{4}/[0-9]{4}-(Ganjil|Genap)$'),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (student_id, course_id, tahun_akademik)
);

CREATE INDEX enrollments_student_id_idx ON enrollments (student_id);
CREATE INDEX enrollments_course_id_idx  ON enrollments (course_id);