package services

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

type ViaCEPService struct{}

type ViaCEPResponse struct {
	Localidade string `json:"localidade"`
	Erro       bool   `json:"erro"`
}

func (s *ViaCEPService) GetCityByCEP(cep string) (string, error) {
	log.Printf("Buscando CEP: %s", cep)

	if len(cep) != 8 {
		return "", fmt.Errorf("invalid zipcode")
	}

	url := fmt.Sprintf("https://viacep.com.br/ws/%s/json/", cep)
	log.Printf("Fazendo requisição para: %s", url)

	resp, err := http.Get(url)
	if err != nil {
		log.Printf("Erro ao fazer requisição: %v", err)
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("Status code inválido: %d", resp.StatusCode)
		return "", fmt.Errorf("can not find zipcode")
	}

	var cepResponse ViaCEPResponse
	if err := json.NewDecoder(resp.Body).Decode(&cepResponse); err != nil {
		log.Printf("Erro ao decodificar resposta: %v", err)
		return "", err
	}

	if cepResponse.Erro || cepResponse.Localidade == "" {
		log.Printf("CEP não encontrado ou resposta inválida")
		return "", fmt.Errorf("can not find zipcode")
	}

	log.Printf("Cidade encontrada: %s", cepResponse.Localidade)
	return cepResponse.Localidade, nil
}
