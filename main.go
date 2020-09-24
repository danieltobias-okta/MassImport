package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Config struct {
	MAX_CONCURRENT_SESSIONS int
	ORG                     string
	API_TOKEN               string
	GROUP_ID                string
	CSV_FILE                string
	NOTIFY                  int
	Password                int
	GEN_PASSWORD            bool
	SPEED	int
}

type Profile struct {
	Profile map[string]string `json:"profile"`
}

var wg sync.WaitGroup

func main() {

	var conf Config    // Settings imported from config.toml
	var groupId string // id of the group to import users into. If there is no GROUP_ID variable in the toml, a group called "import users" will be created for us
	// var all_data [][][]string // We are loading the csv into this variable before splitting it up between various goroutines
	var proceed string // Used for user input, 'y' will begin the import process
	// waitgroup
	configJSON, err := os.Open("config.json")
	if err != nil {
		fmt.Println(`Cannot open configuration file. Make sure you have a "config.json" file in this directoy`)
		writeFile(time.Now().Format("2006-01-02 15:04:05") + ` Cannot open configuration file. Make sure you have a "config.json" file in this directoy`)
	}
	content, err := ioutil.ReadAll(configJSON)
	if err != nil {
		fmt.Println(`Cannot read configuration file. Please check the file and try again.`)
		writeFile(time.Now().Format("2006-01-02 15:04:05") + `Cannot read configuration file. Please check the file and try again.`)

	}

	json.Unmarshal([]byte(content), &conf)

	f, err := os.Open(conf.CSV_FILE)
	defer f.Close()
	if err != nil {
		fmt.Println(conf.CSV_FILE)
		fmt.Println(`Cannot find csv file specified in CSV_FILE variable of the config.json`)
		writeFile(time.Now().Format("2006-01-02 15:04:05") + ` Cannot find csv file specified in CSV_FILE variable of the config.json`)
	}
	// Use or create import group here
	rowEmitter := emitRows(f)
	h, ind, hasPw := parseHeader(<-rowEmitter)

	attributes := Header{h, ind, hasPw}

	fmt.Println("Found profile attributes " + strings.Join(attributes.H, ", ") + ". do you wish to proceed? (y/N)")
	fmt.Scan(&proceed)
	if proceed == "y" || proceed == "Y" {
		start := time.Now()
		if conf.GROUP_ID == "" {
			groupId = createImportGroup(conf)
			if groupId == "FAIL" {
				fmt.Println("Creation of group failed.")
				writeFile(time.Now().Format("2006-01-02 15:04:05") + " failed to create group")
				os.Exit(3)
			}
		} else {
			groupId = conf.GROUP_ID
		}

		results := make(chan string)
		for i := 0; i < conf.MAX_CONCURRENT_SESSIONS; i++ {
			wg.Add(1)
			go worker(rowEmitter, attributes, conf, groupId, &wg, results)
		}

		// var createdProfile Profile
		total := 0

		go func() {
			wg.Wait()
			close(results)
		}()
		for _ = range results {
			total++
			if conf.NOTIFY > 0 {
				if total%conf.NOTIFY == 0 {
					fmt.Printf("imported %d users.\n", total)
				}
			}
		}

		t := time.Now()
		elapsed := t.Sub(start)

		writeFile(time.Now().Format("2006-01-02 15:04:05") + " Finish import of " + strconv.Itoa(total) + " users. Duration was " + elapsed.String())
	}

}
