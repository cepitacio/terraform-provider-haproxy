package haproxy

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
)

const retryDelay = 2 * time.Second

// Transaction executes a function within a transaction, with retry logic.
func (c *HAProxyClient) Transaction(fn func(transactionID string) (*http.Response, error)) (*http.Response, error) {
	retryCount := 0
	for {
		version, err := c.getCurrentConfigurationVersion()
		if err != nil {
			return nil, fmt.Errorf("failed to get configuration version: %v", err)
		}

		id, err := c.createTransactionID(version)
		if err != nil {
			return nil, fmt.Errorf("failed to create transaction ID: %v", err)
		}

		resp, err := fn(id)

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

func (c *HAProxyClient) getCurrentConfigurationVersion() (string, error) {
	req, err := c.newRequest(context.Background(), "GET", "/services/haproxy/configuration/version", nil)
	if err != nil {
		return "", err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d", resp.Body)
	}

	var version int
	if err := json.NewDecoder(resp.Body).Decode(&version); err != nil {
		return "", err
	}

	return fmt.Sprintf("%d", version), nil
}

func (c *HAProxyClient) createTransactionID(version string) (string, error) {
	req, err := c.newRequest(context.Background(), "POST", fmt.Sprintf("/services/haproxy/transactions?version=%s", version), nil)
	if err != nil {
		return "", err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("unexpected status code: %d", resp.Body)
	}

	var transaction TransactionResponse
	if err := json.NewDecoder(resp.Body).Decode(&transaction); err != nil {
		return "", err
	}

	return transaction.ID, nil
}

func (c *HAProxyClient) commitTransactionID(transactionID string) (*http.Response, error) {
	req, err := c.newRequest(context.Background(), "PUT", fmt.Sprintf("/services/haproxy/transactions/%s", transactionID), nil)
	if err != nil {
		return nil, err
	}

	return c.httpClient.Do(req)
}

func TransactionOutdated(err error) bool {
	return strings.Contains(err.Error(), "is outdated and cannot be committed")
}

func TransactionDoesNotExist(err error) bool {
	return strings.Contains(err.Error(), "transaction does not exist")
}

func isVersionMismatch(err error) bool {
	return strings.Contains(err.Error(), "version mismatch")
}

func isVersionOrTransSpecified(err error) bool {
	return strings.Contains(err.Error(), "version or transaction not specified")
}
