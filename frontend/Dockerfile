# Stage 1: Build the application
FROM node:20-alpine AS builder

# Set working directory
WORKDIR /app

# Copy package and configuration files
COPY package.json package-lock.json* ./
COPY next.config.mjs ./
COPY tsconfig.json ./
COPY postcss.config.mjs ./
COPY tailwind.config.ts ./

# Install dependencies
RUN npm install

# Copy source code
COPY public ./public
COPY src ./src

# Build the Next.js application
RUN npm run build

# Stage 2: Production image
FROM node:20-alpine AS runner

WORKDIR /app

# Set environment variables
ENV NODE_ENV=production
# Optionally, you can set NEXT_TELEMETRY_DISABLED=1 to disable telemetry
# ENV NEXT_TELEMETRY_DISABLED 1

# Copy built assets from the builder stage
COPY --from=builder /app/.next ./.next
COPY --from=builder /app/public ./public
COPY --from=builder /app/package.json ./package.json
# If you have a standalone output, you might need to copy node_modules differently
# or rely on the .next/standalone directory if you configure output: 'standalone' in next.config.mjs
# For now, we'll copy the whole node_modules from the builder, which is not ideal for size.
# A better approach for production would be to only copy necessary production dependencies
# or use the standalone output.
COPY --from=builder /app/node_modules ./node_modules
COPY --from=builder /app/next.config.mjs ./next.config.mjs


# Expose the port Next.js runs on
EXPOSE 3000

# Command to start the Next.js application
CMD ["npm", "start"]