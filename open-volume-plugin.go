package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/docker/go-plugins-helpers/volume"
)

type Config struct {
	Mountpoint    string `json:"mountpoint"`
	DefaultSize   string `json:"defaultSize"`
	StorageDriver string `json:"storageDriver"`
}

type OpenVolumePlugin struct {
	Mountpoint    string
	DefaultSize   string
	StorageDriver string
}

func NewOpenVolumePlugin(configFile string) (*OpenVolumePlugin, error) {
	file, err := os.Open(configFile)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var config Config
	err = json.NewDecoder(file).Decode(&config)
	if err != nil {
		return nil, err
	}

	return &OpenVolumePlugin{
		Mountpoint:    config.Mountpoint,
		DefaultSize:   config.DefaultSize,
		StorageDriver: config.StorageDriver,
	}, nil
}



func (p *OpenVolumePlugin) Create(r volume.Request) volume.Response {
	mountpoint := filepath.Join(p.Mountpoint, r.Name)
	if _, err := os.Stat(mountpoint); os.IsNotExist(err) {
		size := r.Options["size"]
		if size == "" {
			size = p.DefaultSize // Use default size from config if size is not specified
		}
		cmd := exec.Command("truncate", "-s", size, filepath.Join(mountpoint, "data.img"))
		err := cmd.Run()
		if err != nil {
			log.Printf("Failed to create volume %s: %s", r.Name, err.Error())
			return volume.Response{Err: fmt.Sprintf("Failed to create volume %s", r.Name)}
		}
		log.Printf("Created volume %s with size %s", r.Name, size)
	} else {
		log.Printf("Volume %s already exists", r.Name)
	}
	return volume.Response{}
}


func (p *OpenVolumePlugin) Remove(r volume.Request) volume.Response {
	mountpoint := filepath.Join(p.mountpoint, r.Name)
	if _, err := os.Stat(mountpoint); err == nil {
		if err := os.RemoveAll(mountpoint); err != nil {
			log.Printf("Failed to remove volume %s: %s", r.Name, err.Error())
			return volume.Response{Err: fmt.Sprintf("Failed to remove volume %s", r.Name)}
		}
		log.Printf("Removed volume %s", r.Name)
	} else if os.IsNotExist(err) {
		log.Printf("Volume %s does not exist", r.Name)
	} else {
		log.Printf("Failed to remove volume %s: %s", r.Name, err.Error())
		return volume.Response{Err: fmt.Sprintf("Failed to remove volume %s", r.Name)}
	}
	return volume.Response{}
}




func (p *OpenVolumePlugin) Mount(r volume.Request) volume.Response {
	mountpoint := filepath.Join(p.mountpoint, r.Name)
	if _, err := os.Stat(mountpoint); os.IsNotExist(err) {
		log.Printf("Volume %s does not exist", r.Name)
		return volume.Response{Err: fmt.Sprintf("Volume %s does not exist", r.Name)}
	}

	return volume.Response{Mountpoint: mountpoint}
}



func (p *OpenVolumePlugin) Unmount(r volume.Request) volume.Response {
	return volume.Response{}
}

func (p *OpenVolumePlugin) Capabilities(r volume.Request) volume.Response {
	return volume.Response{Capabilities: volume.Capability{Scope: "local"}}
}

func (p *OpenVolumePlugin) Resize(r volume.Request) volume.Response {
	mountpoint := filepath.Join(p.Mountpoint, r.Name)
	if _, err := os.Stat(mountpoint); os.IsNotExist(err) {
		log.Printf("Volume %s does not exist", r.Name)
		return volume.Response{Err: fmt.Sprintf("Volume %s does not exist", r.Name)}
	}

	currentSize := p.getVolumeSize(mountpoint)
	if currentSize == -1 {
		log.Printf("Failed to get current size of volume %s", r.Name)
		return volume.Response{Err: fmt.Sprintf("Failed to get current size of volume %s", r.Name)}
	}

	requestedSize, err := strconv.ParseInt(r.Options["size"], 10, 64)
	if err != nil {
		log.Printf("Invalid size specified for volume %s", r.Name)
		return volume.Response{Err: fmt.Sprintf("Invalid size specified for volume %s", r.Name)}
	}

	if currentSize >= requestedSize {
		log.Printf("Current size of volume %s is %d bytes, which is greater than or equal to the requested size %d bytes", r.Name, currentSize, requestedSize)
		return volume.Response{Err: fmt.Sprintf("Cannot resize volume %s to %d bytes because current size is %d bytes or more", r.Name, requestedSize, currentSize)}
	}

	cmd := exec.Command("resize2fs", filepath.Join(mountpoint, "data.img"), fmt.Sprintf("%dG", requestedSize/(1024*1024*1024)))
	err = cmd.Run()
	if err != nil {
		log.Printf("Failed to resize volume %s: %s", r.Name, err.Error())
		return volume.Response{Err: fmt.Sprintf("Failed to resize volume %s", r.Name)}
	}
	log.Printf("Resized volume %s to %dG", r.Name, requestedSize/(1024*1024*1024))
	return volume.Response{}
}

func (p *OpenVolumePlugin) getVolumeSize(mountpoint string) int64 {
	cmd := exec.Command("du", "-sB", "1", mountpoint)
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("Failed to get volume size: %s", err.Error())
		return -1
	}

	fields := strings.Fields(string(output))
	if len(fields) < 1 {
		log.Printf("Unexpected output format from du command")
		return -1
	}

	size, err := strconv.ParseInt(fields[0], 10, 64)
	if err != nil {
		log.Printf("Failed to parse volume size: %s", err.Error())
		return -1
	}

	return size
}

func main() {
	plugin, err := NewOpenVolumePlugin("config.json")
	if err != nil {
		log.Fatalf("Failed to initialize OpenVolumePlugin: %s", err)
	}

	handler := volume.NewHandler(plugin)
	fmt.Println(handler.ServeUnix("openvolume", 0))
}
