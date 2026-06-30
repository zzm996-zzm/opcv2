ARG NODE_BUILDER_IMAGE=node:22-alpine
ARG NGINX_IMAGE=nginx:1.27-alpine

FROM ${NODE_BUILDER_IMAGE} AS builder

WORKDIR /app
ARG NPM_CONFIG_REGISTRY=https://registry.npmjs.org/
ARG VITE_API_BASE_URL=
ENV npm_config_registry=${NPM_CONFIG_REGISTRY}
ENV VITE_API_BASE_URL=${VITE_API_BASE_URL}
COPY package.json package-lock.json ./
RUN npm ci

COPY . .
RUN npm run build

FROM ${NGINX_IMAGE}
COPY nginx.conf /etc/nginx/conf.d/default.conf
COPY --from=builder /app/dist /usr/share/nginx/html

EXPOSE 80
