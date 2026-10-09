# Installation & Usage

## Build instruction

To build the project from source, you need Go installed on your system.

```bash
git clone https://github.com/bropines/awg-wireproxy.git
cd awg-wireproxy
make
```

The compiled binaries will be available in the `build/` directory. Check the `Makefile` for more build options (e.g., `make build-all` for cross-compiling).

## Install via Go

If you have Go configured, you can install the latest version directly:

```bash
go install github.com/bropines/awg-wireproxy/cmd/wireproxy@latest
```

## CLI Usage

```bash
./wireproxy [-c path to config]
```

```text
usage: wireproxy [-h|--help] [-c|--config "<value>"] [-s|--silent]
                 [-d|--daemon] [-i|--info "<value>"] [-v|--version]
                 [-n|--configtest]

                 Userspace wireguard client for proxying

Arguments:

  -h  --help        Print help information
  -c  --config      Path of configuration file
                    Default paths: /etc/wireproxy/wireproxy.conf, $HOME/.config/wireproxy.conf
  -s  --silent      Silent mode
  -d  --daemon      Make wireproxy run in background
  -i  --info        Specify the address and port for exposing health status
  -v  --version     Print version
  -n  --configtest  Configtest mode. Only check the configuration file for
                    validity.
```