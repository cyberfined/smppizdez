package json_storage

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"slices"
	"smppizdez/account"
	"smppizdez/coding"
	"strings"

	"github.com/google/uuid"
)

type Storage struct {
	f *os.File
}

func Open(path string) (Storage, error) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0644)
	return Storage{f: f}, err
}

type accountJson struct {
	Host          string `json:"host"`
	Port          uint16 `json:"port"`
	TLS           bool   `json:"tls"`
	SystemID      string `json:"systemID"`
	Password      string `json:"password"`
	SystemType    string `json:"systemType,omitempty"`
	BindType      string `json:"bindType"`
	DefaultCoding string `json:"defaultCoding"`
}

type bindTypeStr struct {
	typ account.BindType
	str string
}

var bindTypeStrings = []bindTypeStr{
	{typ: account.Transceiver, str: "transceiver"},
	{typ: account.Transmitter, str: "transmitter"},
	{typ: account.Receiver, str: "receiver"},
}

func parseBindType(s string) (account.BindType, error) {
	for _, typStr := range bindTypeStrings {
		if typStr.str == s {
			return typStr.typ, nil
		}
	}

	return account.Transceiver, fmt.Errorf("Unknown bind type %s", s)
}

func bindTypeToString(typ account.BindType) (string, error) {
	for _, typStr := range bindTypeStrings {
		if typ == typStr.typ {
			return typStr.str, nil
		}
	}
	return "", fmt.Errorf("Unknown bind type enum %d", typ)
}

type codingStr struct {
	cod coding.Coding
	str string
}

var codingStrings = []codingStr{
	{cod: coding.GSM7, str: "gsm7"},
	{cod: coding.GSM8, str: "gsm8"},
	{cod: coding.ASCII, str: "ascii"},
	{cod: coding.Octet1, str: "octet1"},
	{cod: coding.Latin1, str: "latin1"},
	{cod: coding.Octet2, str: "octet2"},
	{cod: coding.JIS, str: "jis"},
	{cod: coding.Cyrillic, str: "cyrillic"},
	{cod: coding.Hebrew, str: "hebrew"},
	{cod: coding.UCS2, str: "ucs2"},
	{cod: coding.Pictogram, str: "pictogram"},
	{cod: coding.MusicCodes, str: "musicCodes"},
	{cod: coding.ExtendedJIS, str: "extJis"},
	{cod: coding.KSC5601, str: "ksc5601"},
}

func parseCoding(s string) (coding.Coding, error) {
	for _, codStr := range codingStrings {
		if codStr.str == s {
			return codStr.cod, nil
		}
	}
	return coding.GSM7, fmt.Errorf("Unknown coding %s", s)
}

func codingToString(cod coding.Coding) (string, error) {
	for _, codStr := range codingStrings {
		if codStr.cod == cod {
			return codStr.str, nil
		}
	}
	return "", fmt.Errorf("Unknown coding enum enum %d", cod)
}

func (s Storage) GetAccounts() ([]account.Account, error) {
	accountsMap, err := s.getJsonAccounts()
	if err != nil {
		return nil, err
	}

	accounts := make([]account.Account, 0, len(accountsMap))
	var loadErr error
	for id, accJson := range accountsMap {
		var defaultCoding coding.Coding
		bindType, err := parseBindType(accJson.BindType)
		if err == nil {
			defaultCoding, err = parseCoding(accJson.DefaultCoding)
		}
		if err != nil {
			loadErr = errors.Join(
				loadErr,
				fmt.Errorf("account ID=%s loading error: %w", id, err),
			)
			continue
		}

		acc := account.Account{
			ID:            id,
			Host:          accJson.Host,
			Port:          accJson.Port,
			TLS:           accJson.TLS,
			SystemID:      accJson.SystemID,
			Password:      accJson.Password,
			SystemType:    accJson.SystemType,
			BindType:      bindType,
			DefaultCoding: defaultCoding,
		}
		accounts = append(accounts, acc)
	}

	slices.SortFunc(
		accounts,
		func(a1 account.Account, a2 account.Account) int {
			return strings.Compare(a1.ID, a2.ID)
		},
	)

	return accounts, loadErr
}

func (s Storage) CreateAccount(account *account.Account) error {
	bindTypeStr, err := bindTypeToString(account.BindType)
	if err != nil {
		return err
	}

	defaultCodingStr, err := codingToString(account.DefaultCoding)
	if err != nil {
		return err
	}

	accountsMap, err := s.getJsonAccounts()
	if err != nil {
		return err
	}

	account.ID = uuid.NewString()
	accountsMap[account.ID] = accountJson{
		Host:          account.Host,
		Port:          account.Port,
		TLS:           account.TLS,
		SystemID:      account.SystemID,
		Password:      account.Password,
		SystemType:    account.SystemType,
		BindType:      bindTypeStr,
		DefaultCoding: defaultCodingStr,
	}
	return s.save(accountsMap)
}

func (s Storage) UpdateAccount(account *account.Account) error {
	bindTypeStr, err := bindTypeToString(account.BindType)
	if err != nil {
		return err
	}

	defaultCodingStr, err := codingToString(account.DefaultCoding)
	if err != nil {
		return err
	}

	accountsMap, err := s.getJsonAccounts()
	if err != nil {
		return err
	}

	if _, ok := accountsMap[account.ID]; !ok {
		return fmt.Errorf("Account ID=%s does not exist", account.ID)
	}

	accountsMap[account.ID] = accountJson{
		Host:          account.Host,
		Port:          account.Port,
		TLS:           account.TLS,
		SystemID:      account.SystemID,
		Password:      account.Password,
		SystemType:    account.SystemType,
		BindType:      bindTypeStr,
		DefaultCoding: defaultCodingStr,
	}
	return s.save(accountsMap)
}

func (s Storage) DeleteAccount(id string) error {
	accountsMap, err := s.getJsonAccounts()
	if err != nil {
		return err
	}

	if _, ok := accountsMap[id]; !ok {
		return fmt.Errorf("Account ID=%s does not exist", id)
	}
	delete(accountsMap, id)
	return s.save(accountsMap)
}

func (s Storage) getJsonAccounts() (map[string]accountJson, error) {
	_, err := s.f.Seek(0, io.SeekStart)
	if err != nil {
		return nil, err
	}

	accountsMap := make(map[string]accountJson)
	decoder := json.NewDecoder(s.f)
	err = decoder.Decode(&accountsMap)
	if err != nil {
		if errors.Is(err, io.EOF) {
			return make(map[string]accountJson), nil
		}
		return nil, err
	}
	return accountsMap, nil
}

func (s Storage) save(accountsMap map[string]accountJson) error {
	s.f.Truncate(0)
	_, err := s.f.Seek(0, io.SeekStart)
	if err != nil {
		return err
	}
	encoder := json.NewEncoder(s.f)
	err = encoder.Encode(accountsMap)
	return err
}
