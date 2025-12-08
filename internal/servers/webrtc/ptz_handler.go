package webrtc

import (
	"fmt"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"gopkg.in/yaml.v3"
)

// PortConfig holds port configuration for frontend
type PortConfig struct {
	WebRTC int `json:"webrtc"`
	API    int `json:"api"`
	HLS    int `json:"hls"`
	RTSP   int `json:"rtsp"`
}

// ConfigYAML represents the minimal structure we need from mediamtx.yml
type ConfigYAML struct {
	APIAddress    string `yaml:"apiAddress"`
	HLSAddress    string `yaml:"hlsAddress"`
	RTSPAddress   string `yaml:"rtspAddress"`
	WebRTCAddress string `yaml:"webrtcAddress"`
}

// onConfigPorts returns the configured ports for all services
func (s *httpServer) onConfigPorts(ctx *gin.Context) {
	// Parse WebRTC port from current server address
	webrtcPort := parsePort(s.address)

	// Try to read actual configuration from mediamtx.yml
	apiPort, hlsPort, rtspPort, err := readConfigPorts()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Failed to read port configuration: %v", err),
		})
		return
	}

	// Use MediaMTX default ports if not configured
	if apiPort == 0 {
		apiPort = 9997 // MediaMTX default API port
	}
	if hlsPort == 0 {
		hlsPort = 8888 // MediaMTX default HLS port
	}
	if rtspPort == 0 {
		rtspPort = 8554 // MediaMTX default RTSP port
	}

	ctx.JSON(http.StatusOK, PortConfig{
		WebRTC: webrtcPort,
		API:    apiPort,
		HLS:    hlsPort,
		RTSP:   rtspPort,
	})
}

// readConfigPorts reads port configuration from mediamtx.yml
func readConfigPorts() (apiPort, hlsPort, rtspPort int, err error) {
	// Try common config file locations
	configPaths := []string{
		"/app/mediamtx.yml",
		"./mediamtx.yml",
		"/etc/mediamtx.yml",
	}

	var configData []byte
	var readErr error

	for _, path := range configPaths {
		configData, readErr = os.ReadFile(path)
		if readErr == nil {
			break
		}
	}

	if readErr != nil {
		return 0, 0, 0, fmt.Errorf("failed to read mediamtx.yml from any location: %w", readErr)
	}

	var config ConfigYAML
	if err := yaml.Unmarshal(configData, &config); err != nil {
		return 0, 0, 0, fmt.Errorf("failed to parse YAML configuration: %w", err)
	}

	// Parse ports from configuration (0 if not configured - optional services)
	apiPort = parsePort(config.APIAddress)
	hlsPort = parsePort(config.HLSAddress)
	rtspPort = parsePort(config.RTSPAddress)

	return apiPort, hlsPort, rtspPort, nil
}

// parsePort extracts port number from address string (e.g., ":8119" -> 8119)
func parsePort(address string) int {
	if address == "" {
		return 0
	}
	// Remove leading colon if present
	if address[0] == ':' {
		address = address[1:]
	}
	// Try to parse as integer
	port := 0
	fmt.Sscanf(address, "%d", &port)
	return port
}
