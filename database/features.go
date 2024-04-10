package database

import (
	"context"
	"fmt"
	"strings"

	"gorm.io/gorm"
)

func (db *DB) testFeatures(ctx context.Context) error {
	if err := db.Transaction(ctx, func(tx *gorm.DB) error {
		var res int
		err := tx.Raw("SELECT (1 + 1);").Scan(&res).Error
		if err != nil || res != 2 {
			return fmt.Errorf("connectivity test failed (expected 2, got %v): %w", res, err)
		}

		if db.kind == DB_POSTGRES || db.kind == DB_CRDB {
			err := tx.Raw("SELECT (1 + 2)::bigint;").Scan(&res).Error
			if err != nil || res != 3 {
				return fmt.Errorf("bigint test failed (expected 3, got %v): %w", res, err)
			}

			var encoding string
			err = tx.Raw("SELECT character_set_name FROM information_schema.character_sets;").Scan(&encoding).Error
			if err != nil {
				return fmt.Errorf("querying db encoding: %w", err)
			}
			if strings.ToUpper(encoding) != "UTF8" {
				return fmt.Errorf("db should be using UTF8 encoding, but it uses %v instead", encoding)
			}

			var uuidRes string
			err = tx.Raw("SELECT gen_random_uuid();").Scan(&uuidRes).Error
			if err != nil {
				return fmt.Errorf("gen_random_uuid() is not supported: %w", err)
			}
		}

		return nil
	}); err != nil {
		return fmt.Errorf("running transaction: %w", err)
	}

	db.Logger.Info(ctx, "db: connection ok")
	return nil
}

func (db *DB) GetVersion() (string, error) {
	function := "version"
	if db.kind == DB_SQLITE {
		function = "sqlite_version"
	}

	var version string
	err := db.Raw("SELECT " + function + "();").Scan(&version).Error
	if err != nil {
		return "", fmt.Errorf("getting database version: %w", err)
	}

	return version, nil
}
