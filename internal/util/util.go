package util

import (
	"encoding/json"
	"os"
	"website-monitor/internal/model"
)

func Save(path string, data model.MonitorData) error {
	bytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, bytes, 0644)
}

func Read(path string) (*model.MonitorData, error) {
	bytes, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var data model.MonitorData
	if err := json.Unmarshal(bytes, &data); err != nil {
		return nil, err
	}

	return &data, nil
}
