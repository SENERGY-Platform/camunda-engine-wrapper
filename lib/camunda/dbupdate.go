/*
 * Copyright 2026 InfAI (CC SES)
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package camunda

import (
	"context"
	"database/sql"
	"log/slog"
	"time"

	"github.com/SENERGY-Platform/camunda-engine-wrapper/lib/configuration"

	_ "github.com/lib/pq"
)

func UpdateDatabaseSchema(config configuration.Config) (err error) {
	if config.CamundaDb == "" || config.CamundaDb == "-" {
		return nil
	}
	db, err := sql.Open("postgres", config.CamundaDb)
	if err != nil {
		config.GetLogger().Error("could not open camunda db", "error", err)
		return err
	}

	ctx, _ := context.WithTimeout(context.Background(), time.Minute)
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		config.GetLogger().Error("could not begin camunda db transaction", "error", err)
		return err
	}

	err = updateDatabaseSchema_7_23(config.GetLogger(), tx)
	if err != nil {
		config.GetLogger().Error("could not update camunda db schema to 7.23", "error", err)
		tx.Rollback()
		return err
	}

	err = updateDatabaseSchema_7_24(config.GetLogger(), tx)
	if err != nil {
		config.GetLogger().Error("could not update camunda db schema to 7.24", "error", err)
		tx.Rollback()
		return err
	}

	err = tx.Commit()
	if err != nil {
		config.GetLogger().Error("could not commit camunda db transaction", "error", err)
		return err
	}
	return nil
}

func checkDatabaseSchemaVersion(db *sql.Tx, version string) (hasVersion bool, err error) {
	//query := `select id_, timestamp_, version_ from ACT_GE_SCHEMA_LOG where version_ = $1 limit 1;`
	err = db.QueryRow(`select COUNT(1) from ACT_GE_SCHEMA_LOG where version_ = $1 limit 1;`, version).Scan(&hasVersion)
	return hasVersion, err
}

func updateDatabaseSchema_7_23(logger *slog.Logger, db *sql.Tx) error {
	hasVersion, err := checkDatabaseSchemaVersion(db, "7.23.0")
	if err != nil {
		return err
	}
	if hasVersion {
		return nil
	}

	logger.Info("updating camunda db schema to 7.23")

	_, err = db.Exec(`insert into ACT_GE_SCHEMA_LOG
values ('1200', CURRENT_TIMESTAMP, '7.23.0');`)
	if err != nil {
		return err
	}

	_, err = db.Exec(`alter table ACT_HI_COMMENT
    add column if not exists REV_ integer not null
    default 1;`)
	if err != nil {
		return err
	}

	_, err = db.Exec(`alter table ACT_RU_EXECUTION add column if not exists PROC_DEF_KEY_ varchar(255);`)
	if err != nil {
		return err
	}

	return nil
}

func updateDatabaseSchema_7_24(logger *slog.Logger, db *sql.Tx) error {
	hasVersion, err := checkDatabaseSchemaVersion(db, "7.24.0")
	if err != nil {
		return err
	}
	if hasVersion {
		return nil
	}

	logger.Info("updating camunda db schema to 7.24")

	_, err = db.Exec(`insert into ACT_GE_SCHEMA_LOG
values ('1300', CURRENT_TIMESTAMP, '7.24.0');`)
	if err != nil {
		return err
	}

	return nil
}
