# Piccolo on Flatcar

[Flatcar](https://www.flatcar.org/) is a container-optimized Linux distribution. Piccolo can be configured as a [Systemd-sysext](https://www.freedesktop.org/software/systemd/man/systemd-sysext.html) extension.

## Prepare Piccolo Flatcar extension

1. Build Piccolo for release.

   ```bash
   # from repo root folder
   cd piccolo
   cargo build --release
   ```

1. Copy Piccolo binary to staging folder (currently `0.0.1`).

   ```bash
   cp ./target/release/piccolo ./0.0.1/sysext/piccolo/usr/bin
   ```

1. Create a new Piccolo sysext image.

   ```bash
   cd 0.0.1/sysext
   mksquashfs piccolo piccolo.raw -all-root
   ```

1. Upload the `.raw` file to a GitHub release folder.

   The ignition file under the repo uses a temporary GitHub release at `https://github.com/eclipse-symphony/symphony/releases/download/vtest/piccolo.raw`. To use a different GitHub repo, you'll need to update the ignite.ign file and update the source folder to the repo you want to use.

1. Download Flatcar image if needed.

   ```bash
   wget https://stable.release.flatcar-linux.net/amd64-usr/current/flatcar_production_qemu_image.img.bz2
   bzip2 --decompress --keep flatcar_production_qemu_image.img.bz2
   ```

1. Copy the Flatcar image and the `ignition.ign` file to a `c:\demo\flatcar` folder.

1. Launch Flatcar with Piccolo extension in QEMU.

   ```powershell
   .\qemu-system-x86_64.exe -m 2G -netdev user,id=net0,hostfwd=tcp::2222-:22 -device virtio-net-pci,netdev=net0 -fw_cfg name=opt/org.flatcar-linux/config,file=c:\demo\flatcar\ignition.ign -drive if=virtio,file=c:\demo\flatcar_production_qemu_image.img
   ```

1. Once the Flatcar OS is booted, you can check the Piccolo service status:

   ```bash
   systemctl status piccolo # service should be active
   ```
