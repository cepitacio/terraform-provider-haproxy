package transaction

import (
	"sync"
)

type ConfigTransaction struct {
	Username      string
	Password      string
	BaseURL       string
	SSL           bool
	TransactionID string
	Version       int
}

type Manager struct {
	TransactionID     string
	ConfigTransaction ConfigTransaction
}

type TransactionResponse struct {
	Version int    `json:"_version"`
	ID      string `json:"id"`
	Status  string `json:"status"`
}

var configMutex sync.Mutex
