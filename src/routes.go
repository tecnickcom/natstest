package main

import "github.com/julienschmidt/httprouter"

// Route contains the HTTP route description
type Route struct {
	Method      string            `json:"method"`      // HTTP method
	Path        string            `json:"path"`        // URL path
	Handle      httprouter.Handle `json:"-"`           // Handler function
	Description string            `json:"description"` // Description
}

// Routes is a list of routes
type Routes []Route

// HTTP routes
var routes = Routes{
	Route{
		"GET",
		"/status",
		statusHandler,
		"Check this service status",
	},
	Route{
		"GET",
		"/test/:name",
		test,
		"Execute the specified test",
	},
	Route{
		"GET",
		"/reload",
		reload,
		"Reset and reload the test configuration files",
	},
	Route{
		"PUT",
		"/new/:name",
		newtest,
		"Load and execute the specified test configuration",
	},
	Route{
		"DELETE",
		"/delete/:name",
		deltest,
		"Remove the specified test configuration",
	},
}
