package echopen

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	v310 "github.com/richjyoung/echopen/openapi/v3.1.0"

	"github.com/labstack/echo/v4"
	"gopkg.in/yaml.v3"
)

type APIWrapper struct {
	Schema *v310.Document
	Engine *echo.Echo
}

func New(title string, apiVersion string, config ...WrapperConfigFunc) *APIWrapper {
	wrapper := &APIWrapper{
		Schema: v310.NewDocument(),
		Engine: echo.New(),
	}

	wrapper.Schema.Info.Title = title
	wrapper.Schema.Info.Version = apiVersion
	wrapper.Engine.HTTPErrorHandler = DefaultErrorHandler

	for _, configFunc := range config {
		wrapper = configFunc(wrapper)
	}

	return wrapper
}

func (w *APIWrapper) ServeYAMLSchema(path string) *echo.Route {
	buf, err := yaml.Marshal(w.Schema)

	var handler echo.HandlerFunc = func(c echo.Context) error {
		if err != nil {
			return err
		}
		return c.Blob(http.StatusOK, "application/yaml", buf)
	}

	// Attach directly to the echo engine so the schema is not visible in the schema
	return w.Engine.GET(path, handler)
}

func (w *APIWrapper) ServeJSONSchema(path string) *echo.Route {
	buf, err := json.Marshal(w.Schema)

	var handler echo.HandlerFunc = func(c echo.Context) error {
		if err != nil {
			return err
		}
		return c.Blob(http.StatusOK, "application/json", buf)
	}

	// Attach directly to the echo engine so the schema is not visible in the schema
	return w.Engine.GET(path, handler)
}

func (w *APIWrapper) ServeUI(path string, schemaPath string, uiVersion string) *echo.Route {
	return w.Engine.GET(path, func(c echo.Context) error {
		return c.HTML(http.StatusOK, fmt.Sprintf(`
			<!DOCTYPE html>
			<html lang="en">
				<head>
					<meta charset="UTF-8">
					<title>%[1]s</title>
					<link rel="stylesheet" type="text/css" href="https://cdnjs.cloudflare.com/ajax/libs/swagger-ui/%[3]s/swagger-ui.min.css" />
				</head>

				<body>
					<div id="swagger-ui"></div>
					<script src="https://cdnjs.cloudflare.com/ajax/libs/swagger-ui/%[3]s/swagger-ui-bundle.min.js" charset="UTF-8"> </script>
					<script src="https://cdnjs.cloudflare.com/ajax/libs/swagger-ui/%[3]s/swagger-ui-standalone-preset.min.js" charset="UTF-8"> </script>
					<script>
						window.onload = function() {
							window.ui = SwaggerUIBundle({
								url: "%[2]s",
								dom_id: '#swagger-ui',
								deepLinking: true,
								presets: [
									SwaggerUIBundle.presets.apis,
									SwaggerUIStandalonePreset
								],
								plugins: [
									SwaggerUIBundle.plugins.DownloadUrl
								],
								layout: "StandaloneLayout"
							});
						};
					</script>
				</body>
			</html>
		`, w.Schema.Info.Title, schemaPath, uiVersion))
	})
}

func (w *APIWrapper) Start(addr string) error {
	return w.Engine.Start(addr)
}

func (w *APIWrapper) Licence(lic *v310.License) {
	w.Schema.Info.License = lic
}

func (w *APIWrapper) Contact(c *v310.Contact) {
	w.Schema.Info.Contact = c
}

