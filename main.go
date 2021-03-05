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
	verifiedTokens := make(chan struct{})
	currentBanner := defaultBanner(500, 100)

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

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		userAuthCode := r.URL.Query().Get("code")
		userCreds, err := handlerUserCredentials(userAuthCode, config)
		if err != nil {
			fmt.Fprint(w, err.Error())
			return
		}
		updateSVG(userCreds, config)
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

	lastSVGGeneration := time.Unix(0, 0)
	http.HandleFunc("/stats.svg", func(w http.ResponseWriter, r *http.Request) {
		if time.Since(lastSVGGeneration) > time.Minute*2 {
			currentBanner = updateSVG(userCredentials, config)
			lastSVGGeneration = time.Now()
		}

		w.Header().Set("Content-Type", "image/svg+xml; charset=utf-8")
		w.Header().Set("Cache-Control", "no-store")
		fmt.Fprint(w, currentBanner)
	})

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

func updateSVG(uc UserCredentials, config Config) string {
	hrts, err := heartRateTimesSeries(uc, config)
	if err != nil {
		log.Print("Error grabbing time series", err.Error())
		return defaultBanner(500, 100)
	}
	banner, err := genBanner(hrts, config.DisplayGetSource)
	if err != nil {
		log.Print("Error generating banner: ", err.Error())
		return defaultBanner(500, 100)
	}
	return banner
}
