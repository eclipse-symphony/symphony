# Piccolo on Flatcar

[Flatcar](https://www.flatcar.org/) is a container-optimized Linux distribution. Piccolo can be configured as a [Systemd-sysext](https://www.freedesktop.org/software/systemd/man/systemd-sysext.html) extension.

## Prepare Piccolo Flatcar extension

1. Build Piccolo for release
   ```bash
   # from repo root folder
   cd piccolo
   cargo build --release
   ```
2. Copy Piccolo binary to staging folder (currently ```0.0.1```)
   ```bash
   cp ./target/release/piccolo ./0.0.1/sysext/piccolo/usr/bin
   ```
3. Create new Piccolo sysext image
   ```
   cd 0.0.1/sysext
   mksquashfs piccolo piccolo.raw -all-root
   ```
4. Upload the ```.raw``` file to a GitHub release folder
   The ignition file under the repo uses a temporary GitHub release at ```https://github.com/Haishi2016/Vault818/releases/download/vtest/piccolo.raw```. To use a different GitHub repo, you'll need to update the ignite.ign file and update the source folder to the repo you want to use.
5. Download Flatcar image if needed
   ```bash
   wget https://stable.release.flatcar-linux.net/amd64-usr/current/flatcar_production_qemu_image.img.bz2
   bzip2 --decompress --keep flatcar_production_qemu_image.img.bz2
   ```
5. Launch Flatcar with Piccolo extension in QEMU:
   ```powershell
   .\qemu-system-x86_64.exe -m 2G -netdev user,id=net0,hostfwd=tcp::2222-:22 -device virtio-net-pci,netdev=net0 -fw_cfg name=opt/org.flatcar-linux/config,file=c:\demo\flatcar\ignition.ign -drive if=virtio,file=c:\demo\flatcar_production_qemu_image.img
   ```
   > **NOTE**: This assumes you've copied the Flatcar image and the ```ignition.ign``` file to a ```c:\demo\flatcar``` folder.

6. Once the Flatcar OS is booted, you can check the Piccolo service status:
   ```bash
   systemctl status piccolo # service shoudl be active
   ```