/*
 * Repository Secrets Scanner
 *
 * This is a simple backend API to allow a user to configure repositories for scanning, trigger a scan of those repositories, and retrieve the results.
 *
 * API version: 0.0.1
 * Contact: sean.critica@gmail.com
 * Generated by: Swagger Codegen (https://github.com/swagger-api/swagger-codegen.git)
 */

package swagger

import (
	"net/http"
	"strings"

	"github.com/gorilla/mux"
)

type HttpRouter interface {
	ServeHTTP(w http.ResponseWriter, r *http.Request)
}

type Route struct {
	Name        string
	Method      string
	Pattern     string
	HandlerFunc http.HandlerFunc
}

type Routes []Route

const api_version string = "/v0"

func (a *App) NewRouter() HttpRouter {
	var routes = Routes{
		Route{
			"AddRepository",
			strings.ToUpper("Post"),
			api_version + "/repository",
			a.AddRepository,
		},

		Route{
			"AddScan",
			strings.ToUpper("Post"),
			api_version + "/repository/{id}/startScan",
			a.AddScan,
		},

		Route{
			"DeleteRepository",
			strings.ToUpper("Delete"),
			api_version + "/repository/{id}",
			a.DeleteRepository,
		},

		Route{
			"GetRepository",
			strings.ToUpper("Get"),
			api_version + "/repository/{id}",
			a.GetRepository,
		},

		Route{
			"ListRepositories",
			strings.ToUpper("Get"),
			api_version + "/repositories",
			a.ListRepositories,
		},

		Route{
			"ModifyRepository",
			strings.ToUpper("Put"),
			api_version + "/repository/{id}",
			a.ModifyRepository,
		},

		Route{
			"AddScan",
			strings.ToUpper("Post"),
			api_version + "/repository/{id}/startScan",
			a.AddScan,
		},

		Route{
			"DeleteScan",
			strings.ToUpper("Delete"),
			api_version + "/scan/{id}",
			a.DeleteScan,
		},

		Route{
			"GetScan",
			strings.ToUpper("Get"),
			api_version + "/scan/{id}",
			a.GetScan,
		},

		Route{
			"ListScans",
			strings.ToUpper("Get"),
			api_version + "/scans",
			a.ListScans,
		},
	}

	router := mux.NewRouter().StrictSlash(true)
	for _, route := range routes {
		var handler http.Handler
		handler = route.HandlerFunc
		handler = Logger(handler, route.Name)

		router.
			Methods(route.Method).
			Path(route.Pattern).
			Name(route.Name).
			Handler(handler)
	}

	return router
}
