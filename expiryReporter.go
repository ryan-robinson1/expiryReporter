package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	getCertExpiry "github.com/ryan-robinson1/getCertExpiryPackage"
)

type Request struct {
	url      string
	duration time.Duration
}
type Result struct {
	url    string
	valid  bool
	time   string
	expiry string
	err    error
}

func fileConverter(filepath string) ([]Request, error) {
	var reqs []Request
	file, err := os.Open(filepath)

	if err != nil {
		return []Request{}, nil
	}
	defer file.Close()

	reader := bufio.NewReader(file)

	for {
		line, eoferr := reader.ReadString('\n')
		line = strings.TrimSuffix(line, "\n")

		s := strings.Split(line, ", ")
		dur, err := time.ParseDuration(s[0])
		if err != nil {
			return []Request{}, errors.New("err: invalid file format")
		}
		req := Request{s[1], dur}
		reqs = append(reqs, req)
		if eoferr != nil {
			break
		}
	}
	return reqs, nil
}

func doCheck(r Request, resultsChan chan Result) {
	status, expiry, err := getCertExpiry.GetCertExpiry(r.url, "", "", "", true)
	res := Result{r.url, status == 0, time.Now().UTC().Format("2006-01-02 15:04:05"), expiry, err}
	resultsChan <- res
}
func sleepAndExecute(r Request, resultsChan chan Result) {
	for {
		doCheck(r, resultsChan)
		time.Sleep(r.duration)
	}
}

//Builds a Report from Result struct
func BuildReport(r Result) string {
	if r.err != nil {
		return "-------------------------------------\n" + r.url + " report\n-------------------------------------\ntime        : " + r.time + "\nerror:       " + r.err.Error() + "\n\n"
	}
	return "-------------------------------------\n" + r.url + " report\n-------------------------------------\ntime        : " + r.time + "\nexpired     : " + strconv.FormatBool(!r.valid) + "\nexpiration  : " + r.expiry + "\n\n"
}
func main() {
	reqs, err := fileConverter("input")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	exit := make(chan bool)
	results := make(chan Result, 10)
	ok := make(chan string, 10)
	er := make(chan string, 10)

	for _, r := range reqs {
		go sleepAndExecute(r, results)
	}

	go func() {
		for r := range results {
			R := BuildReport(r)
			if r.valid {
				ok <- R
			} else {
				er <- R
			}
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
