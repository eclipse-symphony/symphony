# Proxy Stage Processor

A proxy stage processor allows a stage to be processed in an isolated environment such as a separate process, container, virtual machine, or physic device. The proxy provider expects a web server that implements the required Symphony stage processor interface. Although you can use your own web server implementations, we recommend using the default Symphony implementation that supports all existing Symphony stage processors to be used over the proxy. 

## Launch a processor web server using Symphony API binary

You can launch a processor web server by launching the `symphony-api` process with a `symphony-processor-server.json` config:
```bash
./symphony-api -c ./symphony-processor-server.json -l Debug
```