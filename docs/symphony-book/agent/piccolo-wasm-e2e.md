# Piccolo E2E Demo
This scenario demonstrates Piccolo's capabilities. You will deploy a WebAssembly module to [Flatcar OS](https://www.flatcar.org/) via cloud, then inject an [eBPF](https://ebpf.io/) module for diagnostics, monitoring the telemetry with [Prometheus](https://prometheus.io/). An approval workflow in Outlook precedes the eBPF deployment. The eBPF is given a time window to run and is then removed from the target.

## Prerequisites 
* **Symphony 0.45.33** or above is deployed to an AKS cluster - this will be your cloud-based control plane.
* **[kubectl](https://kubernetes.io/docs/reference/kubectl/kubectl/)** installed and configured with the above AKS cluster
* **Symphony repo** cloned on your demo machine - currently under the ```wasm-e2e``` branch.
* **[Cargo](https://doc.rust-lang.org/cargo/)** - for building Piccolo binaries.
* **[QEMU](https://www.qemu.org/)** - for running the Flatcar VM.

## Resource Requirements
| Binary | Size of Disk | Memory |
|--------|--------|--------|
| `bpftool` |34.16MB | |
| `containerd` |37.31MB ||
| `containerd-shim-wasmtime-v1`|25.8MB ||
| `ctr`|18.25MB | |
|`docker`|33.14MB | |
|`docker-init`|0.73MB| |
|`docker-proxy`|1.87MB||
|`dockerd`|60.38MB||
|`piccolo`|3.68MB||
|`runc`|14.44MB||
|`wasmtime`|47.57MB||
| **TOTAL**|277.33MB||

## Demo Preparation 
1. Download and uncompress Flatcar QEMU image:
```bash
wget https://stable.release.flatcar-linux.net/amd64-usr/current/flatcar_production_qemu_image.img.bz2
bunzip2 flatcar_production_qemu_image.img.bz2
```
> **NOTE**: [Flatcar Ignition](https://github.com/flatcar/ignition) happens only during first boot. To fully reset the demo, you'll need to restore to the original `.img` file.
2. 

## Demo Setup
1. Make sure your `kubectl` context is set to the correct Kubernetes cluster.
2. Open VSCode under the `docs/samples/piccolo/wasm-ebpf` folder.

## Demo Steps


### I. Examine Symphony artifacts ###

1. Open `target.yaml`. Point out that the targt uses a `providers.target.staging` provider, which stages the artifacts to be deployed on Symphony control plane instead of directly pushing them to the target. The artifact will be later picked up by a polling agent (Piccolo).

> **NOTE**: In many scenarios, you don't have direct access to the tiny edge devices. Hence a polling agent is used to communicate with the control plane through an outbound connection.

## Building maze web server
rustup target add wasm32-wasi
cargo build --target wasm32-wasi

## Building Piccolo
cargo build --release