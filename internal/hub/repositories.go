/*
   Copyright 2020 Docker Inc.

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package hub

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

const (
	// RepositoriesURL path to the Hub API listing the repositories
	RepositoriesURL = "/v2/repositories/%s/"
	// DeleteRepositoryURL path to the Hub API to remove a repository
	DeleteRepositoryURL = "/v2/repositories/%s/"
)

//Repository represents a Docker Hub repository
type Repository struct {
	Name        string
	Description string
	LastUpdated time.Time
	PullCount   int
	StarCount   int
	IsPrivate   bool
}

//GetRepositories lists all the repositories a user can access
func (h *Client) GetRepositories(account string) ([]Repository, error) {
	u, err := url.Parse(h.domain + fmt.Sprintf(RepositoriesURL, account))
	if err != nil {
		return nil, err
	}
	q := url.Values{}
	q.Add("page_size", fmt.Sprintf("%v", itemsPerPage))
	q.Add("page", "1")
	q.Add("ordering", "last_updated")
	u.RawQuery = q.Encode()

	repos, next, err := h.getRepositoriesPage(u.String(), account)
	if err != nil {
		return nil, err
	}

	for next != "" {
		pageRepos, n, err := h.getRepositoriesPage(next, account)
		if err != nil {
			return nil, err
		}
		next = n
		repos = append(repos, pageRepos...)
	}

	return repos, nil
}

//RemoveRepository removes a repository on Hub
func (h *Client) RemoveRepository(repository string) error {
	req, err := http.NewRequest("DELETE", h.domain+fmt.Sprintf(DeleteRepositoryURL, repository), nil)
	if err != nil {
		return err
	}
	req.Header["Authorization"] = []string{fmt.Sprintf("Bearer %s", h.token)}
	_, err = doRequest(req)
	return err
}

func (h *Client) getRepositoriesPage(url, account string) ([]Repository, string, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, "", err
	}
	req.Header["Authorization"] = []string{fmt.Sprintf("Bearer %s", h.token)}
	response, err := doRequest(req)
	if err != nil {
		return nil, "", err
	}
	var hubResponse hubRepositoryResponse
	if err := json.Unmarshal(response, &hubResponse); err != nil {
		return nil, "", err
	}
	var repos []Repository
	for _, result := range hubResponse.Results {
		repo := Repository{
			Name:        fmt.Sprintf("%s/%s", account, result.Name),
			Description: result.Description,
			LastUpdated: result.LastUpdated,
			PullCount:   result.PullCount,
			StarCount:   result.StarCount,
			IsPrivate:   result.IsPrivate,
		}
		repos = append(repos, repo)
	}
	return repos, hubResponse.Next, nil
}

type hubRepositoryResponse struct {
	Count    int                   `json:"count"`
	Next     string                `json:"next,omitempty"`
	Previous string                `json:"previous,omitempty"`
	Results  []hubRepositoryResult `json:"results,omitempty"`
}

type hubRepositoryResult struct {
	Name           string         `json:"name"`
	Namespace      string         `json:"namespace"`
	PullCount      int            `json:"pull_count"`
	StarCount      int            `json:"star_count"`
	RepositoryType RepositoryType `json:"repository_type"`
	CanEdit        bool           `json:"can_edit"`
	Description    string         `json:"description,omitempty"`
	IsAutomated    bool           `json:"is_automated"`
	IsMigrated     bool           `json:"is_migrated"`
	IsPrivate      bool           `json:"is_private"`
	LastUpdated    time.Time      `json:"last_updated"`
	Status         int            `json:"status"`
	User           string         `json:"user"`
}

//RepositoryType lists all the different repository types handled by the Docker Hub
type RepositoryType string

const (
	//ImageType is the classic image type
	ImageType = RepositoryType("image")
)