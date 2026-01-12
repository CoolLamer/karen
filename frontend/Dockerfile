FROM node:20-alpine AS build
WORKDIR /app

ARG VITE_API_BASE_URL=http://localhost:8080
ARG VITE_HOTJAR_ID=
ENV VITE_API_BASE_URL=${VITE_API_BASE_URL}
ENV VITE_HOTJAR_ID=${VITE_HOTJAR_ID}

COPY frontend/package.json ./package.json
COPY frontend/package-lock.json ./package-lock.json
RUN if [ -f package-lock.json ]; then npm ci; else npm install; fi

COPY frontend/ ./
RUN npm run build

FROM nginx:1.27-alpine
COPY frontend/nginx.conf /etc/nginx/conf.d/default.conf
COPY --from=build /app/dist /usr/share/nginx/html

EXPOSE 80


