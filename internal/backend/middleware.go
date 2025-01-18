package backend

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"terraform-provider-haproxy/internal/transaction"
	"terraform-provider-haproxy/internal/utils"
)

func NewManager(config *transaction.ConfigTransaction) *Manager {
	return &Manager{Config: config}
}

// Add a backend configuration
func (c *ConfigBackend) AddBackendConfiguration(payload []byte, TransactionID string) (*http.Response, error) {
	url := fmt.Sprintf("%s/v2/services/haproxy/configuration/backends?transaction_id=%s", c.BaseURL, TransactionID)
	headers := map[string]string{
		"Content-Type": "application/json",
	}

	resp, err := utils.HTTPRequest("POST", url, payload, headers, c.Username, c.Password)

	if err != nil {
		log.Println("backend Create Error sending request:", err, resp)
		return nil, err
	}

	if resp.StatusCode >= 200 && resp.StatusCode <= 299 {
		log.Println("backend Create Succesful sending request:", resp)
	} else {
		body, readErr := io.ReadAll(resp.Body)
		fmt.Printf("Debug: Response Body: %s\n", string(body))
		fmt.Printf("Debug: Response readErr: %s\n", readErr)

		if readErr != nil {
			log.Println("Error reading response body:", readErr)
			return nil, fmt.Errorf("error reading response body: %v", readErr)
		}

		fmt.Printf("Debug: Unexpected backend status code %d\n", resp.StatusCode)
		fmt.Printf("Response Body: %s\n", string(body))

		var apiError utils.APIError
		if jsonErr := json.Unmarshal(body, &apiError); jsonErr == nil {
			return nil, &utils.CustomError{
				StatusCode: resp.StatusCode,
				APIError:   &apiError,
			}
		}
		return nil, &utils.CustomError{
			StatusCode: resp.StatusCode,
			RawMessage: fmt.Sprintf("error: unexpected backend status code %d: %s", resp.StatusCode, string(body)),
		}
	}
	defer resp.Body.Close()
	return resp, nil
}

// Get a backend configuration
func (c *ConfigBackend) GetABackendConfiguration(backendName string) (*http.Response, error) {
	url := fmt.Sprintf("%s/v2/services/haproxy/configuration/backends/%s", c.BaseURL, backendName)
	headers := map[string]string{
		"Content-Type": "application/json",
	}
	resp, err := utils.HTTPRequest("GET", url, nil, headers, c.Username, c.Password)

	if err != nil {
		log.Println("backend Error reading request:", err, resp)
		return nil, err
	}

	if resp.StatusCode >= 200 && resp.StatusCode <= 299 {
		log.Println("backend Read Succesful sending request:", resp)
	} else {
		body, readErr := io.ReadAll(resp.Body)
		fmt.Printf("Debug: Response Body: %s\n", string(body))
		fmt.Printf("Debug: Response readErr: %s\n", readErr)

		if readErr != nil {
			log.Println("Error reading response body:", readErr)
			return nil, fmt.Errorf("error reading response body: %v", readErr)
		}

		fmt.Printf("Debug: Unexpected backend status code %d\n", resp.StatusCode)
		fmt.Printf("Response Body: %s\n", string(body))
		var apiError utils.APIError
		if jsonErr := json.Unmarshal(body, &apiError); jsonErr == nil {
			return nil, &utils.CustomError{
				StatusCode: resp.StatusCode,
				APIError:   &apiError,
			}
		}
		return nil, &utils.CustomError{
			StatusCode: resp.StatusCode,
			RawMessage: fmt.Sprintf("error: unexpected backend status code %d: %s", resp.StatusCode, string(body)),
		}

	}

	return resp, nil
}

// Updated a backend configuration
func (c *ConfigBackend) UpdateBackendConfiguration(backendName string, payload []byte, TransactionID string) (*http.Response, error) {
	url := fmt.Sprintf("%s/v2/services/haproxy/configuration/backends/%s?transaction_id=%s", c.BaseURL, backendName, TransactionID)
	headers := map[string]string{
		"Content-Type": "application/json",
	}
	resp, err := utils.HTTPRequest("PUT", url, payload, headers, c.Username, c.Password)

	if err != nil {
		log.Println("backend Update Error sending request:", err, resp)
		return nil, err
	}

	if resp.StatusCode >= 200 && resp.StatusCode <= 299 {
		log.Println("backend Update Succesful sending request:", resp)
	} else {
		body, readErr := io.ReadAll(resp.Body)
		fmt.Printf("Debug: Response Body: %s\n", string(body))
		fmt.Printf("Debug: Response readErr: %s\n", readErr)

		if readErr != nil {
			log.Println("Error reading response body:", readErr)
			return nil, fmt.Errorf("error reading response body: %v", readErr)
		}

		fmt.Printf("Debug: Unexpected backend status code %d\n", resp.StatusCode)
		fmt.Printf("Response Body: %s\n", string(body))

		var apiError utils.APIError
		if jsonErr := json.Unmarshal(body, &apiError); jsonErr == nil {
			return nil, &utils.CustomError{
				StatusCode: resp.StatusCode,
				APIError:   &apiError,
			}
		}
		return nil, &utils.CustomError{
			StatusCode: resp.StatusCode,
			RawMessage: fmt.Sprintf("error: unexpected backend status code %d: %s", resp.StatusCode, string(body)),
		}
	}

	return resp, nil
}

// Delete a backend configuration
func (c *ConfigBackend) DeleteBackendConfiguration(backendName string, TransactionID string) (*http.Response, error) {
	url := fmt.Sprintf("%s/v2/services/haproxy/configuration/backends/%s?transaction_id=%s", c.BaseURL, backendName, TransactionID)
	headers := map[string]string{
		"Content-Type": "application/json",
	}
	resp, err := utils.HTTPRequest("DELETE", url, nil, headers, c.Username, c.Password)

	if err != nil {
		log.Println("backend Error sending request:", err, resp)
		return nil, err
	}

	if resp.StatusCode >= 200 && resp.StatusCode <= 299 {
		log.Println("backend Delete Succesful sending request:", resp)
	} else {
		body, readErr := io.ReadAll(resp.Body)
		fmt.Printf("Debug: Response Body: %s\n", string(body))
		fmt.Printf("Debug: Response readErr: %s\n", readErr)

		if readErr != nil {
			log.Println("Error reading response body:", readErr)
			return nil, fmt.Errorf("error reading response body: %v", readErr)
		}

		fmt.Printf("Debug: Unexpected backend status code %d\n", resp.StatusCode)
		fmt.Printf("Response Body: %s\n", string(body))

		var apiError utils.APIError
		if jsonErr := json.Unmarshal(body, &apiError); jsonErr == nil {
			return nil, &utils.CustomError{
				StatusCode: resp.StatusCode,
				APIError:   &apiError,
			}
		}
		return nil, &utils.CustomError{
			StatusCode: resp.StatusCode,
			RawMessage: fmt.Sprintf("error: unexpected backend status code %d: %s", resp.StatusCode, string(body)),
		}
	}
	defer resp.Body.Close()
	return resp, nil
}
