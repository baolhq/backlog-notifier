package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/go-toast/toast"
	"github.com/kenzo0107/backlog"
	"github.com/wailsapp/wails/v3/pkg/application"
)

type Service struct{}

type Store struct {
	APIKey string `json:"apiKey"`
	UserID int    `json:"userId"`
	Name   string `json:"name"`
}

func (s *Service) Run() {
	var notifiedIssues = make(map[string]bool)

	store, err := loadStore()
	if err != nil {
		log.Fatalf("Failed to load store: %v", err)
	}

	for {
		log.Printf("Checking for issues assigned to %v", store.Name)
		issues, err := getAssignedIssues(store.APIKey, store.UserID)
		if err != nil {
			log.Printf("Error fetching issues: %v", err)
			time.Sleep(1 * time.Second)
			continue
		}

		if len(issues) == 0 {
			log.Println("There are no new issues assigned to you.")
		} else {
			for _, issue := range issues {
				if issue.Assignee.ID != nil && *issue.Assignee.ID == store.UserID {
					// Check and send notification only if not already notified
					if _, notified := notifiedIssues[*issue.IssueKey]; !notified {
						log.Println("New issue(s) assigned to you, sending notifications...")
						err := sendNotification(issue)
						if err != nil {
							log.Printf("Error sending notification: %v", err)
						} else {
							// Mark the issue as notified
							notifiedIssues[*issue.IssueKey] = true
						}
					} else {
						log.Println("There are no new issues assigned to you.")
					}
				}
			}
		}

		log.Printf("Waiting %d seconds before next check...\n", DELAY_TIME_IN_SEC)
		time.Sleep(DELAY_TIME_IN_SEC * time.Second)
	}
}

// Get backlog user with API Key
func (s *Service) GetUser(apiKey string) string {
	user, err := getOwnUser(apiKey)

	if err != nil || user.ID == nil {
		return "Invalid API Key"
	}
	saveStore(&Store{
		APIKey: apiKey,
		UserID: *user.ID,
		Name:   *user.Name,
	})
	return ""
}

// Hide window
func (s *Service) HideWindow() {
	app := application.Get()
	app.GetWindowByName(APP_TITLE).Hide()
}

// Open URL in default browser
func (s *Service) OpenURL(url string) {
	app := application.Get()
	app.BrowserOpenURL(url)
}

// Retrieves the authenticated user's details from the Backlog API.
func getOwnUser(apiKey string) (*backlog.User, error) {
	apiURL := fmt.Sprintf("%s/api/v2/users/myself", BASE_URL)
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, err
	}

	query := req.URL.Query()
	query.Set("apiKey", apiKey)
	req.URL.RawQuery = query.Encode()

	httpClient := &http.Client{}
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var user backlog.User
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, err
	}

	return &user, nil
}

// Loads the application store from a JSON file. If the file doesn't exist, returns an empty store.
func loadStore() (*Store, error) {
	file, err := os.Open(STORE_FILE)
	if err != nil {
		if os.IsNotExist(err) {
			// Return an empty store if the file doesn't exist
			return &Store{}, nil
		}
		return nil, err
	}
	defer file.Close()

	var store Store
	err = json.NewDecoder(file).Decode(&store)
	if err != nil {
		return nil, err
	}
	return &store, nil
}

// Saves the application store to a JSON file.
func saveStore(store *Store) error {
	file, err := os.Create(STORE_FILE)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	return encoder.Encode(store)
}

// Retrieves issues assigned to a specific user with "In Progress" status from the Backlog API.
func getAssignedIssues(apiKey string, userID int) ([]*backlog.Issue, error) {
	apiURL := fmt.Sprintf("%s/api/v2/issues", BASE_URL)
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, err
	}

	query := req.URL.Query()
	query.Set("apiKey", apiKey)
	query.Set("assigneeId[]", strconv.Itoa(userID))
	// query.Set("statusId[]", strconv.Itoa(IN_PROGRESS_STATUS))
	req.URL.RawQuery = query.Encode()

	httpClient := &http.Client{}
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var issues []*backlog.Issue
	if err := json.NewDecoder(resp.Body).Decode(&issues); err != nil {
		return nil, err
	}

	return issues, nil
}

// Sends a desktop notification for a newly assigned Backlog issue.
func sendNotification(issue *backlog.Issue) error {
	args := fmt.Sprintf("https://eng-sol.backlog.com/view/%s", *issue.IssueKey)
	iconPath, _ := filepath.Abs("./assets/appicon.png")

	notification := toast.Notification{
		AppID:    APP_NAME,
		Audio:    toast.Reminder,
		Duration: toast.Long,
		Title:    fmt.Sprintf("New Backlog Issue Assigned: %s", *issue.IssueKey),
		Message:  *issue.Description,
		Icon:     iconPath,
		Actions: []toast.Action{
			{Type: "protocol", Label: "Open in browser", Arguments: args},
		},
	}

	return notification.Push()
}
