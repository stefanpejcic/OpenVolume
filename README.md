# OpenVolume
Docker Volume plugin for [OpenPanel](https://openpanel.co)


### Installation

1. Update Docker daemon configuration to enable third-party volume plugins. Open the `/etc/docker/daemon.json` file in a text editor with administrative privileges and add the following configuration:

    ```json
    {
      "plugins": {
        "volumes": {
          "enabled": true,
          "plugin_dir": "/var/lib/docker/plugins",
          "scan_on_start": true
        }
      }
    }
    ```

2. Create a folder for the OpenVolume plugin:

    ```bash
    mkdir -p /etc/docker/plugins/openvolume
    ```

3. Download the OpenVolume plugin code:

    ```bash
    git clone https://github.com/stefanpejcic/OpenVolume /etc/docker/plugins/openvolume
    ```

4. Build the plugin binary. Navigate to the plugin directory and compile the Go code:

    ```bash
    cd /etc/docker/plugins/openvolume
    go build openvolume.go
    ```

5. Create the plugin using the compiled binary:

    ```bash
    docker plugin create openvolume /etc/docker/plugins/openvolume/openvolume
    ```

6. Enable the plugin:

    ```bash
    docker plugin enable openvolume
    ```

### Usage

You can now use the OpenVolume plugin to create volumes with specific sizes:


```bash
docker volume create --driver openvolume --name myvolume --opt size=1GB
```

Replace 'myvolume' with the desired name for your volume, and adjust the size as needed.


```bash
docker volume resize myvolume --size=5G
```


## TODO:
- `--private` flag to limit volume to a single contianer.
- 

