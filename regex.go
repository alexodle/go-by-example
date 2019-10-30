package main

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
)

var re = regexp.MustCompile(`Code: ([0-9][0-9][0-9]).`)

func main() {
	ex1 := `error handling request: failed to respond to secret request from pod: failed to read secret from Vault: vault request failed: Error making API request.\n\nURL: GET https://vault.rubix-system.svc.cluster.local:8200/v1/notarealpath_hihi_odle\nCode: 403. Errors:\n\n* 1 error occurred:\n\t* permission denied\n\n, 500`
	ex2 := `error handling request: failed to respond to secret request from pod`

	match(ex1)
	match(ex2)
}

func match(ex string) {
	if code, ok := extractStatusCodeFromVaultError(errors.New(ex)); ok {
		fmt.Printf("Found: %d\n", code)
	} else {
		fmt.Println("not found!")
	}
}

var statusCodeRegex = regexp.MustCompile(`Code: ([0-9][0-9][0-9])\.`)
func extractStatusCodeFromVaultError(err error) (int, bool) {
	if extract := statusCodeRegex.FindStringSubmatch(err.Error()); extract != nil {
		if statusCode, convErr := strconv.Atoi(extract[1]); convErr == nil {
			return statusCode, true
		}
	}
	return 0, false
}
