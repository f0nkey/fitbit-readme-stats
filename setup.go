package main

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// Config holds generated fields when a new app is made at https://dev.fitbit.com/.
type Config struct {
	OAuthClientID    string `json:"oauth_client_id"`
	ClientSecret     string `json:"client_secret"`
	DisplayGetSource bool   `json:"display_get_source"`
}

// UserCredentials holds credentials to authenticate with and request from the FitBit Web API.
type UserCredentials struct {
	// APIToken is used to grab heart rate data.
	APIToken string `json:"access_token"`
	// RefreshToken is used to grab a new APIToken when it expires.
	RefreshToken string `json:"refresh_token"`
	// Scope denotes what data the user is giving us access to.
	Scope string `json:"scope"`
	// UserID is the FitBit user ID.
	UserID string `json:"user_id"`
}

// APIError holds error info from FitBit's API. NOTE: The success field does not exist on 200 responses.
type APIError struct {
	Errors []struct {
		ErrorType string `json:"errorType"`
		Message   string `json:"message"`
	} `json:"errors"`
	Success bool `json:"success"`
}

// HeartRateTimeSeries contains heartrate-time data from Fitbit's API.
type HeartRateTimeSeries struct {
	// Irrelevant struct omitted
	// ActivitiesHeart []struct {}

	// ActivitiesHeartIntraday has minute-by-minute coverage of a user's heart-rate.
	ActivitiesHeartIntraday struct {
		Dataset         []Datapoint `json:"dataset"`
		DatasetInterval int         `json:"datasetInterval"`
		DatasetType     string      `json:"datasetType"`
	} `json:"activities-heart-intraday"`
}

// Dataset holds the heart bpm at a current time in the format provided by FitBit.
type Datapoint struct {
	Time     string    `json:"time"`
	DateTime time.Time // set by us since fitbit only gives hh:mm, not date in Time
	Value    int       `json:"value"`
}

// setupSuccessMsgs provides success messages after the user completes setup.
func setupSuccessMsgs() (plaintext, html string) {
	localLink := "http://localhost:8090/stats.svg"
	plaintext = "Setup complete. Run the binary with the config files on your host.\nUse the following embed on GitHub: ![FitBit Heart Rate Chart](http://HOSTIP:8090/stats.svg)\nView it now at " + localLink
	html = "Setup complete. Run the binary with the config files on your host.\nUse the following embed on GitHub: <code style='color:red'>![FitBit Heart Rate Chart](http://HOSTIP:8090/stats.svg)</code>\nView it now at "
	html = fmt.Sprintf(`<html><body><p>%s<a href="%s">%s</a></p><img src="%s"></body></html>`, html, localLink, localLink, localLink)
	return plaintext, html
}

// heartRateTimesSeries returns the heart rate time series from the past four hours in a plottable format by Banner.GenSVG.
func heartRateTimesSeries(userCreds UserCredentials, appCredentials Config) ([]BannerXY, error) {
	hrts, err := rawHeartRateTimeSeries(userCreds)
	if err != nil {
		if err.Error() == "token must be refreshed" {
			userCreds, err = reqUserCredentials(appCredentials, "", userCreds.RefreshToken)
			if err != nil {
				return nil, fmt.Errorf("error refreshing tokens and credentials: %w", err)
			}
			err = writeUserCredsFile(userCreds)
			if err != nil {
				return nil, fmt.Errorf("error writing user credentials: %w", err)
			}
			hrts, err = rawHeartRateTimeSeries(userCreds)
			if err != nil {
				return nil, fmt.Errorf("error grabbing heartrate data after token refresh: %w", err)
			}
		} else {
			return nil, fmt.Errorf("error grabbing heartrate data: %w", err)
		}
	}

	xy := make([]BannerXY, 0, len(hrts.ActivitiesHeartIntraday.Dataset))
	for _, pt := range hrts.ActivitiesHeartIntraday.Dataset {
		xy = append(xy, BannerXY{
			X: pt.DateTime,
			Y: pt.Value,
		})
	}
	return xy, nil
}

// reqUserCredentials requests from FitBit the fields in the UserCredentials struct.
// userAuthCode is required for normal requests.
// refreshToken is required for refresh requests.
// refreshToken is empty for normal requests.
// userAuthCode is empty for refresh requests.
func reqUserCredentials(appCred Config, userAuthCode string, refreshToken string) (UserCredentials, error) {
	vals := url.Values{}
	vals.Add("clientId", appCred.OAuthClientID)
	vals.Add("grant_type", "authorization_code")
	if refreshToken != "" {
		vals.Set("grant_type", "refresh_token")
		vals.Set("refresh_token", refreshToken)
	}
	vals.Add("redirect_uri", "http://localhost:8090")
	vals.Add("code", userAuthCode)
	r := strings.NewReader(vals.Encode())
	req, err := http.NewRequest("POST", "https://api.fitbit.com/oauth2/token", r)
	if err != nil {
		return UserCredentials{}, err
	}
	authHeader := "Basic " + base64.StdEncoding.EncodeToString([]byte(appCred.OAuthClientID+":"+appCred.ClientSecret))
	req.Header.Add("Authorization", authHeader)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	c := &http.Client{}
	resp, err := c.Do(req)
	if err != nil {
		return UserCredentials{}, err
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return UserCredentials{}, err
	}

	if resp.StatusCode != 200 {
		aErr := APIError{}
		err = json.Unmarshal(b, &aErr)
		if err != nil {
			return UserCredentials{}, err
		}

		if !aErr.Success {
			errArr := make([]string, 0, len(aErr.Errors))
			for _, s := range aErr.Errors {
				errArr = append(errArr, s.Message)
			}
			return UserCredentials{}, errors.New(strings.Join(errArr, ", "))
		}
		return UserCredentials{}, errors.New(resp.Status + " - " + string(b))
	}

	creds := UserCredentials{}
	err = json.Unmarshal(b, &creds)
	if err != nil {
		return UserCredentials{}, err
	}
	if !strings.Contains(creds.Scope, "heartrate") {
		return UserCredentials{}, errors.New("heartrate was not given as a scope permission")
	}
	if creds.APIToken == "" {
		return UserCredentials{}, errors.New("api token empty")
	}
	if creds.RefreshToken == "" {
		return UserCredentials{}, errors.New("refresh token is empty")
	}

	return creds, nil
}

