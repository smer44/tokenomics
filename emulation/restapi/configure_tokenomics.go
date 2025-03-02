// This file is safe to edit. Once it exists it will not be overwritten

package restapi

import (
	"crypto/tls"
	"net/http"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/runtime"
	"github.com/go-openapi/runtime/middleware"
	"github.com/samber/lo"

	"emulation/application"
	"emulation/domain"
	"emulation/models"
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

	emulator := application.NewEmulator()

	api.TokenomicsCompleteCycleHandler = tokenomics.CompleteCycleHandlerFunc(func(params tokenomics.CompleteCycleParams) middleware.Responder {
		result, err := emulator.CompleteCycle()
		if err != nil {
			return middleware.Error(http.StatusBadRequest, err.Error())
		}
		return tokenomics.NewCompleteCycleOK().WithPayload(&models.CycleResult{
			lo.ToPtr(int64(result.Score)),
		})
	})

	api.GetOrderingAgentViewHandler = operations.GetOrderingAgentViewHandlerFunc(func(params operations.GetOrderingAgentViewParams) middleware.Responder {
		result, err := emulator.GetOrderingAgentView(domain.OrderingAgentId(params.ID))
		if err != nil {
			return middleware.Error(http.StatusBadRequest, err.Error())
		}
		return operations.NewGetOrderingAgentViewOK().WithPayload(&models.OrderingAgentView{
			lo.MapEntries(result.Incoming, func(oid domain.OrderId, val map[domain.CapacityType]domain.Capacity) (string, map[string]int64) {
				return string(oid), lo.MapEntries(val, func(ct domain.CapacityType, cap domain.Capacity) (string, int64) {
					return string(ct), int64(cap)
				})
			}),
			lo.MapEntries(result.Producers, func(ct domain.CapacityType, val map[domain.ProducerId]domain.ProducerInfo) (string, map[string]models.ProducingAgentInfo) {
				return string(ct), lo.MapEntries(val, func(pId domain.ProducerId, pInfo domain.ProducerInfo) (string, models.ProducingAgentInfo) {
					return string(pId), models.ProducingAgentInfo{
						int64(pInfo.Capacity),
						string(pInfo.CapacityType),
						int64(pInfo.CutOffPrice),
						string(pInfo.Id),
						int64(pInfo.MaxCapacity),
					}
				})
			}),
		})
	})

	api.GetProducingAgentViewHandler = operations.GetProducingAgentViewHandlerFunc(func(params operations.GetProducingAgentViewParams) middleware.Responder {
		result, err := emulator.GetProducingAgentView(domain.ProducerId(params.ID))
		if err != nil {
			return middleware.Error(http.StatusBadRequest, err.Error())
		}
		return operations.NewGetProducingAgentViewOK().WithPayload(&models.ProducingAgentView{
			int64(result.Capacity),
			int64(result.Degradation),
			string(result.Id),
			int64(result.MaxCapacity),
			int64(result.RequestedCapacity),
			int64(result.Restoration),
			bool(result.RestorationRunning),
			int64(result.Upgrade),
			bool(result.UpgradeRunning),
		})
	})

	api.GetSystemInfoHandler = operations.GetSystemInfoHandlerFunc(func(params operations.GetSystemInfoParams) middleware.Responder {
		info := emulator.GetSystemInfo()
		return operations.NewGetSystemInfoOK().WithPayload([]*models.SystemInfo{
			&info,
		})
	})

	api.ListOrderingAgentsHandler = operations.ListOrderingAgentsHandlerFunc(func(params operations.ListOrderingAgentsParams) middleware.Responder {
		orderingAgents := emulator.GetOrderingAgentInfos()
		result := make([]*models.OrderingAgentInfo, 0, len(orderingAgents))
		for _, info := range orderingAgents {
			result = append(result, &models.OrderingAgentInfo{
				string(info.Id),
			})
		}
		return operations.NewListOrderingAgentsOK().WithPayload(result)
	})

	api.ListProducingAgentsHandler = operations.ListProducingAgentsHandlerFunc(func(params operations.ListProducingAgentsParams) middleware.Responder {
		producerInfos := emulator.GetProducerInfos()
		result := make([]*models.ProducingAgentInfo, 0, len(producerInfos))
		for _, info := range producerInfos {
			result = append(result, &models.ProducingAgentInfo{
				int64(info.Capacity),
				string(info.CapacityType),
				int64(info.CutOffPrice),
				string(info.Id),
				int64(info.MaxCapacity),
			})
		}
		return operations.NewListProducingAgentsOK().WithPayload(result)
	})

	api.TokenomicsResetSystemHandler = tokenomics.ResetSystemHandlerFunc(func(params tokenomics.ResetSystemParams) middleware.Responder {
		emulator.Reset()
		return tokenomics.NewResetSystemOK()
	})

	api.SendOrderingAgentCommandHandler = operations.SendOrderingAgentCommandHandlerFunc(func(params operations.SendOrderingAgentCommandParams) middleware.Responder {
		err := emulator.OrderingAgentAction(domain.OrderingAgentId(params.ID), domain.OrderingAgentCommand{
			lo.MapEntries(params.Body.Orders, func(orderId string, producers map[string]int64) (domain.OrderId, map[domain.ProducerId]domain.Tokens) {
				return domain.OrderId(orderId), lo.MapEntries(producers, func(producerId string, tokens int64) (domain.ProducerId, domain.Tokens) {
					return domain.ProducerId(producerId), domain.Tokens(tokens)
				})
			}),
		})
		if err != nil {
			return middleware.Error(http.StatusBadRequest, err.Error())
		}
		return operations.NewSendOrderingAgentCommandOK()
	})

	api.SendProducingAgentCommandHandler = operations.SendProducingAgentCommandHandlerFunc(func(params operations.SendProducingAgentCommandParams) middleware.Responder {
		err := emulator.ProducingAgentAction(domain.ProducerId(params.ID), domain.ProducingAgentCommand{
			params.Body.DoRestoration,
			params.Body.DoUpgrade,
		})
		if err != nil {
			return middleware.Error(http.StatusBadRequest, err.Error())
		}
		return operations.NewSendProducingAgentCommandOK()
	})

	api.TokenomicsStartOrderingHandler = tokenomics.StartOrderingHandlerFunc(func(params tokenomics.StartOrderingParams) middleware.Responder {
		err := emulator.StartOrdering()
		if err != nil {
			return middleware.Error(http.StatusBadRequest, err.Error())
		}
		return tokenomics.NewStartOrderingOK()
	})

	api.GetConfigHandler = operations.GetConfigHandlerFunc(func(params operations.GetConfigParams) middleware.Responder {
		config := emulator.GetConfig()
		if config == nil {
			return middleware.Error(http.StatusNotFound, "configuration not found")
		}

		return operations.NewGetConfigOK().WithPayload(&models.Configuration{
			CycleEmission: lo.ToPtr(int64(config.CycleEmission)),
			ProcessSheets: lo.Map(config.ProcessSheets, func(ps domain.ProcessSheet, _ int) *models.ProcessSheet {
				return &models.ProcessSheet{
					Product: lo.ToPtr(int64(ps.Product)),
					Require: lo.MapEntries(ps.Require, func(ct domain.CapacityType, cap domain.Capacity) (string, int64) {
						return string(ct), int64(cap)
					}),
				}
			}),
			ProducerConfigs: lo.Map(config.ProducerConfigs, func(pc domain.ProducingAgentConfig, _ int) *models.ProducingAgentConfig {
				return &models.ProducingAgentConfig{
					ID:          lo.ToPtr(string(pc.Id)),
					Type:        lo.ToPtr(string(pc.Type)),
					Capacity:    lo.ToPtr(int64(pc.Capacity)),
					Degradation: lo.ToPtr(int64(pc.Degradation)),
					Restoration: struct{}{},
					Upgrade: &models.Upgrade{
						Product:  int64(pc.Upgrade.Require),
						Capacity: int64(pc.Upgrade.Increases),
					},
				}
			}),
		})
	})

	api.UpdateConfigHandler = operations.UpdateConfigHandlerFunc(func(params operations.UpdateConfigParams) middleware.Responder {
		config := &domain.Configuration{
			CycleEmission: domain.Tokens(lo.FromPtr(params.Body.CycleEmission)),
			ProcessSheets: lo.Map(params.Body.ProcessSheets, func(ps *models.ProcessSheet, _ int) domain.ProcessSheet {
				return domain.ProcessSheet{
					Product: domain.Product(lo.FromPtr(ps.Product)),
					Require: lo.MapEntries(ps.Require, func(ct string, cap int64) (domain.CapacityType, domain.Capacity) {
						return domain.CapacityType(ct), domain.Capacity(cap)
					}),
				}
			}),
			ProducerConfigs: lo.Map(params.Body.ProducerConfigs, func(pc *models.ProducingAgentConfig, _ int) domain.ProducingAgentConfig {
				return domain.ProducingAgentConfig{
					Id:          domain.ProducerId(lo.FromPtr(pc.ID)),
					Type:        domain.CapacityType(lo.FromPtr(pc.Type)),
					Capacity:    domain.Capacity(lo.FromPtr(pc.Capacity)),
					Degradation: domain.DegradationRate(lo.FromPtr(pc.Degradation)),
					Restoration: domain.Restoration{},
					Upgrade: domain.Upgrade{
						Require:   domain.Product(pc.Upgrade.Product),
						Increases: domain.Capacity(pc.Upgrade.Capacity),
					},
				}
			}),
		}

		if err := config.Validate(); err != nil {
			return middleware.Error(http.StatusBadRequest, err.Error())
		}

		if err := emulator.UpdateConfig(config); err != nil {
			return middleware.Error(http.StatusBadRequest, err.Error())
		}

		return operations.NewUpdateConfigOK()
	})

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
