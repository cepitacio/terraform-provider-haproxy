package transaction

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"terraform-provider-haproxy/internal/utils"
	"time"
)

const retryDelay = 2 * time.Second

// Get current configuration, create transaction id and commit transaction id.
// Retry until successful for concurrency because api might fail due to:
// Concurrency error: Tranction is outdate and cannot be commited
// Concurrency error: transaction does not exist
// Concurrency error: version mismatch
// Concurrency error: version or transaction not specified
// Fail if error is different
func (c *ConfigTransaction) Transaction(fn func(transactionID string) (*http.Response, error)) (*http.Response, error) {
	retryCount := 0
	for {
		configMutex.Lock()
		version, err := c.getCurrentConfigurationVersion()
		if err != nil {
			configMutex.Unlock()
			return nil, fmt.Errorf("failed to get configuration version: %v", err)
		}
		fmt.Println("Current Transaction version:", version)
		id, err := c.createTransactionID(version)
		if err != nil {
			configMutex.Unlock()
			return nil, fmt.Errorf("failed to create transaction ID: %v", err)
		}
		fmt.Println("Current Transaction ID:", id)

		resp, err := fn(id)

		fmt.Println("fn callback response:", resp)

		configMutex.Unlock()
		if err != nil {
			if TransactionDoesNotExist(err) {
				log.Printf("Retrying transaction due to not transcation not existing %v", id)
				retryCount++
				time.Sleep(retryDelay)
				continue
			}
			if isVersionOrTransSpecified(err) {
				log.Printf("Retrying transaction due to version or transaction not specified %v", id)
				retryCount++
				time.Sleep(retryDelay)
				continue
			}
			return nil, fmt.Errorf("transaction function failed: %v", err)
		}

		resp, err = c.commitTransactionID(id)

		if err != nil {
			log.Printf("Received error: %v", err)

			if TransactionOutdated(err) {
				log.Printf("Retrying transaction due to outdated version %v", id)
				retryCount++
				time.Sleep(retryDelay)
				continue
			}
			if isVersionMismatch(err) {
				log.Printf("Retrying transaction due to version mismatch %v", id)
				retryCount++
				time.Sleep(retryDelay)
				continue
			}
			return nil, fmt.Errorf("failed to commit transaction after retries: ERR: %v Transaction ID: %v", err, id)
		}

		return resp, nil
	}

}

func TransactionOutdated(err error) bool {
	fmt.Printf("Received error transaction outdated: %T, %+v\n", err, err)
	if customErr, ok := err.(*utils.CustomError); ok && customErr.APIError != nil {
		return customErr.APIError.Code == 406 && strings.Contains(customErr.APIError.Message, "transaction") && strings.Contains(customErr.APIError.Message, "is outdated and cannot be committed")
	}
	return false
}

func TransactionDoesNotExist(err error) bool {
	fmt.Printf("Received error transaction does not exist: %T, %+v\n", err, err)
	if customErr, ok := err.(*utils.CustomError); ok && customErr.APIError != nil {
		return customErr.APIError.Code == 400 && strings.Contains(customErr.APIError.Message, "transaction does not exist")
	}
	return false
}

func isVersionMismatch(err error) bool {
	fmt.Printf("Received error version mismatch: %T, %+v\n", err, err)
	if customErr, ok := err.(*utils.CustomError); ok && customErr.APIError != nil {
		return customErr.APIError.Code == 409 && strings.Contains(customErr.APIError.Message, "version mismatch")
	}
	return false
}

func isVersionOrTransSpecified(err error) bool {
	fmt.Printf("Received error version or transaction not specified: %T, %+v\n", err, err)
	if customErr, ok := err.(*utils.CustomError); ok && customErr.APIError != nil {
		return customErr.APIError.Code == 400 && strings.Contains(customErr.APIError.Message, "version or transaction not specified")
	}
	return false
}
