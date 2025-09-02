// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// Copyright (c) 2025 Maxim Motylkov

package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"visualizer/pkg/database"
	"visualizer/pkg/mainlib"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/template/html/v2"
	"github.com/jackc/pgx/v5/pgxpool"
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
type Config = mainlib.Config

func main() {
	log.Println("Запуск приложения...")

	// Загружаем конфигурацию из общей библиотеки с авто-определением пути
	log.Println("Загрузка конфигурации...")
	cfg, err := mainlib.LoadConfig(mainlib.GetConfigPath())
	if err != nil {
		log.Fatalf("Ошибка загрузки конфигурации: %v", err)
	}
	log.Println("Конфигурация загружена успешно")

	// Подключаемся к PostgreSQL
	log.Println("Подключение к базе данных...")
	db, err := connectDatabase(cfg)
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
		intervals := mainlib.GetAvailableIntervals()

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
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		defer rows.Close()

		var candles []map[string]interface{}
		for rows.Next() {
			var timeVal time.Time
			var open, high, low, close float64
			var figi, interval string
			var volume int64

			err := rows.Scan(&figi, &timeVal, &open, &high, &low, &close, &volume, &interval)
			if err != nil {
				return c.Status(500).JSON(fiber.Map{"error": err.Error()})
			}

			// DXCharts Lite требует timestamp в миллисекундах
			timestamp := timeVal.UnixMilli()

			candles = append(candles, map[string]any{
				"timestamp": timestamp,
				"open":      open,
				"high":      high,
				"low":       low,
				"close":     close,
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
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
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
		intervals := mainlib.GetAvailableIntervals()
		return c.JSON(intervals)
	})

	// Диагностический endpoint для проверки данных
	app.Get("/api/debug", func(c *fiber.Ctx) error {
		// Проверяем доступные FIGI
		figiRows, err := db.Query(c.Context(), `
			SELECT DISTINCT figi FROM candles LIMIT 10
		`)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		defer figiRows.Close()

		var figis []string
		for figiRows.Next() {
			var figi string
			if err := figiRows.Scan(&figi); err != nil {
				return c.Status(500).JSON(fiber.Map{"error": err.Error()})
			}
			figis = append(figis, figi)
		}

		// Проверяем доступные интервалы
		intervalRows, err := db.Query(c.Context(), `
			SELECT DISTINCT interval_type FROM candles
		`)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		defer intervalRows.Close()

		var intervals []string
		for intervalRows.Next() {
			var interval string
			if err := intervalRows.Scan(&interval); err != nil {
				return c.Status(500).JSON(fiber.Map{"error": err.Error()})
			}
			intervals = append(intervals, interval)
		}

		// Проверяем количество записей
		var count int
		err = db.QueryRow(c.Context(), `SELECT COUNT(*) FROM candles`).Scan(&count)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}

		return c.JSON(fiber.Map{
			"available_figis":     figis,
			"available_intervals": intervals,
			"total_records":       count,
		})
	})

	app.Use(func(c *fiber.Ctx) error {
		return c.Status(404).SendString("Страница не найдена")
	})

	// Запуск сервера
	log.Println("Запуск сервера на порту 8080...")
	log.Fatal(app.Listen(":8080"))
}

// локальная функция loadConfig удалена; используется mainlib.LoadConfig

// connectDatabase подключается к PostgreSQL
func connectDatabase(cfg *Config) (*pgxpool.Pool, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.DBName,
		cfg.Database.SSLMode,
	)

	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		return nil, fmt.Errorf("ошибка создания пула подключений: %w", err)
	}

	// Проверяем подключение
	if err := pool.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("ошибка ping к БД: %w", err)
	}

	return pool, nil
}
