package main

import (
	"fmt"
	"net/http"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/labstack/echo/v4"
	"github.com/richjyoung/echopen"
)

func main() {
	// Create a new echOpen wrapper
	api := echopen.New("Swagger Petstore", "1.0.0", "3.0.0")
	api.Description("A sample API that uses a petstore as an example to demonstrate features in the OpenAPI 3.0 specification")
	api.Licence(&openapi3.License{Name: "Apache 2.0", URL: "https://www.apache.org/licenses/LICENSE"})
	api.TermsOfService("http://swagger.io/terms/")
	api.Contact(&openapi3.Contact{
		Name:  "Swagger API Team",
		Email: "apiteam@swagger.io",
		URL:   "http://swagger.io",
	})

	api.AddServer(&openapi3.Server{URL: "https://petstore.swagger.io/v2"})

	api.GET(
		"/pets",
		findPets,
		echopen.WithOperationID("findPets"),
		echopen.WithDescription(`
Returns all pets from the system that the user has access to
Nam sed condimentum est. Maecenas tempor sagittis sapien, nec rhoncus sem sagittis sit amet. Aenean at gravida augue, ac iaculis sem. Curabitur odio lorem, ornare eget elementum nec, cursus id lectus. Duis mi turpis, pulvinar ac eros ac, tincidunt varius justo. In hac habitasse platea dictumst. Integer at adipiscing ante, a sagittis ligula. Aenean pharetra tempor ante molestie imperdiet. Vivamus id aliquam diam. Cras quis velit non tortor eleifend sagittis. Praesent at enim pharetra urna volutpat venenatis eget eget mauris. In eleifend fermentum facilisis. Praesent enim enim, gravida ac sodales sed, placerat id erat. Suspendisse lacus dolor, consectetur non augue vel, vehicula interdum libero. Morbi euismod sagittis libero sed lacinia.

Sed tempus felis lobortis leo pulvinar rutrum. Nam mattis velit nisl, eu condimentum ligula luctus nec. Phasellus semper velit eget aliquet faucibus. In a mattis elit. Phasellus vel urna viverra, condimentum lorem id, rhoncus nibh. Ut pellentesque posuere elementum. Sed a varius odio. Morbi rhoncus ligula libero, vel eleifend nunc tristique vitae. Fusce et sem dui. Aenean nec scelerisque tortor. Fusce malesuada accumsan magna vel tempus. Quisque mollis felis eu dolor tristique, sit amet auctor felis gravida. Sed libero lorem, molestie sed nisl in, accumsan tempor nisi. Fusce sollicitudin massa ut lacinia mattis. Sed vel eleifend lorem. Pellentesque vitae felis pretium, pulvinar elit eu, euismod sapien.
		`),
		echopen.WithQueryStruct(FindPetsQuery{}),
		echopen.WithResponseBody(fmt.Sprint(http.StatusOK), "pet response", []Pet{}),
		echopen.WithResponseBody("default", "unexpected error", Error{}),
	)

	api.POST(
		"/pets",
		addPet,
		echopen.WithOperationID("addPet"),
		echopen.WithDescription("Creates a new pet in the store. Duplicates are allowed"),
		echopen.WithRequestBody("Pet to add to the store", NewPet{}),
		echopen.WithResponseBody(fmt.Sprint(http.StatusOK), "pet response", Pet{}),
		echopen.WithResponseBody("default", "unexpected error", Error{}),
	)

	api.GET(
		"/pets/:id",
		findPetByID,
		echopen.WithOperationID("findPetByID"),
		echopen.WithDescription("Returns a user based on a single ID, if the user does not have access to the pet"),
		echopen.WithPathParameter(&echopen.PathParameter{
			// TODO: needs schema integer/int64
			Name:        "id",
			Description: "ID of pet to fetch",
		}),
		echopen.WithResponseBody(fmt.Sprint(http.StatusOK), "pet response", Pet{}),
		echopen.WithResponseBody("default", "unexpected error", Error{}),
	)

	api.DELETE(
		"/pets/:id",
		deletePet,
		echopen.WithOperationID("deletePet"),
		echopen.WithDescription("deletes a single pet based on the ID supplied"),
		echopen.WithPathParameter(&echopen.PathParameter{
			// TODO: needs schema integer/int64
			Name:        "id",
			Description: "ID of pet to delete",
		}),
		echopen.WithResponse(fmt.Sprint(http.StatusNoContent), "pet deleted"),
		echopen.WithResponseBody("default", "unexpected error", Error{}),
	)

	// Serve the generated schema
	api.ServeSchema("/openapi.yml")
	api.ServeUI("/", "/openapi.yml", "5.10.3")

	// Start the server
	api.Start("localhost:3030")
}

func findPets(c echo.Context) error {
	return nil
}

func addPet(c echo.Context) error {
	return nil
}

func findPetByID(c echo.Context) error {
	return nil
}

func deletePet(c echo.Context) error {
	return nil
}