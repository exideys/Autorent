FROM node:26-alpine AS builder

WORKDIR /app

COPY package*.json ./
RUN npm ci

COPY . .
ARG VITE_API_BASE_URL
ARG VITE_BASE_PATH
ARG VITE_GOOGLE_AUTH_CLIENT_ID
ENV VITE_API_BASE_URL=$VITE_API_BASE_URL
ENV VITE_BASE_PATH=$VITE_BASE_PATH
ENV VITE_GOOGLE_AUTH_CLIENT_ID=$VITE_GOOGLE_AUTH_CLIENT_ID
RUN npm run build

FROM nginx:1.31.2-alpine

COPY nginx.conf /etc/nginx/conf.d/default.conf
COPY --from=builder /app/dist /usr/share/nginx/html

EXPOSE 80
