package main

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/richjyoung/echopen"
	v310 "github.com/richjyoung/echopen/openapi/v3.1.0"
)

type ErrorResponseBody struct {
	Message string `json:"message"`
}

func main() {
	// Create a new echOpen wrapper
	api := echopen.New(
		"Hello World",
		"1.0.0",
		echopen.WithSchemaDescription("Demonstration of routes with security requirements"),
		echopen.WithSchemaLicense(&v310.License{Name: "MIT", URL: "https://example.com/license"}),
	)

	api.Schema.GetComponents().AddJSONResponse("ErrorResponse", "Error response", api.ToSchemaRef(ErrorResponseBody{}))

	api.Schema.GetComponents().AddSecurityScheme("api_key", &v310.SecurityScheme{
		Type: v310.APIKeySecuritySchemeType,
		In:   "header",
		Name: "X-API-Key",
	})

	// Optional security route
	api.GET(
		"/hello",
		hello,
		echopen.WithOptionalSecurity(),
		echopen.WithSecurityRequirement("api_key", []string{}),
		echopen.WithResponse(fmt.Sprint(http.StatusOK), "Successful response"),
		echopen.WithResponseRef("default", "ErrorResponse"),
	)

	// Secured route
	api.GET(
		"/hello_secure",
		helloSecure,
		echopen.WithSecurityRequirement("api_key", []string{}),
		echopen.WithResponse(fmt.Sprint(http.StatusOK), "Successful response"),
		echopen.WithResponseRef("default", "ErrorResponse"),
	)

	// Serve the generated schema
	api.ServeYAMLSchema("/openapi.yml")
	api.ServeUI("/", "/openapi.yml", "5.10.3")

	// Start the server
	api.Start("localhost:3030")
}

func hello(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]interface{}{
		"value":  c.Get("security.api_key"),
		"scopes": c.Get("security.api_key.scopes"),
	})
}

func helloSecure(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]interface{}{
		"value":  c.Get("security.api_key"),
		"scopes": c.Get("security.api_key.scopes"),
	})
}
