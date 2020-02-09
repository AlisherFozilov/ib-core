package core

const (
	managersDDL = `
CREATE TABLE IF NOT EXISTS managers
(
    id INTEGER PRIMARY KEY AUTOINCREMENT ,
    login TEXT NOT NULL UNIQUE ,
    password TEXT NOT NULL
);`

	clientsDDL = `
CREATE TABLE IF NOT EXISTS clients
(
    id INTEGER PRIMARY KEY AUTOINCREMENT ,
    login TEXT NOT NULL UNIQUE ,
    password TEXT NOT NULL ,
	name TEXT NOT NULL ,
	phone TEXT NOT NULL
);`

	managersInitData = `
INSERT INTO managers 
VALUES (1, 'admin', 'top-secret')
ON CONFLICT DO NOTHING ;`


)
