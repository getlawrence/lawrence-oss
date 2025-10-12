# Simple development container with hot-reloading
FROM node:20-alpine

WORKDIR /app

# Install pnpm globally for faster package management
RUN npm install -g pnpm

# Copy package files first (for better caching)
COPY package*.json ./

# Install dependencies
RUN pnpm install

# Expose Vite dev server port
EXPOSE 5173

# Start development server with hot-reload
CMD ["pnpm", "dev", "--host", "0.0.0.0"]
