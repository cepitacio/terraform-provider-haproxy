package acl

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"terraform-provider-haproxy/internal/utils"
)

// Get all acl configurations
func (c *ConfigAcl) GetAllAclConfiguration(parentName string, parentType string) (*http.Response, error) {
	url := fmt.Sprintf("%s/v2/services/haproxy/configuration/acls/?parent_name=%s&parent_type=%s", c.BaseURL, parentName, parentType)
	headers := map[string]string{
		"Content-Type": "application/json",
	}
	resp, err := utils.HTTPRequest("GET", url, nil, headers, c.Username, c.Password)
	if err != nil {
		log.Println("acl Update Error reading request:", err, resp)
		return nil, err
	}

	if resp.StatusCode >= 200 && resp.StatusCode <= 299 {
		log.Println("acl Succesful reading request:", resp)
	} else {
		body, readErr := io.ReadAll(resp.Body)
		fmt.Printf("Debug: Response Body: %s\n", string(body))
		fmt.Printf("Debug: Response readErr: %s\n", readErr)

		if readErr != nil {
			log.Println("Error reading response body:", readErr)
			return nil, fmt.Errorf("error reading response body: %v", readErr)
		}

		fmt.Printf("Debug: Unexpected acl status code %d\n", resp.StatusCode)
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
			RawMessage: fmt.Sprintf("error: unexpected acl status code %d: %s", resp.StatusCode, string(body)),
		}
	}

	return resp, nil
}

// Get an acl configuration
func (c *ConfigAcl) GetAnAclConfiguration(AclIndexName int, parentName string, parentType string) (*http.Response, error) {
	url := fmt.Sprintf("%s/v2/services/haproxy/configuration/acls/%d?parent_name=%s&parent_type=%s", c.BaseURL, AclIndexName, parentName, parentType)
	headers := map[string]string{
		"Content-Type": "application/json",
	}
	resp, err := utils.HTTPRequest("GET", url, nil, headers, c.Username, c.Password)
	if err != nil {
		log.Println("acl Update Error reading request:", err, resp)
		return nil, err
	}

	if resp.StatusCode >= 200 && resp.StatusCode <= 299 {
		log.Println("acl Succesful reading request:", resp)
	} else {
		body, readErr := io.ReadAll(resp.Body)
		fmt.Printf("Debug: Response Body: %s\n", string(body))
		fmt.Printf("Debug: Response readErr: %s\n", readErr)

		if readErr != nil {
			log.Println("Error reading response body:", readErr)
			return nil, fmt.Errorf("error reading response body: %v", readErr)
		}

		fmt.Printf("Debug: Unexpected acl status code %d\n", resp.StatusCode)
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
			RawMessage: fmt.Sprintf("error: unexpected acl status code %d: %s", resp.StatusCode, string(body)),
		}
	}

	return resp, nil
}

// Add an acl configuration
func (c *ConfigAcl) AddAnAclConfiguration(payload []byte, TransactionID string, parentName string, parentType string) (*http.Response, error) {
	url := fmt.Sprintf("%s/v2/services/haproxy/configuration/acls?transaction_id=%s&parent_name=%s&parent_type=%s", c.BaseURL, TransactionID, parentName, parentType)
	headers := map[string]string{
		"Content-Type": "application/json",
	}
	resp, err := utils.HTTPRequest("POST", url, payload, headers, c.Username, c.Password)
	if err != nil {
		log.Println("acl Update Error sending request:", err, resp)
		return nil, err
	}

	if resp.StatusCode >= 200 && resp.StatusCode <= 299 {
		log.Println("acl Create Succesful sending request:", resp)
	} else {
		body, readErr := io.ReadAll(resp.Body)
		fmt.Printf("Debug: Response Body: %s\n", string(body))
		fmt.Printf("Debug: Response readErr: %s\n", readErr)

		if readErr != nil {
			log.Println("Error reading response body:", readErr)
			return nil, fmt.Errorf("error reading response body: %v", readErr)
		}
		fmt.Printf("Debug: Unexpected acl status code %d\n", resp.StatusCode)
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
			RawMessage: fmt.Sprintf("error: unexpected acl status code %d: %s", resp.StatusCode, string(body)),
		}
	}
	defer resp.Body.Close()
	return resp, nil
}

// Update an acl configuration
func (c *ConfigAcl) UpdateAnAclConfiguration(AclIndexName int, payload []byte, TransactionID string, parentName string, parentType string) (*http.Response, error) {
	url := fmt.Sprintf("%s/v2/services/haproxy/configuration/acls/%d?transaction_id=%s&parent_name=%s&parent_type=%s", c.BaseURL, AclIndexName, TransactionID, parentName, parentType)
	headers := map[string]string{
		"Content-Type": "application/json",
	}
	resp, err := utils.HTTPRequest("PUT", url, payload, headers, c.Username, c.Password)
	if err != nil {
		log.Println("acl Update Error sending request:", err, resp)
		return nil, err
	}

	if resp.StatusCode >= 200 && resp.StatusCode <= 299 {
		log.Println("acl Update Succesful sending request:", resp)
	} else {
		body, readErr := io.ReadAll(resp.Body)
		fmt.Printf("Debug: Response Body: %s\n", string(body))
		fmt.Printf("Debug: Response readErr: %s\n", readErr)

		if readErr != nil {
			log.Println("Error reading response body:", readErr)
			return nil, fmt.Errorf("error reading response body: %v", readErr)
		}

		fmt.Printf("Debug: Unexpected acl status code %d\n", resp.StatusCode)
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
			RawMessage: fmt.Sprintf("error: unexpected acl status code %d: %s", resp.StatusCode, string(body)),
		}
	}
	defer resp.Body.Close()

	return resp, nil
}

// Delete an acl configuration
func (c *ConfigAcl) DeleteAnAclConfiguration(AclIndexName int, TransactionID string, parentName string, parentType string) (*http.Response, error) {
	url := fmt.Sprintf("%s/v2/services/haproxy/configuration/acls/%d?transaction_id=%s&parent_name=%s&parent_type=%s", c.BaseURL, AclIndexName, TransactionID, parentName, parentType)
	headers := map[string]string{
		"Content-Type": "application/json",
	}
	resp, err := utils.HTTPRequest("DELETE", url, nil, headers, c.Username, c.Password)
	if err != nil {
		log.Println("acl Update Error sending request:", err, resp)
		return nil, err
	}

	if resp.StatusCode >= 200 && resp.StatusCode <= 299 {
		log.Println("acl Delete Succesful sending request:", resp)
	} else {
		body, readErr := io.ReadAll(resp.Body)
		fmt.Printf("Debug: Response Body: %s\n", string(body))
		fmt.Printf("Debug: Response readErr: %s\n", readErr)

		if readErr != nil {
			log.Println("Error reading response body:", readErr)
			return nil, fmt.Errorf("error reading response body: %v", readErr)
		}

		fmt.Printf("Debug: Unexpected acl status code %d\n", resp.StatusCode)
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
			RawMessage: fmt.Sprintf("error: unexpected acl status code %d: %s", resp.StatusCode, string(body)),
		}
	}
	defer resp.Body.Close()
	return resp, nil
}
