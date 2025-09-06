package haproxy

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"

	"terraform-provider-haproxy/haproxy/utils"
)

// sanitizeResponseBody removes sensitive information from response bodies before logging
func sanitizeResponseBody(body string) string {
	// Remove password from error messages
	passwordRegex := regexp.MustCompile(`"password":\s*"[^"]*"`)
	body = passwordRegex.ReplaceAllString(body, `"password": "***"`)

	// Remove password from "invalid password" messages
	invalidPasswordRegex := regexp.MustCompile(`invalid password:\s*[^\s"]*`)
	body = invalidPasswordRegex.ReplaceAllString(body, `invalid password: ***`)

	// Remove any other potential sensitive fields
	sensitiveFields := []string{"token", "secret", "key", "auth"}
	for _, field := range sensitiveFields {
		fieldRegex := regexp.MustCompile(fmt.Sprintf(`"%s":\s*"[^"]*"`, field))
		body = fieldRegex.ReplaceAllString(body, fmt.Sprintf(`"%s": "***"`, field))
	}

	return body
}

// HAProxyClient is the client for the HAProxy Data Plane API.
type HAProxyClient struct {
	httpClient *http.Client
	baseURL    string
	username   string
	password   string
	apiVersion string
}

// GetAPIVersion returns the API version being used by this client.
func (c *HAProxyClient) GetAPIVersion() string {
	return c.apiVersion
}

// NewHAProxyClient creates a new HAProxy client.
func NewHAProxyClient(httpClient *http.Client, baseURL, username, password, apiVersion string) *HAProxyClient {
	return &HAProxyClient{
		httpClient: httpClient,
		baseURL:    baseURL,
		username:   username,
		password:   password,
		apiVersion: apiVersion,
	}
}

// CreateFrontend creates a new frontend.
// DEPRECATED: Use CreateFrontendInTransaction for new code
func (c *HAProxyClient) CreateFrontend(ctx context.Context, payload *FrontendPayload) error {
	transactionID, err := c.BeginTransaction()
	if err != nil {
		return err
	}
	defer c.RollbackTransaction(transactionID)

	if err := c.CreateFrontendInTransaction(ctx, transactionID, payload); err != nil {
		return err
	}

	return c.CommitTransaction(transactionID)
}

// CreateFrontendInTransaction creates a new frontend using an existing transaction ID.
func (c *HAProxyClient) CreateFrontendInTransaction(ctx context.Context, transactionID string, payload *FrontendPayload) error {
	log.Printf("CreateFrontendInTransaction called with transaction ID: %s, payload: %+v", transactionID, payload)
	req, err := c.newRequest(ctx, "POST", fmt.Sprintf("/services/haproxy/configuration/frontends?transaction_id=%s", transactionID), payload)
	if err != nil {
		return err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("frontend creation failed with status %d", resp.StatusCode)
		}
		return fmt.Errorf("frontend creation failed with status %d: %s", resp.StatusCode, string(body))
	}

	log.Printf("Frontend created successfully in transaction: %s", transactionID)
	return nil
}

// UpdateFrontendInTransaction updates an existing frontend using an existing transaction ID.
func (c *HAProxyClient) UpdateFrontendInTransaction(ctx context.Context, transactionID string, payload *FrontendPayload) error {
	log.Printf("UpdateFrontendInTransaction called with transaction ID: %s, payload: %+v", transactionID, payload)
	req, err := c.newRequest(ctx, "PUT", fmt.Sprintf("/services/haproxy/configuration/frontends/%s?transaction_id=%s", payload.Name, transactionID), payload)
	if err != nil {
		return err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("frontend update failed with status %d", resp.StatusCode)
		}
		return fmt.Errorf("frontend update failed with status %d: %s", resp.StatusCode, string(body))
	}

	log.Printf("Frontend updated successfully in transaction: %s", transactionID)
	return nil
}

// DeleteFrontendInTransaction deletes an existing frontend using an existing transaction ID.
func (c *HAProxyClient) DeleteFrontendInTransaction(ctx context.Context, transactionID string, frontendName string) error {
	log.Printf("DeleteFrontendInTransaction called with transaction ID: %s, frontend: %s", transactionID, frontendName)
	req, err := c.newRequest(ctx, "DELETE", fmt.Sprintf("/services/haproxy/configuration/frontends/%s?transaction_id=%s", frontendName, transactionID), nil)
	if err != nil {
		return err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted && resp.StatusCode != http.StatusNoContent {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("frontend deletion failed with status %d", resp.StatusCode)
		}
		return fmt.Errorf("frontend deletion failed with status %d: %s", resp.StatusCode, string(body))
	}

	log.Printf("Frontend deleted successfully in transaction: %s", transactionID)
	return nil
}

// CreateACL creates a new ACL rule for a frontend.
// DEPRECATED: Use CreateACLInTransaction for new code
func (c *HAProxyClient) CreateACL(ctx context.Context, parentType, parentName string, payload *ACLPayload) error {
	transactionID, err := c.BeginTransaction()
	if err != nil {
		return err
	}
	defer c.RollbackTransaction(transactionID)

	if err := c.CreateACLInTransaction(ctx, transactionID, parentType, parentName, payload); err != nil {
		return err
	}

	return c.CommitTransaction(transactionID)
}

// CreateACLInTransaction creates a new ACL rule using an existing transaction ID.
func (c *HAProxyClient) CreateACLInTransaction(ctx context.Context, transactionID, parentType, parentName string, payload *ACLPayload) error {
	// Debug: Log the ACL payload being sent
	payloadJSON, _ := json.Marshal(payload)
	log.Printf("DEBUG: Creating ACL in transaction %s with payload: %s", transactionID, string(payloadJSON))

	var url string
	if c.apiVersion == "v3" {
		// v3: Use nested endpoint with index-based positioning
		// Use the actual index from the payload for proper ordering
		// Properly pluralize the parent type
		parentTypePlural := parentType + "s"
		url = fmt.Sprintf("/services/haproxy/configuration/%s/%s/acls/%d?transaction_id=%s",
			parentTypePlural, parentName, payload.Index, transactionID)
	} else {
		// v2: Use query parameter approach
		url = fmt.Sprintf("/services/haproxy/configuration/acls?parent_type=%s&parent_name=%s&transaction_id=%s",
			parentType, parentName, transactionID)
	}

	log.Printf("DEBUG: Using ACL endpoint: %s for API version %s", url, c.apiVersion)

	req, err := c.newRequest(ctx, "POST", url, payload)
	if err != nil {
		return err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("ACL creation failed with status %d", resp.StatusCode)
		}
		return fmt.Errorf("ACL creation failed with status %d: %s", resp.StatusCode, string(body))
	}

	log.Printf("ACL created successfully in transaction: %s", transactionID)

	return nil
}

// CreateAllACLsInTransaction creates all ACLs at once using an existing transaction ID
func (c *HAProxyClient) CreateAllACLsInTransaction(ctx context.Context, transactionID, parentType, parentName string, payloads []ACLPayload) error {
	var url string
	var method string

	if c.apiVersion == "v3" {
		// v3: Use nested endpoint under frontends/backends - send all at once
		parentTypePlural := parentType + "s"
		url = fmt.Sprintf("/services/haproxy/configuration/%s/%s/acls?transaction_id=%s",
			parentTypePlural, parentName, transactionID)
		method = "PUT"

		// Debug logging for v3
		payloadJSON, _ := json.Marshal(payloads)
		log.Printf("DEBUG: API %s - Creating all ACLs at once:", c.apiVersion)
		log.Printf("DEBUG: API %s - Method: %s", c.apiVersion, method)
		log.Printf("DEBUG: API %s - Endpoint: %s", c.apiVersion, url)
		log.Printf("DEBUG: API %s - Payload count: %d", c.apiVersion, len(payloads))
		log.Printf("DEBUG: API %s - Payload: %s", c.apiVersion, string(payloadJSON))

		req, err := c.newRequest(ctx, method, url, payloads)
		if err != nil {
			return err
		}

		resp, err := c.httpClient.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				return fmt.Errorf("ACLs creation failed with status %d", resp.StatusCode)
			}
			return fmt.Errorf("ACLs creation failed with status %d: %s", resp.StatusCode, string(body))
		}

		log.Printf("All ACLs created successfully in transaction: %s", transactionID)
		return nil
	} else {
		// v2: Create ACLs individually (v2 doesn't support bulk creation)
		log.Printf("DEBUG: API %s - Creating ACLs individually (v2 limitation):", c.apiVersion)
		log.Printf("DEBUG: API %s - Payload count: %d", c.apiVersion, len(payloads))

		for i, payload := range payloads {
			url := fmt.Sprintf("/services/haproxy/configuration/acls?parent_type=%s&parent_name=%s&transaction_id=%s",
				parentType, parentName, transactionID)
			method := "POST"

			// Debug logging for each individual ACL
			payloadJSON, _ := json.Marshal(payload)
			log.Printf("DEBUG: API %s - Creating ACL %d/%d:", c.apiVersion, i+1, len(payloads))
			log.Printf("DEBUG: API %s - Method: %s", c.apiVersion, method)
			log.Printf("DEBUG: API %s - Endpoint: %s", c.apiVersion, url)
			log.Printf("DEBUG: API %s - Payload: %s", c.apiVersion, string(payloadJSON))

			req, err := c.newRequest(ctx, method, url, payload)
			if err != nil {
				return err
			}

			resp, err := c.httpClient.Do(req)
			if err != nil {
				return err
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
				body, err := io.ReadAll(resp.Body)
				if err != nil {
					return fmt.Errorf("ACL %d creation failed with status %d", i+1, resp.StatusCode)
				}
				return fmt.Errorf("ACL %d creation failed with status %d: %s", i+1, resp.StatusCode, string(body))
			}

			log.Printf("ACL %d/%d created successfully in transaction: %s", i+1, len(payloads), transactionID)
		}

		log.Printf("All %d ACLs created successfully in transaction: %s", len(payloads), transactionID)
		return nil
	}
}

// UpdateACLInTransaction updates an existing ACL rule using an existing transaction ID.
func (c *HAProxyClient) UpdateACLInTransaction(ctx context.Context, transactionID, parentType, parentName string, index int64, payload *ACLPayload) error {
	// Debug: Log the ACL payload being sent
	payloadJSON, _ := json.Marshal(payload)
	log.Printf("DEBUG: Updating ACL in transaction %s with payload: %s", transactionID, string(payloadJSON))

	var url string
	if c.apiVersion == "v3" {
		// v3: Use nested endpoint with index-based positioning
		// Properly pluralize the parent type
		parentTypePlural := parentType + "s"
		url = fmt.Sprintf("/services/haproxy/configuration/%s/%s/acls/%d?transaction_id=%s",
			parentTypePlural, parentName, index, transactionID)
	} else {
		// v2: Use query parameter approach
		url = fmt.Sprintf("/services/haproxy/configuration/acls/%d?parent_type=%s&parent_name=%s&transaction_id=%s",
			index, parentType, parentName, transactionID)
	}

	log.Printf("DEBUG: Using ACL update endpoint: %s for API version %s", url, c.apiVersion)

	req, err := c.newRequest(ctx, "PUT", url, payload)
	if err != nil {
		return err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("ACL update failed with status %d", resp.StatusCode)
		}
		return fmt.Errorf("ACL update failed with status %d: %s", resp.StatusCode, string(body))
	}

	log.Printf("ACL updated successfully in transaction: %s", transactionID)
	return nil
}

// DeleteACLInTransaction deletes an existing ACL rule using an existing transaction ID.
func (c *HAProxyClient) DeleteACLInTransaction(ctx context.Context, transactionID, parentType, parentName string, index int64) error {
	log.Printf("DEBUG: Deleting ACL in transaction %s at index %d", transactionID, index)

	var url string
	if c.apiVersion == "v3" {
		// v3: Use nested endpoint with index-based positioning
		// Properly pluralize the parent type
		parentTypePlural := parentType + "s"
		url = fmt.Sprintf("/services/haproxy/configuration/%s/%s/acls/%d?transaction_id=%s",
			parentTypePlural, parentName, index, transactionID)
	} else {
		// v2: Use query parameter approach
		url = fmt.Sprintf("/services/haproxy/configuration/acls/%d?parent_type=%s&parent_name=%s&transaction_id=%s",
			index, parentType, parentName, transactionID)
	}

	log.Printf("DEBUG: Using ACL delete endpoint: %s for API version %s", url, c.apiVersion)

	req, err := c.newRequest(ctx, "DELETE", url, nil)
	if err != nil {
		return err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted && resp.StatusCode != http.StatusNoContent {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("ACL deletion failed with status %d", resp.StatusCode)
		}
		return fmt.Errorf("ACL deletion failed with status %d: %s", resp.StatusCode, string(body))
	}

	log.Printf("ACL deleted successfully in transaction: %s", transactionID)
	return nil
}

