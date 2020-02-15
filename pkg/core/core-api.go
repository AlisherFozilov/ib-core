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
func checkManagerLoginOnUnique(login string, db *sql.DB) (bool, error) {
	return checkLoginOnUnique(login, getManagerLoginByLogin, db)
}
func checkClientLoginOnUnique(login string, db *sql.DB) (bool, error) {
	return checkLoginOnUnique(login, getClientLoginByLogin, db)
}
func checkLoginOnUnique(login, getLoginByLogin string, db *sql.DB) (bool, error) {
	err := db.QueryRow(getLoginByLogin, login).Scan()
	if err == sql.ErrNoRows {
		return true, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func ReplenishBankAccount(clientId, accountNumber, amount int64,
	db *sql.DB) error {

	var balance int64
	err := db.QueryRow(getBalanceByClientIdAndAccountNumberSQL,
		sql.Named("id", clientId),
		sql.Named("account_number", accountNumber),
	).Scan(&balance)
	if err != nil {
		return err
	}
	increasedBalance := balance + amount
	_, err = db.Exec(updateBalanceByClientIdAndAccountNumberSQL,
		sql.Named("balance", increasedBalance),
		sql.Named("id", clientId),
		sql.Named("account_number", accountNumber),
	)
	return err
}

func AddClient(client Client, db *sql.DB) (err error) {
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
func AddService(service Service, db *sql.DB) (serviceNumber string, err error) {
	result, err := db.Exec(insertServiceWithoutIdSQL, sql.Named("name", service.Name))
	if err != nil {
		return "", err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return "", err
	}
	err = AddBankAccountToService(id, db)
	if err != nil {
		return "", err
	}
	var serviceId, accountNumber int64
	err = db.QueryRow(getServiceIdAndAccountNumberById, id,
	).Scan(&serviceId, &accountNumber)
	if err != nil {
		return "", err
	}

	const zerosForId = 1000_000_000
	const zerosForAccount = 1_0000
	serviceId += zerosForId
	accountNumber += zerosForAccount
	serviceIdStr := strconv.Itoa(int(serviceId))
	accountNumberStr := strconv.Itoa(int(accountNumber))

	serviceNumber = serviceIdStr[1:] + accountNumberStr[1:]
	return serviceNumber, nil
}

func AddManager(manager Manager, db *sql.DB) (err error) {
	_, err = db.Exec(
		insertManagerWithoutIdSQL,
		sql.Named("login", manager.Login),
		sql.Named("password", manager.Password),
	)
	return err
}
func addBankAccount(id int64, getBankAccountsCountByUserIdSQL,
	insertBankAccountToSQL string, db *sql.DB) (err error) {

	var bankAccountsCount int
	err = db.QueryRow(
		getBankAccountsCountByUserIdSQL,
		id,
	).Scan(&bankAccountsCount)
	if err != nil {
		return err
	}

	const startBalance = 0
	_, err = db.Exec(
		insertBankAccountToSQL,
		sql.Named("id", id),
		sql.Named("account_number", bankAccountsCount),
		sql.Named("balance", startBalance),
	)
	if err != nil {
		return err
	}

	return
}
func AddBankAccountToClient(id int64, db *sql.DB) error {
	return addBankAccount(id, getBankAccountsCountByClientIdSQL, insertBankAccountToClientSQL, db)
}
func AddBankAccountToService(id int64, db *sql.DB) error {
	return addBankAccount(id, getBankAccountsServicesCountByServiceIdSQL, insertBankAccountToServiceSQL, db)
}
func AddATM(address string, db *sql.DB) error {
	_, err := db.Exec(insertAtmWithoutIdSQL, address)
	return err
}
func GetClientIdByLogin(login string, db *sql.DB) (id int64, err error) {
	err = db.QueryRow(
		getClientIdByLoginSQL,
		login,
	).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
	}

func TransferToClient(transfer MoneyTransfer, db *sql.DB) error {
	return transferByReceiverAccountId(
		transfer,
		getBalanceByClientIdAndAccountNumberSQL,
		updateBalanceByClientIdAndAccountNumberSQL,
		db)
}

func PayForService(serviceNumber string,
	amount, payerId, payerAccountNumber int64,
	db *sql.DB) error {

	serviceId, accountNumber, err := ServiceNumberToIdAndAccountNumber(serviceNumber)
	if err != nil {
		return err
	}

	transfer := MoneyTransfer{
		Amount:                amount,
		SenderId:              payerId,
		SenderAccountNumber:   payerAccountNumber,
		ReceiverId:            serviceId,
		ReceiverAccountNumber: accountNumber,
	}
	return transferByReceiverAccountId(
		transfer,
		getBalanceByServiceIdAndAccountNumberSQL,
		updateBalanceByServiceIdAndAccountNumberSQL,
		db)
}

const digitLimitForAccount = 4

func ServiceNumberToIdAndAccountNumber(serviceNumber string) (int64, int64, error) {
	serviceIdStr := serviceNumber[:len(serviceNumber)-digitLimitForAccount]
	accountNumberStr := serviceNumber[len(serviceNumber)-digitLimitForAccount:]

	serviceId, err := strconv.Atoi(serviceIdStr)
	if err != nil {
		return 0, 0, err
	}
	accountNumber, err := strconv.Atoi(accountNumberStr)
	if err != nil {
		return 0, 0, err
	}
	return int64(serviceId), int64(accountNumber), nil
}

func transferByReceiverAccountId(
	tfr MoneyTransfer,
	getBalanceByIdAndAccountNumber string,
	updateBalanceByIdAndAccountNumber string,
	db *sql.DB) error {

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
	if tfr.Amount < 1 {
		return errors.New("zero ore less money to transfer")
	}


	var balance int64
	err = tx.QueryRow(
		getBalanceByClientIdAndAccountNumberSQL,
		sql.Named("id", tfr.SenderId),
		sql.Named("account_number", tfr.SenderAccountNumber),
	).Scan(&balance)
	if err != nil {
		return err
	}

	if tfr.Amount > balance {
		return errors.New("no enough money to transfer")
	}

	moneyRest := balance - tfr.Amount
	_, err = tx.Exec(updateBalanceByClientIdAndAccountNumberSQL,
		sql.Named("balance", moneyRest),
		sql.Named("id", tfr.SenderId),
		sql.Named("account_number", tfr.SenderAccountNumber),
	)
	if err != nil {
		return err
	}

	var receiverBalance int64
	err = tx.QueryRow(getBalanceByIdAndAccountNumber,
		sql.Named("id", tfr.ReceiverId),
		sql.Named("account_number", tfr.ReceiverAccountNumber),
	).Scan(&receiverBalance)
	if err != nil {
		return err
	}

	increasedBalance := receiverBalance + tfr.Amount
	_, err = tx.Exec(updateBalanceByIdAndAccountNumber,
		sql.Named("balance", increasedBalance),
		sql.Named("id", tfr.ReceiverId),
		sql.Named("account_number", tfr.ReceiverAccountNumber),
	)
	if err != nil {
		return err
	}

	return nil
}

func GetClientIdByPhoneNumber(phone string, db *sql.DB) (int64, error) {
	var clientId int64
	err := db.QueryRow(getClientIdByPhoneSQL, phone).Scan(&clientId)
	return clientId, err
}

func LoginForManager(login, password string, db *sql.DB) (bool, error) {
	return checkPassword(login, password, getManagerPasswordByLoginSQL, db)
}
func LoginForClient(login, password string, db *sql.DB) (bool, error) {
	return checkPassword(login, password, getClientPasswordByLoginSQL, db)
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

func ExportClientsToJSON(db *sql.DB) error {
	return exportToFile(db, getAllClientsDataSQL, "clients.json",
		mapRowToClient, json.Marshal, mapInterfaceSliceToClients)
}
func ExportAtmsToJSON(db *sql.DB) error {
	return exportToFile(db, getAllAtmDataSQL, "atms.json",
		mapRowToAtm, json.Marshal,
		mapInterfaceSliceToAtms)
}
func ExportBankAccountsToJSON(db *sql.DB) error {
	return exportToFile(db, getAllBankAccountsDataSQL, "bank-accounts.json",
		mapRowToBankAccount, json.Marshal,
		mapInterfaceSliceToBankAccounts)
}

//XML

func ExportClientsToXML(db *sql.DB) error {
	return exportToFile(db, getAllClientsDataSQL, "clients.xml",
		mapRowToClient, xml.Marshal, mapInterfaceSliceToClients)
}
func ExportAtmsToXML(db *sql.DB) error {
	return exportToFile(db, getAllAtmDataSQL, "atms.xml",
		mapRowToAtm, xml.Marshal,
		mapInterfaceSliceToAtms)
}
func ExportBankAccountsToXML(db *sql.DB) error {
	return exportToFile(db, getAllBankAccountsDataSQL, "bank-accounts.xml",
		mapRowToBankAccount, xml.Marshal,
		mapInterfaceSliceToBankAccounts)
}

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
		&bankAccount.UserId,
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

type mapperRowTo func(rows *sql.Rows) (interface{}, error)
type mapperInterfaceSliceTo func([]interface{}) interface{}
type marshaller func(interface{}) ([]byte, error)

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
func ImportClientsFromJSON(db *sql.DB) error {
	return importFromFile(
		db,
		"clients.json",
		func(data []byte) ([]interface{}, error) {
			return mapBytesToClients(data, json.Unmarshal)
		},
		insertClientToDB,
	)
}
func ImportAtmsFromJSON(db *sql.DB) error {
	return importFromFile(
		db,
		"atms.json",
		func(data []byte) ([]interface{}, error) {
			return mapBytesToAtms(data, json.Unmarshal)
		},
		insertAtmToDB,
	)
}
func ImportBankAccountsFromJSON(db *sql.DB) error {
	return importFromFile(
		db,
		"banc-accounts.json",
		func(data []byte) ([]interface{}, error) {
			return mapBytesToBankAccounts(data, json.Unmarshal)
		},
		insertBankAccountToDB,
	)
}

func ImportClientsFromXML(db *sql.DB) error {
	return importFromFile(
		db,
		"clients.xml",
		func(data []byte) ([]interface{}, error) {
			return mapBytesToClients(data, xml.Unmarshal)
		},
		insertClientToDB,
	)
}
func ImportAtmsFromXML(db *sql.DB) error {
	return importFromFile(
		db,
		"atms.xml",
		func(data []byte) ([]interface{}, error) {
			return mapBytesToAtms(data, xml.Unmarshal)
		},
		insertAtmToDB,
	)
}
func ImportBankAccountsFromXML(db *sql.DB) error {
	return importFromFile(
		db,
		"banc-accounts.xml",
		func(data []byte) ([]interface{}, error) {
			return mapBytesToBankAccounts(data, xml.Unmarshal)
		},
		insertBankAccountToDB,
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
		sql.Named("client_id", bankAccount.UserId),
	)
	if err != nil {
		return err
	}
	return nil
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
func AtmsList(db *sql.DB) ([]string, error) {
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
func BankAccountsList(id int64, db *sql.DB) ([]BankAccount, error) {
	rows, err := db.Query(getAllBankAccountsWithoutIdSQL, id)
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
			UserId:    id,
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
func GetAllAccountNumbersByClientId(id int64, db *sql.DB) ([]int64, error) {
	rows, err := db.Query(getAccountNumbersByClientIdSQL, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var accountNumber int64
	var bankAccounts []int64
	for rows.Next() {
		err := rows.Scan(&accountNumber)
		if err != nil {
			return nil, err
		}
		bankAccounts = append(bankAccounts, accountNumber)
	}

	return bankAccounts, nil
}