##
## Copyright (c) Microsoft Corporation.
## Licensed under the MIT license.
## SPDX-License-Identifier: MIT
##

# Use a base image with Rust installed
FROM rust:latest as builder

# Set the working directory
WORKDIR /app

# Copy your Rust project files to the container
COPY . .

# Build your Rust binary
RUN cargo build --release

# Create a minimal runtime image
FROM debian:bookworm-slim
WORKDIR /app
# RUN apt-get update && apt-get install -y libssl-dev

# Copy the compiled binary from the builder stage
COPY --from=builder /app/target/release/piccolo .

# Set the entry point for your application
CMD ["./piccolo"]