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
    (4, "cxk", "dd5fef9c1c1da1394d6d34b248c51be2ad740840"),     # 654321
    (3, "ayaya", "7c4a8d09ca3762af61e59520943dc26494f8941b"),   # 123456
    (3, "lemon", "7c4a8d09ca3762af61e59520943dc26494f8941b");   # 123456

-- ALL FROM springer.com
INSERT INTO Book
    (title, author, isbn, available_count, description, comment)
VALUES
    ("Monte Carlo Methods", "Barbu, Adrian, Zhu, Song-Chun", "978-981-13-2971-5", 2, NULL, NULL),
    ("Compiler Design", "Hack, Sebastian, Wilhelm, Reinhard, Seidl, Helmut", "978-3-642-17637-1", 1, NULL, NULL),
    ("Energy Internet", "Zobaa, Ahmed F, Cao, Junwei (Eds.)", "978-3-030-45452-4", 5, "Provides an ideal resource for students in advanced graduate-level courses and special topics in energy, information and control systems", "5 books"),
    ("Systems Benchmarking", "Kounev, Samuel, Lange, Klaus-Dieter, von Kistowski, Jóakim", "978-3-030-41704-8", 3, "Provides theoretical and practical foundations as well as an in-depth look at modern benchmarks and benchmark development", NULL),
    ("Database Design and Implementation", "Sciore, Edward", "978-3-030-33836-7", 99, NULL, "too many books!"),
    ("Mathematical Modeling and Computational Tools", "Bhattacharya, Somnath, Kumar, Jitendra, Ghoshal, Koeli (Eds.)", "978-981-15-3615-1", 23, "Collects a wide-range of topics in mathematics, statistics, engineering, healthcare, and their applications", "23 books"),
    ("Foundations of Software Science and Computation Structures", "Goubault-Larrecq, Jean, König, Barbara (Eds.)", "978-3-030-45231-5", 1, NULL, "open access"),
    ("Cornerstones", "Birkhäuser Boston", "2197-182X", 2, "Cornerstones comprises textbooks that focus on what students need to know and what faculty should teach regarding various selected topics in pure and applied mathematics and related subjects. Aimed at aspiring young mathematicians at the advanced undergraduate to the second-year graduate level, books that appear in this series are intended to serve as the definitive advanced texts for the next generation of mathematicians. By enlisting only expert mathematicians and leading researchers in each field who are top-notch expositors with established track records, Cornerstones volumes are models of clarity that provide authoritative modern treatments of the essential subjects of pure and applied mathematics while capturing the beauty and excitement of mathematics for the reader. The Series Editors themselves are accomplished researchers with considerable writing experience, and seek to infuse each text with excellence and purpose through a collaborative, yet highly rigorous selection and reviewing protocol.", NULL),
    ("Principles of Mathematics for Economics", "Cerreia-Vioglio, Simone, Marinacci, Massimo, Vigna, Elena", "978-3-319-44715-5", 5, NULL, NULL),
    ("Algebra for Applications", "Slinko, Arkadii", "978-3-030-44073-2", 4, "Suitable for an undergraduate applied algebra course", NULL),
    ("Fundamental Mathematical Analysis", "Magnus, Robert", "978-3-030-46321-2", 0, "Recognises and addresses student difficulties", "lost"),
    ("A Course in Algebraic Error-Correcting Codes", "Ball, Simeon", "978-3-030-41152-7", 6, NULL, "aha!"),
    ("Graph Theory", "Diestel, Reinhard", "978-3-662-53622-3", 1, "Standard textbook of modern graph theory", NULL),
    ("Computational Geometry and Graph Theory", "Ito, H., Kano, M., Katoh, N., Uno, Y. (Eds.)", "978-3-540-89550-3", 1, NULL, "conference"),
    ("Graph Theory", "Bollobas, Bela", "978-1-4612-9967-7", 2, NULL, NULL),
    ("Graph Theory", "Gera, Ralucca, Hedetniemi, Stephen, Larson, Craig (Eds.)", "978-3-319-31940-7", 5, "Describes the origin and history behind conjectures and problems in graph theory", NULL),
    ("Graph Theory and Applications", "Alavi, Y., Lick, D. R., White, A. T. (Eds.)", "978-3-540-38114-3", 1, "Proceedings of the Conference at Western Michigan University, May 10 - 13, 1972", "conference"),
    ("Computational Graph Theory", "Tinhofer, G., Mayr, E.W., Noltemeier, H., Syslo, M.M., Albrecht, R. (Eds.)", "978-3-7091-9076-0", 0, "One ofthe most important aspects in research fields where mathematics is applied is the construction of a formal model of a real system. As for structural relations, graphs have turned out to provide the most appropriate tool for setting up the mathematical model. This is certainly one of the reasons for the rapid expansion in graph theory during the last decades. Furthermore, in recent years it also became clear that the two disciplines of graph theory and computer science have very much in common, and that each one has been capable of assisting significantly in the development of the other. On one hand, graph theorists have found that many of their problems can be solved by the use of com­ puting techniques, and on the other hand, computer scientists have realized that many of their concepts, with which they have to deal, may be conveniently expressed in the lan­ guage of graph theory, and that standard results in graph theory are often very relevant to the solution of problems concerning them. As a consequence, a tremendous number of publications has appeared, dealing with graphtheoretical problems from a computational point of view or treating computational problems using graph theoretical concepts.", "lost"),
    ("Basic Graph Theory", "Rahman, Md. Saidur", "978-3-319-49475-3", 3, "This undergraduate textbook provides an introduction to graph theory, which has numerous applications in modeling problems in science and technology, and has become a vital component to computer science, computer science and engineering, and mathematics curricula of universities all over the world.", NULL),
    ("Combinatorics and Graph Theory", "Harris, John M., Hirst, Jeffry L., Mossinghoff, Michael J.", "978-1-4757-4803-1", 8, NULL, NULL),
    ("Graph Theory and Algorithms", "Saito, N., Nishizeki, T. (Eds.)", "978-3-540-10704-0", 6, "17th Symposium of Research Institute of Electrical Communication, Tohoku University, Sendai, Japan, October 24-25, 1980. Proceedings", NULL),
    ("Algebraic Graph Theory", "Godsil, Chris, Royle, Gordon F.", "978-1-4613-0163-9", 3, NULL, "no description"),
    ("Ten Applications of Graph Theory", "Walther, Hansjoachim", "978-94-009-7154-7", 2, "Growing specialization and diversification have brought a host of monographs and textbooks on increasingly specialized topics. However, the \"tree\" of knowledge of mathematics and related fields does not grow only by putting forth new bran­ ches. It also happens, quite often in fact, that branches which were thought to be completely disparate are suddenly seen to be related. Further, the kind and level of sophistication of mathematics applied in various sciences has changed drastically in recent years: measure theory is used (non-tri­ vially) in regional and theoretical economics; algebraic geometry interacts with physics; the Minkowsky lemma, coding theory and the structure of water meet one another in packing and covering theory; quantum fields, crystal defects and mathematical programming profit from homotopy theory; Lie algebras are relevant to filtering; and prediction and electrical engineering can use Stein spaces. And in addition to this there are such new emerging subdisciplines as \"completely integrable systems\", \"chaos, synergetics and large-scale order\", which are almost impossible to fit into the existing classification schemes. They draw upon widely different sections of mathematics. This program, Mathematics and Its Applications, is devoted to such (new) interrelations as exempla gratia: - a central concept which plays an important role in several different mathe­ matical and/or scientific specialized areas; - new applications of the results and ideas from one area of scientific endeavor into another; - influences which the results, problems and concepts of one field of enquiry have and have had on the development of another.", NULL),
    ("Graph Drawing", "Whitesides, Sue H. (Ed.)", "978-3-540-37623-1", 6, "6th International Symposium, GD '98 Montreal, Canada, August 13-15, 1998 Proceedings", "conference"),
    ("Graph Drawing", "Kratochvil, Jan (Ed.)", "978-3-540-46648-2", 1, "7th International Symposium, GD'99, Stirin Castle, Czech Republic, September 15-19, 1999 Proceedings", NULL),
    ("Encyclopedia of Algorithms", "Kao, Ming-Yang (Ed.)", "978-1-4939-2865-1", 1, "Covers a wealth of problems currently relevant in diverse fields including biology, economics, financial software and computer science, amongst others", "TOO EXPENSIVE!");

INSERT INTO Record
    (user_id, book_id, borrow_date, deadline, final_deadline)
VALUES
    (6, 1, "1926-08-17", "1926-09-17", "2020-02-02"),
    (6, 1, "1926-08-17", "1926-09-17", "2020-02-02"),
    (6, 1, "1926-08-17", "1926-09-17", "2020-02-02"),
    (7, 2, "2020-03-14", "2020-03-15", "2020-03-16"),
    (7, 3, "2020-03-14", "2020-03-15", "2020-03-16"),
    (7, 4, "2020-03-14", "2020-03-15", "2020-03-16"),
    (7, 5, "2020-03-14", "2020-03-15", "2020-03-16"),
    (4, 9, "2020-03-13", "2999-09-26", "2999-09-26"),
    (4, 10, "2020-03-13", "2999-09-26", "2999-09-26"),
    (4, 11, "2020-03-13", "2999-09-26", "2999-09-26"),
    (4, 12, "2020-03-13", "2999-09-26", "2999-09-26"),
    (4, 13, "2020-03-13", "2999-09-26", "2999-09-26");