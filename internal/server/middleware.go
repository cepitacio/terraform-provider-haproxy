package server

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"terraform-provider-haproxy/internal/utils"
)

// Add a server configuration
func (c *ConfigServer) AddServerConfiguration(payload []byte, TransactionID string, parentName string, parentType string) (*http.Response, error) {
	url := fmt.Sprintf("%s/v2/services/haproxy/configuration/servers?transaction_id=%s&parent_name=%s&parent_type=%s", c.BaseURL, TransactionID, parentName, parentType)
	headers := map[string]string{
		"Content-Type": "application/json",
	}
	resp, err := utils.HTTPRequest("POST", url, payload, headers, c.Username, c.Password)

	if err != nil {
		log.Println("Server Create Error sending request:", err, resp)
		return nil, err
	}

	if resp.StatusCode >= 200 && resp.StatusCode <= 299 {
		log.Println("Server Create Succesful sending request:", resp)
	} else {
		body, readErr := io.ReadAll(resp.Body)
		fmt.Printf("Debug: Response Body: %s\n", string(body))
		fmt.Printf("Debug: Response readErr: %s\n", readErr)

		if readErr != nil {
			log.Println("Error reading response body:", readErr)
			return nil, fmt.Errorf("error reading response body: %v", readErr)
		}

		fmt.Printf("Debug: Unexpected server status code %d\n", resp.StatusCode)
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
			RawMessage: fmt.Sprintf("error: unexpected server status code %d: %s", resp.StatusCode, string(body)),
		}
	}
	defer resp.Body.Close()

	return resp, nil
}

// Get a server configuration
func (c *ConfigServer) GetAServerConfiguration(ServerName string, parentName string, parentType string) (*http.Response, error) {
	url := fmt.Sprintf("%s/v2/services/haproxy/configuration/servers/%s?&parent_name=%s&parent_type=%s", c.BaseURL, ServerName, parentName, parentType)
	headers := map[string]string{
		"Content-Type": "application/json",
	}
	resp, err := utils.HTTPRequest("GET", url, nil, headers, c.Username, c.Password)

	if err != nil {
		log.Println("Server Error reading request:", err, resp)
		return nil, err
	}

	if resp.StatusCode >= 200 && resp.StatusCode <= 299 {
		log.Println("Server Read Succesful sending request:", resp)
	} else {
		body, readErr := io.ReadAll(resp.Body)
		fmt.Printf("Debug: Response Body: %s\n", string(body))
		fmt.Printf("Debug: Response readErr: %s\n", readErr)

		if readErr != nil {
			log.Println("Error reading response body:", readErr)
			return nil, fmt.Errorf("error reading response body: %v", readErr)
		}

		fmt.Printf("Debug: Unexpected server status code %d\n", resp.StatusCode)
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
			RawMessage: fmt.Sprintf("error: unexpected server status code %d: %s", resp.StatusCode, string(body)),
		}
	}

	return resp, nil
}

// Update a server configuration
func (c *ConfigServer) UpdateServerConfiguration(ServerName string, payload []byte, TransactionID string, parentName string, parentType string) (*http.Response, error) {
	url := fmt.Sprintf("%s/v2/services/haproxy/configuration/servers/%s?transaction_id=%s&parent_name=%s&parent_type=%s", c.BaseURL, ServerName, TransactionID, parentName, parentType)
	headers := map[string]string{
		"Content-Type": "application/json",
	}
	resp, err := utils.HTTPRequest("PUT", url, payload, headers, c.Username, c.Password)

	if err != nil {
		log.Println("Server Update Error sending request:", err, resp)
		return nil, err
	}

	if resp.StatusCode >= 200 && resp.StatusCode <= 299 {
		log.Println("Server Update Succesful sending request:", resp)
	} else {
		body, readErr := io.ReadAll(resp.Body)
		fmt.Printf("Debug: Response Body: %s\n", string(body))
		fmt.Printf("Debug: Response readErr: %s\n", readErr)

		if readErr != nil {
			log.Println("Error reading response body:", readErr)
			return nil, fmt.Errorf("error reading response body: %v", readErr)
		}

		fmt.Printf("Debug: Unexpected server status code %d\n", resp.StatusCode)
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
			RawMessage: fmt.Sprintf("error: unexpected server status code %d: %s", resp.StatusCode, string(body)),
		}
	}

	defer resp.Body.Close()
	return resp, nil
}

// Delete a server configuration
func (c *ConfigServer) DeleteServerConfiguration(ServerName string, TransactionID string, parentName string, parentType string) (*http.Response, error) {
	url := fmt.Sprintf("%s/v2/services/haproxy/configuration/servers/%s?transaction_id=%s&parent_name=%s&parent_type=%s", c.BaseURL, ServerName, TransactionID, parentName, parentType)
	headers := map[string]string{
		"Content-Type": "application/json",
	}
	resp, err := utils.HTTPRequest("DELETE", url, nil, headers, c.Username, c.Password)

	if err != nil {
		log.Println("Server Error sending request:", err, resp)
		return nil, err
	}

	if resp.StatusCode >= 200 && resp.StatusCode <= 299 {
		log.Println("Server Delete Succesful sending request:", resp)
	} else {
		body, readErr := io.ReadAll(resp.Body)
		fmt.Printf("Debug: Response Body: %s\n", string(body))
		fmt.Printf("Debug: Response readErr: %s\n", readErr)

		if readErr != nil {
			log.Println("Error reading response body:", readErr)
			return nil, fmt.Errorf("error reading response body: %v", readErr)
		}

		fmt.Printf("Debug: Unexpected server status code %d\n", resp.StatusCode)
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
			RawMessage: fmt.Sprintf("error: unexpected server status code %d: %s", resp.StatusCode, string(body)),
		}
	}
	defer resp.Body.Close()
	return resp, nil
}
