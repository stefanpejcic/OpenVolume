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

// configuration values
type Config struct {
	Mountpoint    string `json:"mountpoint"`
	DefaultSize   string `json:"defaultSize"`
	StorageDriver string `json:"storageDriver"`
}

type OpenVolumePlugin struct {
	config Config
}

func NewOpenVolumePlugin(config Config) *OpenVolumePlugin {
	return &OpenVolumePlugin{config: config}
}

func (p *OpenVolumePlugin) Create(r volume.Request) volume.Response {
	mountpoint := filepath.Join(p.config.Mountpoint, r.Name)
	if _, err := os.Stat(mountpoint); err == nil {
		log.Printf("Volume %s already exists", r.Name)
		return volume.Response{}
	}

	sizeStr := p.config.DefaultSize
	if val, ok := r.Options["size"]; ok {
		sizeStr = val
	}

	size, err := strconv.ParseInt(sizeStr, 10, 64)
	if err != nil || size <= 0 {
		log.Printf("Invalid size specified for volume %s", r.Name)
		return volume.Response{Err: "Invalid size specified"}
	}

	if err := os.MkdirAll(mountpoint, 0755); err != nil {
		log.Printf("Failed to create volume %s: %s", r.Name, err.Error())
		return volume.Response{Err: fmt.Sprintf("Failed to create volume %s", r.Name)}
	}

	// Create a sparse file with the specified size
	cmd := exec.Command("truncate", "-s", fmt.Sprintf("%dB", size), filepath.Join(mountpoint, "data.img"))
	if err := cmd.Run(); err != nil {
		log.Printf("Failed to create volume %s: %s", r.Name, err.Error())
		os.RemoveAll(mountpoint)
		return volume.Response{Err: fmt.Sprintf("Failed to create volume %s", r.Name)}
	}

	log.Printf("Created volume %s with size %s bytes", r.Name, sizeStr)
	return volume.Response{}
}



func main() {
	configPath := "config.json"
	configFile, err := os.Open(configPath)
	if err != nil {
		log.Fatalf("Failed to open config file: %s", err)
	}
	defer configFile.Close()

	var config Config
	if err := json.NewDecoder(configFile).Decode(&config); err != nil {
		log.Fatalf("Failed to decode config file: %s", err)
	}

	plugin := NewOpenVolumePlugin(config)
	handler := volume.NewHandler(plugin)
	fmt.Println(handler.ServeUnix("openvolume", 0))
}
