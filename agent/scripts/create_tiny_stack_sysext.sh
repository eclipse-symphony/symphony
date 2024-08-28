#!/usr/bin/env bash
##
## Copyright (c) Microsoft Corporation.
## Licensed under the MIT license.
## SPDX-License-Identifier: MIT
##
set -euo pipefail

export ARCH="${ARCH-x86-64}"
SCRIPTFOLDER="$(dirname "$(readlink -f "$0")")"
ONLY_CONTAINERD="${ONLY_CONTAINERD:-0}"
ONLY_DOCKER="${ONLY_DOCKER:-0}"

if [ $# -lt 2 ] || [ "$1" = "-h" ] || [ "$1" = "--help" ]; then
  echo "Usage: $0 DOCKER_VERSION SYSEXTNAME WASMTIME_VERSION"
  echo "The script will download the Docker release tar ball (e.g., for 20.10.13) and create a sysext squashfs image with the name SYSEXTNAME.raw in the current folder."
  echo "A temporary directory named SYSEXTNAME in the current folder will be created and deleted again."
  echo "All files in the sysext image will be owned by root."
  echo "The necessary systemd services will be created by this script, by default only docker.socket will be enabled."
  echo "To only package containerd without Docker, pass ONLY_CONTAINERD=1 as environment variable (current value is '${ONLY_CONTAINERD}')."
  echo "To only package Docker without containerd and runc, pass ONLY_DOCKER=1 as environment variable (current value is '${ONLY_DOCKER}')."
  echo "To use arm64 pass 'ARCH=arm64' as environment variable (current value is '${ARCH}')."
  "${SCRIPTFOLDER}"/bake.sh --help
  exit 1
fi

if [ "${ONLY_CONTAINERD}" = 1 ] && [ "${ONLY_DOCKER}" = 1 ]; then
  echo "Cannot set both ONLY_CONTAINERD and ONLY_DOCKER" >&2
  exit 1
fi

VERSION="$1"
SYSEXTNAME="$2"
WASM_VERSION="$3"

# The github release uses different arch identifiers, we map them here
# and rely on bake.sh to map them back to what systemd expects
if [ "${ARCH}" = "amd64" ] || [ "${ARCH}" = "x86-64" ]; then
  ARCH="x86_64"
elif [ "${ARCH}" = "arm64" ]; then
  ARCH="aarch64"
fi

# cat <<EOF | docker build --output . -
# FROM rust:latest as build
# RUN apt-get update && apt-get install -y pkg-config libsystemd-dev libdbus-glib-1-dev build-essential libelf-dev libseccomp-dev libclang-dev protobuf-compiler
# RUN cargo install \
# 	--git https://github.com/containerd/runwasi.git \
#     --bin containerd-shim-wasmtime-v1 \
#     --root /out \
#     containerd-shim-wasmtime
# FROM scratch
# COPY --from=build /out/bin /
# EOF

rm -f "docker-${VERSION}.tgz"
curl -o "docker-${VERSION}.tgz" -fsSL "https://download.docker.com/linux/static/stable/${ARCH}/docker-${VERSION}.tgz"

# rm -f "wasmtime-${WASM_VERSION}.tar.xz"
# curl -o "wasmtime-${WASM_VERSION}.tar.xz" -fsSL "https://github.com/bytecodealliance/wasmtime/releases/download/v${WASM_VERSION}/wasmtime-v${WASM_VERSION}-${ARCH}-linux.tar.xz"

# TODO: Also allow to consume upstream containerd and runc release binaries with their respective versions
rm -rf "${SYSEXTNAME}"
mkdir -p "${SYSEXTNAME}"
tar --force-local -xf "docker-${VERSION}.tgz" -C "${SYSEXTNAME}"
rm "docker-${VERSION}.tgz"
mkdir -p "${SYSEXTNAME}"/usr/bin
mv "${SYSEXTNAME}"/docker/* "${SYSEXTNAME}"/usr/bin/
rmdir "${SYSEXTNAME}"/docker
mkdir -p "${SYSEXTNAME}/usr/lib/systemd/system"
if [ "${ONLY_CONTAINERD}" = 1 ]; then
  rm "${SYSEXTNAME}/usr/bin/docker" "${SYSEXTNAME}/usr/bin/dockerd" "${SYSEXTNAME}/usr/bin/docker-init" "${SYSEXTNAME}/usr/bin/docker-proxy"
elif [ "${ONLY_DOCKER}" = 1 ]; then
  rm "${SYSEXTNAME}/usr/bin/containerd" "${SYSEXTNAME}/usr/bin/containerd-shim-runc-v2" "${SYSEXTNAME}/usr/bin/ctr" "${SYSEXTNAME}/usr/bin/runc"
  if [[ "${VERSION%%.*}" -lt 23 ]] ; then
    # Binary releases 23 and higher don't ship containerd-shim
    rm "${SYSEXTNAME}/usr/bin/containerd-shim"
  fi
fi

rm -f "bpftool-v7.2.0-amd64.tar.gz"
curl -o "bpftool-v7.2.0-amd64.tar.gz" -fsSL "https://github.com/libbpf/bpftool/releases/download/v7.2.0/bpftool-v7.2.0-amd64.tar.gz"
tar -zxvf "bpftool-v7.2.0-amd64.tar.gz" -C "${SYSEXTNAME}/usr/bin/"
chmod +x "${SYSEXTNAME}/usr/bin/bpftool"

rm -f "WasmEdge-0.13.5-manylinux2014_x86_64.tar.gz"
curl -o "WasmEdge-0.13.5-manylinux2014_x86_64.tar.gz" -fsSL "https://github.com/WasmEdge/WasmEdge/releases/download/0.13.5/WasmEdge-0.13.5-manylinux2014_x86_64.tar.gz"
tar -zxvf "WasmEdge-0.13.5-manylinux2014_x86_64.tar.gz"
mv WasmEdge-0.13.5-Linux/bin/* "${SYSEXTNAME}/usr/bin/"
mv WasmEdge-0.13.5-Linux/lib64/* "${SYSEXTNAME}/usr/lib/"

rm -f "containerd-shim-wasmedge-x86_64.tar.gz"
curl -o "containerd-shim-wasmedge-x86_64.tar.gz" -fsSL "https://github.com/containerd/runwasi/releases/download/containerd-shim-wasmedge%2Fv0.3.0/containerd-shim-wasmedge-x86_64.tar.gz"
tar -zxvf "containerd-shim-wasmedge-x86_64.tar.gz"
mv containerd-shim-wasmedge-v1 "${SYSEXTNAME}/usr/bin/"
mv containerd-shim-wasmedged-v1 "${SYSEXTNAME}/usr/bin/"
mv containerd-wasmedged "${SYSEXTNAME}/usr/bin/"


if [ "${ONLY_CONTAINERD}" != 1 ]; then
  cat > "${SYSEXTNAME}/usr/lib/systemd/system/docker.socket" <<-'EOF'
	[Unit]
	PartOf=docker.service
	Description=Docker Socket for the API
	[Socket]
	ListenStream=/var/run/docker.sock
	SocketMode=0660
	SocketUser=root
	SocketGroup=docker
	[Install]
	WantedBy=sockets.target
EOF
  mkdir -p "${SYSEXTNAME}/usr/lib/systemd/system/sockets.target.d"
  { echo "[Unit]"; echo "Upholds=docker.socket"; } > "${SYSEXTNAME}/usr/lib/systemd/system/sockets.target.d/10-docker-socket.conf"
  cat > "${SYSEXTNAME}/usr/lib/systemd/system/docker.service" <<-'EOF'
	[Unit]
	Description=Docker Application Container Engine
	After=containerd.service docker.socket network-online.target
	Wants=network-online.target
	Requires=containerd.service docker.socket
	[Service]
	Type=notify
	EnvironmentFile=-/run/flannel/flannel_docker_opts.env
	Environment=DOCKER_SELINUX=--selinux-enabled=true
	ExecStart=/usr/bin/dockerd --host=fd:// --containerd=/run/containerd/containerd.sock $DOCKER_SELINUX $DOCKER_OPTS $DOCKER_CGROUPS $DOCKER_OPT_BIP $DOCKER_OPT_MTU $DOCKER_OPT_IPMASQ
	ExecReload=/bin/kill -s HUP $MAINPID
	LimitNOFILE=1048576
	# Having non-zero Limit*s causes performance problems due to accounting overhead
	# in the kernel. We recommend using cgroups to do container-local accounting.
	LimitNPROC=infinity
	LimitCORE=infinity
	# Uncomment TasksMax if your systemd version supports it.
	# Only systemd 226 and above support this version.
	TasksMax=infinity
	TimeoutStartSec=0
	# set delegate yes so that systemd does not reset the cgroups of docker containers
	Delegate=yes
	# kill only the docker process, not all processes in the cgroup
	KillMode=process
	# restart the docker process if it exits prematurely
	Restart=on-failure
	StartLimitBurst=3
	StartLimitInterval=60s
	[Install]
	WantedBy=multi-user.target
EOF
fi
if [ "${ONLY_DOCKER}" != 1 ]; then
  cat > "${SYSEXTNAME}/usr/lib/systemd/system/containerd.service" <<-'EOF'
	[Unit]
	Description=containerd container runtime
	After=network.target
	[Service]
	Delegate=yes
	Environment=CONTAINERD_CONFIG=/usr/share/containerd/config.toml
	ExecStartPre=mkdir -p /run/docker/libcontainerd
	ExecStartPre=ln -fs /run/containerd/containerd.sock /run/docker/libcontainerd/docker-containerd.sock
	ExecStart=/usr/bin/containerd --config /usr/share/containerd/config.toml
	KillMode=process
	Restart=always
	# (lack of) limits from the upstream docker service unit
	LimitNOFILE=1048576
	LimitNPROC=infinity
	LimitCORE=infinity
	TasksMax=infinity
	[Install]
	WantedBy=multi-user.target
EOF

# cat > "${SYSEXTNAME}/usr/lib/systemd/system/containerd-wasmedge.service" <<-'EOF'
# 	[Unit]
# 	Description=containerd wasm shim service
# 	After=network.target
# 	[Service]
# 	ExecStart=/usr/bin/containerd-wasmedged
# 	Restart=always
# 	[Install]
# 	WantedBy=multi-user.target
# EOF

  mkdir -p "${SYSEXTNAME}/usr/lib/systemd/system/multi-user.target.d"
  { echo "[Unit]"; echo "Upholds=containerd.service"; } > "${SYSEXTNAME}/usr/lib/systemd/system/multi-user.target.d/10-containerd-service.conf"
  { echo "[Unit]"; echo "Upholds=containerd-wasmedge.service"; } > "${SYSEXTNAME}/usr/lib/systemd/system/multi-user.target.d/10-containerd-wasmedge-service.conf"
  mkdir -p "${SYSEXTNAME}/usr/share/containerd"
  cat > "${SYSEXTNAME}/usr/share/containerd/config.toml" <<-'EOF'
	version = 2
	# set containerd's OOM score
	oom_score = -999
	[plugins."io.containerd.grpc.v1.cri".containerd]
	[plugins."io.containerd.grpc.v1.cri".containerd.runtimes]
	[plugins."io.containerd.grpc.v1.cri".containerd.runtimes.wasmedge]
	runtime_type = "io.containerd.wasmedge.v1"
	[plugins."io.containerd.grpc.v1.cri".containerd.runtimes.wasmedge.options]
	BinaryName = "/usr/bin/containerd-shim-wasmedge-v1"
	[plugins."io.containerd.grpc.v1.cri".containerd.runtimes.runc]
	# setting runc.options unsets parent settings
	runtime_type = "io.containerd.runc.v2"
	[plugins."io.containerd.grpc.v1.cri".containerd.runtimes.runc.options]
	SystemdCgroup = true
EOF
  sed 's/SystemdCgroup = true/SystemdCgroup = false/g' "${SYSEXTNAME}/usr/share/containerd/config.toml" > "${SYSEXTNAME}/usr/share/containerd/config-cgroupfs.toml"
fi

# tar --force-local -xf "wasmtime-${WASM_VERSION}.tar.xz" -C "${SYSEXTNAME}"
# rm "wasmtime-${WASM_VERSION}.tar.xz"
# mkdir -p "${SYSEXTNAME}"/usr/bin
# mv "${SYSEXTNAME}"/"wasmtime-v${WASM_VERSION}-${ARCH}-linux"/wasmtime "${SYSEXTNAME}"/usr/bin/
# rm -r "${SYSEXTNAME}"/"wasmtime-v${WASM_VERSION}-${ARCH}-linux"

# mv ./containerd-shim-wasmtime-v1 "${SYSEXTNAME}"/usr/bin/containerd-shim-wasmtime-v1

rm -rf WasmEdge-0.13.5-Linux
rm bpftool-v7.2.0-amd64.tar.gz
rm containerd-shim-wasmedge-x86_64.tar.gz
rm WasmEdge-0.13.5-manylinux2014_x86_64.tar.gz

cp ../target/release/piccolo "${SYSEXTNAME}"/usr/bin/piccolo
{ echo "[Unit]"; echo "Upholds=piccolo.service"; } > "${SYSEXTNAME}/usr/lib/systemd/system/multi-user.target.d/10-piccolo-service.conf"
cat > "${SYSEXTNAME}/usr/lib/systemd/system/piccolo.service" <<-'EOF'
	[Unit]
	Description=Symphony Piccolo Agent
	[Service]
	ExecStart=/usr/bin/piccolo
	Restart=always
	[Install]
	WantedBy=multi-user.target
EOF

"${SCRIPTFOLDER}"/bake.sh "${SYSEXTNAME}"
rm -rf "${SYSEXTNAME}"