// ReadACLs reads all ACL rules for a parent (frontend, backend, etc.).
func (c *HAProxyClient) ReadACLs(ctx context.Context, parentType, parentName string) ([]ACLPayload, error) {
	var url string

	if c.apiVersion == "v3" {
		// v3: Use nested endpoint under frontends/backends
		// Properly pluralize the parent type
		parentTypePlural := parentType + "s"
		url = fmt.Sprintf("/services/haproxy/configuration/%s/%s/acls", parentTypePlural, parentName)
	} else {
		// v2: Use query parameter approach
		url = fmt.Sprintf("/services/haproxy/configuration/acls?parent_type=%s&parent_name=%s", parentType, parentName)
	}

	log.Printf("DEBUG: Using ACL read endpoint: %s for API version %s", url, c.apiVersion)

	req, err := c.newRequest(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var acls []ACLPayload

	if c.apiVersion == "v3" {
		// v3: Response is a direct array, no wrapper
		if err := json.NewDecoder(resp.Body).Decode(&acls); err != nil {
			return nil, fmt.Errorf("failed to decode v3 ACL response: %w", err)
		}
		log.Printf("DEBUG: Raw ACL response from HAProxy: %+v", acls)
	} else {
		// v2: Response has a data wrapper
		var response struct {
			Data []ACLPayload `json:"data"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
			return nil, fmt.Errorf("failed to decode v2 ACL response: %w", err)
		}
		acls = response.Data
		log.Printf("DEBUG: Raw ACL response from HAProxy: %+v", acls)
	}

	return acls, nil
}

// UpdateACL updates an existing ACL rule by index.
// DEPRECATED: Use UpdateACLInTransaction for new code
func (c *HAProxyClient) UpdateACL(ctx context.Context, parentType, parentName string, index int64, payload *ACLPayload) error {
	transactionID, err := c.BeginTransaction()
	if err != nil {
		return err
	}
	defer c.RollbackTransaction(transactionID)

	if err := c.UpdateACLInTransaction(ctx, transactionID, parentType, parentName, index, payload); err != nil {
		return err
	}

	return c.CommitTransaction(transactionID)
}

// DeleteACL deletes an ACL rule by index.
// DEPRECATED: Use DeleteACLInTransaction for new code
func (c *HAProxyClient) DeleteACL(ctx context.Context, parentType, parentName string, index int64) error {
	transactionID, err := c.BeginTransaction()
	if err != nil {
		return err
	}
	defer c.RollbackTransaction(transactionID)

	if err := c.DeleteACLInTransaction(ctx, transactionID, parentType, parentName, index); err != nil {
		return err
	}

	return c.CommitTransaction(transactionID)
}

// CreateAllResourcesInSingleTransaction creates all resources in a single transaction.
// This ensures atomic operations - all resources succeed or all fail together.
// Includes retry mechanism for concurrency issues when multiple workspaces run in parallel.
func (c *HAProxyClient) CreateAllResourcesInSingleTransaction(ctx context.Context, resources *AllResourcesPayload) error {
	log.Printf("Creating all resources in single transaction with retry mechanism")

	const (
		maxRetries = 10
		retryDelay = 2 * time.Second
	)

	retryCount := 0
	for {
		log.Printf("Attempt %d/%d: Creating all resources in single transaction", retryCount+1, maxRetries)

		// Begin transaction
		transactionID, err := c.BeginTransaction()
		if err != nil {
			log.Printf("Attempt %d: Failed to begin transaction: %v", retryCount+1, err)
			if c.isRetryableError(err) {
				retryCount++
				if retryCount >= maxRetries {
					return fmt.Errorf("failed to begin transaction after %d retries: %v", maxRetries, err)
				}
				log.Printf("Attempt %d: Retrying in %v...", retryCount+1, retryDelay)
				time.Sleep(retryDelay)
				continue
			}
			return fmt.Errorf("failed to begin transaction (non-retryable): %v", err)
		}

		log.Printf("Attempt %d: Transaction ID created: %s", retryCount+1, transactionID)

		// Create all resources in the transaction
		err = c.createResourcesInTransaction(ctx, transactionID, resources)
		if err != nil {
			log.Printf("Attempt %d: Resource creation failed in transaction %s: %v", retryCount+1, transactionID, err)
			// Try to rollback the transaction
			if rollbackErr := c.RollbackTransaction(transactionID); rollbackErr != nil {
				log.Printf("Warning: Failed to rollback transaction %s: %v", transactionID, rollbackErr)
			}

			if c.isRetryableError(err) {
				retryCount++
				if retryCount >= maxRetries {
					return fmt.Errorf("resource creation failed after %d retries: %v", maxRetries, err)
				}
				log.Printf("Attempt %d: Retrying in %v...", retryCount+1, retryDelay)
				time.Sleep(retryDelay)
				continue
			}
			return fmt.Errorf("resource creation failed (non-retryable): %v", err)
		}

		// Commit transaction
		log.Printf("Attempt %d: Committing transaction %s", retryCount+1, transactionID)
		err = c.CommitTransaction(transactionID)
		if err != nil {
			log.Printf("Attempt %d: Commit failed for transaction %s: %v", retryCount+1, transactionID, err)

			if c.isRetryableError(err) {
				retryCount++
				if retryCount >= maxRetries {
					return fmt.Errorf("failed to commit transaction after %d retries: %v", maxRetries, err)
				}
				log.Printf("Attempt %d: Retrying in %v...", retryCount+1, retryDelay)
				time.Sleep(retryDelay)
				continue
			}
			return fmt.Errorf("failed to commit transaction (non-retryable): %v", err)
		}

		log.Printf("Success! Transaction %s committed successfully - all resources created in %d attempts", transactionID, retryCount+1)
		return nil
	}
}

// UpdateAllResourcesInSingleTransaction updates all resources in a single transaction.
// This ensures atomic operations - all resources succeed or all fail together.
// Includes retry mechanism for concurrency issues when multiple workspaces run in parallel.
func (c *HAProxyClient) UpdateAllResourcesInSingleTransaction(ctx context.Context, resources *AllResourcesPayload) error {
	log.Printf("Updating all resources in single transaction with retry mechanism")

	const (
		maxRetries = 10
		retryDelay = 2 * time.Second
	)

	retryCount := 0
	for {
		log.Printf("Attempt %d/%d: Updating all resources in single transaction", retryCount+1, maxRetries)

		// Begin transaction
		transactionID, err := c.BeginTransaction()
		if err != nil {
			log.Printf("Attempt %d: Failed to begin transaction: %v", retryCount+1, err)
			if c.isRetryableError(err) {
				retryCount++
				if retryCount >= maxRetries {
					return fmt.Errorf("failed to begin transaction after %d retries: %v", maxRetries, err)
				}
				log.Printf("Attempt %d: Retrying in %v...", retryCount+1, retryDelay)
				time.Sleep(retryDelay)
				continue
			}
			return fmt.Errorf("failed to begin transaction (non-retryable): %v", err)
		}

		log.Printf("Attempt %d: Transaction ID created: %s", retryCount+1, transactionID)

		// Update all resources in the transaction
		err = c.updateResourcesInTransaction(ctx, transactionID, resources)
		if err != nil {
			log.Printf("Attempt %d: Resource update failed in transaction %s: %v", retryCount+1, transactionID, err)
			// Try to rollback the transaction
			if rollbackErr := c.RollbackTransaction(transactionID); rollbackErr != nil {
				log.Printf("Warning: Failed to rollback transaction %s: %v", transactionID, rollbackErr)
			}

			if c.isRetryableError(err) {
				retryCount++
				if retryCount >= maxRetries {
					return fmt.Errorf("resource update failed after %d retries: %v", maxRetries, err)
				}
				log.Printf("Attempt %d: Retrying in %v...", retryCount+1, retryDelay)
				time.Sleep(retryDelay)
				continue
			}
			return fmt.Errorf("resource update failed (non-retryable): %v", err)
		}

		// Commit transaction
		log.Printf("Attempt %d: Committing transaction %s", retryCount+1, transactionID)
		err = c.CommitTransaction(transactionID)
		if err != nil {
			log.Printf("Attempt %d: Commit failed for transaction %s: %v", retryCount+1, transactionID, err)

			if c.isRetryableError(err) {
				retryCount++
				if retryCount >= maxRetries {
					return fmt.Errorf("failed to commit transaction after %d retries: %v", maxRetries, err)
				}
				log.Printf("Attempt %d: Retrying in %v...", retryCount+1, retryDelay)
				time.Sleep(retryDelay)
				continue
			}
			return fmt.Errorf("failed to commit transaction (non-retryable): %v", err)
		}

		log.Printf("Success! Transaction %s committed successfully - all resources updated in %d attempts", transactionID, retryCount+1)
		return nil
	}
}

// DeleteAllResourcesInSingleTransaction deletes all resources in a single transaction.
// This ensures atomic operations - all resources succeed or all fail together.
// Includes retry mechanism for concurrency issues when multiple workspaces run in parallel.
func (c *HAProxyClient) DeleteAllResourcesInSingleTransaction(ctx context.Context, resources *AllResourcesPayload) error {
	log.Printf("Deleting all resources in single transaction with retry mechanism")

	const (
		maxRetries = 10
		retryDelay = 2 * time.Second
	)

	retryCount := 0
	for {
		log.Printf("Attempt %d/%d: Deleting all resources in single transaction", retryCount+1, maxRetries)

		// Begin transaction
		transactionID, err := c.BeginTransaction()
		if err != nil {
			log.Printf("Attempt %d: Failed to begin transaction: %v", retryCount+1, err)
			if c.isRetryableError(err) {
				retryCount++
				if retryCount >= maxRetries {
					return fmt.Errorf("failed to begin transaction after %d retries: %v", maxRetries, err)
				}
				log.Printf("Attempt %d: Retrying in %v...", retryCount+1, retryDelay)
				time.Sleep(retryDelay)
				continue
			}
			return fmt.Errorf("failed to begin transaction (non-retryable): %v", err)
		}

		log.Printf("Attempt %d: Transaction ID created: %s", retryCount+1, transactionID)

		// Delete all resources in the transaction
		err = c.deleteResourcesInTransaction(ctx, transactionID, resources)
		if err != nil {
			log.Printf("Attempt %d: Resource deletion failed in transaction %s: %v", retryCount+1, transactionID, err)
			// Try to rollback the transaction
			if rollbackErr := c.RollbackTransaction(transactionID); rollbackErr != nil {
				log.Printf("Warning: Failed to rollback transaction %s: %v", transactionID, rollbackErr)
			}

			if c.isRetryableError(err) {
				retryCount++
				if retryCount >= maxRetries {
					return fmt.Errorf("resource deletion failed after %d retries: %v", maxRetries, err)
				}
				log.Printf("Attempt %d: Retrying in %v...", retryCount+1, retryDelay)
				time.Sleep(retryDelay)
				continue
			}
			return fmt.Errorf("resource deletion failed (non-retryable): %v", err)
		}

		// Commit transaction
		log.Printf("Attempt %d: Committing transaction %s", retryCount+1, transactionID)
		err = c.CommitTransaction(transactionID)
		if err != nil {
			log.Printf("Attempt %d: Commit failed for transaction %s: %v", retryCount+1, transactionID, err)

			if c.isRetryableError(err) {
				retryCount++
				if retryCount >= maxRetries {
					return fmt.Errorf("failed to commit transaction after %d retries: %v", maxRetries, err)
				}
				log.Printf("Attempt %d: Retrying in %v...", retryCount+1, retryDelay)
				time.Sleep(retryDelay)
				continue
			}
			return fmt.Errorf("failed to commit transaction (non-retryable): %v", err)
		}

		log.Printf("Success! Transaction %s committed successfully - all resources deleted in %d attempts", transactionID, retryCount+1)
		return nil
	}
}

// createResourcesInTransaction creates all resources within an existing transaction
func (c *HAProxyClient) createResourcesInTransaction(ctx context.Context, transactionID string, resources *AllResourcesPayload) error {
	// Create backend first (if provided)
	if resources.Backend != nil {
		log.Printf("Creating backend in transaction %s", transactionID)
		err := c.CreateBackendInTransaction(ctx, transactionID, resources.Backend)
		if err != nil {
			return fmt.Errorf("backend creation failed: %v", err)
		}
		log.Printf("Backend created successfully in transaction %s", transactionID)
	}

	// Create servers (if provided)
	if len(resources.Servers) > 0 {
		for i, server := range resources.Servers {
			log.Printf("Creating server %d/%d in transaction %s", i+1, len(resources.Servers), transactionID)
			err := c.CreateServerInTransaction(ctx, transactionID, server.ParentType, server.ParentName, server.Payload)
			if err != nil {
				return fmt.Errorf("server %d creation failed: %v", i+1, err)
			}
			log.Printf("Server %d created successfully in transaction %s", i+1, transactionID)
		}
	}

	// Create frontend last (if provided)
	if resources.Frontend != nil {
		log.Printf("Creating frontend in transaction %s", transactionID)
		err := c.CreateFrontendInTransaction(ctx, transactionID, resources.Frontend)
		if err != nil {
			return fmt.Errorf("frontend creation failed: %v", err)
		}
		log.Printf("Frontend created successfully in transaction %s", transactionID)
	}

	// Create ACLs after frontend is created (if provided)
	if len(resources.Acls) > 0 {
		for i, acl := range resources.Acls {
			log.Printf("Creating ACL %d/%d in transaction %s", i+1, len(resources.Acls), transactionID)
			err := c.CreateACLInTransaction(ctx, transactionID, acl.ParentType, acl.ParentName, acl.Payload)
			if err != nil {
				return fmt.Errorf("ACL %d creation failed: %v", i+1, err)
			}
			log.Printf("ACL %d created successfully in transaction %s", i+1, transactionID)
		}
	}

	return nil
}

// updateResourcesInTransaction updates all resources within an existing transaction
func (c *HAProxyClient) updateResourcesInTransaction(ctx context.Context, transactionID string, resources *AllResourcesPayload) error {
	// Update backend first (if provided)
	if resources.Backend != nil {
		log.Printf("Updating backend in transaction %s", transactionID)
		err := c.UpdateBackendInTransaction(ctx, transactionID, resources.Backend)
		if err != nil {
			return fmt.Errorf("backend update failed: %v", err)
		}
		log.Printf("Backend updated successfully in transaction %s", transactionID)
	}

	// Update servers (if provided)
	if len(resources.Servers) > 0 {
		for i, server := range resources.Servers {
			log.Printf("Updating server %d/%d in transaction %s", i+1, len(resources.Servers), transactionID)
			err := c.UpdateServerInTransaction(ctx, transactionID, server.ParentType, server.ParentName, server.Payload)
			if err != nil {
				return fmt.Errorf("server %d update failed: %v", i+1, err)
			}
			log.Printf("Server %d updated successfully in transaction %s", i+1, transactionID)
		}
	}

	// Update frontend last (if provided)
	if resources.Frontend != nil {
		log.Printf("Updating frontend in transaction %s", transactionID)
		err := c.UpdateFrontendInTransaction(ctx, transactionID, resources.Frontend)
		if err != nil {
			return fmt.Errorf("frontend update failed: %v", err)
		}
		log.Printf("Frontend updated successfully in transaction %s", transactionID)
	}

	// Update ACLs after frontend is updated (if provided)
	if len(resources.Acls) > 0 {
		for i, acl := range resources.Acls {
			log.Printf("Updating ACL %d/%d in transaction %s", i+1, len(resources.Acls), transactionID)
			err := c.UpdateACLInTransaction(ctx, transactionID, acl.ParentType, acl.ParentName, acl.Payload.Index, acl.Payload)
			if err != nil {
				return fmt.Errorf("ACL %d update failed: %v", i+1, err)
			}
			log.Printf("ACL %d updated successfully in transaction %s", i+1, transactionID)
		}
	}

	return nil
}

// deleteResourcesInTransaction deletes all resources within an existing transaction
func (c *HAProxyClient) deleteResourcesInTransaction(ctx context.Context, transactionID string, resources *AllResourcesPayload) error {
	// Delete ACLs first (they depend on frontend)
	if len(resources.Acls) > 0 {
		log.Printf("Attempting to delete %d ACLs in transaction %s", len(resources.Acls), transactionID)

		for i, acl := range resources.Acls {
			log.Printf("Deleting ACL %d/%d in transaction %s", i+1, len(resources.Acls), transactionID)

			err := c.DeleteACLInTransaction(ctx, transactionID, acl.ParentType, acl.ParentName, acl.Payload.Index)
			if err != nil {
				// Check if this is a "not found" error (ACL already deleted or wrong index)
				if strings.Contains(err.Error(), "404") || strings.Contains(err.Error(), "does not exist") || strings.Contains(err.Error(), "missing object") {
					log.Printf("Warning: ACL %d at index %d not found (likely already deleted): %v", i+1, acl.Payload.Index, err)
					// Continue with deletion - this ACL is already gone
					continue
				}

				// For other errors, log warning but continue (don't fail the entire transaction)
				log.Printf("Warning: ACL %d deletion failed (continuing): %v", i+1, err)
				continue
			}

			log.Printf("ACL %d deleted successfully in transaction %s", i+1, transactionID)
		}

		log.Printf("ACL deletion phase completed in transaction %s", transactionID)
	}

	// Delete frontend (if provided)
	if resources.Frontend != nil {
		log.Printf("Deleting frontend in transaction %s", transactionID)
		err := c.DeleteFrontendInTransaction(ctx, transactionID, resources.Frontend.Name)
		if err != nil {
			return fmt.Errorf("frontend deletion failed: %v", err)
		}
		log.Printf("Frontend deleted successfully in transaction %s", transactionID)
	}

	// Delete servers (if provided)
	if len(resources.Servers) > 0 {
		for i, server := range resources.Servers {
			log.Printf("Deleting server %d/%d in transaction %s", i+1, len(resources.Servers), transactionID)
			err := c.DeleteServerInTransaction(ctx, transactionID, server.ParentType, server.ParentName, server.Payload.Name)
			if err != nil {
				return fmt.Errorf("server %d deletion failed: %v", i+1, err)
			}
			log.Printf("Server %d deleted successfully in transaction %s", i+1, transactionID)
		}
	}

	// Delete backend last (if provided)
	if resources.Backend != nil {
		log.Printf("Deleting backend in transaction %s", transactionID)
		err := c.DeleteBackendInTransaction(ctx, transactionID, resources.Backend.Name)
		if err != nil {
			return fmt.Errorf("backend deletion failed: %v", err)
		}
		log.Printf("Backend deleted successfully in transaction %s", transactionID)
	}

	return nil
}

// isRetryableError determines if an error is retryable based on concurrency issues
func (c *HAProxyClient) isRetryableError(err error) bool {
	// Check for transaction-related concurrency errors
	if TransactionOutdated(err) {
		log.Printf("Retryable error: Transaction outdated")
		return true
	}
	if TransactionDoesNotExist(err) {
		log.Printf("Retryable error: Transaction does not exist")
		return true
	}
	if isVersionMismatch(err) {
		log.Printf("Retryable error: Version mismatch")
		return true
	}
	if isVersionOrTransSpecified(err) {
		log.Printf("Retryable error: Version or transaction not specified")
		return true
	}

	// Check for configuration validation errors that might be transient
	if strings.Contains(err.Error(), "validation error") && strings.Contains(err.Error(), "defaults section") {
		log.Printf("Retryable error: Configuration validation error (may be transient)")
		return true
	}

	return false
}

// ReadFrontend reads a frontend.
func (c *HAProxyClient) ReadFrontend(ctx context.Context, name string) (*FrontendPayload, error) {
	req, err := c.newRequest(ctx, "GET", fmt.Sprintf("/services/haproxy/configuration/frontends/%s", name), nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var frontend struct {
		Data FrontendPayload `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&frontend); err != nil {
		return nil, err
	}

	return &frontend.Data, nil
}

// UpdateFrontend updates a frontend.
func (c *HAProxyClient) UpdateFrontend(ctx context.Context, name string, payload *FrontendPayload) error {
	_, err := c.Transaction(func(transactionID string) (*http.Response, error) {
		req, err := c.newRequest(ctx, "PUT", fmt.Sprintf("/services/haproxy/configuration/frontends/%s?transaction_id=%s", name, transactionID), payload)
		if err != nil {
			return nil, err
		}
		return c.httpClient.Do(req)
	})
	return err
}

// CreateBackend creates a new backend.
func (c *HAProxyClient) CreateBackend(ctx context.Context, payload *BackendPayload) error {
	log.Printf("CreateBackend called with payload: %+v", payload)
	_, err := c.Transaction(func(transactionID string) (*http.Response, error) {
		log.Printf("CreateBackend executing in transaction: %s", transactionID)
		req, err := c.newRequest(ctx, "POST", fmt.Sprintf("/services/haproxy/configuration/backends?transaction_id=%s", transactionID), payload)
		if err != nil {
			return nil, err
		}
		return c.httpClient.Do(req)
	})
	return err
}

// CreateBackendInTransaction creates a new backend using an existing transaction ID.
func (c *HAProxyClient) CreateBackendInTransaction(ctx context.Context, transactionID string, payload *BackendPayload) error {
	log.Printf("CreateBackendInTransaction called with transaction ID: %s, payload: %+v", transactionID, payload)
	req, err := c.newRequest(ctx, "POST", fmt.Sprintf("/services/haproxy/configuration/backends?transaction_id=%s", transactionID), payload)
	if err != nil {
		return err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("backend creation failed with status %d", resp.StatusCode)
		}
		return fmt.Errorf("backend creation failed with status %d: %s", resp.StatusCode, string(body))
	}

	log.Printf("Backend created successfully in transaction: %s", transactionID)
	return nil
}

// UpdateBackendInTransaction updates an existing backend using an existing transaction ID.
func (c *HAProxyClient) UpdateBackendInTransaction(ctx context.Context, transactionID string, payload *BackendPayload) error {
	log.Printf("UpdateBackendInTransaction called with transaction ID: %s, payload: %+v", transactionID, payload)
	req, err := c.newRequest(ctx, "PUT", fmt.Sprintf("/services/haproxy/configuration/backends/%s?transaction_id=%s", payload.Name, transactionID), payload)
	if err != nil {
		return err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("backend update failed with status %d", resp.StatusCode)
		}
		return fmt.Errorf("backend update failed with status %d: %s", resp.StatusCode, string(body))
	}

	log.Printf("Backend updated successfully in transaction: %s", transactionID)
	return nil
}

