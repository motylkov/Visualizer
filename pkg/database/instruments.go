// Package database содержит функции для работы с базой данных свечей
//
// # Copyright (C) 2025 Maxim Motylkov
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.
package database

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Instrument структура инструмента
type Instrument struct {
	Figi              string
	Ticker            string
	Name              string
	InstrumentType    string
	Currency          string
	LotSize           int32
	MinPriceIncrement float64
	TradingStatus     string
	Enabled           bool
}

// GetEnabledInstruments получает включенные инструменты для загрузки свечей
func GetEnabledInstruments(dbpool *pgxpool.Pool) ([]Instrument, error) {
	var query string
	var instruments []Instrument

	// Получаем включенные инструменты определенного типа
	query = `SELECT figi, ticker, name, instrument_type, currency, lot_size, min_price_increment, trading_status, enabled FROM instruments 
				WHERE
					trading_status = 'SECURITY_TRADING_STATUS_NORMAL_TRADING' AND
					enabled = true
				ORDER BY ticker`

	rows, err := dbpool.Query(context.Background(), query)
	if err != nil {
		return nil, fmt.Errorf("ошибка запроса включенных инструментов: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var i Instrument
		err := rows.Scan(
			&i.Figi,
			&i.Ticker,
			&i.Name,
			&i.InstrumentType,
			&i.Currency,
			&i.LotSize,
			&i.MinPriceIncrement,
			&i.TradingStatus,
			&i.Enabled,
		)
		if err != nil {
			return nil, fmt.Errorf("ошибка сканирования инструмента: %w", err)
		}

		instruments = append(instruments, i)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка итерации по инструментам: %w", err)
	}

	return instruments, nil
}
