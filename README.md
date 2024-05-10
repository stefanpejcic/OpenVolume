# OpenVolume
Docker Volume plugin for OpenPanel


## Installation

Create folder:
```bash
mkdir -p /etc/docker/plugins/openvolume
```

Download this:
```bash
git clone https://github.com/stefanpejcic/OpenVolume /etc/docker/plugins/openvolume
```

Create plugin:
```bash
docker plugin create openvolume /path/to/binary
```

Enable plugin:
```bash
docker plugin enable openvolume
```



## Usage

```bash
docker volume create --driver openvolume --name myvolume --opt size=1GB
```
