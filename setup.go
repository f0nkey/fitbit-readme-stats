package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

// Config holds configuration set by command line flags and credentials from FitBit.
type Config struct {
	// Port is the port to serve the SVG on.
	Port int `json:"port"`

	// Timezone is the timezone used for display and calculations. Will get a "no data available" error if different from tz on the FitBit app. https://en.wikipedia.org/wiki/List_of_UTC_time_offsets
	Timezone int `json:"timezone"`

	// TimezoneAbbreviation is the timezone represented in letters e.g., CST, MST.
	TimezoneAbbreviation string `json:"timezone_abbreviation"`

	// BannerTitle is the title at the top of the banner.
	BannerTitle string `json:"banner_title"`

	// CacheInvalidationTime is how long (in seconds) before new heart-rate data should be requested from FitBit's server to make the plot.
	CacheInvalidationTime int `json:"cache_invalidation_time"`

	// PlotRange is the time interval (in hours) to look back for heart-rate data.
	PlotRange int `json:"plot_range"`

	// BannerWidth is the width of the generated .SVG.
	BannerWidth int `json:"banner_width"`

	// BannerHeight is the height of the generated .SVG.
	BannerHeight int `json:"banner_height"`

	// DisplayViewOnGitHub when true displays watermark/link to the GitHub repo in the top left.
	DisplayViewOnGitHub bool `json:"display_view_on_github"`

	// Theme represents the theme of the banner.
	Theme Theme `json:"theme"`

	// AppCredentials holds generated fields when a new app is made at https://dev.fitbit.com/.
	AppCredentials AppCredentials `json:"app_credentials"`

	// UserCredentials holds credentials to authenticate with and request from the FitBit Web API.
	UserCredentials UserCredentials `json:"user_credentials"`
}

