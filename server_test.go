package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

var testDay = "?date1=year_2016%2fmonth_06%2fday_26&offset=0"

func Test_bbHome(t *testing.T) {
	t.Parallel()
	req, err := http.NewRequest("GET", "http://localhost"+getHTTPPort()+"/bb"+testDay, nil)
	if err != nil {
		t.Fatal(err)
	}

	res := httptest.NewRecorder()
	bbHome(res, req)

	// Verify a list of baseball games for that day is returned
	exp := "/bbStream?url=https%3a%2f%2fmediadownloads.mlb.com%2fmlbam%2fmp4%2f"
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

	exp := "/bbStream?url=https%3a%2f%2fmediadownloads.mlb.com%2fmlbam%2fmp4%2f"
	act := res.Body.String()
	if !strings.Contains(act, exp) {
		t.Fatalf("Expected %s got %s", exp, act)
	}
}

func Test_bbAll(t *testing.T) {
	t.Parallel()
	req, err := http.NewRequest("GET", "http://localhost"+getHTTPPort()+"/bbAll"+testDay, nil)
	if err != nil {
		t.Fatal(err)
	}

	res := httptest.NewRecorder()
	bbAll(res, req)

	// Verify a list of baseball games for that day is returned
	exp := "video_list = [\"https://mediadownloads.mlb.com/mlbam/mp4/"
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

	// Verify a redirect is returned
	location := res.Header()["Location"][0]
	if location != URL {
		t.Fatalf("Expected %s got %s", URL, location)
	}
}

func Test_bbStream_normal(t *testing.T) {
	t.Parallel()

	// normal case
	MLB := "https%3a%2f%2fmediadownloads.mlb.com%2fmlbam%2fmp4%2f2016%2f06%2f24%2f849350983%2f1466728732779%2fasset_2500K.mp4"
	req, err := http.NewRequest("GET", "http://localhost"+getHTTPPort()+"/bbStream?url="+MLB, nil)
	if err != nil {
		t.Fatal(err)
	}

	res := httptest.NewRecorder()
	bbStream(res, req)

	// Verify a MLB game is played
	exp := "<source src=\"https://mediadownloads.mlb.com/mlbam/mp4/2016/06/24/849350983/1466728732779/asset_2500K.mp4\" type=\"video/mp4\">"
	act := res.Body.String()
	if !strings.Contains(act, exp) {
		t.Fatalf("Expected %s got %s", exp, act)
	}
}
