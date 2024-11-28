package main

import (
	"bufio"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"sort"
	"strings"
	"sync"

	"golang.org/x/oauth2/google"
	"google.golang.org/api/cloudresourcemanager/v3"
	"google.golang.org/api/option"
)

// downloadGCPPermissions fetches the list of all available GCP permissions from the IAM permissions reference page.
func downloadGCPPermissions() ([]string, error) {
	baseURL := "https://cloud.google.com/iam/docs/permissions-reference"
	resp, err := http.Get(baseURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Extract the iframe URL containing permissions data using regex.
	iframeRegex := regexp.MustCompile(`<iframe src="([^"]+)"`)
	matches := iframeRegex.FindStringSubmatch(string(body))
	if len(matches) < 2 {
		return nil, fmt.Errorf("could not find iframe URL")
	}

	framePageURL := matches[1]
	if strings.HasPrefix(framePageURL, "/") {
		framePageURL = "https://cloud.google.com" + framePageURL
	}

	resp, err = http.Get(framePageURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	frameBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Extract permissions from the table in the iframe content.
	permissionsRegex := regexp.MustCompile(`<td id="([^"]+)">`)
	matches = permissionsRegex.FindAllStringSubmatch(string(frameBody), -1)

	var permissions []string
	for _, match := range matches {
		permissions = append(permissions, match[1])
	}

	return permissions, nil
}

// checkPermissions tests the specified permissions on the given resource using the Cloud Resource Manager API.
func checkPermissions(perms []string, client *cloudresourcemanager.Service, resource string) ([]string, error) {
	request := &cloudresourcemanager.TestIamPermissionsRequest{
		Permissions: perms,
	}

	resp, err := client.Projects.TestIamPermissions(resource, request).Do()
	if err != nil {
		return nil, err
	}

	return resp.Permissions, nil
}

// divideChunks splits a slice of permissions into smaller chunks of a given size.
func divideChunks(perms []string, chunkSize int) [][]string {
	var chunks [][]string
	for i := 0; i < len(perms); i += chunkSize {
		end := i + chunkSize
		if end > len(perms) {
			end = len(perms)
		}
		chunks = append(chunks, perms[i:end])
	}
	return chunks
}

func main() {
	// Define and parse command-line flags.
	project := flag.String("project", "", "GCP project ID")
	folder := flag.String("folder", "", "GCP folder ID")
	organization := flag.String("organization", "", "GCP organization ID")
	credentials := flag.String("credentials", "", "Path to credentials.json")
	verbose := flag.Bool("verbose", false, "Verbose output")
	threads := flag.Int("threads", 3, "Number of threads")
	chunkSize := flag.Int("size", 50, "Chunk size for permission checks")
	flag.Parse()

	// Ensure at least one resource is specified.
	if *project == "" && *folder == "" && *organization == "" {
		fmt.Println("You must specify either a project, folder, or organization.")
		flag.Usage()
		os.Exit(1)
	}

	// Load the credentials file.
	credentialsFile, err := os.Open(*credentials)
	if err != nil {
		fmt.Printf("Error reading credentials file: %v\n", err)
		os.Exit(1)
	}
	defer credentialsFile.Close()

	credentialsBytes, err := ioutil.ReadAll(credentialsFile)
	if err != nil {
		fmt.Printf("Error reading credentials file: %v\n", err)
		os.Exit(1)
	}

	// Parse the credentials to create a JWT config.
	config, err := google.JWTConfigFromJSON(credentialsBytes, "https://www.googleapis.com/auth/cloud-platform")
	if err != nil {
		fmt.Printf("Error parsing credentials file: %v\n", err)
		os.Exit(1)
	}

	// Initialize the Cloud Resource Manager API client.
	ctx := context.Background()
	client, err := cloudresourcemanager.NewService(ctx, option.WithTokenSource(config.TokenSource(ctx)))
	if err != nil {
		fmt.Printf("Error creating Cloud Resource Manager client: %v\n", err)
		os.Exit(1)
	}

	// Download the list of permissions.
	permissions, err := downloadGCPPermissions()
	if err != nil || len(permissions) == 0 {
		fmt.Printf("Error downloading GCP permissions: %v\n", err)
		os.Exit(1)
	}

	sort.Strings(permissions)
	fmt.Printf("Downloaded %d GCP permissions\n", len(permissions))

	// Divide permissions into chunks for parallel processing.
	chunks := divideChunks(permissions, *chunkSize)
	var wg sync.WaitGroup
	var mu sync.Mutex
	havePerms := []string{}

	// Process each chunk in a separate goroutine.
	for _, chunk := range chunks {
		wg.Add(1)
		go func(chunk []string) {
			defer wg.Done()

			// Determine the resource type based on the input flags.
			resource := "projects/" + *project
			if *folder != "" {
				resource = "folders/" + *folder
			} else if *organization != "" {
				resource = "organizations/" + *organization
			}

			// Check permissions for the current chunk.
			foundPerms, err := checkPermissions(chunk, client, resource)
			if err != nil {
				fmt.Printf("Error checking permissions: %v\n", err)
				return
			}

			if *verbose {
				fmt.Printf("Found: %v\n", foundPerms)
			}

			// Append found permissions to the result in a thread-safe manner.
			mu.Lock()
			havePerms = append(havePerms, foundPerms...)
			mu.Unlock()
		}(chunk)
	}

	// Wait for all goroutines to complete.
	wg.Wait()
	fmt.Printf("[+] Your Permissions:\n- %s\n", strings.Join(havePerms, "\n- "))
}
