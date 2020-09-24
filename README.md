# Mass Import Utility

---

## Usage

This utility was created using [Go 1.14](https://golang.org/dl/). You will need Go to do any of the "go run" commands or to use the delete user utility. Otherwise, you can run the binary "MassImport.exe" that I have precompiled in this repo.

The program looks for a `config.json` file which accepts the following options

* `MAX_CONCURRENT_SESSIONS` by default is 75, change this according to the concurrency limit for your Okta org
* `ORG` this is a string (and the value should be placed in "") of the form https://yoursubdomain.okta.com (or oktapreview etc)
* `API_TOKEN` another string for your API token. Create the API token as a super admin and paste the value here
* `CSV_FILE` path to your csv file, by default you should just use a "users.csv" file in the same directory as the program, more on this later
* `GROUP_ID` the program creates new users and places them all into a single import group. Specify the group id you wish to place these users in or simply change it to the empty string "" to have the program automatically create a group for you called "Import Users"
* `NOTIFY` this is an option that will print updates in the console. If the option is set to N then the console will tell you every time it has imported N users (also when a thread is waiting because it hit the API limit). An option of 0 make the program completely silent
* `GEN_PASSWORD` add `"GEN_PASSWORD" = true` if you would like Okta to create temporary passwords and email your users to activate their accounts. You do **not** need to include a 'password' header in the csv if you choose this option. In fact, make sure not to include it
* **NEW** `SPEED` takes values 0 to 100 (but don't put 0). This defines which percentage of your rate limit you would like to use before the program sleeps. For example, SPEED = 75 with a rate limit of 600req/min will make sure the program only uses 75% of your requests which would be 450req/min. Use this feature to avoid hitting rate limits or getting warnings. Alternatively, set it to 100 to hit your rate limit and go fast as possible.

### Csv file formatting
The csv file should be formatted with each profile attribute you want to create users with in the first row. Each column will belong to a single attribute. Each row thereafter will define the values for each user. These values are assumed to be strings. If you include a "password" column, make sure you have `"GEN_PASSWORD" = false` in your `config.json`.

---

### Reverting changes
I have included a utility in the DeleteUsers folder. The utility is definitely not polished, and perhaps a little slower but it will get the job done. In the src to `deleteUsers.go` you will see on line 12:

```go
const (
	org     = "https://myorg.okta.com"
	token   = "mytoken"
	groupId = "mygroupid"
)
```

Simply change the values here. The groupId needs to be the groupId of the group that the utility placed all of the users in. 

---

## Running
You have the option to compile, or more easily on windows cd into the MassImport directory and run

```
go run .
```
The program will ask if you want to create N users with a list of headers that it detects. You need to press "y" and hit the enter key for that process to begin, otherwise the program terminates.

---

## Python Utilities
If you want to play around with the importer, I've included a folder called `pythonUtilites`. You will need [Python 3.7](https://www.python.org/downloads/) to run it. It will generate a list of N users for you. If you want to change the headers, you need to edit the `fieldnames` variable and also include how you want the program to generate these values in your `writer.writerow` function call. It will output a `users.csv` file, which you should move into the parent MassImport directory. The syntax is
```
python createUsers.py N
```
where N is the amount of users you'd like to generate

---

## Binary
I've included a windows binary for your convenience, this way you don't need to install go, dependencies, and `go run`. You can simply execute the program with your `config.json` and `users.csv` file present.

---
## Contact
If you have any questions about the utility, feel free to contact me at daniel.tobias@okta.com. Hope the tool works for you :)
