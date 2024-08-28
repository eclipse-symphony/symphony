set MCUX_WORKSPACE_LOC=C:/Users/hbai/Documents/MCUXpressoIDE_11.9.0_2144/workspace
set MCUX_FLASH_DIR0=C:/nxp/LinkServer_1.4.85/binaries/Flash
set MCUX_FLASH_DIR1=C:/Users/hbai/Documents/MCUXpressoIDE_11.9.0_2144/workspace/.mcuxpressoide_packages_support/MIMXRT1176xxxxx_support/Flash
set LINKSERVER_BIN=C:/nxp/LinkServer_1.4.85/binaries
set MCUX_IDE_BIN=C:/nxp/MCUXpressoIDE_11.9.0_2144/ide/plugins/com.nxp.mcuxpresso.tools.bin.win32_11.9.0.202312111712/binaries/

%LINKSERVER_BIN%/crt_emu_cm_redlink --flash-load-exec "%MCUX_WORKSPACE_LOC%/evkmimxrt1170_hello_world_cm7/Release/evkmimxrt1170_hello_world_cm7.axf" -p MIMXRT1176xxxxx --ConnectScript RT1170_connect_M7_wake_M4.scp --ResetScript RT1170_reset.scp -ProbeHandle=1 -CoreIndex=0 --flash-driver= -x %MCUX_WORKSPACE_LOC%/evkmimxrt1170_hello_world_cm7/Release --flash-dir %MCUX_FLASH_DIR0% --flash-dir %MCUX_FLASH_DIR1% --no-packed --flash-hashing