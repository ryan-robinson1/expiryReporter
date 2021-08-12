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
		return nil, err
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
			return reqs, nil
		}
	}

}

func doCheck(r Request, rc chan Result) {
	status, expiry, err := getCertExpiry.GetCertExpiry(r.url, "", "", "", true)
	rc <- Result{r.url,
		status == 0,
		time.Now().UTC().Format("2006-01-02 15:04:05"),
		expiry,
		err}
}

func sleepAndExecute(r Request, rc chan Result) {
	for {
		doCheck(r, rc)
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
	//	reqs, err := fileConverter("input")
	//	if err != nil {
	//		fmt.Println(err)
	//		os.Exit(1)
	//	}

	reqs := []Request{{"cnn.com:443", 7 * time.Second},
		{"expired.badssl.com:443", 13 * time.Second},
		{"example.com:443", 29 * time.Second},
		{"mainrouter-silkwave.novatr:61617", 22 * time.Second}}

	exit := make(chan bool)
	results := make(chan Result, 10)
	ok := make(chan string, 10)
	er := make(chan string, 10)

	for _, r := range reqs {
		go func(r Request) {
			for {
				doCheck(r, results)
				time.Sleep(r.duration)
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
