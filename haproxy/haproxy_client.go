package haproxy

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"terraform-provider-haproxy/haproxy/utils"
)

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
func (c *HAProxyClient) CreateFrontend(ctx context.Context, payload *FrontendPayload) error {
	_, err := c.Transaction(func(transactionID string) (*http.Response, error) {
		req, err := c.newRequest(ctx, "POST", fmt.Sprintf("/services/haproxy/configuration/frontends?transaction_id=%s", transactionID), payload)
		if err != nil {
			return nil, err
		}
		return c.httpClient.Do(req)
	})
	return err
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

// CreateACL creates a new ACL rule for a frontend.
func (c *HAProxyClient) CreateACL(ctx context.Context, parentType, parentName string, payload *ACLPayload) error {
	// Debug: Log the ACL payload being sent
	payloadJSON, _ := json.Marshal(payload)
	log.Printf("DEBUG: Creating ACL with payload: %s", string(payloadJSON))

	_, err := c.Transaction(func(transactionID string) (*http.Response, error) {
		url := fmt.Sprintf("/services/haproxy/configuration/acls?parent_type=%s&parent_name=%s&transaction_id=%s",
			parentType, parentName, transactionID)
		req, err := c.newRequest(ctx, "POST", url, payload)
		if err != nil {
			return nil, err
		}
		return c.httpClient.Do(req)
	})
	return err
}

// ReadACLs reads all ACL rules for a parent (frontend, backend, etc.).
func (c *HAProxyClient) ReadACLs(ctx context.Context, parentType, parentName string) ([]ACLPayload, error) {
	req, err := c.newRequest(ctx, "GET", fmt.Sprintf("/services/haproxy/configuration/acls?parent_type=%s&parent_name=%s", parentType, parentName), nil)
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

	var response struct {
		Data []ACLPayload `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	return response.Data, nil
}

// UpdateACL updates an existing ACL rule by index.
func (c *HAProxyClient) UpdateACL(ctx context.Context, parentType, parentName string, index int64, payload *ACLPayload) error {
	_, err := c.Transaction(func(transactionID string) (*http.Response, error) {
		url := fmt.Sprintf("/services/haproxy/configuration/acls/%d?parent_type=%s&parent_name=%s&transaction_id=%s",
			index, parentType, parentName, transactionID)
		req, err := c.newRequest(ctx, "PUT", url, payload)
		if err != nil {
			return nil, err
		}
		return c.httpClient.Do(req)
	})
	return err
}

// DeleteACL deletes an ACL rule by index.
func (c *HAProxyClient) DeleteACL(ctx context.Context, parentType, parentName string, index int64) error {
	_, err := c.Transaction(func(transactionID string) (*http.Response, error) {
		url := fmt.Sprintf("/services/haproxy/configuration/acls/%d?parent_type=%s&parent_name=%s&transaction_id=%s",
			index, parentType, parentName, transactionID)
		req, err := c.newRequest(ctx, "DELETE", url, nil)
		if err != nil {
			return nil, err
		}
		return c.httpClient.Do(req)
	})
	return err
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
			if rollbackErr := c.rollbackTransaction(transactionID); rollbackErr != nil {
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
	req, err := c.newRequest(ctx, "GET", fmt.Sprintf("/services/haproxy/configuration/servers?parent_type=%s&parent_name=%s", parentType, parentName), nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return []ServerPayload{}, nil // No servers found is not an error
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var serversWrapper struct {
		Data []ServerPayload `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&serversWrapper); err != nil {
		return nil, err
	}

	return serversWrapper.Data, nil
}

// CreateServerInTransaction creates a new server using an existing transaction ID.
func (c *HAProxyClient) CreateServerInTransaction(ctx context.Context, transactionID, parentType, parentName string, payload *ServerPayload) error {
	log.Printf("CreateServerInTransaction called with transaction ID: %s, parent: %s/%s, payload: %+v", transactionID, parentType, parentName, payload)
	req, err := c.newRequest(ctx, "POST", fmt.Sprintf("/services/haproxy/configuration/servers?parent_type=%s&parent_name=%s&transaction_id=%s", parentType, parentName, transactionID), payload)
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
	req, err := c.newRequest(ctx, "GET", fmt.Sprintf("/services/haproxy/configuration/binds?parent_type=%s&parent_name=%s", parentType, parentName), nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return []BindPayload{}, nil // No binds found is not an error
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var bindsWrapper struct {
		Data []BindPayload `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&bindsWrapper); err != nil {
		return nil, err
	}

	return bindsWrapper.Data, nil
}

// CreateAcl creates a new acl.
func (c *HAProxyClient) CreateAcl(ctx context.Context, parentType, parentName string, payload *AclPayload) error {
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
func (c *HAProxyClient) UpdateAcl(ctx context.Context, index int64, parentType, parentName string, payload *AclPayload) error {
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
func (c *HAProxyClient) CreateHttpRequestRule(ctx context.Context, parentType, parentName string, payload *HttpRequestRulePayload) error {
	_, err := c.Transaction(func(transactionID string) (*http.Response, error) {
		req, err := c.newRequest(ctx, "POST", fmt.Sprintf("/services/haproxy/configuration/http_request_rules?parent_type=%s&parent_name=%s&transaction_id=%s", parentType, parentName, transactionID), payload)
		if err != nil {
			return nil, err
		}
		return c.httpClient.Do(req)
	})
	return err
}

// ReadHttpRequestRules reads all httprequestrules for a given parent.
func (c *HAProxyClient) ReadHttpRequestRules(ctx context.Context, parentType, parentName string) ([]HttpRequestRulePayload, error) {
	req, err := c.newRequest(ctx, "GET", fmt.Sprintf("/services/haproxy/configuration/http_request_rules?parent_type=%s&parent_name=%s", parentType, parentName), nil)
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

	var httpRequestRulesWrapper struct {
		Data []HttpRequestRulePayload `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&httpRequestRulesWrapper); err != nil {
		return nil, err
	}

	return httpRequestRulesWrapper.Data, nil
}

// UpdateHttpRequestRule updates a httprequestrule.
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
func (c *HAProxyClient) DeleteHttpRequestRule(ctx context.Context, index int64, parentType, parentName string) error {
	_, err := c.Transaction(func(transactionID string) (*http.Response, error) {
		req, err := c.newRequest(ctx, "DELETE", fmt.Sprintf("/services/haproxy/configuration/http_request_rules/%d?parent_type=%s&parent_name=%s&transaction_id=%s", index, parentType, parentName, transactionID), nil)
		if err != nil {
			return nil, err
		}
		return c.httpClient.Do(req)
	})
	return err
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
