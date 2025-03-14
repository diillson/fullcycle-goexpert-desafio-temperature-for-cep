package services

import (
	"context"
	"encoding/json"
	"fmt"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

type ViaCEPService struct{}

func (s *ViaCEPService) GetCityByCEP(ctx context.Context, cep string) (string, error) {
	tracer := otel.Tracer("viacep-service")
	ctx, span := tracer.Start(ctx, "ViaCEP-GetCityByCEP")
	defer span.End()

	log.Printf("Buscando CEP: %s", cep)
	span.SetAttributes(attribute.String("cep", cep))

	if len(cep) != 8 {
		return "", fmt.Errorf("invalid zipcode")
	}

	url := fmt.Sprintf("https://viacep.com.br/ws/%s/json/", cep)
	log.Printf("Fazendo requisição para: %s", url)
	span.SetAttributes(attribute.String("url", url))

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		log.Printf("Erro ao criar requisição: %v", err)
		return "", err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Erro ao fazer requisição: %v", err)
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("Status code inválido: %d", resp.StatusCode)
		return "", fmt.Errorf("can not find zipcode")
	}

	// Ler o corpo da resposta como texto
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Erro ao ler corpo da resposta: %v", err)
		return "", fmt.Errorf("internal server error")
	}

	bodyString := string(bodyBytes)
	log.Printf("Resposta da API ViaCEP: %s", bodyString)

	// Verificar se a resposta indica um erro
	if strings.Contains(bodyString, "\"erro\"") || strings.Contains(bodyString, "\"erro\":true") {
		log.Printf("CEP não encontrado: resposta indica erro")
		return "", fmt.Errorf("can not find zipcode")
	}

	// Usar um mapa genérico para decodificar a resposta
	var responseMap map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &responseMap); err != nil {
		log.Printf("Erro ao decodificar resposta JSON: %v", err)
		return "", fmt.Errorf("internal server error")
	}

	// Verificar se o campo localidade existe e é uma string válida
	localidade, ok := responseMap["localidade"].(string)
	if !ok || localidade == "" {
		log.Printf("CEP sem localidade ou com localidade inválida")
		return "", fmt.Errorf("can not find zipcode")
	}

	log.Printf("Cidade encontrada: %s", localidade)
	span.SetAttributes(attribute.String("city", localidade))
	return localidade, nil
}
