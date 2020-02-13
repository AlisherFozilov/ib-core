package core

import (
	"database/sql"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io/ioutil"
	"strconv"
)

func Init(db *sql.DB) (err error) {
	ddls := []string{
		managersDDL,
		clientsDDL,
		bankAccountsDDL,
		bankAccountsServicesDDL,
		servicesDDL,
		atmsDDL,
	}
	err = execQueries(ddls, db)
	if err != nil {
		return err
	}

	initialData := []string{managersInitData}
	err = execQueries(initialData, db)
	if err != nil {
		return err
	}

	return nil
}

func execQueries(queries []string, db *sql.DB) (err error) {
	for _, query := range queries {
		_, err = db.Exec(query)
		if err != nil {
			return fmt.Errorf("can't execute db query '%v': %w", query, err)
		}
	}
	return nil
}

//---------------Manager
func checkLoginOnUnique(login, getLoginByLogin string, db *sql.DB) bool {
	row := db.QueryRow(getLoginByLogin)
}
func addClient(client Client, db *sql.DB) (err error) {
	//TODO: check login on unique
	_, err = db.Exec(
		insertClientWithoutIdSQL,
		sql.Named("login", client.Login),
		sql.Named("password", client.Password),
		sql.Named("name", client.Name),
		sql.Named("phone", client.Phone),
	)
	return err
}
func addManager(manager Manager, db *sql.DB) (err error) {
	_, err = db.Exec(
		insertClientWithoutIdSQL,
		sql.Named("login", manager.Login),
		sql.Named("password", manager.Password),
	)
	return err
}

func addBankAccount(id int64, querySQL string, db *sql.DB) (err error) {
	var bankAccountsCount int
	err = db.QueryRow(
		getBankAccountsCountByClientIdSQL,
		id,
	).Scan(&bankAccountsCount)
	if err != nil {
		return err
	}

	const startBalance = 0
	_, err = db.Exec(
		querySQL,
		id,
		//sql.Named("client_id", id),
		sql.Named("account_number", bankAccountsCount),
		sql.Named("balance", startBalance),
	)
	if err != nil {
		return err
	}

	return
}
func addBankAccountToClient(id int64, db *sql.DB) error {
	return addBankAccount(id, insertBankAccountToClientSQL, db)
}
func addBankAccountToService(id int64, db *sql.DB) error {
	return addBankAccount(id, insertBankAccountToServiceSQL, db)
}
func addATM(address string, db *sql.DB) error {
	_, err := db.Exec(insertAtmWithoutIdSQL, address)
	return err
}
func getClientIdByLogin(login string, db *sql.DB) (id int64, err error) {
	err = db.QueryRow(
		getClientIdByLoginSQL,
		login,
	).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func transferToClient(
	amount, senderId, senderAccountNumber, receiverId,
	receiverAccountNumber int64,
	db *sql.DB) error {

	return transferByReceiverAccountId(amount, senderId,
		senderAccountNumber, receiverId, receiverAccountNumber,
		getBalanceByClientIdAndAccountNumberSQL,
		updateBalanceByClientIdAndAccountNumberSQL,
		db)
}

const digitLimitForAccount = 4

func payForService(serviceNumber string,
	amount, payerId, payerAccountNumber int64,
	db *sql.DB) error {

	serviceIdStr := serviceNumber[:len(serviceNumber)-digitLimitForAccount]
	accountNumberStr := serviceNumber[len(serviceNumber)-digitLimitForAccount:]

	serviceId, err := strconv.Atoi(serviceIdStr)
	if err != nil {
		return err
	}
	accountNumber, err := strconv.Atoi(accountNumberStr)
	if err != nil {
		return err
	}

	return transferByReceiverAccountId(amount, payerId,
		payerAccountNumber, int64(serviceId), int64(accountNumber),
		getBalanceByServiceIdAndAccountNumberSQL,
		updateBalanceByServiceIdAndAccountNumberSQL,
		db)
}

func transferByReceiverAccountId(
	amount, senderId, senderAccountNumber, receiverId,
	receiverAccountNumber int64,
	getBalanceByIdAndAccountNumber string,
	updateBalanceByIdAndAccountNumber string,
	db *sql.DB) error {

	if amount < 1 {
		return errors.New("zero ore less money to transfer")
	}
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			_ = tx.Rollback()
			return
		}
		err = tx.Commit()
	}()

	var balance int64
	err = tx.QueryRow(
		getBalanceByIdAndAccountNumber,
		sql.Named("id", senderId),
		sql.Named("account_number", senderAccountNumber),
	).Scan(&balance)
	if err != nil {
		return err
	}

	if amount < balance {
		return errors.New("no enough money to transfer")
	}

	moneyRest := balance - amount
	_, err = tx.Exec(updateBalanceByIdAndAccountNumber,
		sql.Named("balance", moneyRest),
		sql.Named("id", senderId),
		sql.Named("account_number", senderAccountNumber),
	)
	if err != nil {
		return err
	}

	var receiverBalance int64
	err = tx.QueryRow(getBalanceByClientIdAndAccountNumberSQL,
		sql.Named("id", receiverId),
		sql.Named("account_number", receiverAccountNumber),
	).Scan(&receiverBalance)
	if err != nil {
		return err
	}

	increasedBalance := receiverBalance + amount
	_, err = tx.Exec(updateBalanceByClientIdAndAccountNumberSQL,
		sql.Named("balance", increasedBalance),
		sql.Named("id", receiverId),
		sql.Named("account_number", receiverAccountNumber),
	)
	if err != nil {
		return err
	}

	return nil
}

