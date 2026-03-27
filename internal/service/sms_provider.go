package service

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

type TwilioConfig struct {
	AccountSID string
	AuthToken  string
	FromNumber string
	APIURL     string
}

type TwilioSMSProvider struct {
	config TwilioConfig
}

func NewTwilioSMSProvider(config TwilioConfig) *TwilioSMSProvider {
	apiURL := config.APIURL
	if apiURL == "" {
		apiURL = "https://api.twilio.com/2010-04-01"
	}
	if config.AccountSID != "" && apiURL == "https://api.twilio.com/2010-04-01" {
		apiURL = fmt.Sprintf("https://api.twilio.com/2010-04-01/Accounts/%s", config.AccountSID)
	}
	return &TwilioSMSProvider{config: TwilioConfig{
		AccountSID: config.AccountSID,
		AuthToken:  config.AuthToken,
		FromNumber: config.FromNumber,
		APIURL:     apiURL,
	}}
}

func (p *TwilioSMSProvider) SendSMS(to, message string) error {
	if p.config.AccountSID == "" || p.config.AuthToken == "" {
		return fmt.Errorf("Twilio credentials not configured")
	}

	if p.config.FromNumber == "" {
		return fmt.Errorf("Twilio from number not configured")
	}

	msgData := url.Values{}
	msgData.Set("To", to)
	msgData.Set("From", p.config.FromNumber)
	msgData.Set("Body", message)

	_, err := requestTwilio("POST", p.config.APIURL+"/Messages.json",
		p.config.AccountSID, p.config.AuthToken, msgData)

	return err
}

func requestTwilio(method, urlStr, accountSID, authToken string, data url.Values) ([]byte, error) {
	client := &http.Client{}

	req, err := http.NewRequest(method, urlStr, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}

	req.SetBasicAuth(accountSID, authToken)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Twilio API error: %s", string(body))
	}

	return nil, nil
}
