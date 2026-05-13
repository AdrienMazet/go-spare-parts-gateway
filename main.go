package main

import (
	"log"
	"net/http"

	nethttpmiddleware "github.com/oapi-codegen/nethttp-middleware"

	"github.com/adrienmazet/go-spare-parts-gateway/api"
	"github.com/adrienmazet/go-spare-parts-gateway/internal/controllers"
	"github.com/adrienmazet/go-spare-parts-gateway/internal/repository"
	"github.com/adrienmazet/go-spare-parts-gateway/internal/service"
	"github.com/adrienmazet/go-spare-parts-gateway/internal/service/mapper"
)

func main() {
	sparePartsService := service.NewSparePartsService(
		repository.NewSparePartsRepo(), mapper.NewSparePartsMapper())

	sparePartsHandler := controllers.NewSparePartHandler(sparePartsService)

	mux := http.NewServeMux()
	controllers.SetupRoutes(mux, sparePartsHandler)

	swagger, err := api.GetSwagger()
	if err != nil {
		log.Fatal(err)
	}

	swagger.Servers = nil

	middleware := nethttpmiddleware.OapiRequestValidatorWithOptions(
		swagger,
		&nethttpmiddleware.Options{
			ErrorHandlerWithOpts: controllers.OpenAPIErrorHandler,
		},
	)

	handler := middleware(mux)

	if err := http.ListenAndServe(":8080", handler); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
