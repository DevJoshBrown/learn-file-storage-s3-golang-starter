package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"os/exec"
)

func getVideoAspectRatio(filePath string) (string, error) {
	cmd := exec.Command("ffprobe", "-v", "error", "-print_format", "json", "-show_streams", filePath)
	var b bytes.Buffer
	cmd.Stdout = &b

	err := cmd.Run()
	if err != nil {
		return "", err
	}
	type ffprobeOutput struct {
		Streams []struct {
			Width  int `json:"width"`
			Height int `json:"Height"`
		} `json:"streams"`
	}

	var aspectRatio ffprobeOutput

	err = json.Unmarshal(b.Bytes(), &aspectRatio)
	if err != nil {
		return "", err
	}

	if len(aspectRatio.Streams) == 0 {
		return "", fmt.Errorf("no streams found in video")
	}
	width := aspectRatio.Streams[0].Width
	height := aspectRatio.Streams[0].Height
	ratio := float64(width) / float64(height)

	if math.Abs(ratio-16.0/9.0) < 0.01 {
		return "16:9", nil
	}
	if math.Abs(ratio-9.0/16.0) < 0.01 {
		return "9:16", nil
	}
	return "other", nil

}
