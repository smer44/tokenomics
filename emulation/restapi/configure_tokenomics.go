// This file is safe to edit. Once it exists it will not be overwritten

package restapi

import (
	"crypto/tls"
	"net/http"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/runtime"
	"github.com/go-openapi/runtime/middleware"

	"emulation/restapi/operations"
	"emulation/restapi/operations/tokenomics"
)

//go:generate swagger generate server --target ../../emulation --name Tokenomics --spec ../swagger.yaml --principal models.Principal

func configureFlags(api *operations.TokenomicsAPI) {
	// api.CommandLineOptionsGroups = []swag.CommandLineOptionsGroup{ ... }
}

func configureAPI(api *operations.TokenomicsAPI) http.Handler {
	// configure the api here
	api.ServeError = errors.ServeError

	// Set your custom logger if needed. Default one is log.Printf
	// Expected interface func(string, ...interface{})
	//
	// Example:
	// api.Logger = log.Printf

	api.UseSwaggerUI()
	// To continue using redoc as your UI, uncomment the following line
	// api.UseRedoc()

	api.JSONConsumer = runtime.JSONConsumer()

	api.JSONProducer = runtime.JSONProducer()

	if api.GetOrderingAgentsIDHandler == nil {
		api.GetOrderingAgentsIDHandler = operations.GetOrderingAgentsIDHandlerFunc(func(params operations.GetOrderingAgentsIDParams) middleware.Responder {
			return middleware.NotImplemented("operation operations.GetOrderingAgentsID has not yet been implemented")
		})
	}
	if api.GetProducingAgentsIDHandler == nil {
		api.GetProducingAgentsIDHandler = operations.GetProducingAgentsIDHandlerFunc(func(params operations.GetProducingAgentsIDParams) middleware.Responder {
			return middleware.NotImplemented("operation operations.GetProducingAgentsID has not yet been implemented")
		})
	}
	if api.PostOrderingAgentsIDHandler == nil {
		api.PostOrderingAgentsIDHandler = operations.PostOrderingAgentsIDHandlerFunc(func(params operations.PostOrderingAgentsIDParams) middleware.Responder {
			return middleware.NotImplemented("operation operations.PostOrderingAgentsID has not yet been implemented")
		})
	}
	if api.PostProducingAgentsIDHandler == nil {
		api.PostProducingAgentsIDHandler = operations.PostProducingAgentsIDHandlerFunc(func(params operations.PostProducingAgentsIDParams) middleware.Responder {
			return middleware.NotImplemented("operation operations.PostProducingAgentsID has not yet been implemented")
		})
	}
	if api.GetSystemInfoHandler == nil {
		api.GetSystemInfoHandler = operations.GetSystemInfoHandlerFunc(func(params operations.GetSystemInfoParams) middleware.Responder {
			return middleware.NotImplemented("operation operations.GetSystemInfo has not yet been implemented")
		})
	}
	if api.ListProducersHandler == nil {
		api.ListProducersHandler = operations.ListProducersHandlerFunc(func(params operations.ListProducersParams) middleware.Responder {
			return middleware.NotImplemented("operation operations.ListProducers has not yet been implemented")
		})
	}
	if api.TokenomicsResetSystemHandler == nil {
		api.TokenomicsResetSystemHandler = tokenomics.ResetSystemHandlerFunc(func(params tokenomics.ResetSystemParams) middleware.Responder {
			return middleware.NotImplemented("operation tokenomics.ResetSystem has not yet been implemented")
		})
	}

	api.PreServerShutdown = func() {}

	api.ServerShutdown = func() {}

	return setupGlobalMiddleware(api.Serve(setupMiddlewares))
}

// The TLS configuration before HTTPS server starts.
func configureTLS(tlsConfig *tls.Config) {
	// Make all necessary changes to the TLS configuration here.
}

// As soon as server is initialized but not run yet, this function will be called.
// If you need to modify a config, store server instance to stop it individually later, this is the place.
// This function can be called multiple times, depending on the number of serving schemes.
// scheme value will be set accordingly: "http", "https" or "unix".
func configureServer(s *http.Server, scheme, addr string) {
}

// The middleware configuration is for the handler executors. These do not apply to the swagger.json document.
// The middleware executes after routing but before authentication, binding and validation.
func setupMiddlewares(handler http.Handler) http.Handler {
	return handler
}

// The middleware configuration happens before anything, this middleware also applies to serving the swagger.json document.
// So this is a good place to plug in a panic handling middleware, logging and metrics.
func setupGlobalMiddleware(handler http.Handler) http.Handler {
	return handler
}
