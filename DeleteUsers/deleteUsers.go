package main

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	org     = ""
	token   = ""
	groupId = ""
	n       = 75
)

type DeleteUser struct {
	Id string `json:"id"`
}

func getDelUsers(c *http.Client, groupId string, token string, org string) ([]DeleteUser, bool) {
	var deletedusers []DeleteUser
	url := org + "/api/v1/groups/" + groupId + "/users"

	req, err := http.NewRequest("GET", url, bytes.NewBuffer([]byte("")))
	if err != nil {
		fmt.Println(err)
	}
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "SSWS "+token)
	req.Header.Add("Accept-Encoding", "gzip")
	res, _ := c.Do(req)
	reader, err := gzip.NewReader(res.Body)

	err = json.NewDecoder(reader).Decode(&deletedusers)
	if len(deletedusers) == 0 {
		return deletedusers, false
	}

	return deletedusers, true

}

func jobs(d []DeleteUser) <-chan string {
	out := make(chan string)
	go func() {
		for _, id := range d {
			out <- id.Id
		}
		close(out)
	}()
	return out
}

func worker(ids <-chan string, wg *sync.WaitGroup) {
	client := &http.Client{}

	defer wg.Done()
	for id := range ids {
		url := org + "/api/v1/users/" + id + "/lifecycle/deactivate"
		req, err := http.NewRequest("POST", url, strings.NewReader(""))
		if err != nil {
			fmt.Println(err)
		}
		req.Header.Add("Content-Type", "application/json")
		req.Header.Add("Accept", "application/json")
		req.Header.Add("Authorization", "SSWS "+token)
		r, _ := client.Do(req)

		if r.StatusCode == 429 {
			rest := r.Header.Get("X-Rate-Limit-Reset")
			resttime, _ := strconv.ParseInt(rest, 10, 64)
			ctime := time.Now().Unix()
			time.Sleep(time.Duration(resttime-ctime+1) * time.Second)
			req, err := http.NewRequest("POST", url, strings.NewReader(""))
			if err != nil {
				fmt.Println(err)
			}
			req.Header.Add("Content-Type", "application/json")
			req.Header.Add("Accept", "application/json")
			req.Header.Add("Authorization", "SSWS "+token)
			req.Header.Add("Accept-Encoding", "gzip")
			_, _ = client.Do(req)
		}
		url = org + "/api/v1/users/" + id
		req, err = http.NewRequest("DELETE", url, nil)

		if err != nil {
			fmt.Println(err)
		}
		req.Header.Add("Accept", "application/json")
		req.Header.Add("Content-Type", "application/json")
		req.Header.Add("Authorization", "SSWS "+token)

		n, _ := client.Do(req)
		if n.StatusCode == 429 {
			rest := n.Header.Get("X-Rate-Limit-Reset")
			resttime, _ := strconv.ParseInt(rest, 10, 64)
			ctime := time.Now().Unix()
			time.Sleep(time.Duration(resttime-ctime) * time.Second)
			req, err = http.NewRequest("DELETE", url, nil)

			if err != nil {
				fmt.Println(err)
			}
			req.Header.Add("Accept", "application/json")
			req.Header.Add("Content-Type", "application/json")
			req.Header.Add("Authorization", "SSWS "+token)

		}

	}

}

var wg sync.WaitGroup

func main() {
	client := &http.Client{}
	firstJobs, _ := getDelUsers(client, groupId, token, org)
	job := jobs(firstJobs)

	fmt.Printf("Deleting...\n")
	total := 0
	for len(firstJobs) > 0 {
		for i := 0; i < n; i++ {
			wg.Add(1)
			go worker(job, &wg)
		}
		wg.Wait()
		total += len(firstJobs)
		fmt.Printf("Deleted %d\n", total)
		firstJobs, _ = getDelUsers(client, groupId, token, org)
		job = jobs(firstJobs)
	}
	fmt.Printf("Deleted %d Okta users\n", total)

}
