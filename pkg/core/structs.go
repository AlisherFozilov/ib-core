package core

type Client struct {
	Id       int64
	Login    string
	Password string
	Name     string
	Phone    string
}

type Manager struct {
	Id int64
	Login string
	Password string
}

type BankAccount struct {
	Id        int64
	ClientId  int64
	AccountId int64
	Balance   int64
}

type Atm struct {
	Id      int64
	Address string
}

type ClientsExport struct {
	Clients []Client
}

type BankAccountsExport struct {
	BankAccounts []BankAccount
}

type AtmsExport struct {
	Atms []Atm
}