// AppCredentials holds generated fields when a new app is made at https://dev.fitbit.com/.
type AppCredentials struct {
	OAuthClientID string `json:"oauth_client_id"`
	ClientSecret  string `json:"client_secret"`
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

func setupProcess() {
	file, err := os.OpenFile("config.json", os.O_RDONLY, 0644)
	if !errors.Is(err, os.ErrNotExist) {
		fmt.Println("config.json found. Press y and Enter to continue setup and overwrite this config.json.")
		s := ""
		fmt.Scanln(&s)
		if s != "y" {
			os.Exit(0)
		}
	}
	file.Close()

	abbrev, offset := time.Now().Local().Zone()
	config := Config{
		Port:                  8090,
		Timezone:              offset / 3600,
		TimezoneAbbreviation:  abbrev,
		BannerTitle:           "My Heart Rate From My FitBit Watch (Past 4 Hours)",
		CacheInvalidationTime: 180,
		PlotRange:             4,
		Theme: Theme{
			Background:   "rgba(50, 35, 35, 255)",
			HeartNumber:  "rgba(50, 35, 35, 255)",
			ViewOnGithub: "rgba(230, 225, 196, 255)",
			TimezoneText: "rgba(230, 225, 196, 255)",
			TextTicks:    "rgba(230, 225, 196, 255)",
			CurrentBPM:   "rgba(230, 225, 196, 255)",
			Title:        "rgba(230, 225, 196, 255)",
			Axes:         "rgba(239, 93, 50, 255)",
			PlotLine:     "rgba(239, 172, 50, 255)",
			Heart:        "rgba(239, 172, 50, 255)",
		},
		BannerWidth:         500,
		BannerHeight:        100,
		DisplayViewOnGitHub: false,
		AppCredentials:      AppCredentials{},
		UserCredentials:     UserCredentials{},
	}

	err = writeConfigFile(config)
	if err != nil {
		fmt.Print("Error generating empty config file:", err)
		pressEnterToExit()
	}

	fmt.Print("Entering Setup Mode ...")
	appCreds := askAppCredentials()
	userCreds := askUserCredentials(appCreds)
	config.AppCredentials = appCreds
	config.UserCredentials = userCreds
	err = writeConfigFile(config)
	if err != nil {
		fmt.Println("Error writing to config file:", err)
		pressEnterToExit()
	}
	fmt.Println("\n=========")
	fmt.Println("Step 3. Host")
	fmt.Println("Setup is complete! Run this binary WITHOUT the setup flag to host the banner at http://HOSTIP:8090/stats.svg.")
	fmt.Println("README.md Embed: ![FitBit Heart Rate Chart](http://HOSTIP:" + strconv.Itoa(config.Port) + "/stats.svg)")
	fmt.Println("Press the Enter Key to exit.")
	fmt.Scanln()
}

// askAppCredentials asks the user to register at FitBit's site to get application tokens.
func askAppCredentials() AppCredentials {
	fmt.Println("\n=========")
	fmt.Println("Step 1. Getting App Credentials")
	fmt.Println("1a.")
	fmt.Println("  Visit https://dev.fitbit.com/apps")
	fmt.Println("  Ensure the fields below are set:")
	fmt.Println("  - OAuth 2.0 Application Type: Personal")
	fmt.Println("  - Callback URL: http://localhost:8090")
	fmt.Println("1b.")
	fmt.Println(`  Fill out "oauth_client_id" and "client_secret" fields in the newly generated config.json with tokens from your FitBit app page.`)

	var confirm func()
	confirm = func() {
		fmt.Println("\nPress y and Enter after your credentials.json is filled out.")
		input := ""
		fmt.Scanln(&input)
		if input != "y" {
			confirm()
		}
	}
	confirm()
	conf, err := readConfigFile()
	if err != nil {
		fmt.Println("Error reading config file:", err)
		pressEnterToExit()
	}
	if err = validateAppCredentials(conf.AppCredentials); err != nil {
		fmt.Println("Error in config validation:", err)
		pressEnterToExit()
	}
	return conf.AppCredentials
}

// askAppCredentials asks the user to authenticate over OAuth2 get user tokens.
func askUserCredentials(appCreds AppCredentials) UserCredentials {
	fmt.Println("\n=========")
	fmt.Println("Step 2. Getting User Credentials")
	fmt.Println("Follow this link (leave this binary running): ", tokensLink(appCreds.OAuthClientID))

	userCreds := UserCredentials{}
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		userAuthCode := r.URL.Query().Get("code")
		if userAuthCode == "" {
			return // occurs when user leaves browser open and gets sent to this link again
		}
		var err error
		userCreds, err = reqInitUserCredentials(userAuthCode, appCreds)
		if err != nil {
			fmt.Fprint(w, "Error encountered. See console for further instructions.")
			fmt.Println(w, "Error requesting user credentials", err)
			pressEnterToExit()
		}
		fmt.Fprint(w, "Setup complete! See console for further instructions.")
	})
	go http.ListenAndServe(":8090", nil)

	for {
		if userCreds.APIToken == "" {
			time.Sleep(time.Second)
			continue
		}
		break
	}
	err := validateUserCredentials(userCreds)
	if err != nil {
		fmt.Println("Error validating user credentials:", err)
		pressEnterToExit()
	}
	return userCreds
}

// tokensLink returns the link used to authorize us access to the user's data.
func tokensLink(oauthClientID string) string {
	return fmt.Sprintf("https://www.fitbit.com/oauth2/authorize?response_type=code&client_id=%s&redirect_uri=http://localhost:8090&scope=heartrate&expires_in=604800", oauthClientID)
}

// Needed since Windows CLI closes immediately.
func pressEnterToExit() {
	fmt.Println("Press Enter to exit.")
	fmt.Scanln()
	os.Exit(0)
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

func writeConfigFile(c Config) error {
	b, _ := json.MarshalIndent(&c, "", "	")
	err := ioutil.WriteFile("config.json", b, 0644)
	if err != nil {
		return err
	}
	return nil
}

func validateConfig(c Config) error {
	if err := validateUserCredentials(c.UserCredentials); err != nil {
		return err
	}
	if err := validateAppCredentials(c.AppCredentials); err != nil {
		return err
	}
	return nil
}

func validateAppCredentials(ac AppCredentials) error {
	if ac.ClientSecret == "" {
		return errors.New("client secret in config.json empty")
	}
	if ac.OAuthClientID == "" {
		return errors.New("oauth client id in config.json empty")
	}
	return nil
}

func validateUserCredentials(uc UserCredentials) error {
	if uc.UserID == "" {
		return fmt.Errorf("empty UserID")
	}
	if uc.APIToken == "" {
		return fmt.Errorf("empty APIToken")
	}
	if uc.RefreshToken == "" {
		return fmt.Errorf("empty RefreshToken")
	}
	if uc.Scope == "" {
		return fmt.Errorf("empty Scope")
	}
	return nil
}
