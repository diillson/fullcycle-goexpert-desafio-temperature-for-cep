# Weather API

API RESTful para consulta de temperatura por CEP, que integra o ViaCEP para localização e WeatherAPI para dados meteorológicos.

## Requisitos

- Docker v20.10 ou superior
- Docker Compose v2.0 ou superior
- Conta no [WeatherAPI](https://www.weatherapi.com/) (API key gratuita)
- Conta no Google Cloud Platform (para deploy)

## URL DA APP NO GOOGLE CLOUD RUN

    https://temperature-api-freitas-432475449201.us-central1.run.app

## Tecnologias Utilizadas

- Go 1.24
- Gorilla Mux (Router)
- Docker
- Google Cloud Run

## Configuração Local

1. Clone o repositório:
```bash
git clone https://github.com/diillson/fullcycle-goexpert-desafio-temperature-for-cep.git
cd fullcycle-goexpert-desafio-temperature-for-cep
```

## Configure as variáveis de ambiente:
Crie um arquivo .env na raiz do projeto:
```
WEATHER_API_KEY=sua_api_key_aqui
```

## Execute a aplicação:
```bash
docker-compose up --build
```

## Endpoints

GET /weather/{cep}


Retorna a temperatura atual para a localidade do CEP informado.
Parâmetros:
```
cep: CEP válido com 8 dígitos (somente números)
```

# Exemplos de Requisições:

## CEP válido
curl http://localhost:8080/weather/22450000

## CEP inválido
curl http://localhost:8080/weather/123

## CEP não encontrado
curl http://localhost:8080/weather/99999999

Respostas:

### 200 OK: Sucesso

{
    "temp_C": 25.5,
    "temp_F": 77.9,
    "temp_K": 298.65
}

### 422 Unprocessable Entity: CEP inválido

{
    "error": "invalid zipcode"
}

### 404 Not Found: CEP não encontrado

{
    "error": "can not find zipcode"
}

# Testes

Execute os testes unitários:

### Local
go test -v ./...

### Via Docker
docker-compose run app go test -v ./...

# Deploy no Google Cloud Run

Configure o Google Cloud SDK:

    gcloud init
    gcloud auth configure-docker

Configure as variáveis de ambiente:

    export PROJECT_ID="seu-projeto-id"
    export REGION="us-central1"
    export WEATHER_API_KEY=Sua API KEY

Execute o deploy:

    ./deploy.sh

Acesse a aplicação:

    gcloud run services describe temperature-api --format='value(status.url)'

# Desenvolvimento
Estrutura do Projeto
```bash
weather-api/
├── handlers/        # Handlers HTTP
├── services/        # Lógica de negócios
├── models/          # Modelos de dados
├── main.go         # Ponto de entrada
├── Dockerfile      # Configuração Docker
├── docker-compose.yml
├── deploy.sh
└── README.md
```

Serviços Utilizados

    ViaCEP: API para consulta de endereços por CEP
    WeatherAPI: API para dados meteorológicos
