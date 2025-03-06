#!/bin/bash

# Configurações
PROJECT_ID="temperature-api-freitas"
REGION="us-central1"
SERVICE_NAME="temperature-api-freitas"

# Build e push da imagem
gcloud builds submit --tag gcr.io/$PROJECT_ID/$SERVICE_NAME

# Deploy no Cloud Run
gcloud run deploy $SERVICE_NAME \
  --image gcr.io/$PROJECT_ID/$SERVICE_NAME \
  --platform managed \
  --region $REGION \
  --allow-unauthenticated \
  --set-env-vars WEATHER_API_KEY=$WEATHER_API_KEY