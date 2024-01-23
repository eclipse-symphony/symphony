# Symphony polling agent (Piccolo)

_(last edit: 9/18/2023)_

Piccolo is a lightweight Symphony agent that can be installed on tiny edge devices. It’s about 4MB in size and requires about 430K memory at runtime. It connects back to Symphony control plane with an outbound HTTP/HTTPS connection and reconciles device state with the desired state from the control plane.

Piccolo is written in [Rust](https://www.rust-lang.org/).

## Build Piccolo binary

* Build debug target

  ```bash
  # under the repo root folder
  cd piccolo
  cargo build
  ```

* Build release target

  ```bash
  # under the repo root folder
  cd piccolo
  cargo build --release
  ```

## Related topics

* [Piccolo on Flatcar](./flatcar.md)
* [Piccolo end-to-end demo](./piccolo-wasm-e2e.md)