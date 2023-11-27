package main

import (
	"fmt"
	"net/http"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/labstack/echo/v4"
	"github.com/richjyoung/echopen"
)

const Description = `Basic Example
=============

Test API for basic usage of echOpen.
`

type QueryParams struct {
	Offset *int `query:"offset" description:"Offset into results"`
	Limit  int  `query:"limit"`
}

type RequestBodySettings struct {
	Enabled bool        `json:"enabled" description:"Enabled flag"`
	Other   interface{} `json:"other"`
}

type RequestBody struct {
	FirstName string  `json:"first_name,omitempty" description:"User first name" example:"Joe"`
	LastName  string  `json:"last_name,omitempty" description:"User last name" example:"Bloggs"`
	Email     *string `json:"email,omitempty" description:"Optional email address" example:"joe_bloggs@example.com"`
	Meta      struct {
		TermsAndConditions *int `json:"terms_and_conditions,omitempty" description:"Date of T&Cs acceptance"`
	} `json:"meta,omitempty"`
	Settings *RequestBodySettings `json:"settings,omitempty"`
}

func main() {
	// Create a new echOpen wrapper
	api := echopen.New("Basic Example", "1.0.0", "3.1.0")
	api.Description(Description)
	api.Licence(&openapi3.License{Name: "MIT", URL: "https://example.com/licence"})

	// Configure tags
	api.AddTag(&openapi3.Tag{
		Name:        "hello_world",
		Description: "Hello World API Routes",
	})
	api.AddTag(&openapi3.Tag{
		Name:        "param",
		Description: "Routes with params",
	})

	// Add a group
	helloGroup := api.Group("/hello", echopen.WithGroupTags("hello_world"))
	helloGroup.GET("", hello)
	helloGroup.GET("/:id", helloID, echopen.WithTags("param"), echopen.WithPathParameter(&echopen.PathParameter{
		Name:        "id",
		Description: "ID Parameter",
	}))

	helloGroup.GET("/query", helloQuery, echopen.WithQueryStruct(QueryParams{}))
	helloGroup.PATCH("/body", helloBody, echopen.WithRequestBody("Body params", RequestBody{}))
	helloGroup.PATCH("/body/settings", helloQuery, echopen.WithRequestBody("Body params", RequestBodySettings{}))

	// Serve the generated schema
	api.ServeSchema("/openapi.yml")

	// Start the server
	api.Start("localhost:3030")
}

func hello(c echo.Context) error {
	return c.String(http.StatusOK, "Hello, World!")
}

func helloID(c echo.Context) error {
	id := c.Get("param.id").(string)
	return c.String(http.StatusOK, "Hello, World! - "+id)
}

func helloQuery(c echo.Context) error {
	qry := c.Get("query").(*QueryParams)
	if qry == nil {
		return echo.ErrInternalServerError
	}
	return c.String(http.StatusOK, fmt.Sprintf("Hello, World! - %#v", qry))
}

func helloBody(c echo.Context) error {
	body := c.Get("body").(*RequestBody)
	if body == nil {
		return echo.ErrInternalServerError
	}
	return c.String(http.StatusOK, fmt.Sprintf("Hello, World! - %#v", body))
}