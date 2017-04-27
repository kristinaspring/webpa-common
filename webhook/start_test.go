package webhook

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"testing"
	"time"
	"net/http"
	"net/http/httptest"
	"net/url"
)

type testTransport struct {
	Transport http.RoundTripper
	URL       *url.URL
	Body      string
}

func(tt testTransport) RoundTrip(r *http.Request) (resp *http.Response, err error) {
	r.URL = tt.URL

	resp = &http.Response{
		Header:     make(http.Header),
		Request:    r,
		StatusCode: http.StatusOK,
		Status:     "200 OK",
		Body:       ioutil.NopCloser(bytes.NewBufferString(tt.Body)),
	}
	
	return
}

func testClient(t *testing.T, msg string) http.Client {
	h := func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "get auth response: %s\n", r.URL.String())
	}
	
	s := httptest.NewServer(http.HandlerFunc(h))
	defer s.Close()
	u, err := url.Parse(s.URL)
	if err != nil {
		t.Errorf("unable to parse server url: %v", err)
	}
	
	tt := *new(testTransport)
	tt.URL = u
	tt.Body = msg
	
	return http.Client{
		Transport: tt,
	}
}

func TestGetAuthorization(t *testing.T) {
	sc := NewStartFactory(nil)
	sc.client = testClient(t, "{\"expires_in\": 0, \"serviceAccessToken\": \"Test Token Value\"}")
	
	err := sc.getAuthorization()
	if err != nil {
		t.Errorf("error returned while obtaining authorization: %v", err)
	}
	if sc.Sat.Token == "" {
		t.Error("unable to obtain current hooks")
	}
}

func TestGetPayload(t *testing.T) {
	sc := NewStartFactory(nil)
	sc.client = testClient(t, "What's in the box!")
	resp, err := sc.makeRequest()
	
	body, err := getPayload(resp)
	if err != nil {
		t.Errorf("error return while obtaining payload from response: %v", err)
	}
	if body == nil {
		t.Error("bad payload returned")
	}
}

func TestMakeRequest(t *testing.T) {
	sc := NewStartFactory(nil)
	sc.client = testClient(t, "Making Requests")
	
	resp, err := sc.makeRequest()
	
	if err != nil {
		t.Errorf("error return while performing request: %v", err)
	}
	if resp == nil {
		t.Error("response returned was nil")
	}
}

func TestGetCurrentSystemsHooks(t *testing.T) {
	rc := make(chan Result, 1)

	sc := NewStartFactory(nil)
	sc.client = testClient(t, "{\"expires_in\": 0, \"serviceAccessToken\": \"Test Token Value\"}")
	sc.getAuthorization()

	d := (time.Duration(5)*time.Second).Nanoseconds()
	u := time.Now().Format(time.RFC3339)
	h := fmt.Sprintf(`[
		{
			"config": {
				"url": "http://127.0.0.1/foo",
				"content_type": "json",
				"secret": "icankeepasecret"
			},
			"failure_url": "",
			"events": [
				"myeventtype*"
			],
			"matcher": {
				"device_id": [
					".*"
				]
			},
			"duration": %v,
			"until": "%v",
			"registered_from_address": "127.0.0.2"
		},
		{
			"config": {
				"url": "http://127.0.0.1/boo",
				"content_type": "json",
				"secret": "iforgotthesecret"
			},
			"failure_url": "",
			"events": [
				"yourevent"
			],
			"matcher": {
				"device_id": [
					".*"
				]
			},
			"duration": %v,
			"until": "%v",
			"registered_from_address": "127.0.0.2"
		}
	]`, d, u, d, u)

	sc.client = testClient(t, h)
	go sc.GetCurrentSystemsHooks(rc)
	
	r := <- rc
	if r.err != nil {
		t.Errorf("error returned retrieving system hooks: %v", r.err)
	}
	if r.hooks == nil {
		t.Error("hooks returned was nil")
	}
	
	// test timeout
	h = ``
	sc.client = testClient(t, h)
	go sc.GetCurrentSystemsHooks(rc)
	
	r = <- rc
	if r.err.Error() != "Unable to obtain hook list in allotted time." {
		t.Errorf("test was expected to fail with error \"Unable to obtain hook list in allotted time.\".  got: %v", r.err)
	}
	if r.hooks != nil {
		t.Errorf("expected hooks returned to be nil.  got %v", r.hooks)
	}
}
