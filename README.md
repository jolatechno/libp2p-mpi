# libp2p-mpi

### __WARNING : Development in progress, might contain bug__

### __Please download releases to avoid bug__

[![](https://img.shields.io/badge/project-IPFS-blue.svg?style=flat-square)](https://ipfs.io/)
[![](https://img.shields.io/badge/freenode-%23ipfs-blue.svg?style=flat-square)](http://webchat.freenode.net/?channels=%23ipfs)
[![standard-readme compliant](https://img.shields.io/badge/standard--readme-OK-green.svg?style=flat-square)](https://github.com/RichardLitt/standard-readme)

Message Passing Interface computing using libp2p as a back-bone to build computational pool.

## How to build ?

```
go build -o main
```

## Getting started

You should first launch the ipfs daemon with `ipfs daemon` wich will output :

```
API server listening on /ip4/127.0.0.1/tcp/5001
```

Where `/ip4/127.0.0.1/tcp/5001` is the url of the ipfs daemon.

You can then run the libp2p-mpi dameon using :

```
./main -ipfs-api  /ip4/127.0.0.1/tcp/5001 -ipfs-store SomeIpfsDirectory/
```

For example :

```
./main -ipfs-api  /ip4/127.0.0.1/tcp/5001 -ipfs-store QmRfk8DdfrPQUxxThhgRxpPYvoa9qpjwV1veqXaSYgrrWf/
```

## Usage

See [example](./example) for info on how to design an interpretor for ipfs-mpi.

### WARNING : Development in progress, might contain bug

# ToDo

__main changes:__

- [x] implement basic message passing using libp2p
- [ ] handle fault (like unexpected peer hangup)
- [ ] automatically delete unused file

__'cosmetic' changes:__

- [x] write an example

__long-term goals:__

- [ ] implement a _filecoin-like_ cryptocurrency rewarding computation
