package httprequestrule

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"terraform-provider-haproxy/internal/utils"
)

// Get all httprequestrule configurations
func (c *ConfigHttpRequestRule) GetAllHttpRequestRuleConfiguration(parentName string, parentType string) (*http.Response, error) {
	url := fmt.Sprintf("%s/v2/services/haproxy/configuration/http_request_rules?parent_name=%s&parent_type=%s", c.BaseURL, parentName, parentType)
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
			RawMessage: fmt.Sprintf("error: unexpected httprequestrule status code %d: %s", resp.StatusCode, string(body)),
		}
	}
	// defer resp.Body.Close()
	return resp, nil
}

// Get a httprequest rule configuration
func (c *ConfigHttpRequestRule) GetAHttpRequestRuleConfiguration(HttpRequestRuleIndexName int, parentName string, parentType string) (*http.Response, error) {
	url := fmt.Sprintf("%s/v2/services/haproxy/configuration/http_request_rules/%d?parent_name=%s&parent_type=%s", c.BaseURL, HttpRequestRuleIndexName, parentName, parentType)
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
			RawMessage: fmt.Sprintf("error: unexpected httprequestrule status code %d: %s", resp.StatusCode, string(body)),
		}
	}
	// defer resp.Body.Close()
	return resp, nil
}

// Add a httprequestrule configuration
func (c *ConfigHttpRequestRule) AddAHttpRequestRuleConfiguration(payload []byte, TransactionID string, parentName string, parentType string) (*http.Response, error) {
	url := fmt.Sprintf("%s/v2/services/haproxy/configuration/http_request_rules?transaction_id=%s&parent_name=%s&parent_type=%s", c.BaseURL, TransactionID, parentName, parentType)
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
			RawMessage: fmt.Sprintf("error: unexpected httprequestrule status code %d: %s", resp.StatusCode, string(body)),
		}
	}
	defer resp.Body.Close()

	return resp, nil
}

// Update a httprequestrule configuration
func (c *ConfigHttpRequestRule) UpdateAHttpRequestRuleConfiguration(HttpRequestRuleIndexName int, payload []byte, TransactionID string, parentName string, parentType string) (*http.Response, error) {
	fmt.Println("REQUPID", TransactionID)
	url := fmt.Sprintf("%s/v2/services/haproxy/configuration/http_request_rules/%d?transaction_id=%s&parent_name=%s&parent_type=%s", c.BaseURL, HttpRequestRuleIndexName, TransactionID, parentName, parentType)
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
			RawMessage: fmt.Sprintf("error: unexpected httprequestrule status code %d: %s", resp.StatusCode, string(body)),
		}
	}

	defer resp.Body.Close()
	return resp, nil
}

// Delete a httprequestrule configuration
func (c *ConfigHttpRequestRule) DeleteAHttpRequestRuleConfiguration(HttpRequestRuleIndexName int, TransactionID string, parentName string, parentType string) (*http.Response, error) {
	url := fmt.Sprintf("%s/v2/services/haproxy/configuration/http_request_rules/%d?transaction_id=%s&parent_name=%s&parent_type=%s", c.BaseURL, HttpRequestRuleIndexName, TransactionID, parentName, parentType)
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
			RawMessage: fmt.Sprintf("error: unexpected httprequestrule status code %d: %s", resp.StatusCode, string(body)),
		}
	}
	defer resp.Body.Close()
	return resp, nil
}
