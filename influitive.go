package influitive

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
)

const baseURL = "https://api.influitive.com"

type Client struct {
	Token string `json:"token"`
	OrgID string `json:"orgId"`
}

func NewClient(token, orgID string) (Client, error) {
	return Client{token, orgID}, nil
}

type Contact struct {
	ID             int64  `json:"id"`
	Name           string `json:"name"`
	FirstName      string `json:"first_name"`
	LastName       string `json:"last_name"`
	Email          string `json:"email"`
	Title          string `json:"title"`
	Company        string `json:"company"`
	UUID           string `json:"uuid"`
	Type           string `json:"type"`
	CreatedAt      string `json:"created_at"`
	JoinedAt       string `json:"joined_at"`
	NpsScore       int64  `json:"nps_score"`
	CurrentPoints  int64  `json:"current_points"`
	LifetimePoints int64  `json:"lifetime_points"`
	CRMContactID   string `json:"crm_contact_id"`
	Level          Level  `json:"level"`
}

type Level struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

func httpDo(client Client, method, endpoint string, payload io.Reader) (*http.Response, error) {
	req, err := http.NewRequest(method, fmt.Sprintf("%s/%s", baseURL, endpoint), payload)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Token %s", client.Token))
	req.Header.Set("X_ORG_ID", client.OrgID)
	req.Header.Set("Accept", "application/json")

	httpClient := &http.Client{}
	return httpClient.Do(req)
}

// https://influitive.readme.io/reference#query-for-contacts-from-your-advocatehub
func GetAllMembers(client Client) ([]Contact, error) {
	resp, err := httpDo(client, http.MethodGet, "/contacts", nil)
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve list of members: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		if body, err := ioutil.ReadAll(resp.Body); err == nil {
			fmt.Println(string(body))
		}
		return nil, fmt.Errorf("influitive did not return good status: %s", resp.Status)
	}

	var contacts []Contact
	if err := json.NewDecoder(resp.Body).Decode(&contacts); err != nil {
		return nil, fmt.Errorf("unable to read message body as members: %v", err)
	}

	return contacts, nil
}

// https://influitive.readme.io/reference#get-information-about-your-own-member-record
func GetMe(client Client) (Contact, error) {
	resp, err := httpDo(client, http.MethodGet, "/members/me", nil)
	if err != nil {
		return Contact{}, fmt.Errorf("unable to retrieve details of logged in user: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		if body, err := ioutil.ReadAll(resp.Body); err == nil {
			fmt.Println(string(body))
		}
		return Contact{}, fmt.Errorf("influitive did not return good status: %s", resp.Status)
	}

	var contact Contact
	if err := json.NewDecoder(resp.Body).Decode(&contact); err != nil {
		return Contact{}, fmt.Errorf("unable to read message body as member details: %v", err)
	}

	return contact, nil
}
