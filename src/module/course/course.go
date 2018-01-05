package course

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

func IsEnrolled(userID, scheduleID int64) bool {

	data := url.Values{}
	data.Set("user_id", strconv.FormatInt(userID, 10))
	data.Set("role", "student")
	data.Set("schedule_id", strconv.FormatInt(scheduleID, 10))

	params := data.Encode()
	req, err := http.NewRequest("POST", "http://localhost:9001/api/internal/v1/course/getone", strings.NewReader(params))
	if err != nil {
		return false
	}
	req.Header.Add("Authorization", "abc")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Content-Length", strconv.Itoa(len(params)))

	client := http.Client{
		Timeout: time.Second * 2,
	}

	resp, err := client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return false
	}

	res := GetOneHTTPResponse{}
	err = json.Unmarshal(body, &res)
	if err != nil {
		return false
	}

	if !res.Data.Involved {
		return false
	}

	return true
}

func IsAssistant(userID, scheduleID int64) bool {

	data := url.Values{}
	data.Set("user_id", strconv.FormatInt(userID, 10))
	data.Set("role", "assistant")
	data.Set("schedule_id", strconv.FormatInt(scheduleID, 10))

	params := data.Encode()
	req, err := http.NewRequest("POST", "http://localhost:9001/api/internal/v1/course/getone", strings.NewReader(params))
	if err != nil {
		return false
	}
	req.Header.Add("Authorization", "abc")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Content-Length", strconv.Itoa(len(params)))

	client := http.Client{
		Timeout: time.Second * 2,
	}

	resp, err := client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return false
	}

	res := GetOneHTTPResponse{}
	err = json.Unmarshal(body, &res)
	if err != nil {
		return false
	}
	fmt.Println(res)
	if !res.Data.Involved {
		return false
	}

	return true
}

func SelectEnrolledSchedule(userID int64) ([]CourseSchedule, error) {

	var res []CourseSchedule

	data := url.Values{}
	data.Set("user_id", strconv.FormatInt(userID, 10))
	data.Set("role", "student")

	params := data.Encode()
	req, err := http.NewRequest("POST", "http://localhost:9001/api/internal/v1/course/getall", strings.NewReader(params))
	if err != nil {
		return res, err
	}
	req.Header.Add("Authorization", "abc")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Content-Length", strconv.Itoa(len(params)))

	client := http.Client{
		Timeout: time.Second * 2,
	}

	resp, err := client.Do(req)
	if err != nil {
		return res, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return res, err
	}

	jsn := GetAllHTTPResponse{}
	err = json.Unmarshal(body, &jsn)
	if err != nil {
		return res, err
	}

	if jsn.Code != 200 {
		return res, fmt.Errorf("error request")
	}

	return jsn.Data, nil
}

func SelectTeachedSchedule(userID int64) ([]CourseSchedule, error) {

	var res []CourseSchedule

	data := url.Values{}
	data.Set("user_id", strconv.FormatInt(userID, 10))
	data.Set("role", "assistant")

	params := data.Encode()
	req, err := http.NewRequest("POST", "http://localhost:9001/api/internal/v1/course/getall", strings.NewReader(params))
	if err != nil {
		return res, err
	}
	req.Header.Add("Authorization", "abc")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Content-Length", strconv.Itoa(len(params)))

	client := http.Client{
		Timeout: time.Second * 2,
	}

	resp, err := client.Do(req)
	if err != nil {
		return res, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return res, err
	}

	jsn := GetAllHTTPResponse{}
	err = json.Unmarshal(body, &jsn)
	if err != nil {
		return res, err
	}

	if jsn.Code != 200 {
		return res, fmt.Errorf("error request")
	}

	return jsn.Data, nil
}

func Get(userID, scheduleID int64, isAssistant bool) (GetOne, error) {

	var getOne GetOne

	role := "student"
	if isAssistant {
		role = "assistant"
	}

	data := url.Values{}
	data.Set("user_id", strconv.FormatInt(userID, 10))
	data.Set("role", role)
	data.Set("schedule_id", strconv.FormatInt(scheduleID, 10))

	params := data.Encode()
	req, err := http.NewRequest("POST", "http://localhost:9001/api/internal/v1/course/getone", strings.NewReader(params))
	if err != nil {
		return getOne, err
	}
	req.Header.Add("Authorization", "abc")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Content-Length", strconv.Itoa(len(params)))

	client := http.Client{
		Timeout: time.Second * 2,
	}

	resp, err := client.Do(req)
	if err != nil {
		return getOne, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return getOne, err
	}

	res := GetOneHTTPResponse{}
	err = json.Unmarshal(body, &res)
	if err != nil {
		return getOne, err
	}

	return res.Data, nil
}
