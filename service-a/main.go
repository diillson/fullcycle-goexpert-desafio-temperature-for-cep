package main

import (
	"context"
	"encoding/json"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/zipkin"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	_ "go.opentelemetry.io/otel/trace"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

type CepRequest struct {
	Cep string `json:"cep"`
}

type WeatherResponse struct {
	City  string  `json:"city"`
	TempC float64 `json:"temp_C"`
	TempF float64 `json:"temp_F"`
	TempK float64 `json:"temp_K"`
}

func initTracer() (*sdktrace.TracerProvider, error) {
	zipkinURL := os.Getenv("ZIPKIN_URL")
	if zipkinURL == "" {
		zipkinURL = "http://zipkin:9411/api/v2/spans"
	}

	exporter, err := zipkin.New(zipkinURL)
	if err != nil {
		return nil, err
	}

	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String("service-a"),
		)),
	)

	otel.SetTracerProvider(tracerProvider)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	return tracerProvider, nil
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// Inicializa o tracer
	tp, err := initTracer()
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			log.Printf("Error shutting down tracer provider: %v", err)
		}
	}()

	http.HandleFunc("/weather", handleWeatherRequest)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Service-A starting on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func handleWeatherRequest(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tracer := otel.Tracer("service-a")
	ctx, span := tracer.Start(ctx, "HandleWeatherRequest")
	defer span.End()

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "only POST method is allowed"})
		return
	}

	w.Header().Set("Content-Type", "application/json")

	var req CepRequest
	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid request body"})
		return
	}

	if err := json.Unmarshal(body, &req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid request format"})
		return
	}

	cep := req.Cep
	span.SetAttributes(attribute.String("cep", cep))

	// Validar CEP
	if !isValidCEP(cep) {
		w.WriteHeader(http.StatusUnprocessableEntity)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid zipcode"})
		return
	}

	// Chamar o serviço B
	response, statusCode, err := callServiceB(ctx, cep)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "error calling service B"})
		return
	}

	// Repassar a resposta do serviço B
	w.WriteHeader(statusCode)
	w.Write(response)
}

func isValidCEP(cep string) bool {
	// Verificar se o CEP é uma string de 8 dígitos
	if len(cep) != 8 {
		return false
	}

	// Verificar se contém apenas dígitos
	for _, c := range cep {
		if c < '0' || c > '9' {
			return false
		}
	}

	return true
}

func callServiceB(ctx context.Context, cep string) ([]byte, int, error) {
	tracer := otel.Tracer("service-a")
	ctx, span := tracer.Start(ctx, "CallServiceB")
	defer span.End()

	span.SetAttributes(attribute.String("cep", cep))

	serviceBURL := os.Getenv("SERVICE_B_URL")
	if serviceBURL == "" {
		serviceBURL = "http://service-b:8081/weather"
	}

	reqData := CepRequest{Cep: cep}
	reqBody, err := json.Marshal(reqData)
	if err != nil {
		return nil, 0, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", serviceBURL, strings.NewReader(string(reqBody)))
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set("Content-Type", "application/json")

	client := http.Client{
		Transport: otelhttp.NewTransport(http.DefaultTransport),
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, 0, err
	}

	span.SetAttributes(attribute.Int("status_code", resp.StatusCode))

	return respBody, resp.StatusCode, nil
}
