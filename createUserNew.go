package main

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Header struct {
	H     []string
	Pw    int
	hasPw bool
}

func parseHeader(c []string) (header []string, k int, f bool) {

	i, found := Find("password", c)
	return c, i, found
}

func prepareRequest(req *http.Request, token string) {
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "SSWS "+token)
}

func Find(v string, l []string) (int, bool) {
	for i := 0; i < len(l); i++ {
		if v == l[i] {
			return i, true
		}
	}
	return -1, false
}

func emitRows(f *os.File) <-chan []string {
	out := make(chan []string)
	r := csv.NewReader(f)

	go func() {
		for {
			record, err := r.Read()
			if err == io.EOF {

				close(out)
				break
			} else if err != nil {

				fmt.Println(err)
			}
			out <- record

		}

	}()
	return out
}

func formRequest(s []string, h Header, c Config, groupId string) []byte {
	out := `{"profile":{`
	N := len(h.H)
	for i := 0; i < N; i++ {
		if i != h.Pw {
			out += `"` + h.H[i] + `":"` + s[i] + `"`
		}

		out += `,`

	}

	out = out[0 : len(out)-1]

	if !c.GEN_PASSWORD && h.hasPw {
		out = out[0 : len(out)-1]
	}

	out += `},`
	if h.hasPw {
		out += `"credentials":{"password":{"value":"` + s[h.Pw] + `"}},`
	} else {
		if c.GEN_PASSWORD == true {

		} else {

			out += `"credentials":{"password":{"hook":{"type":"default"}}},`
		}

	}
	out += `"groupIds":["` + groupId + `"]}`

	return []byte(out)

}

func worker(rows <-chan []string, h Header, c Config, groupId string, wg *sync.WaitGroup, results chan string) {
	client := &http.Client{}
	defer wg.Done()
	var newRate, limit, rem int
	for user := range rows {
		url := c.ORG + "/api/v1/users?activate=true"
		if c.GEN_PASSWORD {
			url += "&sendEmail=true"
		}
		reqBody := formRequest(user, h, c, groupId)
		req, err := http.NewRequest("POST", url, bytes.NewBuffer(reqBody))
		if err != nil {
			fmt.Println(err)
			writeFile(time.Now().Format("2006-01-02 15:04:05") + ` ` + err.Error())
		}
		prepareRequest(req, c.API_TOKEN)
		res, err := client.Do(req)
		

		if c.SPEED != 100 {
			limit, _ = strconv.Atoi(res.Header["X-Rate-Limit-Limit"][0])
			rem, _ = strconv.Atoi(res.Header["X-Rate-Limit-Remaining"][0])
			newRate = int(float32(c.SPEED) / float32(100) * float32(limit))
			if newRate < 1 {
				newRate = 1
			}
			if rem < (limit-newRate) {
				rest, _ := strconv.ParseInt(res.Header.Get("X-Rate-Limit-Reset"), 10, 64)
				time.Sleep(time.Duration(rest-time.Now().Unix()+3) * time.Second)
			}
		}

		if res.StatusCode == 429 {
			rest, _ := strconv.ParseInt(res.Header.Get("X-Rate-Limit-Reset"), 10, 64)
			time.Sleep(time.Duration(rest-time.Now().Unix()+3) * time.Second)
			req, err := http.NewRequest("POST", url, bytes.NewBuffer(reqBody))
			if err != nil {
				fmt.Println(err)
				writeFile(time.Now().Format("2006-01-02 15:04:05") + ` ` + err.Error())
			}
			prepareRequest(req, c.API_TOKEN)
			_, _ = client.Do(req)

		} else {
			if res.StatusCode != 200 {
				r, _ := ioutil.ReadAll(res.Body)
				writeFile(time.Now().Format("2006-01-02 15:04:05") + " Failed user " + strings.Join(user, ", ") + " " + string(r))
			}
		}
		if res.StatusCode == 200 {

		}
		r, _ := ioutil.ReadAll(res.Body)
		results <- string(r)
	}
}
