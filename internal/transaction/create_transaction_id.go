package transaction

import (
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"terraform-provider-haproxy/internal/utils"
)

func (c *ConfigTransaction) createTransactionID(version int) (string, error) {

	versionStr := strconv.Itoa(version)

	url := fmt.Sprintf("%s/v2/services/haproxy/transactions?version=%s", c.BaseURL, versionStr)
	headers := map[string]string{
		"Content-Type": "application/json",
	}
	resp, err := utils.HTTPRequest("POST", url, nil, headers, c.Username, c.Password)
	if err != nil {
		return "", fmt.Errorf("error sending request: %v", err)
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading response body: %v", err)
	}

	var responseData TransactionResponse
	if err := json.Unmarshal(body, &responseData); err == nil {
		return responseData.ID, nil
	}

	var apiError utils.APIError
	if jsonErr := json.Unmarshal(body, &apiError); jsonErr == nil {
		return "", &utils.CustomError{
			StatusCode: resp.StatusCode,
			APIError:   &apiError,
		}
	}

	return "", &utils.CustomError{
		StatusCode: resp.StatusCode,
		RawMessage: fmt.Sprintf("error: unexpected transaction commit status code %d: %s", resp.StatusCode, string(body)),
	}
}

func (m *Manager) CreateTransactionID(version int) (string, error) {
	versionStr := strconv.Itoa(version)

	url := fmt.Sprintf("%s/v2/services/haproxy/transactions?version=%s", m.ConfigTransaction.BaseURL, versionStr)
	headers := map[string]string{
		"Content-Type": "application/json",
	}
	resp, err := utils.HTTPRequest("POST", url, nil, headers, m.ConfigTransaction.Username, m.ConfigTransaction.Password)
	if err != nil {
		return "", fmt.Errorf("error sending request: %v", err)
	}

	defer resp.Body.Close()

	var responseData TransactionResponse
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading response body: %v", err)
	}

	err = json.Unmarshal(body, &responseData)
	if err != nil {
		return "", fmt.Errorf("error decoding response JSON: %v", err)
	}

	return responseData.ID, nil
}
