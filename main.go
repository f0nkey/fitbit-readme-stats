package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"
)

func main() {
	port := flag.String("port", "8090", "port to serve the banner on")
	setupMode := flag.Bool("setup", false, "true will make the binary run through the setup process to generate credentials.json, false will serve the SVG normally")
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
	currentBanner := defaultBanner(500, 100)
	http.HandleFunc("/stats.svg", func(w http.ResponseWriter, r *http.Request) {
		if time.Since(lastSVGGeneration) > time.Minute*2 {
			currentBanner = updateSVG(config)
			lastSVGGeneration = time.Now()
		}

		w.Header().Set("Content-Type", "image/svg+xml; charset=utf-8")
		w.Header().Set("Cache-Control", "no-store")
		fmt.Fprint(w, currentBanner)
	})
	fmt.Println("Ensure Bluetooth is enabled on your phone so data can sync to FitBit's servers.")
	fmt.Println("Use the following README embed:", "![FitBit Heart Rate Chart](http://HOSTIP:8090/stats.svg)")
	fmt.Println("Serving on port", *port + ".")
	http.ListenAndServe(":" + *port, nil)
}
