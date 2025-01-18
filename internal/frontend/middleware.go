package frontend

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"terraform-provider-haproxy/internal/utils"
)

// Add a frontend configuration
func (c *ConfigFrontend) AddFrontendConfiguration(payload []byte, TransactionID string) (*http.Response, error) {
	url := fmt.Sprintf("%s/v2/services/haproxy/configuration/frontends?transaction_id=%s", c.BaseURL, TransactionID)
	headers := map[string]string{
		"Content-Type": "application/json",
	}

	resp, err := utils.HTTPRequest("POST", url, payload, headers, c.Username, c.Password)
	if err != nil {
		log.Println("fronted Create Error sending request:", err, resp)
		return nil, err
	}

	if resp.StatusCode >= 200 && resp.StatusCode <= 299 {
		log.Println("fronted Create Succesful sending request:", resp)
	} else {
		body, readErr := io.ReadAll(resp.Body)
		fmt.Printf("Debug: Response Body: %s\n", string(body))
		fmt.Printf("Debug: Response readErr: %s\n", readErr)

		if readErr != nil {
			log.Println("Error reading response body:", readErr)
			return nil, fmt.Errorf("error reading response body: %v", readErr)
		}

		fmt.Printf("Debug: Unexpected fronted status code %d\n", resp.StatusCode)
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
			RawMessage: fmt.Sprintf("error: unexpected frontend status code %d: %s", resp.StatusCode, string(body)),
		}
	}
	defer resp.Body.Close()
	return resp, nil
}

// Get a frontend configuration
func (c *ConfigFrontend) GetAFrontendConfiguration(FrontendName string) (*http.Response, error) {
	url := fmt.Sprintf("%s/v2/services/haproxy/configuration/frontends/%s", c.BaseURL, FrontendName)
	headers := map[string]string{
		"Content-Type": "application/json",
	}
	resp, err := utils.HTTPRequest("GET", url, nil, headers, c.Username, c.Password)

	if err != nil {
		log.Println("fronted Error reading request:", err, resp)
		return nil, err
	}

	if resp.StatusCode >= 200 && resp.StatusCode <= 299 {
		log.Println("fronted Read Succesful sending request:", resp)
	} else {
		body, readErr := io.ReadAll(resp.Body)
		fmt.Printf("Debug: Response Body: %s\n", string(body))
		fmt.Printf("Debug: Response readErr: %s\n", readErr)

		if readErr != nil {
			log.Println("Error reading response body:", readErr)
			return nil, fmt.Errorf("error reading response body: %v", readErr)
		}

		fmt.Printf("Debug: Unexpected fronted status code %d\n", resp.StatusCode)
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
			RawMessage: fmt.Sprintf("error: unexpected frontend status code %d: %s", resp.StatusCode, string(body)),
		}
	}

	return resp, nil
}

// Update a frontend configuration
func (c *ConfigFrontend) UpdateFrontendConfiguration(FrontendName string, payload []byte, TransactionID string) (*http.Response, error) {
	url := fmt.Sprintf("%s/v2/services/haproxy/configuration/frontends/%s?transaction_id=%s", c.BaseURL, FrontendName, TransactionID)
	headers := map[string]string{
		"Content-Type": "application/json",
	}
	resp, err := utils.HTTPRequest("PUT", url, payload, headers, c.Username, c.Password)

	if err != nil {
		log.Println("fronted Update Error sending request:", err, resp)
		return nil, err
	}

	if resp.StatusCode >= 200 && resp.StatusCode <= 299 {
		log.Println("fronted Update Succesful sending request:", resp)
	} else {
		body, readErr := io.ReadAll(resp.Body)
		fmt.Printf("Debug: Response Body: %s\n", string(body))
		fmt.Printf("Debug: Response readErr: %s\n", readErr)

		if readErr != nil {
			log.Println("Error reading response body:", readErr)
			return nil, fmt.Errorf("error reading response body: %v", readErr)
		}

		fmt.Printf("Debug: Unexpected fronted status code %d\n", resp.StatusCode)
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
			RawMessage: fmt.Sprintf("error: unexpected frontend status code %d: %s", resp.StatusCode, string(body)),
		}
	}
	defer resp.Body.Close()

	return resp, nil
}

// Delete a frontend configuration
func (c *ConfigFrontend) DeleteFrontendConfiguration(FrontendName string, TransactionID string) (*http.Response, error) {
	url := fmt.Sprintf("%s/v2/services/haproxy/configuration/frontends/%s?transaction_id=%s", c.BaseURL, FrontendName, TransactionID)
	headers := map[string]string{
		"Content-Type": "application/json",
	}
	resp, err := utils.HTTPRequest("DELETE", url, nil, headers, c.Username, c.Password)

	if err != nil {
		log.Println("fronted Error sending request:", err, resp)
		return nil, err
	}

	if resp.StatusCode >= 200 && resp.StatusCode <= 299 {
		log.Println("fronted Delete Succesful sending request:", resp)
	} else {
		body, readErr := io.ReadAll(resp.Body)
		fmt.Printf("Debug: Response Body: %s\n", string(body))
		fmt.Printf("Debug: Response readErr: %s\n", readErr)

		if readErr != nil {
			log.Println("Error reading response body:", readErr)
			return nil, fmt.Errorf("error reading response body: %v", readErr)
		}

		fmt.Printf("Debug: Unexpected fronted status code %d\n", resp.StatusCode)
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
			RawMessage: fmt.Sprintf("error: unexpected frontend status code %d: %s", resp.StatusCode, string(body)),
		}
	}
	defer resp.Body.Close()
	return resp, nil
}
