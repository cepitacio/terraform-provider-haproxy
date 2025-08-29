package haproxy

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"terraform-provider-haproxy/haproxy/utils"
)

const retryDelay = 2 * time.Second

var configMutex sync.Mutex

// Transaction executes a function within a transaction, with retry logic.
// This is the old behavior for backward compatibility.
func (c *HAProxyClient) Transaction(fn func(transactionID string) (*http.Response, error)) (*http.Response, error) {
	retryCount := 0
	for {
		configMutex.Lock()
		version, err := c.getCurrentConfigurationVersion()
		if err != nil {
			configMutex.Unlock()
			return nil, fmt.Errorf("failed to get configuration version: %v", err)
		}
		log.Printf("Current Transaction version: %s", version)

		// Try to create transaction ID with retry logic for version conflicts
		var id string
		for createRetry := 0; createRetry < 3; createRetry++ {
			id, err = c.createTransactionID(version)
			if err != nil {
				// Check if it's a version mismatch error that we can retry
				if customErr, ok := err.(*utils.CustomError); ok && customErr.APIError != nil {
					if customErr.APIError.Code == 409 && strings.Contains(customErr.APIError.Message, "version mismatch") {
						log.Printf("Version mismatch creating transaction, retrying with fresh version (attempt %d)", createRetry+1)
						// Get fresh version and retry
						version, err = c.getCurrentConfigurationVersion()
						if err != nil {
							configMutex.Unlock()
							return nil, fmt.Errorf("failed to get fresh configuration version: %v", err)
						}
						log.Printf("Fresh Transaction version: %s", version)
						time.Sleep(retryDelay)
						continue
					}
				}
				// If not a retryable error or max retries reached, break
				break
			}
			// Successfully created transaction ID
			break
		}

		if err != nil {
			configMutex.Unlock()
			return nil, fmt.Errorf("failed to create transaction ID after retries: %v", err)
		}

		log.Printf("Current Transaction ID: %s", id)

		log.Printf("Executing transaction function for ID: %s", id)
		resp, err := fn(id)

		log.Printf("Transaction function completed for ID: %s, response: %+v", id, resp)

		configMutex.Unlock()
		if err != nil {
			// 🔥 CRITICAL: Rollback transaction on any error to prevent orphaned resources
			log.Printf("Resource creation failed, rolling back transaction %s", id)
			rollbackErr := c.rollbackTransaction(id)
			if rollbackErr != nil {
				log.Printf("Warning: Failed to rollback transaction %s: %v", id, rollbackErr)
			}

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

		// 🔥 CRITICAL: Check if the resource creation was successful before committing
		if resp != nil {
			log.Printf("Resource creation response status: %d", resp.StatusCode)
			log.Printf("Resource creation response headers: %+v", resp.Header)

			if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
				// 🔥 CRITICAL: Resource creation failed - rollback transaction to prevent orphaned resources
				log.Printf("Resource creation failed with status %d, rolling back transaction %s", resp.StatusCode, id)
				rollbackErr := c.rollbackTransaction(id)
				if rollbackErr != nil {
					log.Printf("Warning: Failed to rollback transaction %s: %v", id, rollbackErr)
				}

				// Clone the response body since we need to read it
				bodyBytes, _ := io.ReadAll(resp.Body)
				log.Printf("Resource creation failed with status %d: %s", resp.StatusCode, string(bodyBytes))
				return nil, fmt.Errorf("resource creation failed with status %d: %s", resp.StatusCode, string(bodyBytes))
			}

			// Log successful response details
			if resp.StatusCode == http.StatusCreated || resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusAccepted {
				log.Printf("Resource creation successful with status %d", resp.StatusCode)
			}

			log.Printf("Resource created successfully in transaction %s", id)
		} else {
			log.Printf("Warning: Transaction function returned nil response")
		}

		log.Printf("About to commit transaction %s", id)
		log.Printf("Transaction %s: All resources created successfully, proceeding to commit", id)
		log.Printf("Transaction %s: Calling commitTransactionID with ID: '%s'", id, id)
		resp, err = c.commitTransactionID(id)

		if err != nil {
			log.Printf("Received commit error: %v", err)

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

		// Log successful commit
		log.Printf("Transaction %s committed successfully with status: %d", id, resp.StatusCode)
		log.Printf("Transaction %s response body: %+v", id, resp)
		log.Printf("Transaction %s response headers: %+v", id, resp.Header)

		// Log the commit response body for debugging
		if resp != nil && resp.Body != nil {
			commitBody, _ := io.ReadAll(resp.Body)
			log.Printf("Transaction %s commit response body content: %s", id, string(commitBody))
		}

		return resp, nil
	}
}

// BeginTransaction creates a new transaction and returns its ID.
func (c *HAProxyClient) BeginTransaction() (string, error) {
	configMutex.Lock()
	defer configMutex.Unlock()

	version, err := c.getCurrentConfigurationVersion()
	if err != nil {
		return "", fmt.Errorf("failed to get configuration version: %v", err)
	}
	log.Printf("Creating transaction with version: %s", version)

	// Try to create transaction ID with retry logic for version conflicts
	for createRetry := 0; createRetry < 3; createRetry++ {
		id, err := c.createTransactionID(version)
		if err != nil {
			// Check if it's a version mismatch error that we can retry
			if customErr, ok := err.(*utils.CustomError); ok && customErr.APIError != nil {
				if customErr.APIError.Code == 409 && strings.Contains(customErr.APIError.Message, "version mismatch") {
					log.Printf("Version mismatch creating transaction, retrying with fresh version (attempt %d)", createRetry+1)
					// Get fresh version and retry
					version, err = c.getCurrentConfigurationVersion()
					if err != nil {
						return "", fmt.Errorf("failed to get fresh configuration version: %v", err)
					}
					log.Printf("Fresh transaction version: %s", version)
					time.Sleep(retryDelay)
					continue
				}
			}
			// If not a retryable error or max retries reached, return error
			return "", fmt.Errorf("failed to create transaction ID after retries: %v", err)
		}
		// Successfully created transaction ID
		log.Printf("Transaction created successfully with ID: %s", id)
		return id, nil
	}

	return "", fmt.Errorf("failed to create transaction after max retries")
}

// RollbackTransaction rolls back a transaction by its ID.
func (c *HAProxyClient) rollbackTransaction(transactionID string) error {
	log.Printf("Rolling back transaction: %s", transactionID)

	// HAProxy Data Plane API doesn't have a rollback endpoint
	// Instead, we need to delete the transaction without committing
	// This effectively "rolls back" by removing the uncommitted changes

	// Delete the transaction ID to clean up
	req, err := c.newRequest(context.Background(), "DELETE", fmt.Sprintf("/services/haproxy/transactions/%s", transactionID), nil)
	if err != nil {
		return fmt.Errorf("failed to rollback transaction %s: %v", transactionID, err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute rollback request for transaction %s: %v", transactionID, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to rollback transaction %s, status: %d, body: %s", transactionID, resp.StatusCode, string(bodyBytes))
	}

	log.Printf("Transaction %s rolled back successfully", transactionID)
	return nil
}

// CommitTransaction commits a transaction by its ID.
func (c *HAProxyClient) CommitTransaction(transactionID string) error {
	log.Printf("Committing transaction: %s", transactionID)
	resp, err := c.commitTransactionID(transactionID)
	if err != nil {
		return fmt.Errorf("failed to commit transaction %s: %v", transactionID, err)
	}

	log.Printf("Transaction %s committed successfully with status: %d", transactionID, resp.StatusCode)
	return nil
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
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
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
		// Try to read error body for more details
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
		}

		// Try to parse as JSON error response
		var errorResp struct {
			Message string `json:"message"`
			Code    int    `json:"code"`
		}
		if json.Unmarshal(body, &errorResp) == nil && errorResp.Message != "" {
			// Return CustomError that can be detected by retry logic
			apiError := &utils.APIError{
				Code:    errorResp.Code,
				Message: errorResp.Message,
			}
			return "", utils.NewCustomError("Failed to create transaction", apiError)
		}

		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
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

	log.Printf("Committing transaction %s", transactionID)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		log.Printf("Error committing transaction %s: %v", transactionID, err)
		return nil, err
	}

	log.Printf("Transaction %s commit response status: %d", transactionID, resp.StatusCode)

	// Check if commit was successful
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		// Try to read error body for more details
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("transaction commit failed with status %d", resp.StatusCode)
		}

		// Try to parse as JSON error response
		var errorResp struct {
			Message string `json:"message"`
			Code    int    `json:"code"`
		}
		if json.Unmarshal(body, &errorResp) == nil && errorResp.Message != "" {
			apiError := &utils.APIError{
				Code:    errorResp.Code,
				Message: errorResp.Message,
			}
			return nil, utils.NewCustomError("Transaction commit failed", apiError)
		}

		return nil, fmt.Errorf("transaction commit failed with status %d: %s", resp.StatusCode, string(body))
	}

	return resp, nil
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
	return strings.Contains(err.Error(), "version or transaction not specified")
}
