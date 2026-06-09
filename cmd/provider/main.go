package main

import (
	"log"
	"net/http"
	"os"

	"github.com/adrienmazet/go-spare-parts-gateway/external/atdistri"
	"github.com/adrienmazet/go-spare-parts-gateway/external/globalcar"
	"github.com/adrienmazet/go-spare-parts-gateway/external/hexapart"
	"github.com/adrienmazet/go-spare-parts-gateway/external/partspec"
)

func main() {
	providerName := os.Getenv("PROVIDER_NAME")
	port := os.Getenv("PROVIDER_PORT")
	if port == "" {
		port = "8081"
	}

	handler, ok := handlerFor(providerName)
	if !ok {
		log.Fatalf("unknown provider %q", providerName)
	}

	if err := http.ListenAndServe(":"+port, handler); err != nil {
		log.Fatalf("provider server failed: %v", err)
	}
}

func handlerFor(providerName string) (http.Handler, bool) {
	switch providerName {
	case "atdistri":
		return atdistri.NewHandler(), true
	case "globalcar":
		return globalcar.NewHandler(), true
	case "hexapart":
		return hexapart.NewHandler(), true
	case "partspec":
		return partspec.NewHandler(), true
	default:
		return nil, false
	}
}
