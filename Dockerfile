ARG NODE_VERSION=24.7.0-alpine
ARG NGINX_VERSION=alpine3.22

# =========================================
# Stage 0: Build the GO WebAssembly Module
# =========================================
FROM golang:1.24.2 AS go-builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN env GOOS=js GOARCH=wasm go build -o ./src/assets/main.wasm

# =========================================
# Stage 1: Build the Angular Application
# =========================================

# Use a lightweight Node.js image for building (customizable via ARG)
FROM node:${NODE_VERSION} AS builder

# Set the working directory inside the container
WORKDIR /app

# Copy package-related files first to leverage Docker's caching mechanism
COPY package.json package-lock.json ./

# Install project dependencies using npm ci (ensures a clean, reproducible install)
RUN --mount=type=cache,target=/root/.npm npm ci

# Copy the rest of the application source code into the container
COPY . .

# Ensure the WebAssembly artifact built in the Go stage is available
# to the Node/Angular build. The go-builder stage produces
# /app/src/assets/main.wasm; copy it into this stage before building
# so it gets bundled into the Angular output.
COPY --from=go-builder /app/src/assets/main.wasm ./src/assets/main.wasm

# Build the Angular application
RUN npm run build:angular

# =========================================
# Stage 2: Prepare Nginx to Serve Static Files
# =========================================

FROM nginxinc/nginx-unprivileged:${NGINX_VERSION} AS runner

# Use a built-in non-root user for security best practices
USER nginx

# Copy custom Nginx config
COPY nginx.conf /etc/nginx/nginx.conf

# Copy the static build output from the build stage to Nginx's default HTML serving directory
COPY --chown=nginx:nginx --from=builder /app/dist/*/browser /usr/share/nginx/html

# Expose port 8080 to allow HTTP traffic
# Note: The default NGINX container now listens on port 8080 instead of 80 
EXPOSE 8080

# Start Nginx directly with custom config
ENTRYPOINT ["nginx", "-c", "/etc/nginx/nginx.conf"]
CMD ["-g", "daemon off;"]