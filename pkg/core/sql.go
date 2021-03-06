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
);`  //bankAccountsNumber INTEGER NOT NULL
	bankAccountsDDL = `
CREATE TABLE IF NOT EXISTS bank_accounts
(
    id        INTEGER PRIMARY KEY AUTOINCREMENT,
    client_id INTEGER NOT NULL REFERENCES clients,
    account_number INTEGER NOT NULL,
    balance   INTEGER NOT NULL
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
    account_number  INTEGER NOT NULL,
    balance    INTEGER NOT NULL
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
	insertManagerWithoutIdSQL = `
INSERT INTO managers(login, password)
VALUES (:login, :password);`
	getBankAccountsCountByClientIdSQL = `
SELECT count(ba.id)
FROM bank_accounts ba
    JOIN clients c on ba.client_id = ?;
`
	getBankAccountsServicesCountByServiceIdSQL = `
SELECT count(bas.id)
FROM bank_accounts_services bas
         JOIN services s on bas.service_id = ?;
`
	insertBankAccountToClientSQL = `
INSERT INTO bank_accounts (client_id, account_number, balance)
VALUES (:id, :account_number, :balance);`
	insertBankAccountToServiceSQL = `
INSERT INTO bank_accounts_services (service_id, account_number, balance)
VALUES (:id, :account_number, :balance);`
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
SELECT balance, account_number
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
ON CONFLICT DO NOTHING ;`
	getAllBankAccountsDataSQL = `
SELECT *
FROM bank_accounts;`
	getAllBankAccountsServicesDataSQL = `
SELECT *
FROM bank_accounts_services;`

	insertBankAccountSQL = `
INSERT INTO bank_accounts
VALUES (:id, :client_id, :account_number, :balance)
ON CONFLICT DO NOTHING ;
`

	getBalanceByClientIdAndAccountNumberSQL = `
SELECT ba.balance
FROM bank_accounts ba
WHERE ba.client_id = :id
AND ba.account_number  = :account_number;`

	updateBalanceByClientIdAndAccountNumberSQL = `
UPDATE bank_accounts
SET balance = :balance
WHERE client_id = :id
  AND account_number = :account_number;`

	getClientIdByPhoneSQL = `
SELECT c.id
FROM clients c
WHERE c.phone = ?;`

	getBalanceByServiceIdAndAccountNumberSQL = `
SELECT balance
FROM bank_accounts_services bas
WHERE bas.service_id = :id
  AND bas.account_number  = :account_number;`

	updateBalanceByServiceIdAndAccountNumberSQL = `
UPDATE bank_accounts_services
SET balance = :balance
WHERE service_id = :id
  AND account_number = :account_number;
`
	getClientPasswordByLoginSQL = `
SELECT password
FROM clients
WHERE login = ?;`

	getManagerPasswordByLoginSQL = `
SELECT password
FROM managers
WHERE login = ?;`

	getManagerLoginByLogin = `
SELECT login
FROM managers
WHERE login = ?;`

	getClientLoginByLogin = `
SELECT login
FROM clients
WHERE login = ?;`

	insertServiceWithoutIdSQL = `
INSERT INTO services (name)
VALUES (:name)
ON CONFLICT DO NOTHING;`

	getServiceIdAndAccountNumberById = `
SELECT service_id, account_number
FROM bank_accounts_services
WHERE service_id = ?;`

	getAccountNumbersByClientIdSQL = `
SELECT account_number
FROM bank_accounts
WHERE client_id = ?;`
)
