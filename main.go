package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type urlFlag struct {
	value *url.URL
}

func (u *urlFlag) Set(v string) error {
	parsedURL, err := url.Parse(v)
	if err != nil || parsedURL.Scheme == "" || parsedURL.Host == "" {
		return errors.New("invalid url")
	}
	u.value = parsedURL
	return nil
}

func (u *urlFlag) String() string {
	if u.value == nil {
		return ""
	}

	return u.value.String()
}

var verbose = flag.Bool("v", false, "Verbose mode")

func verboseLog(msg string) {
	if *verbose {
		log.Println("-> " + msg)
	}
}

func main() {
	origin := urlFlag{}
	flag.Var(&origin, "from", "origin of the file")
	destination := flag.String("to", ".\\", "Destination of the file (with filename)")

	flag.Parse()

	start := time.Now()
	verboseLog("parsed flags")

	fp, err := filepath.Abs(*destination)
	if err != nil {
		verboseLog("unable to get filepath: " + err.Error())
	}

	verboseLog("generated filepath")

	if strings.HasSuffix(fp, "/") {
		log.Fatalln("no filename was provided on destination")
	}

	// Ensure the file destination exists
	verboseLog("Ensuring file destination exists")
	_, err = os.Stat(filepath.Dir(fp))
	if os.IsNotExist(err) {
		verboseLog("\tfile destination didn't exist yet, creating it")
		err := os.MkdirAll(filepath.Dir(fp), 0755)
		if err != nil {
			log.Fatalln("unable to create directory: " + err.Error() + "\n aborting..")
		}
	} else if err != nil {
		log.Fatalln("error checking existence of directory:  " + err.Error() + "\naborting..")
	}

	// Tries to retrieve file from server
	verboseLog("retrieving file from server")
	resp, err := http.Get(origin.String())
	if err != nil {
		verboseLog("an error occurred after get request.")
		log.Fatalln("Error getting the file")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		verboseLog("status code expected to be 200 (ok) but was " + strconv.Itoa(resp.StatusCode))
		log.Fatalln("Error getting the file")
	}

	// Create the file locally
	verboseLog("creating the file locally")
	out, err := os.Create(fp)
	if err != nil {
		verboseLog("the system couldn't create the file: " + err.Error())
		log.Fatalln("unable to create the file")
	}

	defer out.Close()

	// Actually writes the file
	verboseLog("writing the file")
	bwrt, err := io.Copy(out, resp.Body)
	if err != nil {
		verboseLog("the system couldn't write the file: " + err.Error())
		log.Fatalln("unable to write the file")
	}

	elapsed := time.Since(start)

	verboseLog(fmt.Sprintf("downloaded %d bytes in %v\nsaved to %s", bwrt, elapsed, fp))
	fmt.Println("Done!")

}
