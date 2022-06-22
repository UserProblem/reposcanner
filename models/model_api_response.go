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

type ApiResponse struct {

	// id produced by the operation
	Id int64 `json:"id,omitempty"`

	// result of the operation
	Message string `json:"message,omitempty"`
}