func getClientIdByPhoneNumber(phone string, db *sql.DB) (int64, error) {
	var clientId int64
	err := db.QueryRow(getClientIdByPhoneSQL, phone).Scan(&clientId)
	return clientId, err
}

func loginForManager(login, password string, db *sql.DB) (bool, error) {
	return checkPassword(login, password, getClientPasswordByLoginSQL, db)
}
func loginForClient(login, password string, db *sql.DB) (bool, error) {
	return checkPassword(login, password, getManagerPasswordByLoginSQL, db)
}
func checkPassword(login, password, getPasswordByLogin string, db *sql.DB) (bool, error) {
	var dbPassword string

	err := db.QueryRow(
		getPasswordByLogin,
		login).Scan(&dbPassword)

	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}

		return false, queryError(getPasswordByLogin, err)
	}

	if dbPassword != password {
		return false, ErrInvalidPass
	}

	return true, nil
}

func (receiver *QueryError) Unwrap() error {
	return receiver.Err
}

func (receiver *QueryError) Error() string {
	return fmt.Sprintf("can't execute query %s: %s", receiver.Query, receiver.Err.Error())
}

func queryError(query string, err error) *QueryError {
	return &QueryError{Query: query, Err: err}
}

var ErrInvalidPass = errors.New("invalid password")

type QueryError struct { // alt + enter
	Query string
	Err   error
}

//----------------------------JSON && XML----------------

//Export
//JSON

func exportClientsToJSON(db *sql.DB) error {
	return exportToFile(db, getAllClientsDataSQL, "clients.json",
		mapRowToClient, json.Marshal, mapInterfaceSliceToClients)
}
func exportAtmsToJSON(db *sql.DB) error {
	return exportToFile(db, getAllAtmDataSQL, "atms.json",
		mapRowToAtm, json.Marshal,
		mapInterfaceSliceToAtms)
}
func exportBankAccountsToJSON(db *sql.DB) error {
	return exportToFile(db, getAllBankAccountsDataSQL, "bank-accounts.json",
		mapRowToBankAccount, json.Marshal,
		mapInterfaceSliceToBankAccounts)
}

//XML

func exportClientsToXML(db *sql.DB) error {
	return exportToFile(db, getAllClientsDataSQL, "clients.xml",
		mapRowToClient, xml.Marshal, mapInterfaceSliceToClients)
}
func exportAtmsToXML(db *sql.DB) error {
	return exportToFile(db, getAllAtmDataSQL, "atms.xml",
		mapRowToAtm, xml.Marshal,
		mapInterfaceSliceToAtms)
}
func exportBankAccountsToXML(db *sql.DB) error {
	return exportToFile(db, getAllBankAccountsDataSQL, "bank-accounts.xml",
		mapRowToBankAccount, xml.Marshal,
		mapInterfaceSliceToBankAccounts)
}

type mapperRowTo func(rows *sql.Rows) (interface{}, error)
type mapperInterfaceSliceTo func([]interface{}) interface{}
type marshaller func(interface{}) ([]byte, error)

func mapRowToClient(rows *sql.Rows) (interface{}, error) {
	client := Client{}
	err := rows.Scan(&client.Id, &client.Login, &client.Password,
		&client.Name, &client.Phone)
	if err != nil {
		return nil, err
	}
	return client, nil
}
func mapRowToAtm(rows *sql.Rows) (interface{}, error) {
	atm := Atm{}
	err := rows.Scan(&atm.Id, &atm.Address)
	if err != nil {
		return nil, err
	}
	return atm, nil
}
func mapRowToBankAccount(rows *sql.Rows) (interface{}, error) {
	bankAccount := BankAccount{}
	err := rows.Scan(
		&bankAccount.Id,
		&bankAccount.ClientId,
		&bankAccount.AccountId,
		&bankAccount.Balance,
	)
	if err != nil {
		return nil, err
	}
	return bankAccount, nil
}

