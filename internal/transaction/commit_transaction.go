package transaction

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"terraform-provider-haproxy/internal/utils"
)

// Commit a transaction id
func (c *ConfigTransaction) commitTransactionID(TransactionID string) (*http.Response, error) {
	url := fmt.Sprintf("%s/v2/services/haproxy/transactions/%s", c.BaseURL, TransactionID)

	fmt.Println("----------url----------", url)
	headers := map[string]string{
		"Content-Type": "application/json",
	}
	resp, err := utils.HTTPRequest("PUT", url, nil, headers, c.Username, c.Password)

	if err != nil {
		log.Println("Error sending request:", err)
		return nil, err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %s", err)
	}
	bodyStr := string(body)
	if resp.StatusCode != 200 && resp.StatusCode != 202 {
		log.Println("error commiting transaction", string(bodyStr))

		var apiError utils.APIError
		if jsonErr := json.Unmarshal(body, &apiError); jsonErr == nil {
			return nil, &utils.CustomError{
				StatusCode: resp.StatusCode,
				APIError:   &apiError,
			}
		}
		return nil, &utils.CustomError{
			StatusCode: resp.StatusCode,
			RawMessage: fmt.Sprintf("error: unexpected transaction commit status code %d: %s", resp.StatusCode, string(body)),
		}
	}
	defer resp.Body.Close()
	fmt.Println(string(body))
	return resp, nil
}

func NewManager(TransactionID string, config ConfigTransaction) *Manager {
	return &Manager{
		TransactionID:     TransactionID,
		ConfigTransaction: config,
	}
}

func (m *Manager) CommitTransactionID(TransactionID string) (*http.Response, error) {
	url := fmt.Sprintf("%s/v2/services/haproxy/transactions/%s", m.ConfigTransaction.BaseURL, TransactionID)

	fmt.Println("----------URL----------", url)

	headers := map[string]string{
		"Content-Type": "application/json",
	}
	resp, err := utils.HTTPRequest("PUT", url, nil, headers, m.ConfigTransaction.Username, m.ConfigTransaction.Password)
	if err != nil {
		log.Println("Error sending request:", err)
		return nil, err
	}
	if resp == nil {
		return nil, fmt.Errorf("received nil response")
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			fmt.Println("Error closing response body:", err)
		}
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %s", err)
	}

	bodyStr := string(body)
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		log.Println("Error committing transaction:", bodyStr)
		var apiError utils.APIError
		if jsonErr := json.Unmarshal(body, &apiError); jsonErr == nil {
			return nil, &utils.CustomError{
				StatusCode: resp.StatusCode,
				APIError:   &apiError,
			}
		}
		return nil, &utils.CustomError{
			StatusCode: resp.StatusCode,
			RawMessage: fmt.Sprintf("error: unexpected transaction commit status code %d: %s", resp.StatusCode, string(body)),
		}
	}

	fmt.Println("Transaction commit response:", bodyStr)

	return resp, nil
}
