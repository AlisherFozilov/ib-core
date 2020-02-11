package core

type Client struct {
	Id       int64
	Login    string
	Password string
	Name     string
	Phone    string
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
