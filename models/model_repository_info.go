/*
 * Repository Secrets Scanner
 *
 * This is a simple backend API to allow a user to configure repositories for scanning, trigger a scan of those repositories, and retrieve the results.
 *
 * API version: 0.0.1
 * Contact: sean.critica@gmail.com
 * Generated by: Swagger Codegen (https://github.com/swagger-api/swagger-codegen.git)
 */

package models

type RepositoryInfo struct {

	// short name for the repository
	Name string `json:"name"`

	// public URL of the repository
	Url string `json:"url"`

	// branch of the repository
	Branch string `json:"branch,omitempty"`
}

func DefaultRepositoryInfo() *RepositoryInfo {
	return &RepositoryInfo{
		Name:   "new repo",
		Url:    "http://example.com/repo",
		Branch: "main",
	}
}

func (ri *RepositoryInfo) Clone() *RepositoryInfo {
	return &RepositoryInfo{
		Name:   ri.Name,
		Url:    ri.Url,
		Branch: ri.Branch,
	}
}