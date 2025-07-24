# Base image for building the Rust library
FROM ubuntu:22.04 AS build

# Install necessary build tools and dependencies
RUN apt-get update && apt-get install -y \
    gcc g++ make curl wget git \
    libssl-dev pkg-config \
    ca-certificates \
    build-essential protobuf-compiler \
    binutils \
    libc6-dev \
    zlib1g-dev \
    && apt-get clean && rm -rf /var/lib/apt/lists/*

# Install Rust and rustup
RUN curl https://sh.rustup.rs -sSf | sh -s -- -y --no-modify-path \
    && . "$HOME/.cargo/env" \
    && rustup target add x86_64-unknown-linux-gnu

# Set up Rust environment
ENV PATH=/root/.cargo/bin:$PATH

# Work directory
WORKDIR /workspace

# Customers will mount their source code to /workspace and run cargo build
CMD ["sh", "-c", "cargo build --release"]