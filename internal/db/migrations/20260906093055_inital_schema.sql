-- +goose Up
CREATE TABLE
    journal (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        journalname TEXT UNIQUE NOT NULL,
        password TEXT NOT NULL,
        created TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
        modified TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
    );

CREATE TABLE
    tag (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        title TEXT UNIQUE NOT NULL
    );

CREATE TABLE
    memory (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        journal_id INTEGER NOT NULL,
        memorydate TIMESTAMP NOT NULL,
        title TEXT NOT NULL,
        body TEXT NOT NULL,
        created TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
        modified TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
        FOREIGN KEY (journal_id) REFERENCES journal (id) ON DELETE CASCADE
    );

CREATE TABLE
    memory_tag (
        journal_id INTEGER NOT NULL,
        memory_id INTEGER NOT NULL,
        tag_id INTEGER NOT NULL,
        PRIMARY KEY (journal_id, memory_id, tag_id),
        FOREIGN KEY (journal_id) REFERENCES journal (id) ON DELETE CASCADE,
        FOREIGN KEY (memory_id) REFERENCES memory (id) ON DELETE CASCADE,
        FOREIGN KEY (tag_id) REFERENCES tag (id) ON DELETE CASCADE
    );

-- +goose Down
DROP TABLE memory_tag;

DROP TABLE memory;

DROP TABLE tag;

DROP TABLE journal;