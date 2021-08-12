package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"time"

	getCertExpiry "github.com/ryan-robinson1/getCertExpiryPackage"
)

type Requests struct {
	Requests []Request `json:"requests"`
}

type Request struct {
	Url      string `json:"url"`
	Duration string `json:"duration"`
}

type Result struct {
	Url    string
	valid  bool
	time   string
	expiry string
	err    error
}

func parseJson(filepath string) ([]Request, error) {
	jsonFile, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer jsonFile.Close()

	byteValue, _ := ioutil.ReadAll(jsonFile)

	var requests Requests

	json.Unmarshal(byteValue, &requests)
	return requests.Requests, nil
}

func doCheck(r Request, rc chan Result) {
	status, expiry, err := getCertExpiry.GetCertExpiry(r.Url, "", "", "", true)
	rc <- Result{r.Url,
		status == 0,
		time.Now().UTC().Format("2006-01-02 15:04:05"),
		expiry,
		err}
}

//Builds a Report from Result struct
func BuildReport(r Result) string {
	if r.err != nil {
		return "-------------------------------------\n" + r.Url + " report\n-------------------------------------\ntime        : " + r.time + "\nerror       : " + r.err.Error() + "\n\n"
	}
	return "-------------------------------------\n" + r.Url + " report\n-------------------------------------\ntime        : " + r.time + "\nexpired     : " + strconv.FormatBool(!r.valid) + "\nexpiration  : " + r.expiry + "\n\n"
}

func main() {
	reqs, err := parseJson("input.json")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	exit := make(chan bool)
	results := make(chan Result, 10)
	ok := make(chan string, 10)
	er := make(chan string, 10)

	for _, r := range reqs {
		go func(r Request) {
			for {
				doCheck(r, results)
				d, err := time.ParseDuration(r.Duration)
				if err != nil {
					fmt.Println("err: invalid duration " + r.Duration)
					os.Exit(1)
				}
				time.Sleep(d)
			}
		}(r)
	}

	go func() {
		for r := range results {
			r := r // YES there is a valid and needed reason this is here. Left as an exercise to the student to find out why
			go func() {
				if r.valid {
					ok <- BuildReport(r)
				} else {
					er <- BuildReport(r)
				}
			}()
		}
	}()

	go func() {
		for R := range ok {
			fmt.Fprintln(os.Stdout, R)
		}
	}()

	go func() {
		for R := range er {
			fmt.Fprintln(os.Stderr, R)
		}
	}()

	<-exit
}
