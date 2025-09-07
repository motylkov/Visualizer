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

	"visualizer/pkg/config"

	"github.com/jackc/pgx/v5/pgxpool"
)

// ConnectToDatabase подключается к базе данных и инициализирует её
func ConnectToDatabase(ctx context.Context, dbConfig *config.DatabaseConfig) (*pgxpool.Pool, error) {
	dbURL := fmt.Sprintf("postgresql://%s:%s@%s:%d/%s?sslmode=%s",
		dbConfig.User, dbConfig.Password, dbConfig.Host, dbConfig.Port, dbConfig.DBName, dbConfig.SSLMode)

	dbpool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		return nil, fmt.Errorf("ошибка подключения к БД: %w", err)
	}

	return dbpool, nil
}
