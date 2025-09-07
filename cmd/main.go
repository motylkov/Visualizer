// Package main
//
// # Copyright (C) 2025 Maxim Motylkov
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.
package main

import (
	"context"
	"log"
	"time"

	"visualizer/pkg/config"
	"visualizer/pkg/database"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/template/html/v2"
)

type Candle struct {
	Figi     string    `json:"figi"`
	Time     time.Time `json:"time"`
	Open     float64   `json:"open"`
	High     float64   `json:"high"`
	Low      float64   `json:"low"`
	Close    float64   `json:"close"`
	Volume   int64     `json:"volume"`
	Interval string    `json:"interval"`
}

// Use shared config type from mainlib
type Config = config.Config

func main() {
	log.Println("Запуск приложения...")

	// Создаем контекст
	ctx := context.Background()

	// Загружаем конфигурацию из общей библиотеки с авто-определением пути
	log.Println("Загрузка конфигурации...")
	cfg, err := config.LoadConfig(config.GetConfigPath())
	if err != nil {
		log.Fatalf("Ошибка загрузки конфигурации: %v", err)
	}
	log.Println("Конфигурация загружена успешно")

	// Подключаемся к PostgreSQL
	log.Println("Подключение к базе данных...")
	db, err := database.ConnectToDatabase(ctx, &cfg.Database)
	if err != nil {
		log.Fatalf("ERR: Ошибка подключения к БД: %v", err)
	}
	log.Println("Подключение к БД успешно")
	defer db.Close()

	// Настройка шаблонизатора (для HTML)
	log.Println("Инициализация веб-сервера...")
	engine := html.New("./templates", ".html")
	app := fiber.New(fiber.Config{
		Views: engine,
	})
	log.Println("Веб-сервер инициализирован")

	// Маршрут: главная страница
	app.Get("/", func(c *fiber.Ctx) error {
		// Получаем доступные инструменты
		instruments, err := database.GetEnabledInstruments(db)
		if err != nil {
			log.Printf("Ошибка получения инструментов: %v", err)
			instruments = []database.Instrument{}
		}

		// Получаем доступные интервалы
		intervals := config.GetAvailableIntervals()

		return c.Render("chart", fiber.Map{
			"Title":       "График акций",
			"Instruments": instruments,
			"Intervals":   intervals,
		})
	})

	// API: /api/candles?figi=SBER&interval=day
	app.Get("/api/candles", func(c *fiber.Ctx) error {
		figi := c.Query("figi", "BBG004730N88") // по умолчанию — Сбер
		interval := c.Query("interval", "CANDLE_INTERVAL_DAY")

		rows, err := db.Query(c.Context(), `
            SELECT figi, time, open_price, high_price, low_price, close_price, volume, interval_type
            FROM candles
            WHERE figi = $1 AND interval_type = $2
            ORDER BY time
        `, figi, interval)

		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}
		defer rows.Close()

		var candles []map[string]interface{}
		for rows.Next() {
			var timeVal time.Time
			var open, high, low, closePrice float64
			var figi, interval string
			var volume int64

			err := rows.Scan(&figi, &timeVal, &open, &high, &low, &closePrice, &volume, &interval)
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
			}

			// DXCharts Lite требует timestamp в миллисекундах
			timestamp := timeVal.UnixMilli()

			candles = append(candles, map[string]any{
				"timestamp": timestamp,
				"open":      open,
				"high":      high,
				"low":       low,
				"close":     closePrice,
				"volume":    volume,
			})
		}

		// Всегда возвращаем массив, даже если пустой
		if candles == nil {
			candles = []map[string]interface{}{}
		}

		return c.JSON(candles)
	})

	// API: получение доступных инструментов
	app.Get("/api/instruments", func(c *fiber.Ctx) error {
		log.Println("Запрос инструментов...")
		instruments, err := database.GetEnabledInstruments(db)
		if err != nil {
			log.Printf("Ошибка получения инструментов: %v", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}
		log.Printf("Получено %d инструментов", len(instruments))

		// Всегда возвращаем массив, даже если пустой
		if instruments == nil {
			instruments = []database.Instrument{}
		}

		return c.JSON(instruments)
	})

	// API: получение доступных интервалов
	app.Get("/api/intervals", func(c *fiber.Ctx) error {
		intervals := config.GetAvailableIntervals()
		return c.JSON(intervals)
	})

	// Диагностический endpoint для проверки данных
	app.Get("/api/debug", func(c *fiber.Ctx) error {
		// Проверяем доступные FIGI
		figiRows, err := db.Query(c.Context(), `
			SELECT DISTINCT figi FROM candles LIMIT 10
		`)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}
		defer figiRows.Close()

		var figis []string
		for figiRows.Next() {
			var figi string
			if err := figiRows.Scan(&figi); err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
			}
			figis = append(figis, figi)
		}

		// Проверяем доступные интервалы
		intervalRows, err := db.Query(c.Context(), `
			SELECT DISTINCT interval_type FROM candles
		`)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}
		defer intervalRows.Close()

		var intervals []string
		for intervalRows.Next() {
			var interval string
			if err := intervalRows.Scan(&interval); err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
			}
			intervals = append(intervals, interval)
		}

		// Проверяем количество записей
		var count int
		err = db.QueryRow(c.Context(), `SELECT COUNT(*) FROM candles`).Scan(&count)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}

		return c.JSON(fiber.Map{
			"available_figis":     figis,
			"available_intervals": intervals,
			"total_records":       count,
		})
	})

	app.Use(func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusNotFound).SendString("Страница не найдена")
	})

	// Запуск сервера
	log.Println("Запуск сервера на порту 8080...")
	if err := app.Listen(":8080"); err != nil {
		log.Println("Ошибка запуска сервера:", err)
		return
	}
}