// dateHour returns a time.Time as YYYY-MM-DD and HH.
func dateHour(t time.Time) (date, hour string) {
	min := prependZero(t.Minute())
	hrStr := strconv.Itoa(t.Hour()) + ":" + min
	mo := t.Month()
	moStr := prependZero(int(mo))
	day := t.Day()
	dayStr := prependZero(day)
	return strconv.Itoa(t.Year()) + "-" + moStr + "-" + dayStr, hrStr
}

func prependZero(i int) string {
	str := strconv.Itoa(i)
	if i < 10 && str[0] != '0' {
		return "0" + str
	}
	return str
}

// rawHeartRateTimeSeries returns heartrate-time data from FitBit.
func rawHeartRateTimeSeries(userCreds UserCredentials) (HeartRateTimeSeries, error) {
	u := `https://api.fitbit.com/1/user/%s/activities/heart/date/%s/%s/1min/time/%s/%s.json`
	// todo: query Get Profile endpoint to offset timezone by their UTC offset: GET https://api.fitbit.com/1/user/[user-id]/profile.json
	var hourRange int = 4
	tRange := time.Hour * time.Duration(hourRange)
	endDate, endHr := dateHour(time.Now())
	startDate, startHr := dateHour(time.Now().Add(-tRange))
	uri := fmt.Sprintf(u, userCreds.UserID, startDate, endDate, startHr, endHr)

	r, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		return HeartRateTimeSeries{}, err
	}
	r.Header.Add("Authorization", "Bearer "+userCreds.APIToken)
	c := http.Client{}
	resp, err := c.Do(r)
	if err != nil {
		return HeartRateTimeSeries{}, err
	}
	if resp.StatusCode == 401 {
		return HeartRateTimeSeries{}, errors.New("token must be refreshed")
	}
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return HeartRateTimeSeries{}, err
	}
	ts := HeartRateTimeSeries{}
	err = json.Unmarshal(b, &ts)
	if err != nil {
		return HeartRateTimeSeries{}, err
	}

	dataset := make([]Datapoint, 0, len(ts.ActivitiesHeartIntraday.Dataset))
	for _, entry := range ts.ActivitiesHeartIntraday.Dataset {
		sp := strings.Split(entry.Time, ":")
		hr, _ := strconv.Atoi(sp[0])
		min, _ := strconv.Atoi(sp[1])

		today := time.Now()
		yesterday := time.Now().Add(time.Hour * -24)
		actualDay := today
		if startDate != endDate { // determining actual date since fitbit does not include date in Datapoint.Time
			if hr > hourRange {
				actualDay = yesterday
			}
		}
		dataset = append(dataset, Datapoint{
			Time:     entry.Time,
			DateTime: time.Date(actualDay.Year(), actualDay.Month(), actualDay.Day(), hr, min, 0, 0, time.UTC),
			Value:    entry.Value,
		})
	}
	ts.ActivitiesHeartIntraday.Dataset = dataset

	return ts, nil
}

// tokensLink returns the link used to authorize us access to the user's data.
func tokensLink(oauthClientID string) string {
	return fmt.Sprintf("https://www.fitbit.com/oauth2/authorize?response_type=code&client_id=%s&redirect_uri=http://localhost:8090&scope=heartrate&expires_in=604800", oauthClientID)
}

// genConfigFile writes an empty config file.
func genConfigFile() {
	b, _ := json.MarshalIndent(&Config{
		OAuthClientID:    "",
		ClientSecret:     "",
		DisplayGetSource: false,
	}, "", "	")
	err := ioutil.WriteFile("config.json", b, 0644)
	if err != nil {
		log.Fatal("error creating config.json", err)
	}
}

// genUserCredsFile writes an empty user credentials file.
func genUserCredsFile() {
	b, _ := json.MarshalIndent(&UserCredentials{}, "", "	")
	err := ioutil.WriteFile("credentials.json", b, 0644)
	if err != nil {
		log.Fatal("error creating config.json", err)
	}
}

func readUserCredsFile() (UserCredentials, error) {
	b, err := ioutil.ReadFile("credentials.json")
	if err != nil {
		return UserCredentials{}, err
	}
	tok := UserCredentials{}
	err = json.Unmarshal(b, &tok)
	if err != nil {
		log.Fatal("error parsing config.json", err)
	}
	return tok, nil
}

func writeUserCredsFile(uc UserCredentials) error {
	b, _ := json.MarshalIndent(&uc, "", "	")
	err := ioutil.WriteFile("credentials.json", b, 0644)
	if err != nil {
		log.Fatal("error creating credentials.json", err)
	}
	return nil
}

func readConfigFile() (Config, error) {
	b, err := ioutil.ReadFile("config.json")
	if err != nil {
		return Config{}, err
	}
	conf := Config{}
	err = json.Unmarshal(b, &conf)
	if err != nil {
		log.Fatal("error parsing config.json", err)
	}
	return conf, err
}

func validateConfig(c Config) error {
	if c.ClientSecret == "" {
		return errors.New("client secret in config.json empty")
	}
	if c.OAuthClientID == "" {
		return errors.New("oauth client id in config.json empty")
	}
	return nil
}
