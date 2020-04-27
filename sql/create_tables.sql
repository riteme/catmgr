CREATE TABLE IF NOT EXISTS UserType(
    type_id INT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    type_name VARCHAR(32),
    can_update BOOLEAN NOT NULL DEFAULT false,
    can_adduser BOOLEAN NOT NULL DEFAULT false,
    can_borrow BOOLEAN NOT NULL DEFAULT false,
    can_inspect BOOLEAN NOT NULL DEFAULT false
);

CREATE TABLE IF NOT EXISTS User(
    user_id INT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    type_id INT NOT NULL,
    name VARCHAR(128) NOT NULL,
    token CHAR(40) NOT NULL,
    FOREIGN KEY (type_id)
        REFERENCES UserType(type_id)
);

CREATE TABLE IF NOT EXISTS Book(
    book_id INT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    title VARCHAR(256) NOT NULL,
    author VARCHAR(128) NOT NULL,
    isbn VARCHAR(128) NOT NULL,
    available_count INT NOT NULL CHECK(available_count >= 0),
    comment TEXT
);

CREATE TABLE IF NOT EXISTS Record(
    record_id INT NOT NULL AUTO_INCREMENT,
    user_id INT NOT NULL,
    book_id INT NOT NULL,
    return_date DATE DEFAULT NULL,
    borrow_date DATE NOT NULL,
    deadline DATE NOT NULL CHECK(deadline >= borrow_date),
    final_deadline DATE NOT NULL CHECK(final_deadline >= deadline),
    PRIMARY KEY (record_id, user_id, book_id),
    FOREIGN KEY (user_id)
        REFERENCES User(user_id),
    FOREIGN KEY (book_id)
        REFERENCES Book(book_id)
);