// DeleteBackendInTransaction deletes an existing backend using an existing transaction ID.
func (c *HAProxyClient) DeleteBackendInTransaction(ctx context.Context, transactionID string, backendName string) error {
	log.Printf("DeleteBackendInTransaction called with transaction ID: %s, backend: %s", transactionID, backendName)
	req, err := c.newRequest(ctx, "DELETE", fmt.Sprintf("/services/haproxy/configuration/backends/%s?transaction_id=%s", backendName, transactionID), nil)
	if err != nil {
		return err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted && resp.StatusCode != http.StatusNoContent {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("backend deletion failed with status %d", resp.StatusCode)
		}
		return fmt.Errorf("backend deletion failed with status %d: %s", resp.StatusCode, string(body))
	}

	log.Printf("Backend deleted successfully in transaction: %s", transactionID)
	return nil
}

// ReadBackend reads a backend.
func (c *HAProxyClient) ReadBackend(ctx context.Context, name string) (*BackendPayload, error) {
	req, err := c.newRequest(ctx, "GET", fmt.Sprintf("/services/haproxy/configuration/backends/%s", name), nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var backend struct {
		Data BackendPayload `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&backend); err != nil {
		return nil, err
	}

	return &backend.Data, nil
}

// UpdateBackend updates a backend.
func (c *HAProxyClient) UpdateBackend(ctx context.Context, name string, payload *BackendPayload) error {
	_, err := c.Transaction(func(transactionID string) (*http.Response, error) {
		req, err := c.newRequest(ctx, "PUT", fmt.Sprintf("/services/haproxy/configuration/backends/%s?transaction_id=%s", name, transactionID), payload)
		if err != nil {
			return nil, err
		}
		return c.httpClient.Do(req)
	})
	return err
}

// DeleteBackend deletes a backend.
func (c *HAProxyClient) DeleteBackend(ctx context.Context, name string) error {
	_, err := c.Transaction(func(transactionID string) (*http.Response, error) {
		req, err := c.newRequest(ctx, "DELETE", fmt.Sprintf("/services/haproxy/configuration/backends/%s?transaction_id=%s", name, transactionID), nil)
		if err != nil {
			return nil, err
		}
		return c.httpClient.Do(req)
	})
	return err
}

// CreateServer creates a new server.
func (c *HAProxyClient) CreateServer(ctx context.Context, parentType, parentName string, payload *ServerPayload) error {
	resp, err := c.Transaction(func(transactionID string) (*http.Response, error) {
		req, err := c.newRequest(ctx, "POST", fmt.Sprintf("/services/haproxy/configuration/servers?parent_type=%s&parent_name=%s&transaction_id=%s", parentType, parentName, transactionID), payload)
		if err != nil {
			return nil, err
		}
		return c.httpClient.Do(req)
	})
	if err != nil {
		return err
	}

	// Check if the server creation was successful
	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		// Try to read the error response body for more details
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("failed to create server: unexpected status code: %d (could not read error body)", resp.StatusCode)
		}

		// Try to parse as JSON error response
		var errorResp struct {
			Message string `json:"message"`
			Code    int    `json:"code"`
		}
		if json.Unmarshal(body, &errorResp) == nil && errorResp.Message != "" {
			// Return CustomError that transaction retry can detect
			apiError := &utils.APIError{
				Code:    errorResp.Code,
				Message: errorResp.Message,
			}
			return utils.NewCustomError("Server creation failed", apiError)
		}

		// Fallback to raw body if not JSON
		if len(body) > 0 {
			apiError := &utils.APIError{
				Code:    resp.StatusCode,
				Message: string(body),
			}
			return utils.NewCustomError("Server creation failed", apiError)
		}

		apiError := &utils.APIError{
			Code:    resp.StatusCode,
			Message: "Unknown error",
		}
		return utils.NewCustomError("Server creation failed", apiError)
	}

	return nil
}

// ReadServers reads all servers for a given parent.
func (c *HAProxyClient) ReadServers(ctx context.Context, parentType, parentName string) ([]ServerPayload, error) {
	// Use version-aware URL construction
	apiVersion := c.GetAPIVersion()
	var url string
	if apiVersion == "v3" {
		// For v3, use the correct endpoint structure: /services/haproxy/configuration/backends/{parent_name}/servers
		// Note: newRequest() already adds the /v3 prefix, so we don't include it here
		if parentType == "backend" {
			url = fmt.Sprintf("/services/haproxy/configuration/backends/%s/servers", parentName)
		} else {
			// For other parent types, use the generic endpoint
			url = fmt.Sprintf("/services/haproxy/configuration/servers?parent_type=%s&parent_name=%s", parentType, parentName)
		}
	} else {
		url = fmt.Sprintf("/services/haproxy/configuration/servers?parent_type=%s&parent_name=%s", parentType, parentName)
	}
	log.Printf("DEBUG: ReadServers URL: %s (API version: %s)", url, apiVersion)
	log.Printf("DEBUG: ReadServers parentType: %s, parentName: %s", parentType, parentName)

	req, err := c.newRequest(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	// Log the full request details
	log.Printf("DEBUG: ReadServers full request URL: %s", req.URL.String())
	log.Printf("DEBUG: ReadServers request method: %s", req.Method)
	log.Printf("DEBUG: ReadServers request headers: %v", req.Header)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	log.Printf("DEBUG: ReadServers response status: %d", resp.StatusCode)

	// Read response body for debugging
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	log.Printf("DEBUG: ReadServers response body: %s", sanitizeResponseBody(string(bodyBytes)))

	if resp.StatusCode == http.StatusNotFound {
		log.Printf("DEBUG: ReadServers - no servers found (404)")
		return []ServerPayload{}, nil // No servers found is not an error
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Try to parse as direct array first (HAProxy v3 format)
	var servers []ServerPayload
	if err := json.Unmarshal(bodyBytes, &servers); err != nil {
		// If that fails, try to parse as wrapper object (HAProxy v2 format)
		var serversWrapper struct {
			Data []ServerPayload `json:"data"`
		}
		if err := json.Unmarshal(bodyBytes, &serversWrapper); err != nil {
			return nil, err
		}
		servers = serversWrapper.Data
	}

	log.Printf("DEBUG: ReadServers found %d servers", len(servers))
	for i, server := range servers {
		disabledStr := "nil"
		if server.Disabled != nil {
			disabledStr = fmt.Sprintf("%t", *server.Disabled)
		}
		log.Printf("DEBUG: Server %d: %s (%s:%d) - Check:'%s' Maxconn:%d Weight:%d Disabled:%s",
			i, server.Name, server.Address, server.Port, server.Check, server.Maxconn, server.Weight, disabledStr)
	}

	return servers, nil
}

// CreateServerInTransaction creates a new server using an existing transaction ID.
func (c *HAProxyClient) CreateServerInTransaction(ctx context.Context, transactionID, parentType, parentName string, payload *ServerPayload) error {
	log.Printf("CreateServerInTransaction called with transaction ID: %s, parent: %s/%s, payload: %+v", transactionID, parentType, parentName, payload)

	var url string
	if c.apiVersion == "v3" {
		// v3: Use nested endpoint under backends
		// Properly pluralize the parent type
		parentTypePlural := parentType + "s"
		url = fmt.Sprintf("/services/haproxy/configuration/%s/%s/servers?transaction_id=%s",
			parentTypePlural, parentName, transactionID)
	} else {
		// v2: Use query parameter approach
		url = fmt.Sprintf("/services/haproxy/configuration/servers?parent_type=%s&parent_name=%s&transaction_id=%s",
			parentType, parentName, transactionID)
	}

	log.Printf("DEBUG: Using server endpoint: %s for API version %s", url, c.apiVersion)

	req, err := c.newRequest(ctx, "POST", url, payload)
	if err != nil {
		return err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("server creation failed with status %d", resp.StatusCode)
		}
		return fmt.Errorf("server creation failed with status %d: %s", resp.StatusCode, string(body))
	}

	log.Printf("Server created successfully in transaction: %s", transactionID)
	return nil
}

// UpdateServerInTransaction updates an existing server using an existing transaction ID.
func (c *HAProxyClient) UpdateServerInTransaction(ctx context.Context, transactionID, parentType, parentName string, payload *ServerPayload) error {
	log.Printf("UpdateServerInTransaction called with transaction ID: %s, parent: %s/%s, payload: %+v", transactionID, parentType, parentName, payload)

	var url string
	if c.apiVersion == "v3" {
		// v3: Use nested endpoint under backends
		// Properly pluralize the parent type
		parentTypePlural := parentType + "s"
		url = fmt.Sprintf("/services/haproxy/configuration/%s/%s/servers/%s?transaction_id=%s",
			parentTypePlural, parentName, payload.Name, transactionID)
	} else {
		// v2: Use query parameter approach
		url = fmt.Sprintf("/services/haproxy/configuration/servers/%s?parent_type=%s&parent_name=%s&transaction_id=%s",
			payload.Name, parentType, parentName, transactionID)
	}

	log.Printf("DEBUG: Using server update endpoint: %s for API version %s", url, c.apiVersion)

	req, err := c.newRequest(ctx, "PUT", url, payload)
	if err != nil {
		return err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("server update failed with status %d", resp.StatusCode)
		}
		return fmt.Errorf("server update failed with status %d: %s", resp.StatusCode, string(body))
	}

	log.Printf("Server updated successfully in transaction: %s", transactionID)
	return nil
}

// DeleteServerInTransaction deletes an existing server using an existing transaction ID.
func (c *HAProxyClient) DeleteServerInTransaction(ctx context.Context, transactionID, parentType, parentName string, serverName string) error {
	log.Printf("DeleteServerInTransaction called with transaction ID: %s, parent: %s/%s, server: %s", transactionID, parentType, parentName, serverName)

	var url string
	if c.apiVersion == "v3" {
		// v3: Use nested endpoint under backends
		// Properly pluralize the parent type
		parentTypePlural := parentType + "s"
		url = fmt.Sprintf("/services/haproxy/configuration/%s/%s/servers/%s?transaction_id=%s",
			parentTypePlural, parentName, serverName, transactionID)
	} else {
		// v2: Use query parameter approach
		url = fmt.Sprintf("/services/haproxy/configuration/servers/%s?parent_type=%s&parent_name=%s&transaction_id=%s",
			serverName, parentType, parentName, transactionID)
	}

	log.Printf("DEBUG: Using server delete endpoint: %s for API version %s", url, c.apiVersion)

	req, err := c.newRequest(ctx, "DELETE", url, nil)
	if err != nil {
		return err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted && resp.StatusCode != http.StatusNoContent {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("server deletion failed with status %d", resp.StatusCode)
		}
		return fmt.Errorf("server deletion failed with status %d: %s", resp.StatusCode, string(body))
	}

	log.Printf("Server deleted successfully in transaction: %s", transactionID)
	return nil
}

// ReadServer reads a server.
func (c *HAProxyClient) ReadServer(ctx context.Context, name, parentType, parentName string) (*ServerPayload, error) {
	req, err := c.newRequest(ctx, "GET", fmt.Sprintf("/services/haproxy/configuration/servers/%s?parent_type=%s&parent_name=%s", name, parentType, parentName), nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var server struct {
		Data ServerPayload `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&server); err != nil {
		return nil, err
	}

	return &server.Data, nil
}

// UpdateServer updates a server.
func (c *HAProxyClient) UpdateServer(ctx context.Context, name, parentType, parentName string, payload *ServerPayload) error {
	_, err := c.Transaction(func(transactionID string) (*http.Response, error) {
		req, err := c.newRequest(ctx, "PUT", fmt.Sprintf("/services/haproxy/configuration/servers/%s?parent_type=%s&parent_name=%s&transaction_id=%s", name, parentType, parentName, transactionID), payload)
		if err != nil {
			return nil, err
		}
		return c.httpClient.Do(req)
	})
	return err
}

// DeleteServer deletes a server.
func (c *HAProxyClient) DeleteServer(ctx context.Context, name, parentType, parentName string) error {
	_, err := c.Transaction(func(transactionID string) (*http.Response, error) {
		req, err := c.newRequest(ctx, "DELETE", fmt.Sprintf("/services/haproxy/configuration/servers/%s?parent_type=%s&parent_name=%s&transaction_id=%s", name, parentType, parentName, transactionID), nil)
		if err != nil {
			return nil, err
		}
		return c.httpClient.Do(req)
	})
	return err
}

// CreateBind creates a new bind.
func (c *HAProxyClient) CreateBind(ctx context.Context, parentType, parentName string, payload *BindPayload) error {
	_, err := c.Transaction(func(transactionID string) (*http.Response, error) {
		req, err := c.newRequest(ctx, "POST", fmt.Sprintf("/services/haproxy/configuration/binds?parent_type=%s&parent_name=%s&transaction_id=%s", parentType, parentName, transactionID), payload)
		if err != nil {
			return nil, err
		}
		return c.httpClient.Do(req)
	})
	return err
}

// CreateBindInTransaction creates a new bind within an existing transaction.
func (c *HAProxyClient) CreateBindInTransaction(ctx context.Context, transactionID, parentType, parentName string, payload *BindPayload) (*http.Response, error) {
	var url string

	// Use version-aware endpoint structure
	if c.apiVersion == "v3" {
		// v3: nested under parent resource
		url = fmt.Sprintf("/services/haproxy/configuration/%ss/%s/binds?transaction_id=%s", parentType, parentName, transactionID)
	} else {
		// v2: query parameters
		url = fmt.Sprintf("/services/haproxy/configuration/binds?parent_type=%s&parent_name=%s&transaction_id=%s", parentType, parentName, transactionID)
	}

	req, err := c.newRequest(ctx, "POST", url, payload)
	if err != nil {
		return nil, err
	}
	return c.httpClient.Do(req)
}

// UpdateBindInTransaction updates a bind within an existing transaction.
func (c *HAProxyClient) UpdateBindInTransaction(ctx context.Context, transactionID, name, parentType, parentName string, payload *BindPayload) (*http.Response, error) {
	var url string

	// Use version-aware endpoint structure
	if c.apiVersion == "v3" {
		// v3: nested under parent resource
		url = fmt.Sprintf("/services/haproxy/configuration/%ss/%s/binds/%s?transaction_id=%s", parentType, parentName, name, transactionID)
	} else {
		// v2: query parameters
		url = fmt.Sprintf("/services/haproxy/configuration/binds/%s?parent_type=%s&parent_name=%s&transaction_id=%s", name, parentType, parentName, transactionID)
	}

	req, err := c.newRequest(ctx, "PUT", url, payload)
	if err != nil {
		return nil, err
	}
	return c.httpClient.Do(req)
}

// ReadBind reads a bind.
func (c *HAProxyClient) ReadBind(ctx context.Context, name, parentType, parentName string) (*BindPayload, error) {
	req, err := c.newRequest(ctx, "GET", fmt.Sprintf("/services/haproxy/configuration/binds/%s?parent_type=%s&parent_name=%s", name, parentType, parentName), nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var bind struct {
		Data BindPayload `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&bind); err != nil {
		return nil, err
	}

	return &bind.Data, nil
}

// UpdateBind updates a bind.
func (c *HAProxyClient) UpdateBind(ctx context.Context, name, parentType, parentName string, payload *BindPayload) error {
	_, err := c.Transaction(func(transactionID string) (*http.Response, error) {
		req, err := c.newRequest(ctx, "PUT", fmt.Sprintf("/services/haproxy/configuration/binds/%s?parent_type=%s&parent_name=%s&transaction_id=%s", name, parentType, parentName, transactionID), payload)
		if err != nil {
			return nil, err
		}
		return c.httpClient.Do(req)
	})
	return err
}

// DeleteBind deletes a bind.
func (c *HAProxyClient) DeleteBind(ctx context.Context, name, parentType, parentName string) error {
	_, err := c.Transaction(func(transactionID string) (*http.Response, error) {
		req, err := c.newRequest(ctx, "DELETE", fmt.Sprintf("/services/haproxy/configuration/binds/%s?parent_type=%s&parent_name=%s&transaction_id=%s", name, parentType, parentName, transactionID), nil)
		if err != nil {
			return nil, err
		}
		return c.httpClient.Do(req)
	})
	return err
}

// DeleteBindInTransaction deletes a bind within an existing transaction.
func (c *HAProxyClient) DeleteBindInTransaction(ctx context.Context, transactionID, name, parentType, parentName string) (*http.Response, error) {
	var url string

	// Use version-aware endpoint structure
	if c.apiVersion == "v3" {
		// v3: nested under parent resource
		url = fmt.Sprintf("/services/haproxy/configuration/%ss/%s/binds/%s?transaction_id=%s", parentType, parentName, name, transactionID)
	} else {
		// v2: query parameters
		url = fmt.Sprintf("/services/haproxy/configuration/binds/%s?parent_type=%s&parent_name=%s&transaction_id=%s", name, parentType, parentName, transactionID)
	}

	req, err := c.newRequest(ctx, "DELETE", url, nil)
	if err != nil {
		return nil, err
	}
	return c.httpClient.Do(req)
}

// DeleteFrontend deletes a frontend.
func (c *HAProxyClient) DeleteFrontend(ctx context.Context, name string) error {
	_, err := c.Transaction(func(transactionID string) (*http.Response, error) {
		req, err := c.newRequest(ctx, "DELETE", fmt.Sprintf("/services/haproxy/configuration/frontends/%s?transaction_id=%s", name, transactionID), nil)
		if err != nil {
			return nil, err
		}
		return c.httpClient.Do(req)
	})
	return err
}

// ReadBinds reads all binds for a given parent.
func (c *HAProxyClient) ReadBinds(ctx context.Context, parentType, parentName string) ([]BindPayload, error) {
	var url string

	// Construct URL based on API version
	if c.apiVersion == "v3" {
		// v3: nested under parent resource (note: frontend -> frontends, backend -> backends)
		if parentType == "frontend" {
			parentType = "frontends"
		} else if parentType == "backend" {
			parentType = "backends"
		}
		// Use same format as CreateBindsInTransaction (without /v3 prefix)
		url = fmt.Sprintf("/services/haproxy/configuration/%s/%s/binds", parentType, parentName)
	} else {
		// v2: query parameters (no version prefix needed)
		url = fmt.Sprintf("/services/haproxy/configuration/binds?parent_type=%s&parent_name=%s", parentType, parentName)
	}

	// Debug: Log the URL being constructed
	log.Printf("DEBUG: ReadBinds constructing URL: %s (API version: %s)", url, c.apiVersion)

	// Debug: Log the full request details
	log.Printf("DEBUG: ReadBinds base URL: %s", c.baseURL)
	log.Printf("DEBUG: ReadBinds full URL: %s%s", c.baseURL, url)

	req, err := c.newRequest(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	// Debug: Log the actual request being sent
	log.Printf("DEBUG: ReadBinds request method: %s", req.Method)
	log.Printf("DEBUG: ReadBinds request URL: %s", req.URL.String())
	log.Printf("DEBUG: ReadBinds request headers: %+v", req.Header)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Debug: Log the response status
	log.Printf("DEBUG: ReadBinds response status: %d", resp.StatusCode)

	if resp.StatusCode == http.StatusNotFound {
		if c.apiVersion == "v3" {
			// v3: 404 means the endpoint doesn't exist (configuration error)
			return nil, fmt.Errorf("binds endpoint not found for v3 API - check URL construction. URL attempted: %s", url)
		} else {
			// v2: 404 might mean "no binds found" (legitimate)
			log.Printf("DEBUG: ReadBinds: 404 - No binds found (v2)")
			return []BindPayload{}, nil
		}
	}

	if resp.StatusCode != http.StatusOK {
		// Debug: Log error response body
		bodyBytes, _ := io.ReadAll(resp.Body)
		log.Printf("DEBUG: ReadBinds error response body: %s", sanitizeResponseBody(string(bodyBytes)))
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var binds []BindPayload

	if c.apiVersion == "v3" {
		// v3: binds are returned directly as an array
		if err := json.NewDecoder(resp.Body).Decode(&binds); err != nil {
			return nil, err
		}
	} else {
		// v2: binds are wrapped in a "data" field
		var bindsWrapper struct {
			Data []BindPayload `json:"data"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&bindsWrapper); err != nil {
			return nil, err
		}
		binds = bindsWrapper.Data
	}

	return binds, nil
}

// CreateAcl creates a new acl.
func (c *HAProxyClient) CreateAcl(ctx context.Context, parentType, parentName string, payload *ACLPayload) error {
	_, err := c.Transaction(func(transactionID string) (*http.Response, error) {
		req, err := c.newRequest(ctx, "POST", fmt.Sprintf("/services/haproxy/configuration/acls?parent_type=%s&parent_name=%s&transaction_id=%s", parentType, parentName, transactionID), payload)
		if err != nil {
			return nil, err
		}
		return c.httpClient.Do(req)
	})
	return err
}

// ReadAcls reads all acls for a given parent.
func (c *HAProxyClient) ReadAcls(ctx context.Context, parentType, parentName string) ([]AclPayload, error) {
	req, err := c.newRequest(ctx, "GET", fmt.Sprintf("/services/haproxy/configuration/acls?parent_type=%s&parent_name=%s", parentType, parentName), nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return []AclPayload{}, nil // No acls found is not an error
	}

	if resp.StatusCode == http.StatusUnprocessableEntity {
		// 422 usually means invalid parameters, treat as no ACLs found
		return []AclPayload{}, nil
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var aclsWrapper struct {
		Data []AclPayload `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&aclsWrapper); err != nil {
		return nil, err
	}

	return aclsWrapper.Data, nil
}

// UpdateAcl updates a acl.
func (c *HAProxyClient) UpdateAcl(ctx context.Context, index int64, parentType, parentName string, payload *ACLPayload) error {
	_, err := c.Transaction(func(transactionID string) (*http.Response, error) {
		req, err := c.newRequest(ctx, "PUT", fmt.Sprintf("/services/haproxy/configuration/acls/%d?parent_type=%s&parent_name=%s&transaction_id=%s", index, parentType, parentName, transactionID), payload)
		if err != nil {
			return nil, err
		}
		return c.httpClient.Do(req)
	})
	return err
}

// DeleteAcl deletes a acl.
func (c *HAProxyClient) DeleteAcl(ctx context.Context, index int64, parentType, parentName string) error {
	_, err := c.Transaction(func(transactionID string) (*http.Response, error) {
		req, err := c.newRequest(ctx, "DELETE", fmt.Sprintf("/services/haproxy/configuration/acls/%d?parent_type=%s&parent_name=%s&transaction_id=%s", index, parentType, parentName, transactionID), nil)
		if err != nil {
			return nil, err
		}
		return c.httpClient.Do(req)
	})
	return err
}

// CreateHttpRequestRule creates a new httprequestrule.
// DEPRECATED: Use CreateHttpRequestRuleInTransaction for new code
func (c *HAProxyClient) CreateHttpRequestRule(ctx context.Context, parentType, parentName string, payload *HttpRequestRulePayload) error {
	transactionID, err := c.BeginTransaction()
	if err != nil {
		return err
	}
	defer c.RollbackTransaction(transactionID)

	if err := c.CreateHttpRequestRuleInTransaction(ctx, transactionID, parentType, parentName, payload); err != nil {
		return err
	}

	return c.CommitTransaction(transactionID)
}

// ReadHttpRequestRules reads all httprequestrules for a given parent.
func (c *HAProxyClient) ReadHttpRequestRules(ctx context.Context, parentType, parentName string) ([]HttpRequestRulePayload, error) {
	var url string
	if c.apiVersion == "v3" {
		// v3: Use nested endpoint under frontends/backends
		parentTypePlural := parentType + "s"
		url = fmt.Sprintf("/services/haproxy/configuration/%s/%s/http_request_rules", parentTypePlural, parentName)
	} else {
		// v2: Use query parameter approach
		url = fmt.Sprintf("/services/haproxy/configuration/http_request_rules?parent_type=%s&parent_name=%s", parentType, parentName)
	}

	req, err := c.newRequest(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return []HttpRequestRulePayload{}, nil // No httprequestrules found is not an error
	}

	if resp.StatusCode == http.StatusUnprocessableEntity {
		// 422 usually means invalid parameters, treat as no rules found
		return []HttpRequestRulePayload{}, nil
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var httpRequestRules []HttpRequestRulePayload
	if c.apiVersion == "v3" {
		// v3: Response is a direct array, no wrapper
		if err := json.NewDecoder(resp.Body).Decode(&httpRequestRules); err != nil {
			return nil, err
		}
	} else {
		// v2: Response is wrapped in {"data": [...]}
		var httpRequestRulesWrapper struct {
			Data []HttpRequestRulePayload `json:"data"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&httpRequestRulesWrapper); err != nil {
			return nil, err
		}
		httpRequestRules = httpRequestRulesWrapper.Data
	}

	return httpRequestRules, nil
}

// UpdateHttpRequestRule updates a httprequestrule.
// DEPRECATED: Use individual resource management for new code
func (c *HAProxyClient) UpdateHttpRequestRule(ctx context.Context, index int64, parentType, parentName string, payload *HttpRequestRulePayload) error {
	_, err := c.Transaction(func(transactionID string) (*http.Response, error) {
		req, err := c.newRequest(ctx, "PUT", fmt.Sprintf("/services/haproxy/configuration/http_request_rules/%d?parent_type=%s&parent_name=%s&transaction_id=%s", index, parentType, parentName, transactionID), payload)
		if err != nil {
			return nil, err
		}
		return c.httpClient.Do(req)
	})
	return err
}

// DeleteHttpRequestRule deletes a httprequestrule.
// DEPRECATED: Use DeleteHttpRequestRuleInTransaction for new code
func (c *HAProxyClient) DeleteHttpRequestRule(ctx context.Context, index int64, parentType, parentName string) error {
	transactionID, err := c.BeginTransaction()
	if err != nil {
		return err
	}
	defer c.RollbackTransaction(transactionID)

	if err := c.DeleteHttpRequestRuleInTransaction(ctx, transactionID, index, parentType, parentName); err != nil {
		return err
	}

	return c.CommitTransaction(transactionID)
}

// CreateHttpResponseRule creates a new httpresponserule.
func (c *HAProxyClient) CreateHttpResponseRule(ctx context.Context, parentType, parentName string, payload *HttpResponseRulePayload) error {
	_, err := c.Transaction(func(transactionID string) (*http.Response, error) {
		req, err := c.newRequest(ctx, "POST", fmt.Sprintf("/services/haproxy/configuration/http_response_rules?parent_type=%s&parent_name=%s&transaction_id=%s", parentType, parentName, transactionID), payload)
		if err != nil {
			return nil, err
		}
		return c.httpClient.Do(req)
	})
	return err
}

// ReadHttpResponseRules reads all httpresponserules for a given parent.
func (c *HAProxyClient) ReadHttpResponseRules(ctx context.Context, parentType, parentName string) ([]HttpResponseRulePayload, error) {
	req, err := c.newRequest(ctx, "GET", fmt.Sprintf("/services/haproxy/configuration/http_response_rules?parent_type=%s&parent_name=%s", parentType, parentName), nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return []HttpResponseRulePayload{}, nil // No httpresponserules found is not an error
	}

	if resp.StatusCode == http.StatusUnprocessableEntity {
		// 422 usually means invalid parameters, treat as no rules found
		return []HttpResponseRulePayload{}, nil
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var httpResponseRulesWrapper struct {
		Data []HttpResponseRulePayload `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&httpResponseRulesWrapper); err != nil {
		return nil, err
	}

	return httpResponseRulesWrapper.Data, nil
}

// UpdateHttpResponseRule updates a httpresponserule.
func (c *HAProxyClient) UpdateHttpResponseRule(ctx context.Context, index int64, parentType, parentName string, payload *HttpResponseRulePayload) error {
	_, err := c.Transaction(func(transactionID string) (*http.Response, error) {
		req, err := c.newRequest(ctx, "PUT", fmt.Sprintf("/services/haproxy/configuration/http_response_rules/%d?parent_type=%s&parent_name=%s&transaction_id=%s", index, parentType, parentName, transactionID), payload)
		if err != nil {
			return nil, err
		}
		return c.httpClient.Do(req)
	})
	return err
}

// DeleteHttpResponseRule deletes a httpresponserule.
func (c *HAProxyClient) DeleteHttpResponseRule(ctx context.Context, index int64, parentType, parentName string) error {
	_, err := c.Transaction(func(transactionID string) (*http.Response, error) {
		req, err := c.newRequest(ctx, "DELETE", fmt.Sprintf("/services/haproxy/configuration/http_response_rules/%d?parent_type=%s&parent_name=%s&transaction_id=%s", index, parentType, parentName, transactionID), nil)
		if err != nil {
			return nil, err
		}
		return c.httpClient.Do(req)
	})
	return err
}

// CreateResolver creates a new resolver.
func (c *HAProxyClient) CreateResolver(ctx context.Context, payload *ResolverPayload) error {
	_, err := c.Transaction(func(transactionID string) (*http.Response, error) {
		req, err := c.newRequest(ctx, "POST", fmt.Sprintf("/services/haproxy/configuration/resolvers?transaction_id=%s", transactionID), payload)
		if err != nil {
			return nil, err
		}
		return c.httpClient.Do(req)
	})
	return err
}

// ReadResolver reads a resolver.
func (c *HAProxyClient) ReadResolver(ctx context.Context, name string) (*ResolverPayload, error) {
	req, err := c.newRequest(ctx, "GET", fmt.Sprintf("/services/haproxy/configuration/resolvers/%s", name), nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var resolverWrapper struct {
		Data ResolverPayload `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&resolverWrapper); err != nil {
		return nil, err
	}

	return &resolverWrapper.Data, nil
}

// UpdateResolver updates a resolver.
func (c *HAProxyClient) UpdateResolver(ctx context.Context, name string, payload *ResolverPayload) error {
	_, err := c.Transaction(func(transactionID string) (*http.Response, error) {
		req, err := c.newRequest(ctx, "PUT", fmt.Sprintf("/services/haproxy/configuration/resolvers/%s?transaction_id=%s", name, transactionID), payload)
		if err != nil {
			return nil, err
		}
		return c.httpClient.Do(req)
	})
	return err
}

// DeleteResolver deletes a resolver.
func (c *HAProxyClient) DeleteResolver(ctx context.Context, name string) error {
	_, err := c.Transaction(func(transactionID string) (*http.Response, error) {
		req, err := c.newRequest(ctx, "DELETE", fmt.Sprintf("/services/haproxy/configuration/resolvers/%s?transaction_id=%s", name, transactionID), nil)
		if err != nil {
			return nil, err
		}
		return c.httpClient.Do(req)
	})
	return err
}

// CreateNameserver creates a new nameserver.
func (c *HAProxyClient) CreateNameserver(ctx context.Context, resolver string, payload *NameserverPayload) error {
	_, err := c.Transaction(func(transactionID string) (*http.Response, error) {
		req, err := c.newRequest(ctx, "POST", fmt.Sprintf("/services/haproxy/configuration/nameservers?resolver=%s&transaction_id=%s", resolver, transactionID), payload)
		if err != nil {
			return nil, err
		}
		return c.httpClient.Do(req)
	})
	return err
}

// ReadNameserver reads a nameserver.
func (c *HAProxyClient) ReadNameserver(ctx context.Context, name, resolver string) (*NameserverPayload, error) {
	req, err := c.newRequest(ctx, "GET", fmt.Sprintf("/services/haproxy/configuration/nameservers/%s?resolver=%s", name, resolver), nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var nameserverWrapper struct {
		Data NameserverPayload `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&nameserverWrapper); err != nil {
		return nil, err
	}

	return &nameserverWrapper.Data, nil
}

// ReadNameservers reads all nameservers for a resolver.
func (c *HAProxyClient) ReadNameservers(ctx context.Context, resolver string) ([]NameserverPayload, error) {
	req, err := c.newRequest(ctx, "GET", fmt.Sprintf("/services/haproxy/configuration/nameservers?resolver=%s", resolver), nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return []NameserverPayload{}, nil // No nameservers found is not an error
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var nameserversWrapper struct {
		Data []NameserverPayload `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&nameserversWrapper); err != nil {
		return nil, err
	}

	return nameserversWrapper.Data, nil
}

// UpdateNameserver updates a nameserver.
func (c *HAProxyClient) UpdateNameserver(ctx context.Context, name, resolver string, payload *NameserverPayload) error {
	_, err := c.Transaction(func(transactionID string) (*http.Response, error) {
		req, err := c.newRequest(ctx, "PUT", fmt.Sprintf("/services/haproxy/configuration/nameservers/%s?resolver=%s&transaction_id=%s", name, resolver, transactionID), payload)
		if err != nil {
			return nil, err
		}
		return c.httpClient.Do(req)
	})
	return err
}

// DeleteNameserver deletes a nameserver.
func (c *HAProxyClient) DeleteNameserver(ctx context.Context, name, resolver string) error {
	_, err := c.Transaction(func(transactionID string) (*http.Response, error) {
		req, err := c.newRequest(ctx, "DELETE", fmt.Sprintf("/services/haproxy/configuration/nameservers/%s?resolver=%s&transaction_id=%s", name, resolver, transactionID), nil)
		if err != nil {
			return nil, err
		}
		return c.httpClient.Do(req)
	})
	return err
}

// CreatePeers creates a new peers.
func (c *HAProxyClient) CreatePeers(ctx context.Context, payload *PeersPayload) error {
	_, err := c.Transaction(func(transactionID string) (*http.Response, error) {
		req, err := c.newRequest(ctx, "POST", fmt.Sprintf("/services/haproxy/configuration/peers?transaction_id=%s", transactionID), payload)
		if err != nil {
			return nil, err
		}
		return c.httpClient.Do(req)
	})
	return err
}

// ReadPeers reads a peers.
func (c *HAProxyClient) ReadPeers(ctx context.Context, name string) (*PeersPayload, error) {
	req, err := c.newRequest(ctx, "GET", fmt.Sprintf("/services/haproxy/configuration/peers/%s", name), nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var peersWrapper struct {
		Data PeersPayload `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&peersWrapper); err != nil {
		return nil, err
	}

	return &peersWrapper.Data, nil
}

// UpdatePeers updates a peers.
func (c *HAProxyClient) UpdatePeers(ctx context.Context, name string, payload *PeersPayload) error {
	_, err := c.Transaction(func(transactionID string) (*http.Response, error) {
		req, err := c.newRequest(ctx, "PUT", fmt.Sprintf("/services/haproxy/configuration/peers/%s?transaction_id=%s", name, transactionID), payload)
		if err != nil {
			return nil, err
		}
		return c.httpClient.Do(req)
	})
	return err
}

// DeletePeers deletes a peers.
func (c *HAProxyClient) DeletePeers(ctx context.Context, name string) error {
	_, err := c.Transaction(func(transactionID string) (*http.Response, error) {
		req, err := c.newRequest(ctx, "DELETE", fmt.Sprintf("/services/haproxy/configuration/peers/%s?transaction_id=%s", name, transactionID), nil)
		if err != nil {
			return nil, err
		}
		return c.httpClient.Do(req)
	})
	return err
}

// CreatePeerEntry creates a new peer_entry.
func (c *HAProxyClient) CreatePeerEntry(ctx context.Context, peers string, payload *PeerEntryPayload) error {
	_, err := c.Transaction(func(transactionID string) (*http.Response, error) {
		req, err := c.newRequest(ctx, "POST", fmt.Sprintf("/services/haproxy/configuration/peer_entries?peers=%s&transaction_id=%s", peers, transactionID), payload)
		if err != nil {
			return nil, err
		}
		return c.httpClient.Do(req)
	})
	return err
}

// ReadPeerEntry reads a peer_entry.
func (c *HAProxyClient) ReadPeerEntry(ctx context.Context, name, peers string) (*PeerEntryPayload, error) {
	req, err := c.newRequest(ctx, "GET", fmt.Sprintf("/services/haproxy/configuration/peer_entries/%s?peers=%s", name, peers), nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var peerEntryWrapper struct {
		Data PeerEntryPayload `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&peerEntryWrapper); err != nil {
		return nil, err
	}

	return &peerEntryWrapper.Data, nil
}

// ReadPeerEntries reads all peer entries for a peers group.
func (c *HAProxyClient) ReadPeerEntries(ctx context.Context, peers string) ([]PeerEntryPayload, error) {
	req, err := c.newRequest(ctx, "GET", fmt.Sprintf("/services/haproxy/configuration/peer_entries?peers=%s", peers), nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return []PeerEntryPayload{}, nil // No peer entries found is not an error
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var peerEntriesWrapper struct {
		Data []PeerEntryPayload `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&peerEntriesWrapper); err != nil {
		return nil, err
	}

	return peerEntriesWrapper.Data, nil
}

// UpdatePeerEntry updates a peer_entry.
func (c *HAProxyClient) UpdatePeerEntry(ctx context.Context, name, peers string, payload *PeerEntryPayload) error {
	_, err := c.Transaction(func(transactionID string) (*http.Response, error) {
		req, err := c.newRequest(ctx, "PUT", fmt.Sprintf("/services/haproxy/configuration/peer_entries/%s?peers=%s&transaction_id=%s", name, peers, transactionID), payload)
		if err != nil {
			return nil, err
		}
		return c.httpClient.Do(req)
	})
	return err
}

// DeletePeerEntry deletes a peer_entry.
func (c *HAProxyClient) DeletePeerEntry(ctx context.Context, name, peers string) error {
	_, err := c.Transaction(func(transactionID string) (*http.Response, error) {
		req, err := c.newRequest(ctx, "DELETE", fmt.Sprintf("/services/haproxy/configuration/peer_entries/%s?peers=%s&transaction_id=%s", name, peers, transactionID), nil)
		if err != nil {
			return nil, err
		}
		return c.httpClient.Do(req)
	})
	return err
}

// CreateStickRule creates a new stick_rule.
func (c *HAProxyClient) CreateStickRule(ctx context.Context, backend string, payload *StickRulePayload) error {
	_, err := c.Transaction(func(transactionID string) (*http.Response, error) {
		req, err := c.newRequest(ctx, "POST", fmt.Sprintf("/services/haproxy/configuration/stick_rules?backend=%s&transaction_id=%s", backend, transactionID), payload)
		if err != nil {
			return nil, err
		}
		return c.httpClient.Do(req)
	})
	return err
}

// ReadStickRule reads a stick_rule.
func (c *HAProxyClient) ReadStickRule(ctx context.Context, index int64, backend string) (*StickRulePayload, error) {
	req, err := c.newRequest(ctx, "GET", fmt.Sprintf("/services/haproxy/configuration/stick_rules/%d?backend=%s", index, backend), nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var stickRuleWrapper struct {
		Data StickRulePayload `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&stickRuleWrapper); err != nil {
		return nil, err
	}

	return &stickRuleWrapper.Data, nil
}

// ReadStickRules reads all stick rules for a backend.
func (c *HAProxyClient) ReadStickRules(ctx context.Context, backend string) ([]StickRulePayload, error) {
	req, err := c.newRequest(ctx, "GET", fmt.Sprintf("/services/haproxy/configuration/stick_rules?backend=%s", backend), nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return []StickRulePayload{}, nil // No stick rules found is not an error
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var stickRulesWrapper struct {
		Data []StickRulePayload `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&stickRulesWrapper); err != nil {
		return nil, err
	}

	return stickRulesWrapper.Data, nil
}

// UpdateStickRule updates a stick_rule.
func (c *HAProxyClient) UpdateStickRule(ctx context.Context, index int64, backend string, payload *StickRulePayload) error {
	_, err := c.Transaction(func(transactionID string) (*http.Response, error) {
		req, err := c.newRequest(ctx, "PUT", fmt.Sprintf("/services/haproxy/configuration/stick_rules/%d?backend=%s&transaction_id=%s", index, backend, transactionID), payload)
		if err != nil {
			return nil, err
		}
		return c.httpClient.Do(req)
	})
	return err
}

// DeleteStickRule deletes a stick_rule.
func (c *HAProxyClient) DeleteStickRule(ctx context.Context, index int64, backend string) error {
	_, err := c.Transaction(func(transactionID string) (*http.Response, error) {
		req, err := c.newRequest(ctx, "DELETE", fmt.Sprintf("/services/haproxy/configuration/stick_rules/%d?backend=%s&transaction_id=%s", index, backend, transactionID), nil)
		if err != nil {
			return nil, err
		}
		return c.httpClient.Do(req)
	})
	return err
}

func (c *HAProxyClient) newRequest(ctx context.Context, method, path string, body interface{}) (*http.Request, error) {
	var buf bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			return nil, err
		}
	}

	url := fmt.Sprintf("%s/%s%s", c.baseURL, c.apiVersion, path)
	req, err := http.NewRequestWithContext(ctx, method, url, &buf)
	if err != nil {
		return nil, err
	}

	req.SetBasicAuth(c.username, c.password)
	req.Header.Set("Content-Type", "application/json")

	return req, nil
}

// ReadBackends reads all backends.
func (c *HAProxyClient) ReadBackends(ctx context.Context) ([]BackendPayload, error) {
	req, err := c.newRequest(ctx, "GET", "/services/haproxy/configuration/backends", nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return []BackendPayload{}, nil // No backends found is not an error
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var backendsWrapper struct {
		Data []BackendPayload `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&backendsWrapper); err != nil {
		return nil, err
	}

	return backendsWrapper.Data, nil
}

// ReadFrontends reads all frontends.
func (c *HAProxyClient) ReadFrontends(ctx context.Context) ([]FrontendPayload, error) {
	req, err := c.newRequest(ctx, "GET", "/services/haproxy/configuration/frontends", nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return []FrontendPayload{}, nil // No frontends found is not an error
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var frontendsWrapper struct {
		Data []FrontendPayload `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&frontendsWrapper); err != nil {
		return nil, err
	}

	return frontendsWrapper.Data, nil
}

// CreateHttpcheck creates a new httpcheck.
func (c *HAProxyClient) CreateHttpcheck(ctx context.Context, parentType, parentName string, payload *HttpcheckPayload) error {
	_, err := c.Transaction(func(transactionID string) (*http.Response, error) {
		req, err := c.newRequest(ctx, "POST", fmt.Sprintf("/services/haproxy/configuration/http_checks?parent_type=%s&parent_name=%s&transaction_id=%s", parentType, parentName, transactionID), payload)
		if err != nil {
			return nil, err
		}
		return c.httpClient.Do(req)
	})
	return err
}

// ReadHttpchecks reads all httpchecks for a given parent.
func (c *HAProxyClient) ReadHttpchecks(ctx context.Context, parentType, parentName string) ([]HttpcheckPayload, error) {
	req, err := c.newRequest(ctx, "GET", fmt.Sprintf("/services/haproxy/configuration/http_checks?parent_type=%s&parent_name=%s", parentType, parentName), nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return []HttpcheckPayload{}, nil // No httpchecks found is not an error
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var httpchecksWrapper struct {
		Data []HttpcheckPayload `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&httpchecksWrapper); err != nil {
		return nil, err
	}

	return httpchecksWrapper.Data, nil
}

// UpdateHttpcheck updates a httpcheck.
func (c *HAProxyClient) UpdateHttpcheck(ctx context.Context, index int64, parentType, parentName string, payload *HttpcheckPayload) error {
	_, err := c.Transaction(func(transactionID string) (*http.Response, error) {
		req, err := c.newRequest(ctx, "PUT", fmt.Sprintf("/services/haproxy/configuration/http_checks/%d?parent_type=%s&parent_name=%s&transaction_id=%s", index, parentType, parentName, transactionID), payload)
		if err != nil {
			return nil, err
		}
		return c.httpClient.Do(req)
	})
	return err
}

// DeleteHttpcheck deletes a httpcheck.
func (c *HAProxyClient) DeleteHttpcheck(ctx context.Context, index int64, parentType, parentName string) error {
	_, err := c.Transaction(func(transactionID string) (*http.Response, error) {
		req, err := c.newRequest(ctx, "DELETE", fmt.Sprintf("/services/haproxy/configuration/http_checks/%d?parent_type=%s&parent_name=%s&transaction_id=%s", index, parentType, parentName, transactionID), nil)
		if err != nil {
			return nil, err
		}
		return c.httpClient.Do(req)
	})
	return err
}

// CreateStickTable creates a new stick_table.
func (c *HAProxyClient) CreateStickTable(ctx context.Context, payload *StickTablePayload) error {
	_, err := c.Transaction(func(transactionID string) (*http.Response, error) {
		req, err := c.newRequest(ctx, "POST", fmt.Sprintf("/services/haproxy/configuration/stick_tables?transaction_id=%s", transactionID), payload)
		if err != nil {
			return nil, err
		}
		return c.httpClient.Do(req)
	})
	return err
}

// ReadStickTable reads a stick_table.
func (c *HAProxyClient) ReadStickTable(ctx context.Context, name string) (*StickTablePayload, error) {
	req, err := c.newRequest(ctx, "GET", fmt.Sprintf("/services/haproxy/configuration/stick_tables/%s", name), nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var stickTableWrapper struct {
		Data StickTablePayload `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&stickTableWrapper); err != nil {
		return nil, err
	}

	return &stickTableWrapper.Data, nil
}

// UpdateStickTable updates a stick_table.
func (c *HAProxyClient) UpdateStickTable(ctx context.Context, name string, payload *StickTablePayload) error {
	_, err := c.Transaction(func(transactionID string) (*http.Response, error) {
		req, err := c.newRequest(ctx, "PUT", fmt.Sprintf("/services/haproxy/configuration/stick_tables/%s?transaction_id=%s", name, transactionID), payload)
		if err != nil {
			return nil, err
		}
		return c.httpClient.Do(req)
	})
	return err
}

// DeleteStickTable deletes a stick_table.
func (c *HAProxyClient) DeleteStickTable(ctx context.Context, name string) error {
	_, err := c.Transaction(func(transactionID string) (*http.Response, error) {
		req, err := c.newRequest(ctx, "DELETE", fmt.Sprintf("/services/haproxy/configuration/stick_tables/%s?transaction_id=%s", name, transactionID), nil)
		if err != nil {
			return nil, err
		}
		return c.httpClient.Do(req)
	})
	return err
}

// CreateTcpCheck creates a new tcp_check.
func (c *HAProxyClient) CreateTcpCheck(ctx context.Context, parentType, parentName string, payload *TcpCheckPayload) error {
	_, err := c.Transaction(func(transactionID string) (*http.Response, error) {
		req, err := c.newRequest(ctx, "POST", fmt.Sprintf("/services/haproxy/configuration/tcp_checks?parent_type=%s&parent_name=%s&transaction_id=%s", parentType, parentName, transactionID), payload)
		if err != nil {
			return nil, err
		}
		return c.httpClient.Do(req)
	})
	return err
}

// ReadTcpChecks reads all tcp_checks for a given parent.
func (c *HAProxyClient) ReadTcpChecks(ctx context.Context, parentType, parentName string) ([]TcpCheckPayload, error) {
	req, err := c.newRequest(ctx, "GET", fmt.Sprintf("/services/haproxy/configuration/tcp_checks?parent_type=%s&parent_name=%s", parentType, parentName), nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return []TcpCheckPayload{}, nil // No tcp_checks found is not an error
	}

	if resp.StatusCode == http.StatusUnprocessableEntity {
		// 422 usually means invalid parameters, treat as no checks found
		return []TcpCheckPayload{}, nil
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var tcpChecksWrapper struct {
		Data []TcpCheckPayload `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tcpChecksWrapper); err != nil {
		return nil, err
	}

	return tcpChecksWrapper.Data, nil
}

// UpdateTcpCheck updates a tcp_check.
func (c *HAProxyClient) UpdateTcpCheck(ctx context.Context, index int64, parentType, parentName string, payload *TcpCheckPayload) error {
	_, err := c.Transaction(func(transactionID string) (*http.Response, error) {
		req, err := c.newRequest(ctx, "PUT", fmt.Sprintf("/services/haproxy/configuration/tcp_checks/%d?parent_type=%s&parent_name=%s&transaction_id=%s", index, parentType, parentName, transactionID), payload)
		if err != nil {
			return nil, err
		}
		return c.httpClient.Do(req)
	})
	return err
}

// DeleteTcpCheck deletes a tcp_check.
func (c *HAProxyClient) DeleteTcpCheck(ctx context.Context, index int64, parentType, parentName string) error {
	_, err := c.Transaction(func(transactionID string) (*http.Response, error) {
		req, err := c.newRequest(ctx, "DELETE", fmt.Sprintf("/services/haproxy/configuration/tcp_checks/%d?parent_type=%s&parent_name=%s&transaction_id=%s", index, parentType, parentName, transactionID), nil)
		if err != nil {
			return nil, err
		}
		return c.httpClient.Do(req)
	})
	return err
}

// CreateTcpRequestRule creates a new tcp_request_rule.
func (c *HAProxyClient) CreateTcpRequestRule(ctx context.Context, parentType, parentName string, payload *TcpRequestRulePayload) error {
	_, err := c.Transaction(func(transactionID string) (*http.Response, error) {
		req, err := c.newRequest(ctx, "POST", fmt.Sprintf("/services/haproxy/configuration/tcp_request_rules?parent_type=%s&parent_name=%s&transaction_id=%s", parentType, parentName, transactionID), payload)
		if err != nil {
			return nil, err
		}
		return c.httpClient.Do(req)
	})
	return err
}

// ReadTcpRequestRules reads all tcp_request_rules for a given parent.
func (c *HAProxyClient) ReadTcpRequestRules(ctx context.Context, parentType, parentName string) ([]TcpRequestRulePayload, error) {
	req, err := c.newRequest(ctx, "GET", fmt.Sprintf("/services/haproxy/configuration/tcp_request_rules?parent_type=%s&parent_name=%s", parentType, parentName), nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return []TcpRequestRulePayload{}, nil // No tcp_request_rules found is not an error
	}

	if resp.StatusCode == http.StatusUnprocessableEntity {
		// 422 usually means invalid parameters, treat as no rules found
		return []TcpRequestRulePayload{}, nil
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var tcpRequestRulesWrapper struct {
		Data []TcpRequestRulePayload `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tcpRequestRulesWrapper); err != nil {
		return nil, err
	}

	return tcpRequestRulesWrapper.Data, nil
}

// UpdateTcpRequestRule updates a tcp_request_rule.
func (c *HAProxyClient) UpdateTcpRequestRule(ctx context.Context, index int64, parentType, parentName string, payload *TcpRequestRulePayload) error {
	_, err := c.Transaction(func(transactionID string) (*http.Response, error) {
		req, err := c.newRequest(ctx, "PUT", fmt.Sprintf("/services/haproxy/configuration/tcp_request_rules/%d?parent_type=%s&parent_name=%s&transaction_id=%s", index, parentType, parentName, transactionID), payload)
		if err != nil {
			return nil, err
		}
		return c.httpClient.Do(req)
	})
	return err
}

// DeleteTcpRequestRule deletes a tcp_request_rule.
func (c *HAProxyClient) DeleteTcpRequestRule(ctx context.Context, index int64, parentType, parentName string) error {
	_, err := c.Transaction(func(transactionID string) (*http.Response, error) {
		req, err := c.newRequest(ctx, "DELETE", fmt.Sprintf("/services/haproxy/configuration/tcp_request_rules/%d?parent_type=%s&parent_name=%s&transaction_id=%s", index, parentType, parentName, transactionID), nil)
		if err != nil {
			return nil, err
		}
		return c.httpClient.Do(req)
	})
	return err
}

// CreateTcpResponseRule creates a new tcp_response_rule.
func (c *HAProxyClient) CreateTcpResponseRule(ctx context.Context, parentType, parentName string, payload *TcpResponseRulePayload) error {
	_, err := c.Transaction(func(transactionID string) (*http.Response, error) {
		req, err := c.newRequest(ctx, "POST", fmt.Sprintf("/services/haproxy/configuration/tcp_response_rules?parent_type=%s&parent_name=%s&transaction_id=%s", parentType, parentName, transactionID), payload)
		if err != nil {
			return nil, err
		}
		return c.httpClient.Do(req)
	})
	return err
}

// ReadTcpResponseRules reads all tcp_response_rules for a given parent.
func (c *HAProxyClient) ReadTcpResponseRules(ctx context.Context, parentType, parentName string) ([]TcpResponseRulePayload, error) {
	req, err := c.newRequest(ctx, "GET", fmt.Sprintf("/services/haproxy/configuration/tcp_response_rules?parent_type=%s&parent_name=%s", parentType, parentName), nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return []TcpResponseRulePayload{}, nil // No tcp_response_rules found is not an error
	}

	if resp.StatusCode == http.StatusUnprocessableEntity {
		// 422 usually means invalid parameters, treat as no rules found
		return []TcpResponseRulePayload{}, nil
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var tcpResponseRulesWrapper struct {
		Data []TcpResponseRulePayload `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tcpResponseRulesWrapper); err != nil {
		return nil, err
	}

	return tcpResponseRulesWrapper.Data, nil
}

// UpdateTcpResponseRule updates a tcp_response_rule.
func (c *HAProxyClient) UpdateTcpResponseRule(ctx context.Context, index int64, parentType, parentName string, payload *TcpResponseRulePayload) error {
	_, err := c.Transaction(func(transactionID string) (*http.Response, error) {
		req, err := c.newRequest(ctx, "PUT", fmt.Sprintf("/services/haproxy/configuration/tcp_response_rules/%d?parent_type=%s&parent_name=%s&transaction_id=%s", index, parentType, parentName, transactionID), payload)
		if err != nil {
			return nil, err
		}
		return c.httpClient.Do(req)
	})
	return err
}

// DeleteTcpResponseRule deletes a tcp_response_rule.
func (c *HAProxyClient) DeleteTcpResponseRule(ctx context.Context, index int64, parentType, parentName string) error {
	_, err := c.Transaction(func(transactionID string) (*http.Response, error) {
		req, err := c.newRequest(ctx, "DELETE", fmt.Sprintf("/services/haproxy/configuration/tcp_response_rules/%d?parent_type=%s&parent_name=%s&transaction_id=%s", index, parentType, parentName, transactionID), nil)
		if err != nil {
			return nil, err
		}
		return c.httpClient.Do(req)
	})
	return err
}

// CreateLogForward creates a new log_forward.
func (c *HAProxyClient) CreateLogForward(ctx context.Context, payload *LogForwardPayload) error {
	_, err := c.Transaction(func(transactionID string) (*http.Response, error) {
		req, err := c.newRequest(ctx, "POST", fmt.Sprintf("/services/haproxy/configuration/log_forwards?transaction_id=%s", transactionID), payload)
		if err != nil {
			return nil, err
		}
		return c.httpClient.Do(req)
	})
	return err
}

// ReadLogForward reads a log_forward.
func (c *HAProxyClient) ReadLogForward(ctx context.Context, name string) (*LogForwardPayload, error) {
	req, err := c.newRequest(ctx, "GET", fmt.Sprintf("/services/haproxy/configuration/log_forwards/%s", name), nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var logForwardWrapper struct {
		Data LogForwardPayload `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&logForwardWrapper); err != nil {
		return nil, err
	}

	return &logForwardWrapper.Data, nil
}

// UpdateLogForward updates a log_forward.
func (c *HAProxyClient) UpdateLogForward(ctx context.Context, name string, payload *LogForwardPayload) error {
	_, err := c.Transaction(func(transactionID string) (*http.Response, error) {
		req, err := c.newRequest(ctx, "PUT", fmt.Sprintf("/services/haproxy/configuration/log_forwards/%s?transaction_id=%s", name, transactionID), payload)
		if err != nil {
			return nil, err
		}
		return c.httpClient.Do(req)
	})
	return err
}

// DeleteLogForward deletes a log_forward.
func (c *HAProxyClient) DeleteLogForward(ctx context.Context, name string) error {
	_, err := c.Transaction(func(transactionID string) (*http.Response, error) {
		req, err := c.newRequest(ctx, "DELETE", fmt.Sprintf("/services/haproxy/configuration/log_forwards/%s?transaction_id=%s", name, transactionID), nil)
		if err != nil {
			return nil, err
		}
		return c.httpClient.Do(req)
	})
	return err
}

// CreateGlobal creates a new global.
func (c *HAProxyClient) CreateGlobal(ctx context.Context, payload *GlobalPayload) error {
	_, err := c.Transaction(func(transactionID string) (*http.Response, error) {
		req, err := c.newRequest(ctx, "PUT", fmt.Sprintf("/services/haproxy/configuration/global?transaction_id=%s", transactionID), payload)
		if err != nil {
			return nil, err
		}
		return c.httpClient.Do(req)
	})
	return err
}

// ReadGlobal reads a global.
func (c *HAProxyClient) ReadGlobal(ctx context.Context) (*GlobalPayload, error) {
	req, err := c.newRequest(ctx, "GET", "/services/haproxy/configuration/global", nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var globalWrapper struct {
		Data GlobalPayload `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&globalWrapper); err != nil {
		return nil, err
	}

	return &globalWrapper.Data, nil
}

// UpdateGlobal updates a global.
func (c *HAProxyClient) UpdateGlobal(ctx context.Context, payload *GlobalPayload) error {
	_, err := c.Transaction(func(transactionID string) (*http.Response, error) {
		req, err := c.newRequest(ctx, "PUT", fmt.Sprintf("/services/haproxy/configuration/global?transaction_id=%s", transactionID), payload)
		if err != nil {
			return nil, err
		}
		return c.httpClient.Do(req)
	})
	return err
}

// DeleteGlobal deletes a global.
func (c *HAProxyClient) DeleteGlobal(ctx context.Context) error {
	_, err := c.Transaction(func(transactionID string) (*http.Response, error) {
		req, err := c.newRequest(ctx, "DELETE", fmt.Sprintf("/services/haproxy/configuration/global?transaction_id=%s", transactionID), nil)
		if err != nil {
			return nil, err
		}
		return c.httpClient.Do(req)
	})
	return err
}

// CreateHttpRequestRuleInTransaction creates a new httprequestrule using an existing transaction ID.
func (c *HAProxyClient) CreateHttpRequestRuleInTransaction(ctx context.Context, transactionID, parentType, parentName string, payload *HttpRequestRulePayload) error {
	var url string
	var method string
	var requestPayload interface{}

	if c.apiVersion == "v3" {
		// v3: Use nested endpoint under frontends/backends
		// v3 doesn't support POST for individual rules - only PUT to replace entire list
		// v3 expects an array of rules, not a single rule
		parentTypePlural := parentType + "s"
		url = fmt.Sprintf("/services/haproxy/configuration/%s/%s/http_request_rules?transaction_id=%s",
			parentTypePlural, parentName, transactionID)
		method = "PUT"

		// For v3, we need to read existing rules first, then add the new one
		// This is a limitation of v3 - we can't create individual rules
		existingRules, err := c.ReadHttpRequestRules(ctx, parentType, parentName)
		if err != nil {
			return fmt.Errorf("failed to read existing HTTP request rules for v3: %w", err)
		}

		// Add the new rule to existing rules
		allRules := append(existingRules, *payload)
		requestPayload = allRules
	} else {
		// v2: Use query parameter approach with POST
		url = fmt.Sprintf("/services/haproxy/configuration/http_request_rules?parent_type=%s&parent_name=%s&transaction_id=%s",
			parentType, parentName, transactionID)
		method = "POST"
		requestPayload = payload
	}

	log.Printf("DEBUG: Using HTTP request rule endpoint: %s with method %s for API version %s", url, method, c.apiVersion)

	req, err := c.newRequest(ctx, method, url, requestPayload)
	if err != nil {
		return err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("HTTP request rule creation failed with status %d", resp.StatusCode, string(body))
		}
		return fmt.Errorf("HTTP request rule creation failed with status %d: %s", resp.StatusCode, string(body))
	}

	log.Printf("HTTP request rule created successfully in transaction: %s", transactionID)
	return nil
}

// CreateAllHttpRequestRulesInTransaction creates all HTTP request rules at once using an existing transaction ID
func (c *HAProxyClient) CreateAllHttpRequestRulesInTransaction(ctx context.Context, transactionID, parentType, parentName string, payloads []HttpRequestRulePayload) error {
	var url string
	var method string

	if c.apiVersion == "v3" {
		// v3: Use nested endpoint under frontends/backends - send all at once
		parentTypePlural := parentType + "s"
		url = fmt.Sprintf("/services/haproxy/configuration/%s/%s/http_request_rules?transaction_id=%s",
			parentTypePlural, parentName, transactionID)
		method = "PUT"

		// Debug logging for v3
		payloadJSON, _ := json.Marshal(payloads)
		log.Printf("DEBUG: API %s - Creating all HTTP request rules at once:", c.apiVersion)
		log.Printf("DEBUG: API %s - Method: %s", c.apiVersion, method)
		log.Printf("DEBUG: API %s - Endpoint: %s", c.apiVersion, url)
		log.Printf("DEBUG: API %s - Payload count: %d", c.apiVersion, len(payloads))
		log.Printf("DEBUG: API %s - Payload: %s", c.apiVersion, string(payloadJSON))

		req, err := c.newRequest(ctx, method, url, payloads)
		if err != nil {
			return err
		}

		resp, err := c.httpClient.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				return fmt.Errorf("HTTP request rules creation failed with status %d", resp.StatusCode)
			}
			return fmt.Errorf("HTTP request rules creation failed with status %d: %s", resp.StatusCode, string(body))
		}

		log.Printf("All HTTP request rules created successfully in transaction: %s", transactionID)
		return nil
	} else {
		// v2: Create HTTP request rules individually (v2 doesn't support bulk creation)
		log.Printf("DEBUG: API %s - Creating HTTP request rules individually (v2 limitation):", c.apiVersion)
		log.Printf("DEBUG: API %s - Payload count: %d", c.apiVersion, len(payloads))

		for i, payload := range payloads {
			url := fmt.Sprintf("/services/haproxy/configuration/http_request_rules?parent_type=%s&parent_name=%s&transaction_id=%s",
				parentType, parentName, transactionID)
			method := "POST"

			// Debug logging for each individual HTTP request rule
			payloadJSON, _ := json.Marshal(payload)
			log.Printf("DEBUG: API %s - Creating HTTP request rule %d/%d:", c.apiVersion, i+1, len(payloads))
			log.Printf("DEBUG: API %s - Method: %s", c.apiVersion, method)
			log.Printf("DEBUG: API %s - Endpoint: %s", c.apiVersion, url)
			log.Printf("DEBUG: API %s - Payload: %s", c.apiVersion, string(payloadJSON))

			req, err := c.newRequest(ctx, method, url, payload)
			if err != nil {
				return err
			}

			resp, err := c.httpClient.Do(req)
			if err != nil {
				return err
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
				body, err := io.ReadAll(resp.Body)
				if err != nil {
					return fmt.Errorf("HTTP request rule %d creation failed with status %d", i+1, resp.StatusCode)
				}
				return fmt.Errorf("HTTP request rule %d creation failed with status %d: %s", i+1, resp.StatusCode, string(body))
			}

			log.Printf("HTTP request rule %d/%d created successfully in transaction: %s", i+1, len(payloads), transactionID)
		}

		log.Printf("All %d HTTP request rules created successfully in transaction: %s", len(payloads), transactionID)
		return nil
	}
}

// DeleteHttpRequestRuleInTransaction deletes an existing httprequestrule using an existing transaction ID.
func (c *HAProxyClient) DeleteHttpRequestRuleInTransaction(ctx context.Context, transactionID string, index int64, parentType, parentName string) error {
	var url string
	if c.apiVersion == "v3" {
		// v3: Use nested endpoint under frontends/backends
		// Properly pluralize the parent type
		parentTypePlural := parentType + "s"
		url = fmt.Sprintf("/services/haproxy/configuration/%s/%s/http_request_rules/%d?transaction_id=%s",
			parentTypePlural, parentName, index, transactionID)
	} else {
		// v2: Use query parameter approach
		url = fmt.Sprintf("/services/haproxy/configuration/http_request_rules/%d?parent_type=%s&parent_name=%s&transaction_id=%s",
			index, parentType, parentName, transactionID)
	}

	log.Printf("DEBUG: Using HTTP request rule delete endpoint: %s for API version %s", url, c.apiVersion)

	req, err := c.newRequest(ctx, "DELETE", url, nil)
	if err != nil {
		return err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted && resp.StatusCode != http.StatusNoContent {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("HTTP request rule deletion failed with status %d", resp.StatusCode)
		}
		return fmt.Errorf("HTTP request rule deletion failed with status %d: %s", resp.StatusCode, string(body))
	}

	log.Printf("HTTP request rule deleted successfully in transaction: %s", transactionID)
	return nil
}
