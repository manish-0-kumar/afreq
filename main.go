package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"sync"
)

func main() {
	var hideJson, hideContentType bool
	var searchText string

	// Parse command-line arguments and flags manually
	args := os.Args[1:]
	for i := 0; i < len(args); i++ {
		arg := args[i]

		switch arg {
		case "-hj", "--hide-json":
			hideJson = true
		case "-hct", "--hide-content-type":
			hideContentType = true
		default:
			// Assume it's the search text
			searchText = arg
		}
	}

	// Split the searchText by the "|" delimiter
	searchStrings := strings.Split(searchText, "|")

	// Create a channel to communicate results.
	jobs := make(chan string)
	var wg sync.WaitGroup

	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for domain := range jobs {
				resp, err := http.Get(domain)
				if err != nil {
					continue
				}
				defer resp.Body.Close()

				body, err := ioutil.ReadAll(resp.Body)
				if err != nil {
					fmt.Println(err)
					continue
				}
				sb := strings.ToLower(string(body)) // Convert response content to lowercase

				// Iterate over each search string and check if any of them are present (case-insensitive)
				for _, searchString := range searchStrings {
					searchString = strings.ToLower(searchString) // Convert search string to lowercase
					checkResult := strings.Contains(sb, searchString)
					if checkResult {
						// If any search string is found, print the content type and break
						contentType := resp.Header.Get("Content-Type")

						// If hideJson is true and the content type is "application/json," skip printing.
						if hideJson && strings.Contains(contentType, "application/json") {
							continue
						}

						// If hideContentType is false, print the content type.
						if !hideContentType {
							fmt.Println("MIME Type:", contentType)
						}
						fmt.Println(domain)
						break
					}
				}
			}
		}()
	}

	// Read domains from standard input.
	sc := bufio.NewScanner(os.Stdin)
	for sc.Scan() {
		domain := sc.Text()
		jobs <- domain
	}

	close(jobs)
	wg.Wait()
}
