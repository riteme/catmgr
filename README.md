# CATMGR

## Database Schemas

```
User(user_id, type_id, name, token)
UserType(type_id, type_name, can_update, can_adduser, can_borrow, can_inspect)
Book(book_id, title, author, isbn, available_count, comment)
Record(record_id, user_id, book_id, return_date, borrow_date, deadline, final_deadline)
```

## Commands

* update: update book information (add/remove)
* adduser: add new user account
* show: checkout book information
* borrow: interactive? cli to borrow book
* extend: extend deadline
* return: interactive? cli to return book
* list: full information about one account records (may have reformatting options)

search?
