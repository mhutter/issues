package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"regexp"
)

func main() {
	// determine GIT remotes
	cmd := exec.Command("git", "remote", "-v")
	remotes, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Print(string(remotes))
		os.Exit(1)
	}

	// make sure we have a GitHub repo
	matched, _ := regexp.Match("github.com", remotes)
	if !matched {
		fmt.Println("Not a GitHub repository")
		os.Exit(1)
	}

	// determine push remote name
	re := regexp.MustCompile("github.com[:/]([^.]+)\\.git\\s+\\(push\\)")
	matches := re.FindSubmatch(remotes)
	if len(matches) < 2 {
		fmt.Printf("Repo name not found for remotes:\n%s", string(remotes))
		os.Exit(1)
	}

	// API request
	uri := "https://api.github.com/repos/" + string(matches[1]) + "/issues"
	resp, err := http.Get(uri)
	exitMaybe(err)

	if resp.StatusCode == 403 {
		// 403 not allowed --> Rate limit! :-(
		fmt.Println("--> Hit rate-limit, displaying cached results")
		cache, err := ioutil.ReadFile(".issues-cache")
		exitMaybe(err)
		fmt.Print(string(cache))
		os.Exit(0)
	}

	// parse the data
	body, err := ioutil.ReadAll(resp.Body)
	exitMaybe(err)
	var issues []githubIssue
	err = json.Unmarshal(body, &issues)
	exitMaybe(err)

	var out string
	if len(issues) < 1 {
		out = "No open issues! \\o/"
	} else {

		// construct output
		// out = ""
		for _, issue := range issues {
			out += fmt.Sprintf("#%-3d %s\n", issue.Number, issue.Title)
		}
	}

	// cache output
	ioutil.WriteFile(".issues-cache", []byte(out), 0644)

	// oh, and actually print the output!
	fmt.Print(out)
}

func exitMaybe(err error) {
	if err != nil {
		fmt.Println("!!! Error:", err.Error())
		os.Exit(1)
	}
}

type githubIssue struct {
	Number  int    `json:"number"`
	Title   string `json:"title"`
	Message string `json:"message"`
}
