FROM node:24-alpine
WORKDIR /app

# 1) Install deps
COPY package.json package-lock.json* ./
RUN npm ci

# 2) Copy your app
COPY next.config.mjs tsconfig.json postcss.config.mjs tailwind.config.ts ./
COPY public ./public
COPY src ./src

# 3) At container startup run build (with env injected by Compose) then start
ENTRYPOINT ["sh", "-c", "npm run build && npm run start"]