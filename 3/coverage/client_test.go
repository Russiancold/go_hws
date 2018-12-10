package main

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"
)

//invalid requests
func TestSearchClient_FindUsers(t *testing.T) {
	type TestCase struct {
		req      SearchRequest
		response *SearchResponse
		err      SearchErrorResponse
	}

	cases := []TestCase{
		//invalid requests
		TestCase{
			SearchRequest{-1, 0, "name:Hitler", "age", 1},
			&SearchResponse{},
			SearchErrorResponse{"limit must be > 0"},
		},
		TestCase{
			SearchRequest{5, -1, "name:Hitler", "age", 1},
			&SearchResponse{},
			SearchErrorResponse{"offset must be > 0"},
		},
		TestCase{
			SearchRequest{26, 0, "name:Hitler", "age", 1},
			&SearchResponse{[]User{}, false},
			SearchErrorResponse{},
		},
		TestCase{
			SearchRequest{5, 0, "name:Hitler", "age", -2},
			&SearchResponse{},
			SearchErrorResponse{"unknown bad request error: invalid request"},
		},
		TestCase{
			SearchRequest{5, 0, "name:Hitler", "about", -2},
			&SearchResponse{},
			SearchErrorResponse{"OrderField about invalid"},
		},
	}

	ts := httptest.NewServer(http.HandlerFunc(SearchServer))

	for caseNum, item := range cases {
		c := &SearchClient{
			AccessToken: "42",
			URL:         ts.URL,
		}
		result, err := c.FindUsers(item.req)

		if err != nil && err.Error() != item.err.Error {
			t.Errorf("[%d] expected error: %+v , got: %+v", caseNum, item.err.Error, err)
		}
		if err == nil {
			if !reflect.DeepEqual(result, item.response) {
				t.Errorf("[%d] expected result: %+v , got: %+v", caseNum, item.response, result)
			}
		}
	}
	ts.Close()
}

//broken file, timeout and token
func TestSearchClient_FindUsers2(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(SearchServer))
	c := &SearchClient{
		AccessToken: "42",
		URL:         ts.URL,
	}
	req := SearchRequest{5, 0, "name:il", "age", 1}
	_, err := c.FindUsers(req)
	if err != nil {
		t.Error("unexpected error")
	}
	testTimeout = func() { time.Sleep(3 * time.Second) }
	_, err = c.FindUsers(req)
	expectedError := "timeout for limit=6&offset=0&order_by=1&order_field=age&query=name%3Ail"
	if err.Error() != expectedError {
		t.Errorf("expected: %+v, got:  %+v", expectedError, err)
	}
	testTimeout = func() {}
	fileURI = "broken.xml"
	expectedError = "SearchServer fatal error"
	_, err = c.FindUsers(req)
	if err.Error() != expectedError {
		t.Errorf("expected: %+v, got:  %+v", expectedError, err)
	}
	fileURI = "/home/alexandr/Documents/golang-2018-2/3/99_hw/coverage/dataset.xml"
	c.AccessToken = "invalid_token"
	expectedError = "Bad AccessToken"
	_, err = c.FindUsers(req)
	if err.Error() != expectedError {
		t.Errorf("expected: %+v, got:  %+v", expectedError, err)
	}
}

//JSON marshalling and Client.Do
func TestSearchClient_FindUsers3(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(getBrokenBodyJSON))
	c := &SearchClient{
		AccessToken: "42",
		URL:         ts.URL,
	}
	req := SearchRequest{5, 0, "name:il", "age", 1}
	_, err := c.FindUsers(req)
	expectedError := "cant unpack result json: invalid character 'b' looking for beginning of object key string"
	if err.Error() != expectedError {
		t.Errorf("expected: %+v, got:  %+v", expectedError, err)
	}

	ts = httptest.NewServer(http.HandlerFunc(getBrokenErrorJSON))
	c.URL = ts.URL
	expectedError = "cant unpack error json: invalid character 'b' looking for beginning of object key string"
	_, err = c.FindUsers(req)
	if err.Error() != expectedError {
		t.Errorf("expected: %+v, got:  %+v", expectedError, err)
	}

	client.CheckRedirect = func(req *http.Request, reqs []*http.Request) error { return errors.New("Client error") }
	ts = httptest.NewServer(http.HandlerFunc(checkRedirectBroken))
	c.URL = ts.URL
	expectedError = "unknown error Get /fuck_up: Client error"
	_, err = c.FindUsers(req)
	if err.Error() != expectedError {
		t.Errorf("expected: %+v, got:  %+v", expectedError, err)
	}
}

func checkRedirectBroken(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "fuck_up", http.StatusSeeOther)
}

func getBrokenErrorJSON(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusBadRequest)
	w.Write([]byte("{broken: JSON}"))
}

func getBrokenBodyJSON(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("{broken: JSON}"))
}

