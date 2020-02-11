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
);` //bankAccountsNumber INTEGER NOT NULL
	bankAccountsDDL = `
CREATE TABLE IF NOT EXISTS bank_accounts
(
    id        INTEGER PRIMARY KEY AUTOINCREMENT,
    client_id INTEGER NOT NULL REFERENCES clients,
    AccountId INTEGER NOT NULL,
    Balance   INTEGER NOT NULL
);`
	servicesDDL = `
CREATE TABLE IF NOT EXISTS services
(
    id INTEGER PRIMARY KEY AUTOINCREMENT ,
    name TEXT NOT NULL
);`
	bankAccountsServicesDDL = `
CREATE TABLE IF NOT EXISTS bank_accounts_services
(
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    service_id INTEGER NOT NULL REFERENCES services,
    AccountId  INTEGER NOT NULL,
    Balance    INTEGER NOT NULL
);`
	atmsDDL = `
CREATE TABLE IF NOT EXISTS atms
(
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    address TEXT
);`

	managersInitData = `
INSERT INTO managers 
VALUES (1, 'admin', 'top-secret')
ON CONFLICT DO NOTHING ;`
	insertClientWithoutIdSQL = `
INSERT INTO clients(login, password, name, phone)
VALUES (:login, :password, :name, :phone);`
	getBankAccountsCountByClientIdSQL = `
SELECT count(ba.id)
FROM bank_accounts ba
    JOIN clients c on ba.client_id = ?;
`
	insertBankAccountToClientSQL = `
INSERT INTO bank_accounts (client_id, AccountId, Balance)
VALUES (?, :AccountId, :Balance);`
	insertBankAccountToServiceSQL = `
INSERT INTO bank_accounts_services (service_id, AccountId, Balance)
VALUES (?, :AccountId, :Balance);`
	getClientIdByLoginSQL = `
SELECT id
FROM clients
WHERE login = ?;`
	insertAtmWithoutIdSQL = `
INSERT INTO atms(address)
VALUES (:address);`
	getAllAtmAddressesSQL = `
SELECT atms.address
FROM atms;`
	getAllBankAccountsWithoutIdSQL = `
SELECT Balance, AccountId
FROM bank_accounts
WHERE client_id = ?;`
	getAllClientsDataSQL = `
SELECT *
FROM clients;`
	insertClientSQL = `
INSERT INTO clients
VALUES (:id, :login, :password, :name, :phone)
ON CONFLICT DO NOTHING;`
	getAllAtmDataSQL = `
SELECT *
FROM atms;`
	insertAtmSQL = `
INSERT INTO atms
VALUES (:id, :address)
ON CONFLICT DO NOTHING ;
`
	getAllBankAccountsDataSQL = `
SELECT *
FROM bank_accounts;`
)
