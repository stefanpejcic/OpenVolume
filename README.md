# OpenVolume
Docker Volume plugin for OpenPanel


## Installation

```bash
docker plugin create openvolume /path/to/binary
```

```bash
docker plugin enable openvolume
```



## Usage

```bash
docker volume create --driver openvolume --name myvolume --opt size=1GB
```
