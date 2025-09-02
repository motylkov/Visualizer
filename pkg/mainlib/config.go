// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// Copyright (c) 2025 Maxim Motylkov

package mainlib

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// DatabaseConfig структура конфигурации базы данных
type DatabaseConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	DBName   string `yaml:"dbname"`
	SSLMode  string `yaml:"sslmode"`
}

// Config структура конфигурации
type Config struct {
	Database DatabaseConfig `yaml:"database"`
}

// LoadConfig загружает конфигурацию из YAML файла
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("не удалось прочитать файл конфигурации: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("ошибка парсинга YAML: %w", err)
	}

	return &cfg, nil
}

// GetConfigPath определяет путь к файлу конфигурации
func GetConfigPath() string {
	// Получаем путь к исполняемому файлу
	execPath, err := os.Executable()
	if err != nil {
		// Если не удалось получить путь к исполняемому файлу, используем относительный путь
		return "config/config.yaml"
	}

	// Определяем директорию исполняемого файла
	execDir := filepath.Dir(execPath)

	// Если исполняемый файл в папке bin, то конфиг на один уровень выше
	if filepath.Base(execDir) == "bin" {
		return filepath.Join(filepath.Dir(execDir), "config", "config.yaml")
	}

	// Иначе используем относительный путь (для go run)
	return "config/config.yaml"
}

// IntervalOption структура для выбора интервала
type IntervalOption struct {
	Value string `json:"value"`
	Label string `json:"label"`
}

// GetAvailableIntervals возвращает доступные интервалы с человекочитаемыми названиями
func GetAvailableIntervals() []IntervalOption {
	return []IntervalOption{
		{Value: "CANDLE_INTERVAL_1_MIN", Label: "1 минута"},
		{Value: "CANDLE_INTERVAL_5_MIN", Label: "5 минут"},
		{Value: "CANDLE_INTERVAL_15_MIN", Label: "15 минут"},
		{Value: "CANDLE_INTERVAL_30_MIN", Label: "30 минут"},
		{Value: "CANDLE_INTERVAL_HOUR", Label: "1 час"},
		{Value: "CANDLE_INTERVAL_DAY", Label: "1 день"},
		{Value: "CANDLE_INTERVAL_WEEK", Label: "1 неделя"},
		{Value: "CANDLE_INTERVAL_MONTH", Label: "1 месяц"},
	}
}