func (w *APIWrapper) Add(method string, path string, handler echo.HandlerFunc, config ...RouteConfigFunc) *RouteWrapper {
	// Construct a new operation for this path and method
	op := &v310.Operation{}

	// Convert echo format to OpenAPI path
	oapiPath := echoRouteToOpenAPI(path)

	// Get the PathItem for this route
	pathItemRef, ok := w.Schema.Paths[oapiPath]
	if !ok {
		pathItemRef = &v310.Ref[v310.PathItem]{Value: &v310.PathItem{}}
		w.Schema.Paths[oapiPath] = pathItemRef
	}
	pathItem := pathItemRef.Value

	// Find or create the path item for this entry
	switch strings.ToLower(method) {
	case "delete":
		pathItem.Delete = op
	case "get":
		pathItem.Get = op
	case "head":
		pathItem.Head = op
	case "options":
		pathItem.Options = op
	case "patch":
		pathItem.Patch = op
	case "post":
		pathItem.Post = op
	case "put":
		pathItem.Put = op
	case "trace":
		pathItem.Trace = op
	default:
		panic(fmt.Sprintf("echopen: unknown method %s", method))
	}

	// Start populating return wrapper
	wrapper := &RouteWrapper{
		API:       w,
		Operation: op,
		PathItem:  pathItem,
		Handler:   handler,
	}

	// Set default operation ID
	wrapper.Operation.OperationID = genOpID(method, path)

	// Apply config transforms
	for _, configFunc := range config {
		wrapper = configFunc(wrapper)
	}

	// Add validation middleware to the start of the chain
	middlewares := append([]echo.MiddlewareFunc{wrapper.validationMiddleware()}, wrapper.Middlewares...)

	// Add the route in to the echo engine
	wrapper.Route = w.Engine.Add(method, path, wrapper.Handler, middlewares...)

	// Give the echo route the same name
	wrapper.Route.Name = wrapper.Operation.OperationID

	return wrapper
}

func (w *APIWrapper) Group(prefix string, config ...GroupConfigFunc) *GroupWrapper {
	wrapper := &GroupWrapper{
		Prefix: prefix,
		API:    w,
	}

	// Apply config transforms
	for _, configFunc := range config {
		wrapper = configFunc(wrapper)
	}

	group := w.Engine.Group(prefix, wrapper.Middlewares...)
	wrapper.Group = group
	return wrapper
}

func (w *APIWrapper) CONNECT(path string, handler echo.HandlerFunc, config ...RouteConfigFunc) *RouteWrapper {
	return w.Add("CONNECT", path, handler, config...)
}

func (w *APIWrapper) DELETE(path string, handler echo.HandlerFunc, config ...RouteConfigFunc) *RouteWrapper {
	return w.Add("DELETE", path, handler, config...)
}

func (w *APIWrapper) GET(path string, handler echo.HandlerFunc, config ...RouteConfigFunc) *RouteWrapper {
	return w.Add("GET", path, handler, config...)
}

func (w *APIWrapper) HEAD(path string, handler echo.HandlerFunc, config ...RouteConfigFunc) *RouteWrapper {
	return w.Add("HEAD", path, handler, config...)
}

func (w *APIWrapper) OPTIONS(path string, handler echo.HandlerFunc, config ...RouteConfigFunc) *RouteWrapper {
	return w.Add("OPTIONS", path, handler, config...)
}

func (w *APIWrapper) PATCH(path string, handler echo.HandlerFunc, config ...RouteConfigFunc) *RouteWrapper {
	return w.Add("PATCH", path, handler, config...)
}

func (w *APIWrapper) POST(path string, handler echo.HandlerFunc, config ...RouteConfigFunc) *RouteWrapper {
	return w.Add("POST", path, handler, config...)
}

func (w *APIWrapper) PUT(path string, handler echo.HandlerFunc, config ...RouteConfigFunc) *RouteWrapper {
	return w.Add("PUT", path, handler, config...)
}

func (w *APIWrapper) TRACE(path string, handler echo.HandlerFunc, config ...RouteConfigFunc) *RouteWrapper {
	return w.Add("TRACE", path, handler, config...)
}

func (w *APIWrapper) ErrorHandler(h echo.HTTPErrorHandler) {
	w.Engine.HTTPErrorHandler = h
}
