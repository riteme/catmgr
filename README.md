# CATMGR

## Prerequisites

* MySQL / MariaDB
* `catmgrd`
    * Go 1.14
    * make (<https://www.gnu.org/software/make/>)
    * Go MySQL driver (<https://github.com/go-sql-driver/mysql>)
* `catmgr-cli` (`catmgr-cli.py`)
    * Python 3.7+
    * click (<https://click.palletsprojects.com/en/7.x/>)
    * requests (<https://requests.readthedocs.io/en/master/>)

## Setup

### Database

Execute `sql/setup.sql` or `sql/setup_test.sql`.

```
mysql < sql/setup_test.sql
```

Default database name for `sql/setup.sql` is `library`. `sql/setup_test.sql` is used for test only, which contains sample records in `sql/samples.sql`. The default database name for test is `library_test`.

CAUTION: `sql/setup_test.sql` will drop database `library_test`!

### `catmgrd`

Build & run the server:

```
make serve
```

If you just want to build the executable binary file, type `make` instead. The server program is placed in `build` directory.

`catmgrd` requires a config file named `catmgrd.json` in the working directory:

```json
{
    "username": "root",
    "password": "root",
    "protocol": "tcp",
    "address": "localhost",
    "port": 3306,
    "database": "library_test"
}
```

where you can fill the username and password in the first two fileds. By default, `catmgrd` will listen the local port 10777 (i.e. `localhost:10777`), you can specify the listen address in command line:

```
./build/catmgrd -listen :12345  # listen on port 12345
```

### `catmgr-cli`

The file `catmgr-cli.py` is a simple CLI interface to communicate with `catmgrd` server. Type `python3 catmgr-cli.py --help` to see all available commands:

```
Usage: catmgr-cli.py [OPTIONS] COMMAND [ARGS]...

Options:
  -v, --verbose  Show more information.
  --help         Show this message and exit.

Commands:
  adduser  Add a new user.
  borrow   Borrow a book.
  extend   Extend deadline.
  list     List borrow history.
  new      Add a new book.
  return   Return a book.
  show     Search for books.
  update   Update book information.
```

Normally, `catmgr-cli` try to connect to `localhost:10777`, which can be overriden by config file `catmgr.json` in the working directory:

```json
{
    "server_url": "http://localhost:12345",
    "user": "root",
    "password": "root"
}
```

The config file `catmgr.json` is not required. You can provide default user name and password so that you don't type them every time an authentication is required.

## Unit Tests

Run unit tests for `catmgrd` (`go test`):

```
make test
```

You may need to setup `library_test` first to pass all unit tests.

## Security

NO SECURITY. User names and passwords are not encrypted during authentication for simplicity. HTTPS may help.

TODO: fix it.

## Database Schemas

```
User(user_id, type_id, name, token)
UserType(type_id, type_name, can_update, can_adduser, can_borrow, can_inspect)
Book(book_id, title, author, isbn, available_count, description, comment)
Record(record_id, user_id, book_id, return_date, borrow_date, deadline, final_deadline)
```

See `sql/create_tables.sql` for details.

## NOTE

This project has nothing to do with cats. It's a book/library management system.