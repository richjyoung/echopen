package echopen

import (
	"fmt"
	"strings"

	oa3 "github.com/getkin/kin-openapi/openapi3"
	"github.com/labstack/echo/v4"
)

type GroupWrapper struct {
	API                  *APIWrapper
	Prefix               string
	Middlewares          []echo.MiddlewareFunc
	Tags                 []string
	SecurityRequirements oa3.SecurityRequirements
	Group                *echo.Group
}

type GroupConfigFunc func(*GroupWrapper) *GroupWrapper

func (g *GroupWrapper) Add(method string, path string, handler echo.HandlerFunc, config ...RouteConfigFunc) *RouteWrapper {
	// Construct a new operation for this path and method
	op := &oa3.Operation{
		Responses: map[string]*oa3.ResponseRef{},
	}

	// Get full path from group
	fullPath := g.Prefix + path

	// Convert echo format to OpenAPI path
	oapiPath := echoRouteToOpenAPI(fullPath)

	// Get the PathItem for this route
	pathItem := g.API.Schema.Paths.Find(oapiPath)
	if pathItem == nil {
		pathItem = &oa3.PathItem{}
		g.API.Schema.Paths[oapiPath] = pathItem
	}

	// Find or create the path item for this entry
	switch strings.ToLower(method) {
	case "connect":
		pathItem.Connect = op
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
		API:       g.API,
		Group:     g,
		Operation: op,
		PathItem:  pathItem,
		Handler:   handler,
	}

	// Add group tags
	wrapper = WithTags(g.Tags...)(wrapper)

	for _, req := range g.SecurityRequirements {
		wrapper = WithSecurityRequirement(req)(wrapper)
	}

	// Apply config transforms
	for _, configFunc := range config {
		wrapper = configFunc(wrapper)
	}

	// Add the route in to the group (non-prefixed path)
	wrapper.Route = g.Group.Add(method, path, wrapper.Handler, wrapper.Middlewares...)

	// Ensure the operation ID is set, and the echo route is given the same name
	if wrapper.Operation.OperationID == "" {
		wrapper.Operation.OperationID = genOpID(method, fullPath)
	}
	wrapper.Route.Name = wrapper.Operation.OperationID

	return wrapper
}

func (g *GroupWrapper) CONNECT(path string, handler echo.HandlerFunc, config ...RouteConfigFunc) *RouteWrapper {
	return g.Add("CONNECT", path, handler, config...)
}

func (g *GroupWrapper) DELETE(path string, handler echo.HandlerFunc, config ...RouteConfigFunc) *RouteWrapper {
	return g.Add("DELETE", path, handler, config...)
}

func (g *GroupWrapper) GET(path string, handler echo.HandlerFunc, config ...RouteConfigFunc) *RouteWrapper {
	return g.Add("GET", path, handler, config...)
}

func (g *GroupWrapper) HEAD(path string, handler echo.HandlerFunc, config ...RouteConfigFunc) *RouteWrapper {
	return g.Add("HEAD", path, handler, config...)
}

func (g *GroupWrapper) OPTIONS(path string, handler echo.HandlerFunc, config ...RouteConfigFunc) *RouteWrapper {
	return g.Add("OPTIONS", path, handler, config...)
}

func (g *GroupWrapper) PATCH(path string, handler echo.HandlerFunc, config ...RouteConfigFunc) *RouteWrapper {
	return g.Add("PATCH", path, handler, config...)
}

func (g *GroupWrapper) POST(path string, handler echo.HandlerFunc, config ...RouteConfigFunc) *RouteWrapper {
	return g.Add("POST", path, handler, config...)
}

func (g *GroupWrapper) PUT(path string, handler echo.HandlerFunc, config ...RouteConfigFunc) *RouteWrapper {
	return g.Add("PUT", path, handler, config...)
}

func (g *GroupWrapper) TRACE(path string, handler echo.HandlerFunc, config ...RouteConfigFunc) *RouteWrapper {
	return g.Add("TRACE", path, handler, config...)
}

func WithEchoGroupMiddlewares(m ...echo.MiddlewareFunc) GroupConfigFunc {
	return func(gw *GroupWrapper) *GroupWrapper {
		gw.Middlewares = append(gw.Middlewares, m...)
		return gw
	}
}

func WithGroupTags(tags ...string) GroupConfigFunc {
	return func(gw *GroupWrapper) *GroupWrapper {
		gw.Tags = append(gw.Tags, tags...)
		return gw
	}
}

func WithGroupSecurityRequirement(req oa3.SecurityRequirement) GroupConfigFunc {
	return func(gw *GroupWrapper) *GroupWrapper {
		gw.SecurityRequirements = append(gw.SecurityRequirements, req)
		return gw
	}
}
