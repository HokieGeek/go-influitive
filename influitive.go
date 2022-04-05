package influitive

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
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
	ID              int64             `json:"id"`
	Name            string            `json:"name"`
	FirstName       string            `json:"first_name"`
	LastName        string            `json:"last_name"`
	Email           string            `json:"email"`
	Title           string            `json:"title"`
	Company         string            `json:"company"`
	UUID            string            `json:"uuid"`
	Type            string            `json:"type"`
	CreatedAt       string            `json:"created_at"`
	JoinedAt        string            `json:"joined_at"`
	LockedAt        string            `json:"locked_at"`
	ExternalIDS     map[string]string `json:"external_ids"`
	MatchCategories map[string]string `json:"match_categories"`
	CustomFields    map[string]string `json:"custom_fields"`
	NpsScore        int64             `json:"nps_score"`
	CurrentPoints   int64             `json:"current_points"`
	LifetimePoints  int64             `json:"lifetime_points"`
	CRMContactID    string            `json:"crm_contact_id"`
	SalesforceID    string            `json:"salesforce_id"`
	InviteLink      string            `json:"invite_link"`
	Language        string            `json:"language"`
	Address         string            `json:"address"`
	Level           Level             `json:"level"`
	Source          string            `json:"source"`
	Thumb           string            `json:"thumb"`
}

type Level struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type contactsResponse struct {
	Links   links    `json:"links"`
	Members []Member `json:"contacts"`
}

type links struct {
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
	req.Header.Set("Content-Type", "application/json")

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
		qp.Set("q[status]", "all")
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

/*
type eventMember struct {
	ID           string `json:"id"`
	Email        string `json:"email"`
	CRMContactID string `json:"crm_contact_id"`
	FirstName    string `json:"first_name"`
	LastName     string `json:"last_name"`
}
*/

type logEventRequest struct {
	Type   string `json:"type"`
	Member Member `json:"member"`
	Notes  string `json:"notes"`
	Link   string `json:"link"`
	Points string `json:"points"`
}

type logCustomEventRequest struct {
	Type    string  `json:"type"`
	Points  string  `json:"points"`
	Contact contact `json:"contact"`
	Stage   stage   `json:"stage"`
}

type contact struct {
	ID    string `json:"id"`
	Email string `json:"email"`
}

type stage struct {
	Code string `json:"code"`
}

type logCustomEventResponse struct {
	ID            int64      `json:"id"`
	EventTypeCode string     `json:"event_type_code"`
	Points        int64      `json:"points"`
	Member        Member     `json:"contact"`
	Parameters    parameters `json:"parameters"`
}

type parameters struct {
}

const (
	EventReferralSubmitted = "referral_submitted"
)

// https://influitive.readme.io/reference#post-reference-type-events
func logEvent(client Client, eventType string, memberID, points int64) error {
	req := logEventRequest{
		Type:   eventType,
		Member: Member{ID: memberID},
		Points: strconv.FormatInt(points, 10),
	}
	buf, err := json.Marshal(req)
	if err != nil {
		return err
	}

	resp, err := httpDo(client, http.MethodPost, fmt.Sprintf("%s/references/events", baseURL), bytes.NewBuffer(buf))
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

	var evResp logCustomEventResponse
	if err := json.NewDecoder(resp.Body).Decode(&evResp); err != nil {
		return fmt.Errorf("unable to read message body as member details: %v", err)
	}

	return nil

}

func LogEvent(client Client, eventType string, memberID, points int64) error {
	return logEvent(client, eventType, memberID, points)
}

// https://influitive.readme.io/reference#events
func logCustomEvent(client Client, eventType, challengeCode string, memberID, points int64) error {
	req := logCustomEventRequest{
		Type:    eventType,
		Contact: contact{ID: strconv.FormatInt(memberID, 10)},
		Points:  strconv.FormatInt(points, 10),
		Stage:   stage{Code: challengeCode},
	}
	buf, err := json.Marshal(req)
	if err != nil {
		return err
	}

	resp, err := httpDo(client, http.MethodPost, fmt.Sprintf("%s/events", baseURL), bytes.NewBuffer(buf))
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

	var evResp logCustomEventResponse
	if err := json.NewDecoder(resp.Body).Decode(&evResp); err != nil {
		return fmt.Errorf("unable to read message body as member details: %v", err)
	}

	return nil

}

func LogCustomEvent(client Client, eventType string, memberID, points int64) error {
	return logCustomEvent(client, eventType, "", memberID, points)
}

func LogCustomChallengeEvent(client Client, eventType, challengeCode string, memberID, points int64) error {
	return logCustomEvent(client, eventType, challengeCode, memberID, points)
}

type createMemberRequest struct {
	Email         string            `json:"email"`
	Name          string            `json:"name"`
	Source        string            `json:"source"`
	Title         string            `json:"title"`
	Company       string            `json:"company"`
	SalesforceID  string            `json:"salesforce_id"`
	MatchCriteria map[string]string `json:"match_criteria"`
	Type          string            `json:"type"`
}

// https://influitive.readme.io/reference#create-a-member-identified-by-email
func CreateMemberByEmail(client Client, email, name, source string) (Member, error) {
	create := createMemberRequest{Email: email, Name: name, Source: source, Type: "Nominee"}
	buf, err := json.Marshal(create)
	if err != nil {
		return Member{}, err
	}

	resp, err := httpDo(client, http.MethodPost, fmt.Sprintf("%s/members", baseURL), bytes.NewBuffer(buf))
	if err != nil {
		return Member{}, fmt.Errorf("unable to create member: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
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

type inviteResponse struct {
	Status     string `json:"status"`
	InviteLink string `json:"invite_link"`
}

// https://influitive.readme.io/reference#invite-a-member-identified-by-id
func InviteMember(client Client, id int64, sendEmail bool) error {
	reqBody := fmt.Sprintf(`{"deliver_emails":%v}`, sendEmail)

	resp, err := httpDo(client, http.MethodPost, fmt.Sprintf("%s/members/%d/invitations", baseURL, id), bytes.NewBufferString(reqBody))
	if err != nil {
		return fmt.Errorf("unable to invite member: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		if body, err := ioutil.ReadAll(resp.Body); err == nil {
			fmt.Println(string(body))
		}
		return fmt.Errorf("influitive did not return good status: %s", resp.Status)
	}

	var invResp inviteResponse
	if err := json.NewDecoder(resp.Body).Decode(&invResp); err != nil {
		return fmt.Errorf("unable to read message body as invitation response: %v", err)
	}

	return nil

}

// DeleteMemberByID is not implemented as there is no API to do this
func DeleteMemberByID(client Client, id int64) error {
	return errors.New("NOT IMPLEMENTED")
	/*
		resp, err := httpDo(client, http.MethodDelete, fmt.Sprintf("%s/members/%d", baseURL, id), nil)
		if err != nil {
			return fmt.Errorf("unable to retrieve details of logged in user: %v", err)
		}
		defer resp.Body.Close()

		if body, err := ioutil.ReadAll(resp.Body); err == nil {
			fmt.Println(string(body))
		}

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("influitive did not return good status: %s", resp.Status)
		}

		return nil
	*/
}
