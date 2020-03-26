# ipfs-mpi-shell

[![](https://img.shields.io/badge/project-IPFS-blue.svg?style=flat-square)](https://ipfs.io/)
[![](https://img.shields.io/badge/freenode-%23ipfs-blue.svg?style=flat-square)](http://webchat.freenode.net/?channels=%23ipfs)
[![standard-readme compliant](https://img.shields.io/badge/standard--readme-OK-green.svg?style=flat-square)](https://github.com/RichardLitt/standard-readme)

> mpi plugin using go-ipfs and go-libp2p

## Usage

To create a new shell use `Shell, err := shell.NewShell(api_url)` where `api_url` is the url of the ipfs-mpi api.

`Shell.List(file)` will return `host, peers` where `host` is the host address and `peers` is a list of the addresses of all peers listening fore the `file` interpreter.

`Shell.Send(pid, expected, messages)` will return a list of response messages of length `expected`.

### WARNING : Development in progress, might contain bug

## License

MIT