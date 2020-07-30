package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

type Group struct {
	Id string `json:"id"`
}

func createImportGroup(c Config) string {

	url := c.ORG + "/api/v1/groups"
	method := "POST"

	payload := []byte(`{"profile":{"name":"Import Group","description":"Users imported with the mass import utility"}}`)

	client := &http.Client{}
	req, err := http.NewRequest(method, url, bytes.NewBuffer(payload))

	if err != nil {
		fmt.Println(err)
		writeFile(time.Now().Format("2006-01-02 15:04:05") + ` ` + err.Error())
	}
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "SSWS "+c.API_TOKEN)

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		writeFile(time.Now().Format("2006-01-02 15:04:05") + ` ` + err.Error())
	}
	if res.StatusCode != 200 {
		return "FAIL"
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		writeFile(time.Now().Format("2006-01-02 15:04:05") + ` ` + err.Error())
	}
	var group Group
	json.Unmarshal(body, &group)

	return group.Id
}
