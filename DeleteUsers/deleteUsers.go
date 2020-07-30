package main

import (
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

func jobs(d []DeleteUser) <-chan string {
	out := make(chan string)
	go func() {
		for _, id := range d {
			out <- id.Id
		}
	}()
	return out
}

func worker(ids <-chan string, results chan<- int, wg *sync.WaitGroup) {
	client := &http.Client{}
	defer wg.Done()
	for id := range ids {
		fmt.Println(id)
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
		results <- 1

	}
	close(results)

}

var wg sync.WaitGroup

func main() {
	stime := time.Now()
	url := org + "/api/v1/groups/" + groupId + "/users"
	client := &http.Client{}
	var deletedusers []DeleteUser
	fmt.Println("Time to delete!")
	results := make(chan int)

	total := 0
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println(err)
	}
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "SSWS "+token)
	res, _ := client.Do(req)
	err = json.NewDecoder(res.Body).Decode(&deletedusers)

	for len(deletedusers) != 0 {

		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			fmt.Println(err)
		}
		req.Header.Add("Accept", "application/json")
		req.Header.Add("Content-Type", "application/json")
		req.Header.Add("Authorization", "SSWS "+token)
		res, _ := client.Do(req)
		if res.StatusCode == 429 {
			rest := res.Header.Get("X-Rate-Limit-Reset")
			resttime, _ := strconv.ParseInt(rest, 10, 64)
			ctime := time.Now().Unix()
			time.Sleep(time.Duration(resttime-ctime) * time.Second)

		} else {
			err = json.NewDecoder(res.Body).Decode(&deletedusers)
			jobEmitter := jobs(deletedusers)
			for x := 0; x < n; x++ {
				wg.Add(1)
				go worker(jobEmitter, results, &wg)
			}
		}

		// deletedusers = nil

	}
	go func() {
		wg.Wait()
		close(results)
	}()
	for _ = range results {
		total++
		_ = <-results
		fmt.Printf("Deleted %d users.\n", total)
	}

	fmt.Println("Completed " + time.Since(stime).String())

}
