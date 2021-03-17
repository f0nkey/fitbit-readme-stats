package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"
)

func main() {
	setupMode := flag.Bool("setup", false, "run through the setup process to generate credentials.json, instead of serving the SVG normally")
	flag.Parse()

	if *setupMode {
		setupProcess()
		os.Exit(0)
	}

	config, err := readConfigFile()
	if err != nil {
		fmt.Println("Error reading config file (use -setup flag on this binary if you have not already):", err)
		pressEnterToExit()
	}
	if err = validateConfig(config); err != nil {
		fmt.Println("Error validating config file (use -setup flag on this binary if you have not already):", err)
		pressEnterToExit()
	}

	lastSVGGeneration := time.Unix(0, 0)
	currentBanner := defaultBanner(config)
	http.HandleFunc("/stats.svg", func(w http.ResponseWriter, r *http.Request) {
		if time.Since(lastSVGGeneration) > time.Second*time.Duration(config.CacheInvalidationTime) {
			currentBanner = updateSVG(&config)
			lastSVGGeneration = time.Now()
		}

		w.Header().Set("Content-Type", "image/svg+xml; charset=utf-8")
		w.Header().Set("Cache-Control", "no-store, no-cache, max-age=0")
		fmt.Fprint(w, currentBanner)
	})
	fmt.Println("Ensure Bluetooth is enabled on your phone so data can sync to FitBit's servers, as well as Battery Saver mode being off.")
	fmt.Println("Use the following README embed:", "![FitBit Heart Rate Chart](http://HOSTIP:"+strconv.Itoa(config.Port)+"/stats.svg)")
	fmt.Println("Serving on port", strconv.Itoa(config.Port)+".")
	http.ListenAndServe(":"+strconv.Itoa(config.Port), nil)
}
