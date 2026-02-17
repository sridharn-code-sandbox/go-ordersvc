FROM node:20-alpine AS build
WORKDIR /app
COPY package.json package-lock.json* ./
RUN npm ci
COPY . .
RUN npm run build

FROM nginx:1.27-alpine
COPY --from=build /app/dist /usr/share/nginx/html
COPY nginx/nginx.conf.template /etc/nginx/templates/default.conf.template
EXPOSE 80
ENV API_HOST=localhost API_PORT=8080
