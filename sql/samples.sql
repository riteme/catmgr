INSERT INTO UserType
    (type_name, can_update, can_adduser, can_borrow, can_inspect)
VALUES
    ("root", true, true, true, true),
    ("admin", true, true, false, true),
    ("student", false, false, true, false),
    ("guest", false, false, false, false);

INSERT INTO User
    (type_id, name, token)
VALUES
    (1, "root", "dc76e9f0c0006e8f919e0c515c66dbba3982f785"),    # root
    (2, "admin", "d033e22ae348aeb5660fc2140aec35850c4da997"),   # admin
    (3, "riteme", "7c4a8d09ca3762af61e59520943dc26494f8941b"),  # 123456
    (3, "nano", "7c4a8d09ca3762af61e59520943dc26494f8941b"),    # 123456
    (4, "cxk", "dd5fef9c1c1da1394d6d34b248c51be2ad740840")      # 654321