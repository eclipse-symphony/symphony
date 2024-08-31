# Step 1: Build the base image with node.js
FROM node:16-alpine as builder

# Set the working directory in the container
WORKDIR /app

# Copy package.json and package-lock.json
COPY package.json package-lock.json ./

# Install dependencies
RUN npm ci
# Copy the rest of the application code
COPY . .

# Build the application
RUN npm run build

# Step 2: Use a smaller base image for the production environment
FROM node:16-alpine

# Set the working directory in the container
WORKDIR /app

# Install the Next.js production server
RUN npm install next

# Copy the build artifacts from the builder stage
COPY --from=builder /app/public ./public
COPY --from=builder /app/.next ./.next
COPY --from=builder /app/node_modules ./node_modules
COPY --from=builder /app/package.json ./package.json

# Expose the port Next.js runs on
EXPOSE 3000

# Command to run the application
CMD ["npm", "start"]
