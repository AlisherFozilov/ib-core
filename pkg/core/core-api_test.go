package core

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	_ "github.com/mattn/go-sqlite3"
	"testing"
)

func Test_exportClientsToJSON(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("can't open db: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			t.Fatalf("can't close db: %v", err)
		}
	}()
	_, err = db.Exec(clientsDDL)
	if err != nil {
		t.Fatal(err)
	}

	_, err = db.Exec(`INSERT INTO clients (login, password, name, phone)
VALUES ('loginOne', 'secret1', 'Alisher', '123'),
       ('loginTwo', 'secret2', 'Fozilov', '456');
`)
	if err != nil {
		t.Fatal(err)
	}

	err = exportClientsToJSON(db)
	if err != nil {
		t.Error(err)
	}
}

func Test_importClientsFromJSON(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("can't open db: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			t.Fatalf("can't close db: %v", err)
		}
	}()
	_, err = db.Exec(clientsDDL)
	if err != nil {
		t.Fatal(err)
	}
	err = importClientsFromJSON(db)
	if err != nil {
		t.Error(err)
	}

	rows, err := db.Query(`SELECT *
FROM clients;`)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()

	client := Client{}
	for rows.Next() {
		err = rows.Scan(&client.Id, &client.Login, &client.Password,
			&client.Name, &client.Phone)
		if err != nil {
			t.Error(err)
		}
		fmt.Println(client)
	}
}