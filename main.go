package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

func main() {
	var currentBanner = NewBanner()
	var verifiedTokens = make(chan struct{})

	config, err := readConfigFile()
	if err != nil {
		if strings.Contains(err.Error(), "The system cannot find the file specified.") || strings.Contains(err.Error(), "no such file") {
			genConfigFile()
			fmt.Println("Generated config.json file. Fill the file out with tokens from creating a new app at https://dev.fitbit.com/apps then relaunch this binary.\nEnsure the app type is Personal, and Callback URL is http://localhost:8090/.")
			os.Exit(0)
		}
		log.Fatalln(err)
	}
	if err = validateConfig(config); err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/stats.svg", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/svg+xml; charset=utf-8")
		w.Header().Set("Cache-Control", "no-store")
		fmt.Fprint(w, currentBanner.svg)
	})
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		userAuthCode := r.URL.Query().Get("code")
		userCreds, err := handlerUserCredentials(userAuthCode, config)
		if err != nil {
			fmt.Fprint(w, err.Error())
			return
		}
		updateSVG(userCreds, config, currentBanner)
		_, html := setupSuccessMsgs()
		fmt.Fprint(w, html)
		verifiedTokens <- struct{}{}
	})
	go http.ListenAndServe(":8090", nil)

	userCredentials, err := readUserCredsFile()
	if err != nil {
		if !strings.Contains(err.Error(), "no such file") || strings.Contains(err.Error(), "The system cannot find the file specified.") {
			log.Fatalln(err)
		}
		genUserCredsFile()
		fmt.Println("Generated user credentials file. Follow this link (leave this binary running): ", tokensLink(config.OAuthClientID))
		<-verifiedTokens
		userCredentials, err = readUserCredsFile()
		if err != nil {
			log.Fatalln("error reading user credentials after generating a second time", err)
		}
	}

	go func() {
		for {
			updateSVG(userCredentials, config, currentBanner)
			time.Sleep(time.Minute * 2)
		}
	}()
	pt, _ := setupSuccessMsgs()
	fmt.Println(pt)
	noExit := make(chan bool)
	<-noExit
}

// handlerUserCredentials requests user credentials from FitBit and writes them to disk.
func handlerUserCredentials(userAuthCode string, config Config) (UserCredentials, error) {
	if userAuthCode == "" {
		return UserCredentials{}, fmt.Errorf("no user auth code provided")
	}
	userCreds, err := reqUserCredentials(config, userAuthCode, "")
	if err != nil {
		return UserCredentials{}, fmt.Errorf("error grabbing user tokens and credentials: %w", err)
	}
	err = writeUserCredsFile(userCreds)
	if err != nil {
		return UserCredentials{}, fmt.Errorf("error writing user credentials: %w", err)
	}
	return userCreds, nil
}

func updateSVG(uc UserCredentials, config Config, banner *Banner) {
	hrts, err := heartRateTimesSeries(uc, config)
	if err != nil {
		log.Print("Error grabbing time series", err.Error())
		return
	}
	_, err = banner.GenSVG(hrts, config.DisplayGetSource)
	if err != nil {
		log.Print("Error generating SVG: ", err.Error())
	}
}
