package auth

import (
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/asepnur/meiko_information/src/util/helper"

	"github.com/asepnur/meiko_information/src/webserver/template"
	"github.com/julienschmidt/httprouter"
)

type (
	Config struct {
		SessionKey string `json:"sessionkey"`
	}
)

var (
	c                  Config
	errSessionNotlogin = errors.New("SessionNotLogin")
)

func Init(cfg Config) {
	c = cfg
}

// MustAuthorize you must provide the Bearer token on header if you're using this middleware
func MustAuthorize(h httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		cookie, err := r.Cookie(c.SessionKey)
		if err != nil {
			template.RenderJSONResponse(w, new(template.Response).
				SetCode(http.StatusForbidden).
				AddError("Invalid Session"))
			return
		}

		userData, err := getUserInfo(cookie.Value)
		if err != nil {
			template.RenderJSONResponse(w, new(template.Response).
				SetCode(http.StatusForbidden).
				AddError("Invalid Session"))
			return
		}

		r = r.WithContext(context.WithValue(r.Context(), "User", userData))

		h(w, r, ps)
	}
}

// OptionalAuthorize you don't really have to pass the Bearer token if using this middleware
func OptionalAuthorize(h httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		var userData *User
		cookie, err := r.Cookie(c.SessionKey)
		if err == nil {
			userData, _ = getUserInfo(cookie.Value)
		}

		r = r.WithContext(context.WithValue(r.Context(), "User", userData))
		h(w, r, ps)
	}
}

func getUserInfo(session string) (*User, error) {

	session = strings.Trim(session, " ")
	data := url.Values{}
	data.Set("cookie", session)

	params := data.Encode()
	req, err := http.NewRequest("POST", "http://localhost:9000/api/v1/user/exhange-profile", strings.NewReader(params))
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", "abc")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Content-Length", strconv.Itoa(len(params)))

	client := http.Client{
		Timeout: time.Second * 2,
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, nil
	}

	res := &SessionHTTPResponse{}
	err = json.Unmarshal(body, res)
	if err != nil {
		return nil, err
	}
	if res == nil {
		return nil, errSessionNotlogin
	}

	usr := &res.Data
	return usr, nil
}

func Oauth(h httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		token := r.Header.Get("Authorization")
		if token != "abc" {
			template.RenderJSONResponse(w, new(template.Response).
				SetCode(http.StatusForbidden).
				AddError("Invalid Session"))
			return
		}
		h(w, r, ps)
	}
}

func (u User) IsHasRoles(module string, roles ...string) bool {

	if len(roles) < 1 || len(u.Roles[module]) < 1 {
		return false
	}

	for _, val := range roles {
		if helper.IsStringInSlice(val, u.Roles[module]) {
			return true
		}
	}

	return false
}
