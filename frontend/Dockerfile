FROM node:22-alpine AS build
WORKDIR /app
RUN npm install -g bun@1.4.4
COPY package.json bun.lock ./
RUN bun install --frozen-lockfile
COPY . .
ARG VITE_API_URL=""
ENV VITE_API_URL=$VITE_API_URL NITRO_PRESET=node-server
RUN npm run build

FROM node:22-alpine
WORKDIR /app
ENV NODE_ENV=production PORT=3000
COPY --from=build --chown=node:node /app/.output ./
USER node
EXPOSE 3000
CMD ["node", "server/index.mjs"]
