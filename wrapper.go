package echopen

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"reflect"
	"strings"

	v310 "github.com/richjyoung/echopen/openapi/v3.1.0"

	"github.com/labstack/echo/v4"
	"gopkg.in/yaml.v3"
)

type Config struct {
	BaseURL                  string
	DisableDefaultMiddleware bool
}

type APIWrapper struct {
	Spec   *v310.Specification
	Engine *echo.Echo
	Config *Config

	schemaMap map[reflect.Type]string
}

func New(title string, apiVersion string, config ...WrapperConfigFunc) *APIWrapper {
	wrapper := &APIWrapper{
		Spec:   v310.NewSpecification(),
		Engine: echo.New(),
		Config: &Config{},

		schemaMap: map[reflect.Type]string{},
	}

	wrapper.Spec.Info.Title = title
	wrapper.Spec.Info.Version = apiVersion
	wrapper.Engine.HTTPErrorHandler = DefaultErrorHandler

	for _, configFunc := range config {
		wrapper = configFunc(wrapper)
	}

	return wrapper
}

func (w *APIWrapper) WriteYAMLSpec(path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}

	buf, err := yaml.Marshal(w.Spec)
	if err != nil {
		return err
	}

	f.Write([]byte("# Specification generated by echOpen\n\n"))
	f.Write(buf)

	return f.Close()
}

func (w *APIWrapper) ServeYAMLSpec(path string, filters ...SpecFilterFunc) *echo.Route {
	s := w.Spec

	if len(filters) > 0 {
		s = w.Spec.Copy()
		for _, f := range filters {
			s = f(s)
		}
	}

	buf, err := yaml.Marshal(s)

	var handler echo.HandlerFunc = func(c echo.Context) error {
		if err != nil {
			return err
		}
		return c.Blob(http.StatusOK, "application/yaml", buf)
	}

	// Attach directly to the echo engine so the schema is not visible in the schema
	return w.Engine.GET(path, handler)
}

func (w *APIWrapper) ServeJSONSpec(path string, filters ...SpecFilterFunc) *echo.Route {
	s := w.Spec

	if len(filters) > 0 {
		s = w.Spec.Copy()
		for _, f := range filters {
			s = f(s)
		}
	}

	buf, err := json.Marshal(s)

	var handler echo.HandlerFunc = func(c echo.Context) error {
		if err != nil {
			return err
		}
		return c.Blob(http.StatusOK, "application/json", buf)
	}

	// Attach directly to the echo engine so the schema is not visible in the schema
	return w.Engine.GET(path, handler)
}

func (w *APIWrapper) ServeSwaggerUI(path string, schemaPath string, version string) *echo.Route {
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
		`, w.Spec.Info.Title, schemaPath, version))
	})
}

func (w *APIWrapper) ServeRapidoc(path string, schemaPath string) *echo.Route {
	return w.Engine.GET(path, func(c echo.Context) error {
		return c.HTML(http.StatusOK, fmt.Sprintf(`
			<!doctype html> <!-- Important: must specify -->
			<html>
			<head>
				<meta charset="utf-8"> <!-- Important: rapi-doc uses utf8 characters -->
				<script type="module" src="https://unpkg.com/rapidoc/dist/rapidoc-min.js"></script>
			</head>
			<body>
				<rapi-doc
					spec-url="%[1]s"
					theme = "dark"
				> </rapi-doc>
			</body>
			</html>
		`, schemaPath))
	})
}

// Start starts an HTTP server
func (w *APIWrapper) Start(addr string) error {
	return w.Engine.Start(addr)
}

// Register a new route with the given method and path
func (w *APIWrapper) Add(method string, path string, handler echo.HandlerFunc, config ...RouteConfigFunc) *RouteWrapper {
	// Construct a new operation for this path and method
	op := &v310.Operation{}

	// Convert echo format to OpenAPI path
	oapiPath := echoRouteToOpenAPI(path)

	// Get full path from configured Base URL
	fullPath := w.Config.BaseURL + path

	// Get the PathItem for this route
	pathItemRef, ok := w.Spec.Paths[oapiPath]
	if !ok {
		pathItemRef = &v310.Ref[v310.PathItem]{Value: &v310.PathItem{}}
		w.Spec.Paths[oapiPath] = pathItemRef
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
		API:               w,
		Operation:         op,
		PathItem:          pathItem,
		Handler:           handler,
		RequestBodySchema: map[string]*v310.Schema{},
	}

	// Set default operation ID
	wrapper.Operation.OperationID = genOpID(method, path)

	// Apply config transforms
	for _, configFunc := range config {
		wrapper = configFunc(wrapper)
	}

	// Add validation middleware to the start of the chain
	middlewares := []echo.MiddlewareFunc{}
	if !w.Config.DisableDefaultMiddleware {
		middlewares = append(middlewares, wrapper.middleware())
	}
	middlewares = append(middlewares, wrapper.Middlewares...)

	// Add the route in to the echo engine
	wrapper.Route = w.Engine.Add(method, fullPath, wrapper.Handler, middlewares...)

	// Give the echo route the same name
	wrapper.Route.Name = wrapper.Operation.OperationID

	return wrapper
}

// Create a new group with prefix and optional group-specific configuration
func (w *APIWrapper) Group(prefix string, config ...GroupConfigFunc) *GroupWrapper {
	wrapper := &GroupWrapper{
		Prefix: prefix,
		API:    w,
	}

	// Apply config transforms
	for _, configFunc := range config {
		wrapper = configFunc(wrapper)
	}

	fullPath := w.Config.BaseURL + prefix

	group := w.Engine.Group(fullPath, wrapper.Middlewares...)
	wrapper.RouterGroup = group
	return wrapper
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

// Extend the default echo handler to cover errors defined by echopen
func DefaultErrorHandler(err error, c echo.Context) {
	if errors.Is(err, ErrSecurityRequirementsNotMet) {
		c.JSON(http.StatusUnauthorized, map[string]interface{}{
			"message": http.StatusText(http.StatusUnauthorized),
		})
	} else if errors.Is(err, ErrRequiredParameterMissing) {
		c.JSON(http.StatusBadRequest, map[string]interface{}{
			"message": http.StatusText(http.StatusBadRequest),
		})
	} else if errors.Is(err, ErrContentTypeNotSupported) {
		c.JSON(http.StatusUnsupportedMediaType, map[string]interface{}{
			"message": http.StatusText(http.StatusUnsupportedMediaType),
		})
	} else if he, ok := err.(*echo.HTTPError); ok {
		if c.Echo().Debug && he.Internal != nil {
			c.JSON(he.Code, map[string]interface{}{
				"message": he.Internal.Error(),
			})
		} else {
			c.JSON(he.Code, map[string]interface{}{
				"message": he.Message,
			})
		}
	} else {
		c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"message": http.StatusText(http.StatusInternalServerError),
		})
	}
}
