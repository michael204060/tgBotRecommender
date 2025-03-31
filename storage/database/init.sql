CREATE TABLE if not exists users (
    id SERIAL PRIMARY KEY,
    sender INT NOT NULL
);

CREATE TABLE if not exists messages (
    id SERIAL PRIMARY KEY,
    content VARCHAR(255),
    priority INT,
    user_id INT NOT NULL,
    flag int,
    FOREIGN KEY (user_id) REFERENCES users(id)
);