func mapInterfaceSliceToClients(ifaces []interface{}) interface{} {
	clients := make([]Client, len(ifaces))
	for i := range ifaces {
		clients[i] = ifaces[i].(Client)
	}
	clientsExport := ClientsExport{Clients: clients}
	return clientsExport
}
func mapInterfaceSliceToAtms(ifaces []interface{}) interface{} {
	atms := make([]Atm, len(ifaces))
	for i := range ifaces {
		atms[i] = ifaces[i].(Atm)
	}
	atmsExport := AtmsExport{Atms: atms}
	return atmsExport
}
func mapInterfaceSliceToBankAccounts(ifaces []interface{}) interface{} {
	bankAccounts := make([]BankAccount, len(ifaces))
	for i := range ifaces {
		bankAccounts[i] = ifaces[i].(BankAccount)
	}
	bankAccountsExport := BankAccountsExport{BankAccounts: bankAccounts}
	return bankAccountsExport
}

func exportToFile(db *sql.DB, querySQL string, filename string,
	mapRow mapperRowTo, marshal marshaller,
	mapDataSlice mapperInterfaceSliceTo) error {

	rows, err := db.Query(querySQL)
	if err != nil {
		return err
	}
	defer func() {
		err = rows.Close()
	}()
	var dataSlice []interface{}
	for rows.Next() {
		dataElement, err := mapRow(rows)
		if err != nil {
			return err
		}
		dataSlice = append(dataSlice, dataElement)
	}
	exportData := mapDataSlice(dataSlice)
	data, err := marshal(exportData)
	err = ioutil.WriteFile(filename, data, 0666)
	if err != nil {
		return err
	}
	return nil
}

//---------simple way
/*
func importClientsFromJSON(db *sql.DB) error {
	clientsData, err := ioutil.ReadFile("clients.json")
	if err != nil {
		return err
	}
	//TODO: var clients []Client
	clients := []Client{}
	err = json.Unmarshal(clientsData, &clients)
	if err != nil {
		return err
	}

	for _, client := range clients {
		_, err = db.Exec(
			insertClientSQL,
			sql.Named("id", client.Id),
			sql.Named("name", client.Name),
			sql.Named("login", client.Login),
			sql.Named("password", client.Password),
			sql.Named("phone", client.Phone),
		)
		if err != nil {
			return err
		}
	}

	return nil
}
func importAtmsFromJSON(db *sql.DB) error {
	atmsData, err := ioutil.ReadFile("atms.json")
	if err != nil {
		return err
	}

	atms := []Atm{}
	err = json.Unmarshal(atmsData, &atms)
	if err != nil {
		return err
	}

	for _, atm := range atms {
		_, err = db.Exec(
			insertAtmSQL,
			sql.Named("id", atm.Id),
			sql.Named("name", atm.Address),
		)
		if err != nil {
			return err
		}
	}

	return nil
}
*/

//Import

//JSON
func importClientsFromJSON(db *sql.DB) error {
	return importFromFile(
		db,
		"clients.json",
		func(data []byte) ([]interface{}, error) {
			return mapBytesToClients(data, json.Unmarshal)
		},
		insertClientToDB,
	)
}
func mapBytesToClients(data []byte,
	unmarshal func([]byte, interface{}) error,
) ([]interface{}, error) {
	clientsExport := ClientsExport{}
	err := unmarshal(data, &clientsExport)
	if err != nil {
		return nil, err
	}
	ifaces := make([]interface{}, len(clientsExport.Clients))
	for index := range ifaces {
		ifaces[index] = clientsExport.Clients[index]
	}
	return ifaces, nil
}
func insertClientToDB(iface interface{}, db *sql.DB) error {
	client := iface.(Client)
	_, err := db.Exec(
		insertClientSQL,
		sql.Named("id", client.Id),
		sql.Named("name", client.Name),
		sql.Named("login", client.Login),
		sql.Named("password", client.Password),
		sql.Named("phone", client.Phone),
	)
	if err != nil {
		return err
	}
	return nil
}

