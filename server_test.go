package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

var testDay = "?date1=year_2021%2fmonth_06%2fday_24&offset=0"
var testCatchAllDay = "?date1=06%2f24%2f2021"

func init() {
	parseTemplates()
}

func Test_bbHome(t *testing.T) {
	t.Parallel()
	req, err := http.NewRequest("GET", "http://localhost"+getHTTPPort()+"/bb"+testDay, nil)
	if err != nil {
		t.Fatal(err)
	}

	res := httptest.NewRecorder()
	bbHome(res, req)

	// Verify a list of baseball games for that day is returned
	exp := "/bbStream?url=%2fapi%2fv1%2fgame%2f"

	act := res.Body.String()
	if !strings.Contains(act, exp) {
		t.Fatalf("Expected %s got %s", exp, act)
	}
}

func Test_bbAjaxDay(t *testing.T) {
	t.Parallel()
	req, err := http.NewRequest("GET", "http://localhost"+getHTTPPort()+"/bbAjaxDay"+testDay, nil)
	if err != nil {
		t.Fatal(err)
	}

	res := httptest.NewRecorder()
	bbAjaxDay(res, req)

	exp := "/bbStream?url=%2fapi%2fv1%2fgame%2f"
	act := res.Body.String()
	if !strings.Contains(act, exp) {
		t.Fatalf("Expected %s got %s", exp, act)
	}
}

func Test_bbAll(t *testing.T) {
	t.Parallel()

	// time for GottaCatchEmAll isn't formatted how we expect
	req, err := http.NewRequest("GET", "http://localhost"+getHTTPPort()+"/bbAll"+testCatchAllDay, nil)
	if err != nil {
		t.Fatal(err)
	}

	res := httptest.NewRecorder()
	bbAll(res, req)

	// Verify a list of baseball games for that day is returned
	exp := "video_list = [\"/api/v1/game"
	act := res.Body.String()
	if !strings.Contains(act, exp) {
		t.Fatalf("Expected %s got %s", exp, act)
	}
}

func Test_bbStream_redirect(t *testing.T) {
	t.Parallel()

	// redirect case
	URL := "https://www.youtube.com/user/MLB"
	req, err := http.NewRequest("GET", "http://localhost"+getHTTPPort()+"/bbStream?url="+URL, nil)
	if err != nil {
		t.Fatal(err)
	}

	res := httptest.NewRecorder()
	bbStream(res, req)

	// Verify redirect and BadRequest is returned
	location := res.Header()["Location"][0]
	if res.Result().StatusCode != http.StatusBadRequest {
		t.Fatalf("Expected %d got %d", http.StatusBadRequest, res.Result().StatusCode)
	}
	if location != URL {
		t.Fatalf("Expected %s got %s", URL, location)
	}
}

func Test_bbStream_normal(t *testing.T) {
	t.Parallel()

	// normal case
	MLB := "%2fapi%2fv1%2fgame%2f661182%2fcontent"
	req, err := http.NewRequest("GET", "http://localhost"+getHTTPPort()+"/bbStream?url="+MLB, nil)
	if err != nil {
		t.Fatal(err)
	}

	res := httptest.NewRecorder()
	bbStream(res, req)

	// Verify a MLB game is played
	exp := "<source src=\"https://mlb-cuts-diamond.mlb.com/FORGE/2022/2022-08/09/7336868f-a25c598f-946a8f60-csvm-diamondx64-asset_1280x720_59_4000K.mp4\" type=\"video/mp4\">"
	act := res.Body.String()
	if !strings.Contains(act, exp) {
		t.Fatalf("Expected %s got %s", exp, act)
	}
}
