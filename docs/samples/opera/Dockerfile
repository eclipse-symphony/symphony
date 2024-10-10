# Step 1: Build the base image with Node.js
FROM node:18-alpine as builder

# Set the working directory in the container
WORKDIR /app

# Copy package.json and package-lock.json
COPY package.json package-lock.json ./

# Install dependencies (including dev dependencies for build)
RUN npm install --legacy-peer-deps

# Copy the rest of the application code
COPY . .

# Build the application
RUN npm run build

# Step 2: Use a smaller base image for the production environment
FROM node:18-alpine

# Set the working directory in the container
WORKDIR /app

# Set NODE_ENV to production
ENV NODE_ENV=production

# Copy the build artifacts from the builder stage
COPY --from=builder /app /app

# Install production dependencies
RUN npm install --only=production

# Expose the port Next.js runs on
EXPOSE 3000

# Command to run the application
CMD ["npm", "start"]