func importAtmsFromJSON(db *sql.DB) error {
	return importFromFile(
		db,
		"atms.json",
		func(data []byte) ([]interface{}, error) {
			return mapBytesToAtms(data, json.Unmarshal)
		},
		insertAtmToDB,
	)
}
func mapBytesToAtms(data []byte,
	unmarshal func([]byte, interface{}) error,
) ([]interface{}, error) {
	atmsExport := AtmsExport{}
	err := unmarshal(data, &atmsExport)
	if err != nil {
		return nil, err
	}
	ifaces := make([]interface{}, len(atmsExport.Atms))
	for index := range ifaces {
		ifaces[index] = atmsExport.Atms[index]
	}
	return ifaces, nil
}
func insertAtmToDB(iface interface{}, db *sql.DB) error {
	atm := iface.(Atm)
	_, err := db.Exec(
		insertAtmSQL,
		sql.Named("id", atm.Id),
		sql.Named("name", atm.Address),
	)
	if err != nil {
		return err
	}
	return nil
}

func importBankAccountsFromJSON(db *sql.DB) error {
	return importFromFile(
		db,
		"banc-accounts.json",
		func(data []byte) ([]interface{}, error) {
			return mapBytesToBankAccounts(data, json.Unmarshal)
		},
		insertBankAccountToDB,
	)
}
func mapBytesToBankAccounts(data []byte,
	unmarshal func([]byte, interface{}) error,
) ([]interface{}, error) {
	bankAccountsExport := BankAccountsExport{}
	err := unmarshal(data, &bankAccountsExport)
	if err != nil {
		return nil, err
	}
	ifaces := make([]interface{}, len(bankAccountsExport.BankAccounts))
	for index := range ifaces {
		ifaces[index] = bankAccountsExport.BankAccounts[index]
	}
	return ifaces, nil
}
func insertBankAccountToDB(iface interface{}, db *sql.DB) error {
	bankAccount := iface.(BankAccount)
	_, err := db.Exec(
		insertBankAccountSQL,
		sql.Named("id", bankAccount.Id),
		sql.Named("balance", bankAccount.Balance),
		sql.Named("account_number", bankAccount.AccountId),
		sql.Named("client_id", bankAccount.ClientId),
	)
	if err != nil {
		return err
	}
	return nil
}

func importClientsFromXML(db *sql.DB) error {
	return importFromFile(
		db,
		"clients.xml",
		func(data []byte) ([]interface{}, error) {
			return mapBytesToClients(data, xml.Unmarshal)
		},
		insertClientToDB,
	)
}
func importAtmsFromXML(db *sql.DB) error {
	return importFromFile(
		db,
		"atms.xml",
		func(data []byte) ([]interface{}, error) {
			return mapBytesToAtms(data, xml.Unmarshal)
		},
		insertAtmToDB,
	)
}
func importBankAccountsFromXML(db *sql.DB) error {
	return importFromFile(
		db,
		"banc-accounts.xml",
		func(data []byte) ([]interface{}, error) {
			return mapBytesToBankAccounts(data, xml.Unmarshal)
		},
		insertBankAccountToDB,
	)
}

func importFromFile(db *sql.DB, filename string,
	mapBytesToInterfaces func([]byte) ([]interface{}, error),
	insertToDB func(interface{}, *sql.DB) error,
) error {
	itemsData, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	sliceData, err := mapBytesToInterfaces(itemsData)

	for _, datum := range sliceData {
		err = insertToDB(datum, db)
		if err != nil {
			return err
		}
	}

	return nil
}

//---------------Client
func atmsList(db *sql.DB) ([]string, error) {
	atms := make([]string, 0)
	rows, err := db.Query(getAllAtmAddressesSQL)
	if err != nil {
		return nil, err
	}
	defer func() {
		err = rows.Close()
		if err != nil {
			atms = nil
		}
	}()
	var address string
	for rows.Next() {
		err = rows.Scan(&address)
		if err != nil {
			return nil, err
		}
		atms = append(atms, address)
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return atms, nil
}
func bankAccountsList(id int64, db *sql.DB) ([]BankAccount, error) {
	rows, err := db.Query(getAllBankAccountsWithoutIdSQL)
	if err != nil {
		return nil, err
	}
	bankAccounts := make([]BankAccount, 0)
	defer func() {
		err = rows.Close()
		if err != nil {
			bankAccounts = nil
		}
	}()
	var balance, accountId int64
	for rows.Next() {
		err = rows.Scan(&balance, &accountId)
		if err != nil {
			return nil, err
		}
		bankAccounts = append(bankAccounts, BankAccount{
			ClientId:  id,
			AccountId: accountId,
			Balance:   balance,
		})
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return bankAccounts, nil
}
