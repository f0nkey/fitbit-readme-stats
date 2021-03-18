package main

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// APIError holds error info from FitBit's API.
// NOTE: The success field does not exist on 200 responses (which you wouldn't see this struct in a 200 response).
type APIError struct {
	Errors []struct {
		ErrorType string `json:"errorType"`
		Message   string `json:"message"`
	} `json:"errors"`
	Success bool `json:"success"`
}

// HeartRateTimeSeries contains heartrate-time data from FitBit's API.
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

// reqInitUserCredentials requests user credentials from FitBit for the first time.
func reqInitUserCredentials(userAuthCode string, config AppCredentials) (UserCredentials, error) {
	if userAuthCode == "" {
		return UserCredentials{}, fmt.Errorf("no user auth code provided")
	}
	userCreds, err := reqUserCredentials(config, userAuthCode, "")
	if err != nil {
		return UserCredentials{}, fmt.Errorf("error grabbing user tokens and credentials: %w", err)
	}
	return userCreds, nil
}

// reqUserCredentials requests from FitBit the fields in the UserCredentials struct.
// If requesting a refresh, userAuthCode must be empty and refreshToken filled out.
// If not requesting a refresh, userAuthCode must be filled and refreshToken empty.
func reqUserCredentials(appCred AppCredentials, userAuthCode string, refreshToken string) (UserCredentials, error) {
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

// heartRateTimesSeries returns the heart rate time series from the past four hours in a plottable format.
// Side Effects: May write to config.json and edit the config argument with a refresh token if token expired.
func heartRateTimesSeries(config *Config) ([]BannerXY, error) {
	hrts, err := rawHeartRateTimeSeries(config.UserCredentials, *config)
	if err != nil {
		if err.Error() == "token must be refreshed" {
			userCreds, err := reqUserCredentials(config.AppCredentials, "", config.UserCredentials.RefreshToken)
			if err != nil {
				return nil, fmt.Errorf("error refreshing tokens and credentials: %w", err)
			}
			config.UserCredentials = userCreds
			err = writeConfigFile(*config)
			if err != nil {
				return nil, fmt.Errorf("error writing to config file after getting refresh token: %w", err)
			}
			hrts, err = rawHeartRateTimeSeries(config.UserCredentials, *config)
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

// dateHourMin returns a time.Time as YYYY-MM-DD and HH.
func dateHourMin(t time.Time) (date, hourMin string) {
	min := prependZero(t.Minute())
	hrStr := prependZero(t.Hour()) + ":" + min
	mo := t.Month()
	moStr := prependZero(int(mo))
	day := t.Day()
	dayStr := prependZero(day)
	return strconv.Itoa(t.Year()) + "-" + moStr + "-" + dayStr, hrStr
}

func prependZero(i int) string {
	str := strconv.Itoa(i)
	if i > 9 {
		return str
	}
	return "0" + str
}

// rawHeartRateTimeSeries returns heartrate-time data from FitBit.
func rawHeartRateTimeSeries(userCreds UserCredentials, config Config) (HeartRateTimeSeries, error) {
	hourRange := config.PlotRange
	tRange := time.Hour * time.Duration(hourRange)
	loc := time.FixedZone("zone", config.Timezone*3600)
	now := time.Now().UTC().In(loc)
	endDate, endHr := dateHourMin(now)
	startDate, startHr := dateHourMin(now.Add(-tRange))
	u := `https://api.fitbit.com/1/user/%s/activities/heart/date/%s/%s/1min/time/%s/%s.json`
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
		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return HeartRateTimeSeries{}, err
		}
		apiErr := APIError{}
		err = json.Unmarshal(b, &apiErr)
		if err != nil {
			return HeartRateTimeSeries{}, err
		}
		if strings.Contains(apiErr.Errors[0].Message, "Access token expired") {
			return HeartRateTimeSeries{}, fmt.Errorf("token must be refreshed")
		}
		return HeartRateTimeSeries{}, fmt.Errorf(apiErr.Errors[0].Message)
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

		today := now
		yesterday := now.Add(time.Hour * -24)
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

	continuousDataset := fillInGaps(dataset, 60)
	ts.ActivitiesHeartIntraday.Dataset = continuousDataset

	return ts, nil
}

// fillInGaps fills in gaps in the dataset with the previous datapoint, so ticks can be generated at hour marks.
func fillInGaps(data []Datapoint, gapAmt int64) []Datapoint {
	if len(data) == 0 {
		return []Datapoint{}
	}
	start := data[0].DateTime.Unix() + gapAmt
	for i, validTime := 1, start;
		i < len(data);
		i, validTime = i+1, validTime+gapAmt {
			if data[i].DateTime.Unix() != validTime {
				newEntry := Datapoint{
					Time:     "",
					DateTime: time.Unix(validTime, 0),
					Value:    data[i-1].Value,
				}
				data = append(data[:i], append([]Datapoint{newEntry}, data[i:]...)...)
			}
	}
	return data
}