type Row struct {
	Id        int    `xml:"id"`
	FirstName string `xml:"first_name"`
	LastName  string `xml:"last_name"`
	Age       int    `xml:"age"`
	About     string `xml:"about"`
	Gender    string `xml:"gender"`
}

type Root struct {
	Version string `xml:"version,attr"`
	List    []Row  `xml:"row"`
}

type ResponseUser struct {
	Id     int
	Name   string
	Age    int
	About  string
	Gender string
}

var fileURI = "/home/alexandr/Documents/golang-2018-2/3/99_hw/coverage/dataset.xml"
var testTimeout = func() { return }

func SearchServer(w http.ResponseWriter, r *http.Request) {
	testTimeout()
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	token := r.Header.Get("AccessToken")
	if !validateToken(token) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"Error": "Bad AccessToken"}`))
		return
	}
	params := r.URL.Query()
	legalParams := []string{"limit", "offset", "query", "order_field", "order_by"}
	for _, val := range legalParams {
		if !checkParam(val, params) {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"Error": "invalid request"}`))
			return
		}
	}
	limit, e := strconv.Atoi(params["limit"][0])
	offset, er := strconv.Atoi(params["offset"][0])
	orderField := params["order_field"][0]
	if !checkOrderField(orderField) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"Error": "ErrorBadOrderField"}`))
		return
	}
	orderBy, err := strconv.Atoi(params["order_by"][0])
	// query representation - column:value
	query := strings.Split(params["query"][0], ":")
	if len(query) != 2 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"Error": "invalid query"}`))
		return
	}
	queryColumn, queryValue := query[0], query[1]
	//looks shitty
	if e != nil || er != nil || err != nil || orderBy < -1 || orderBy > 1 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"Error": "invalid request"}`))
		return
	}
	xmlFile, err := os.Open(fileURI)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"Error": "data not found""}`))
		return
	}
	defer xmlFile.Close()
	var users Root
	byteValue, err := ioutil.ReadAll(xmlFile)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"Error": "data can't be read"}`))
		return
	}
	xml.Unmarshal(byteValue, &users)
	responseUsers := castSrvToResponseUser(users)
	responseUsers = findByField(responseUsers, queryColumn, queryValue)
	switch orderBy {
	case 0:
		return
	case -1:
		sortUsers(&responseUsers, orderField)
		//reverse desc
		for i, j := 0, len(responseUsers)-1; i < j; i, j = i+1, j-1 {
			responseUsers[i], responseUsers[j] = responseUsers[j], responseUsers[i]
		}
	case 1:
		sortUsers(&responseUsers, orderField)
	}
	if len(responseUsers) > offset {
		if len(responseUsers) > offset+limit {
			responseUsers = responseUsers[offset : offset+limit]
		} else {
			responseUsers = responseUsers[offset:]
		}
	} else {
		responseUsers = []ResponseUser{}
	}
	jsonUsers, _ := json.Marshal(responseUsers)
	w.Write(jsonUsers)
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
}

func validateToken(token string) bool {
	validTokens := []string{"42", "2018", "foo"}
	for _, val := range validTokens {
		if token == val {
			return true
		}
	}
	return false
}

func findByField(users []ResponseUser, field, value string) []ResponseUser {
	foundUsers := []ResponseUser{}
	switch field {
	case "name":
		for _, val := range users {
			if strings.Contains(strings.ToUpper(val.Name), strings.ToUpper(value)) {
				foundUsers = append(foundUsers, val)
			}
		}
	case "about":
		for _, val := range users {
			if strings.Contains(strings.ToUpper(val.About), strings.ToUpper(value)) {
				foundUsers = append(foundUsers, val)
			}
		}
	default:
		return users
	}
	return foundUsers
}

func castSrvToResponseUser(users Root) []ResponseUser {
	responseUsers := []ResponseUser{}
	for _, user := range users.List {
		responseUsers = append(responseUsers, ResponseUser{
			user.Id, user.FirstName + " " + user.LastName, user.Age, user.About, user.Gender})
	}
	return responseUsers
}

func sortUsers(users *[]ResponseUser, orderField string) {
	var less func(i, j int) bool
	switch orderField {
	case "age":
		less = func(i, j int) bool {
			return (*users)[i].Age < (*users)[j].Age
		}
	case "name":
		less = func(i, j int) bool {
			return (*users)[i].Name < (*users)[j].Name
		}
	case "id":
		less = func(i, j int) bool {
			return (*users)[i].Id < (*users)[j].Id
		}
	default:
		less = func(i, j int) bool {
			return (*users)[i].Name < (*users)[j].Name
		}
	}
	sort.Slice(*users, less)
}

func checkOrderField(orderField string) bool {
	validFields := [6]string{"age", "name", "id"}
	for _, val := range validFields {
		if orderField == val {
			return true
		}
	}
	return false
}

func checkParam(key string, params map[string][]string) bool {
	param, exist := params[key]
	if !exist {
		return false
	}
	if len(param) != 1 {
		return false
	}
	return true
}
