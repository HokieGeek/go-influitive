package influitive

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
)

const baseURL = "https://api.influitive.com"

type Client struct {
	Token string `json:"token"`
	OrgID string `json:"orgId"`
}

func NewClient(token, orgID string) (Client, error) {
	return Client{token, orgID}, nil
}

type Member struct {
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
	// SalesforceID   interface{} `json:"salesforce_id"`
	Level  Level  `json:"level"`
	Source string `json:"source"`
	Thumb  string `json:"thumb"`
}

type Level struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type contactsResponse struct {
	Links   Links    `json:"links"`
	Members []Member `json:"contacts"`
}

type Links struct {
	Self string `json:"self"`
	Next string `json:"next"`
}

func httpDo(client Client, method, url string, payload io.Reader) (*http.Response, error) {
	req, err := http.NewRequest(method, url, payload)
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
func QueryMembersByField(client Client, field, value string) ([]Member, error) {
	members := make([]Member, 0)

	// TODO: better error handling

	next := fmt.Sprintf("%s/contacts", baseURL)
	qp := url.Values{}
	if len(field) > 0 {
		qp.Set(fmt.Sprintf("q[%s]", field), value)
		next += "?" + qp.Encode()
	}

	for {
		if len(next) == 0 {
			break
		}

		resp, err := httpDo(client, http.MethodGet, next, nil)
		if err != nil {
			return nil, fmt.Errorf("unable to retrieve list of members: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			if body, err := ioutil.ReadAll(resp.Body); err == nil {
				fmt.Println(string(body))
			}
			return nil, fmt.Errorf("influitive did not return a good response status: %s", resp.Status)
		}

		var contactsResp contactsResponse
		if err := json.NewDecoder(resp.Body).Decode(&contactsResp); err != nil {
			return nil, fmt.Errorf("unable to read message body as members: %v", err)
		}

		members = append(members, contactsResp.Members...)
		next = contactsResp.Links.Next
	}

	return members, nil
}

func GetAllMembers(client Client) ([]Member, error) {
	return QueryMembersByField(client, "", "")
}

// https://influitive.readme.io/reference#get-information-about-your-own-member-record
func GetMe(client Client) (Member, error) {
	resp, err := httpDo(client, http.MethodGet, fmt.Sprintf("%s/members/me", baseURL), nil)
	if err != nil {
		return Member{}, fmt.Errorf("unable to retrieve details of logged in user: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		if body, err := ioutil.ReadAll(resp.Body); err == nil {
			fmt.Println(string(body))
		}
		return Member{}, fmt.Errorf("influitive did not return good status: %s", resp.Status)
	}

	var member Member
	if err := json.NewDecoder(resp.Body).Decode(&member); err != nil {
		return Member{}, fmt.Errorf("unable to read message body as member details: %v", err)
	}

	return member, nil
}

type eventResponse struct {
	ID            int64      `json:"id"`
	EventTypeCode string     `json:"event_type_code"`
	Points        int64      `json:"points"`
	Member        Member     `json:"contact"`
	Parameters    Parameters `json:"parameters"`
}

type Parameters struct {
}

// https://influitive.readme.io/reference#events
func LogEvent(client Client, eventType, memberID string) error {
	resp, err := httpDo(client, http.MethodPost, fmt.Sprintf("%s/events", baseURL), nil)
	if err != nil {
		return fmt.Errorf("unable to retrieve details of logged in user: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		if body, err := ioutil.ReadAll(resp.Body); err == nil {
			fmt.Println(string(body))
		}
		return fmt.Errorf("influitive did not return good status: %s", resp.Status)
	}

	var evResp eventResponse
	if err := json.NewDecoder(resp.Body).Decode(&evResp); err != nil {
		return fmt.Errorf("unable to read message body as member details: %v", err)
	}

	return nil

}
