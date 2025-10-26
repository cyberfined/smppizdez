package account

import (
	"fmt"
	"smppizdez/coding"
)

type Account struct {
	ID            string
	Host          string
	Port          uint16
	TLS           bool
	SystemID      string
	Password      string
	SystemType    string
	BindType      BindType
	DefaultCoding coding.Coding
}

type BindType int

const (
	Transceiver BindType = iota + 1
	Transmitter
	Receiver
)

func (t BindType) String() string {
	switch t {
	case Transceiver:
		return "Transceiver"
	case Transmitter:
		return "Transmitter"
	case Receiver:
		return "Receiver"
	default:
		return fmt.Sprintf("unknown bind type (%d)", t)
	}
}

type Repository interface {
	GetAccounts() ([]Account, error)
	CreateAccount(account *Account) error
	UpdateAccount(account *Account) error
	DeleteAccount(id string) error
}
