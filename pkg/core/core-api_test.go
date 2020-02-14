package core

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"io/ioutil"
	"log"
	"testing"
)

func createDBinMemory(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite3", "file::memory:?cache=shared")
	if err != nil {
		t.Fatal(err)
	}
	return db
}

func Test_exportClientsToJSON(t *testing.T) {
	db := createDBinMemory(t)
	defer db.Close()
	_, err := db.Exec(clientsDDL)
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

	bytesWant, err := ioutil.ReadFile("testData/clients.json")
	if err != nil {
		log.Fatal(err)
	}
	bytesGot, err := ioutil.ReadFile("clients.json")
	if err != nil {
		log.Fatal(err)
	}

	if !(string(bytesWant) == string(bytesGot)) {
		t.Error("Files don't match")
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

func Test_checkManagerLoginOnUnique(t *testing.T) {
	db := createDBinMemory(t)
	ok, err := checkManagerLoginOnUnique("", db)

	if ok {
		t.Error("want false, got true")
	}

	ok, err = checkManagerLoginOnUnique("abc", db)

	if ok {
		t.Error("want false, got true")
	}

	_, err = db.Exec(managersDDL)
	if err != nil {
		t.Fatal(err)
	}

	_, err = db.Exec(`INSERT INTO managers (login, password) 
VALUES ('test', 'test-password')`)
	if err != nil {
		t.Fatal(err)
	}

	ok, _ = checkManagerLoginOnUnique("test", db)

	if ok {
		t.Error("want false, got true")
	}

	ok, err = checkManagerLoginOnUnique("coolLogin", db)
	if err != nil {
		if err != sql.ErrNoRows {
			t.Errorf("want error type ErrNoRows, got: %v", err)
		}
	}
	if !ok {
		t.Error("want true, got false")
	}
}
func Test_checkClientLoginOnUnique(t *testing.T) {
	db := createDBinMemory(t)
	ok, err := checkClientLoginOnUnique("", db)

	if ok {
		t.Error("want false, got true")
	}

	ok, err = checkClientLoginOnUnique("abc", db)

	if ok {
		t.Error("want false, got true")
	}

	_, err = db.Exec(clientsDDL)
	if err != nil {
		t.Fatal(err)
	}

	_, err = db.Exec(`INSERT INTO clients (login, password, name, phone) 
VALUES ('test', '', '', '')`)
	if err != nil {
		t.Fatal(err)
	}

	ok, _ = checkClientLoginOnUnique("test", db)

	if ok {
		t.Error("want false, got true")
	}

	ok, err = checkClientLoginOnUnique("coolLogin", db)
	if err != nil {
		if err != sql.ErrNoRows {
			t.Errorf("want error type ErrNoRows, got: %v", err)
		}
	}
	if !ok {
		t.Error("want true, got false")
	}
}

func Test_addClient(t *testing.T) {
	db := createDBinMemory(t)
	defer db.Close()

	client := Client{
		Login:    "a",
		Password: "b",
		Name:     "c",
		Phone:    "d",
	}
	err := addClient(client, db)
	if err == nil {
		t.Error("want not nil error")
	}

	_, err = db.Exec(clientsDDL)
	if err != nil {
		t.Fatal(err)
	}

	err = addClient(client, db)
	if err != nil {
		t.Error("want nil error")
	}

	err = db.QueryRow(`SELECT * FROM clients`).Scan(&client.Id, &client.Login, &client.Password, &client.Name, &client.Phone)
	if err != nil {
		t.Fatal(err)
	}
	if client.Id != 1 {
		t.Errorf("want: %v, got: %v", client.Id, 1)
	}
	if client.Login != "a" {
		t.Errorf("want: %v, got: %v", client.Login, "a")
	}
	if client.Password != "b" {
		t.Errorf("want: %v, got: %v", client.Password, "b")
	}
	if client.Name != "c" {
		t.Errorf("want: %v, got: %v", client.Name, "c")
	}
	if client.Phone != "d" {
		t.Errorf("want: %v, got: %v", client.Phone, "d")
	}
}
func Test_addManager(t *testing.T) {
	db := createDBinMemory(t)
	defer db.Close()

	manager := Manager{
		Login:    "a",
		Password: "b",
	}
	err := addManager(manager, db)
	if err == nil {
		t.Error("want not nil error")
	}

	_, err = db.Exec(managersDDL)
	if err != nil {
		t.Fatal(err)
	}

	err = addManager(manager, db)
	if err != nil {
		t.Error("want nil error, got: ", err)
	}

	err = db.QueryRow(`SELECT * FROM managers`).Scan(&manager.Id, &manager.Login, &manager.Password)
	if err != nil {
		t.Fatal(err)
	}
	if manager.Id != 1 {
		t.Errorf("want: %v, got: %v", manager.Id, 1)
	}
	if manager.Login != "a" {
		t.Errorf("want: %v, got: %v", manager.Login, "a")
	}
	if manager.Password != "b" {
		t.Errorf("want: %v, got: %v", manager.Password, "b")
	}
}

func Test_addBankAccountToClient(t *testing.T) {
	db := createDBinMemory(t)
	defer db.Close()

	err := addBankAccountToClient(0, db)
	if err == nil {
		t.Error("want not nil error")
	}

	_, err = db.Exec(clientsDDL)
	if err != nil {
		t.Fatal(err)
	}

	err = addBankAccountToClient(0, db)
	if err == nil {
		t.Error("want not nil error")
	}

	client := Client{
		Id:       1,
		Login:    "test",
		Password: "test-password",
		Name:     "test-guy",
		Phone:    "123",
	}
	_, err = db.Exec(insertClientSQL,
		sql.Named("id", client.Id),
		sql.Named("password", client.Password),
		sql.Named("phone", client.Phone),
		sql.Named("name", client.Name),
		sql.Named("login", client.Login),
	)
	if err != nil {
		t.Fatal(err)
	}

	err = addBankAccountToClient(1, db)
	if err == nil {
		t.Error("want not nil error")
	}

	_, err = db.Exec(bankAccountsDDL)
	if err != nil {
		t.Fatal(err)
	}

	err = addBankAccountToClient(1, db)
	if err != nil {
		t.Error("want nil error, got: ", err)
	}

	bankAccountGot := BankAccount{}
	bankAccountWant := BankAccount{
		Id:        1,
		UserId:    1,
		AccountId: 0,
		Balance:   0,
	}
	err = db.QueryRow(getAllBankAccountsDataSQL).Scan(&bankAccountGot.Id, &bankAccountGot.UserId, &bankAccountGot.AccountId, &bankAccountGot.Balance)
	if err != nil {
		t.Fatal(err)
	}

	if bankAccountGot != bankAccountWant {
		t.Errorf("want: \n%v\ngot: \n%v\n", bankAccountWant, bankAccountGot)
	}

	err = addBankAccountToClient(1, db)
	if err != nil {
		t.Error("want nil error, got: ", err)
	}

	err = db.QueryRow(`SELECT * FROM bank_accounts
ORDER BY 1 DESC LIMIT 1`,
	).Scan(&bankAccountGot.Id, &bankAccountGot.UserId, &bankAccountGot.AccountId, &bankAccountGot.Balance)
	if err != nil {
		t.Fatal(err)
	}
	bankAccountWant.Id = 2
	bankAccountWant.AccountId = 1
	if bankAccountGot != bankAccountWant {
		t.Errorf("want: \n%v\ngot: \n%v\n", bankAccountWant, bankAccountGot)
		rows, err := db.Query(`SELECT * FROM bank_accounts`)
		if err != nil {
			t.Fatal(err)
		}
		defer rows.Close()
		for rows.Next() {
			err := rows.Scan(&bankAccountGot.Id, &bankAccountGot.UserId, &bankAccountGot.AccountId, &bankAccountGot.Balance)
			if err != nil {
				t.Fatal(err)
			}
			t.Error(bankAccountGot)
		}
	}
}
func Test_addBankAccountToService(t *testing.T) {
	db := createDBinMemory(t)
	defer db.Close()

	err := addBankAccountToService(0, db)
	if err == nil {
		t.Error("want not nil error")
	}

	_, err = db.Exec(servicesDDL)
	if err != nil {
		t.Fatal(err)
	}

	err = addBankAccountToService(0, db)
	if err == nil {
		t.Error("want not nil error")
	}

	service := Service{
		Name: "test",
	}

	_, err = db.Exec(insertServiceWithoutIdSQL,
		sql.Named("name", service.Name),
	)
	if err != nil {
		t.Fatal(err)
	}

	err = addBankAccountToService(1, db)
	if err == nil {
		t.Error("want not nil error")
	}

	_, err = db.Exec(bankAccountsServicesDDL)
	if err != nil {
		t.Fatal(err)
	}

	err = addBankAccountToService(1, db)
	if err != nil {
		t.Error("want nil error, got: ", err)
	}

	bankAccountGot := BankAccount{}
	bankAccountWant := BankAccount{
		Id:        1,
		UserId:    1,
		AccountId: 0,
		Balance:   0,
	}
	err = db.QueryRow(getAllBankAccountsServicesDataSQL).Scan(&bankAccountGot.Id, &bankAccountGot.UserId, &bankAccountGot.AccountId, &bankAccountGot.Balance)
	if err != nil {
		t.Fatal(err)
	}

	if bankAccountGot != bankAccountWant {
		t.Errorf("want: \n%v\ngot: \n%v\n", bankAccountWant, bankAccountGot)
	}

	err = addBankAccountToService(1, db)
	if err != nil {
		t.Error("want nil error, got: ", err)
	}

	err = db.QueryRow(`SELECT * FROM bank_accounts_services
ORDER BY 1 DESC LIMIT 1`,
	).Scan(&bankAccountGot.Id, &bankAccountGot.UserId, &bankAccountGot.AccountId, &bankAccountGot.Balance)
	if err != nil {
		t.Fatal(err)
	}
	bankAccountWant.Id = 2
	bankAccountWant.AccountId = 1
	if bankAccountGot != bankAccountWant {
		t.Errorf("want: \n%v\ngot: \n%v\n", bankAccountWant, bankAccountGot)
		rows, err := db.Query(`SELECT * FROM bank_accounts`)
		if err != nil {
			t.Fatal(err)
		}
		defer rows.Close()
		for rows.Next() {
			err := rows.Scan(&bankAccountGot.Id, &bankAccountGot.UserId, &bankAccountGot.AccountId, &bankAccountGot.Balance)
			if err != nil {
				t.Fatal(err)
			}
			t.Error(bankAccountGot)
		}
	}
}

func Test_addATM(t *testing.T) {
	db := createDBinMemory(t)
	defer db.Close()

	err := addATM("atm-address", db)
	if err == nil {
		t.Error("want not nil error")
	}

	_, err = db.Exec(atmsDDL)
	if err != nil {
		t.Fatal(err)
	}

	addressWant := "Rogun"
	err = addATM(addressWant, db)
	if err != nil {
		t.Error("want nil error, got: ", err)
	}

	var addressGot string

	err = db.QueryRow(getAllAtmAddressesSQL).Scan(&addressGot)
	if err != nil {
		if err == sql.ErrNoRows {
			t.Fatal("nothing added")
		}
		t.Fatal(err)
	}

	if addressGot != addressWant {
		t.Errorf("want: %v, got: %v", addressWant, addressGot)
	}

	addressWant = "Dushanbe"
	err = addATM(addressWant, db)
	if err != nil {
		t.Error("want nil error, got: ", err)
	}

	err = db.QueryRow(`SELECT address 
					FROM atms WHERE id = 2`).Scan(&addressGot)
	if err != nil {
		if err == sql.ErrNoRows {
			t.Fatal("nothing added")
		}
		t.Fatal(err)
	}

	if addressGot != addressWant {
		t.Errorf("want: %v, got: %v", addressWant, addressGot)
	}
}
func Test_addService(t *testing.T) {
	db := createDBinMemory(t)
	defer db.Close()

	service := Service{
		Name: "taxes",
	}
	err := addService(service, db)
	if err == nil {
		t.Error("want not nil error")
	}

	_, err = db.Exec(servicesDDL)
	if err != nil {
		t.Fatal(err)
	}

	err = addService(service, db)
	if err != nil {
		t.Error("want nil error, got: ", err)
	}

	serviceWant := Service{
		Id:   1,
		Name: "taxes",
	}
	serviceGot := Service{}
	err = db.QueryRow(`SELECT * FROM services`).Scan(&serviceGot.Id, &serviceGot.Name)
	if err != nil {
		if err == sql.ErrNoRows {
			t.Fatal("nothing added")
		}
		t.Fatal(err)
	}

	if serviceGot != serviceWant {
		t.Errorf("want: %v, got: %v", serviceWant, serviceGot)
	}

	service.Name = "big-taxes"
	err = addService(service, db)
	if err != nil {
		t.Error("want nil error, got: ", err)
	}

	serviceWant = Service{
		Id:   2,
		Name: "big-taxes",
	}
	err = db.QueryRow(`SELECT * FROM services WHERE id = 2`).Scan(&serviceGot.Id, &serviceGot.Name)
	if err != nil {
		if err == sql.ErrNoRows {
			t.Fatal("nothing added")
		}
		t.Fatal(err)
	}

	if serviceGot != serviceWant {
		t.Errorf("want: %v, got: %v", serviceWant, serviceGot)

		rows, err := db.Query(`SELECT * FROM services`)
		if err != nil {
			t.Fatal(err)
		}
		defer rows.Close()
		for rows.Next() {
			err := rows.Scan(&serviceGot.Id, &serviceGot.Name)
			if err != nil {
				t.Fatal(err)
			}
			t.Error(serviceGot)
		}
	}
}

func Test_getClientIdByLogin(t *testing.T) {
	db := createDBinMemory(t)
	defer db.Close()

	_, err := getClientIdByLogin("test", db)
	if err == nil {
		t.Error("want not nil error")
	}

	_, err = db.Exec(clientsDDL)
	if err != nil {
		t.Fatal(err)
	}

	_, err = getClientIdByLogin("test", db)
	if err == nil {
		t.Error("want not nil error")
	}

	client := Client{
		Login:    "abc",
		Password: "def",
		Name:     "ghi",
		Phone:    "xyz",
	}
	_, err = db.Exec(insertClientWithoutIdSQL,
		sql.Named("name", client.Name),
		sql.Named("password", client.Password),
		sql.Named("phone", client.Phone),
		sql.Named("login", client.Login),
	)
	if err != nil {
		t.Fatal(err)
	}

	id, err := getClientIdByLogin("abc", db)
	if err != nil {
		t.Error("want nil error")
	}

	if id != 1 {
		t.Errorf("want: %v, got: %v", 1, id)
	}
}
func Test_getClientIdByPhoneNumber(t *testing.T) {
	db := createDBinMemory(t)
	defer db.Close()

	_, err := getClientIdByPhoneNumber("123", db)
	if err == nil {
		t.Error("want not nil error")
	}

	_, err = db.Exec(clientsDDL)
	if err != nil {
		t.Fatal(err)
	}

	_, err = getClientIdByPhoneNumber("123", db)
	if err == nil {
		t.Error("want not nil error")
	}

	client := Client{
		Login:    "a",
		Password: "b",
		Name:     "c",
		Phone:    "123",
	}
	_, err = db.Exec(insertClientWithoutIdSQL,
		sql.Named("login", client.Login),
		sql.Named("password", client.Password),
		sql.Named("name", client.Name),
		sql.Named("phone", client.Phone),
	)
	if err != nil {
		t.Fatal(err)
	}

	_, err = getClientIdByPhoneNumber("456", db)
	if err == nil {
		t.Error("want not nil error")
	}

	clientId, err := getClientIdByPhoneNumber("123", db)
	if err != nil {
		t.Error("want nil error, got: ", err)
	}

	if clientId != 1 {
		t.Error("want 1, got: ", clientId)
	}

	err = db.QueryRow(`SELECT id FROM clients`).Scan(&clientId)
	if err != nil {
		t.Fatal(err)
	}

	if clientId != 1 {
		t.Error("want 1, got: ", clientId)
	}

	client2 := Client{
		Login: "aa",
		Phone: "987",
	}
	_, err = db.Exec(insertClientWithoutIdSQL,
		sql.Named("login", client2.Login),
		sql.Named("password", client2.Password),
		sql.Named("name", client2.Name),
		sql.Named("phone", client2.Phone),
	)
	if err != nil {
		t.Fatal(err)
	}

	clientId, err = getClientIdByPhoneNumber("01", db)
	if err == nil {
		t.Error("want not nil error, got: ", err)
	}
	if clientId != 0 {
		t.Error("want 0, got: ", clientId)
	}

	clientId, err = getClientIdByPhoneNumber("987", db)
	if err != nil {
		t.Error("want nil error, got: ", err)
	}
	if clientId != 2 {
		t.Error("want 2, got: ", clientId)
	}
}

func Test_loginForManager(t *testing.T) {
	db := createDBinMemory(t)
	defer db.Close()

	ok, err := loginForManager("hello", "golang", db)
	if err == nil {
		t.Error("want not nil error")
	}
	if ok == true {
		t.Error("want: false, got true")
	}

	_, err = db.Exec(managersDDL)
	if err != nil {
		t.Fatal(err)
	}

	ok, err = loginForManager("hello", "golang", db)
	if err != nil {
		t.Error("want not nil error")
	}
	if ok == true {
		t.Error("want: false, got true")
	}

	manager := Manager{
		Login:    "i-am-manager",
		Password: "dont-forget-me",
	}
	_, err = db.Exec(insertManagerWithoutIdSQL,
		sql.Named("login", manager.Login),
		sql.Named("password", manager.Password),
	)

	ok, err = loginForManager("hello", "golang", db)
	if err != nil {
		t.Error("want not nil error")
	}
	if ok == true {
		t.Error("want: false, got true")
	}

	ok, err = loginForManager("i-am-manager",
		"dont-forget-me", db)
	if err != nil {
		t.Error("want nil error")
	}
	if ok == false {
		t.Error("want: true, got false")
	}
}
func Test_loginForClient(t *testing.T) {
	db := createDBinMemory(t)
	defer db.Close()

	ok, err := loginForClient("hello", "golang", db)
	if err == nil {
		t.Error("want not nil error")
	}
	if ok == true {
		t.Error("want: false, got true")
	}

	_, err = db.Exec(clientsDDL)
	if err != nil {
		t.Fatal(err)
	}

	ok, err = loginForClient("hello", "golang", db)
	if err != nil {
		t.Error("want not nil error")
	}
	if ok == true {
		t.Error("want: false, got true")
	}

	client := Client{
		Login:    "i-am-client",
		Password: "dont-forget-me",
	}
	_, err = db.Exec(insertClientWithoutIdSQL,
		sql.Named("login", client.Login),
		sql.Named("password", client.Password),
		sql.Named("name", client.Name),
		sql.Named("phone", client.Phone),
	)

	ok, err = loginForClient("hello", "golang", db)
	if err != nil {
		t.Error("want not nil error")
	}
	if ok == true {
		t.Error("want: false, got true")
	}

	ok, err = loginForClient("i-am-client",
		"dont-forget-me", db)
	if err != nil {
		t.Error("want nil error")
	}
	if ok == false {
		t.Error("want: true, got false")
	}
}

func Test_transferToClient(t *testing.T) {
	db := createDBinMemory(t)
	defer db.Close()

	transfer := MoneyTransfer{
		Amount:                100,
		SenderId:              1,
		SenderAccountNumber:   0,
		ReceiverId:            2,
		ReceiverAccountNumber: 0,
	}

	err := transferToClient(transfer, db)
	if err == nil {
		t.Error("want not nil error")
	}

	_, err = db.Exec(clientsDDL)
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec(bankAccountsDDL)
	if err != nil {
		t.Fatal(err)
	}

	err = transferToClient(transfer, db)
	if err == nil {
		t.Error("want not nil error")
	}

	clientSender := Client{
		Login: "first",
	}

	err = addClient(clientSender, db)
	if err != nil {
		t.Fatal(err)
	}

	clientReceiver := Client{
		Login: "second",
	}
	err = addClient(clientReceiver, db)
	if err != nil {
		t.Fatal(err)
	}

	id1, err := getClientIdByLogin("first", db)
	if err != nil {
		t.Fatal(err)
	}
	err = addBankAccountToClient(id1, db)
	if err != nil {
		t.Fatal(err)
	}
	id2, err := getClientIdByLogin("second", db)
	if err != nil {
		t.Fatal(err)
	}
	err = addBankAccountToClient(id2, db)
	if err != nil {
		t.Fatal(err)
	}

	err = transferToClient(transfer, db)
	if err == nil {
		t.Error("want not nil error")
	}

	transfer.Amount = 0
	err = transferToClient(transfer, db)
	if err == nil {
		t.Error("want not nil error")
	}

	err = replenishBankAccount(id1, 0, 100, db)
	if err != nil {
		t.Fatal(err)
	}

	transfer.Amount = 40
	err = transferToClient(transfer, db)
	if err != nil {
		t.Error("want: nil, got: ", err)
	}

	var clientReceiverBalance int64
	err = db.QueryRow(getBalanceByClientIdAndAccountNumberSQL,
		sql.Named("id", id2),
		sql.Named("account_number", 0),
	).Scan(&clientReceiverBalance)
	if err != nil {
		t.Fatal(err)
	}
	if clientReceiverBalance != 40 {
		t.Error("want: 40, got: ", clientReceiverBalance)
	}


	var clientSenderBalance int
	err = db.QueryRow(getBalanceByClientIdAndAccountNumberSQL,
		sql.Named("id", id1),
		sql.Named("account_number", 0),
	).Scan(&clientSenderBalance)
	if err != nil {
		t.Fatal(err)
	}
	if clientSenderBalance != 60 {
		t.Error("want: 60, got: ", clientSenderBalance)
	}
}
