package main

import (
	"fmt"
	"os"

	"website-monitor/internal/model"
	"website-monitor/internal/service"
	"website-monitor/internal/util"
)

func main() {
	// * 存成隱藏檔
	filepath := ".webMonitor.json"

	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		// * 生成初始檔案
		defaultData := model.MonitorData{
			List: []model.Website{},
			Config: model.Config{
				Host:     "",
				Port:     587,
				Username: "",
				Password: "",
				From:     "",
				To:       []string{},
				Enabled:  false,
			},
		}

		if err := util.Save(filepath, defaultData); err != nil {
			fmt.Printf("Failed to create init file: %v\n", err)
			os.Exit(1)
		}
	}

	data, err := util.Read(filepath)
	if err != nil {
		fmt.Printf("Failed to read data: %v\n", err)
		os.Exit(1)
	}

	var list []string
	for _, e := range data.List {
		list = append(list, e.URL)
	}

	monitor := service.New(list, data.Config)

	view := service.NewUI(monitor)
	if err := view.Run(); err != nil {
		fmt.Printf("Failed to run: %v\n", err)
		os.Exit(1)
	}
}
