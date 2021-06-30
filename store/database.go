package store

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/bytebase/bytebase"
	"github.com/bytebase/bytebase/api"
	"go.uber.org/zap"
)

var (
	_ api.DatabaseService = (*DatabaseService)(nil)
)

// DatabaseService represents a service for managing database.
type DatabaseService struct {
	l  *zap.Logger
	db *DB
}

// NewDatabaseService returns a new instance of DatabaseService.
func NewDatabaseService(logger *zap.Logger, db *DB) *DatabaseService {
	return &DatabaseService{l: logger, db: db}
}

// CreateDatabase creates a new database.
func (s *DatabaseService) CreateDatabase(ctx context.Context, create *api.DatabaseCreate) (*api.Database, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, FormatError(err)
	}
	defer tx.Rollback()

	database, err := s.createDatabase(ctx, tx.Tx, create)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, FormatError(err)
	}

	return database, nil
}

func (s *DatabaseService) CreateDatabaseTx(ctx context.Context, tx *sql.Tx, create *api.DatabaseCreate) (*api.Database, error) {
	return s.createDatabase(ctx, tx, create)
}

// FindDatabaseList retrieves a list of databases based on find.
func (s *DatabaseService) FindDatabaseList(ctx context.Context, find *api.DatabaseFind) ([]*api.Database, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, FormatError(err)
	}
	defer tx.Rollback()

	list, err := s.findDatabaseList(ctx, tx, find)
	if err != nil {
		return []*api.Database{}, err
	}

	return list, nil
}

// FindDatabase retrieves a single database based on find.
// Returns ENOTFOUND if no matching record.
// Returns the first matching one and prints a warning if finding more than 1 matching records.
func (s *DatabaseService) FindDatabase(ctx context.Context, find *api.DatabaseFind) (*api.Database, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, FormatError(err)
	}
	defer tx.Rollback()

	list, err := s.findDatabaseList(ctx, tx, find)
	if err != nil {
		return nil, err
	} else if len(list) == 0 {
		return nil, &bytebase.Error{Code: bytebase.ENOTFOUND, Message: fmt.Sprintf("database not found: %v", find)}
	} else if len(list) > 1 {
		s.l.Warn(fmt.Sprintf("found %d databases with filter %v, expect 1. ", len(list), find))
	}
	return list[0], nil
}

// PatchDatabase updates an existing database by ID.
// Returns ENOTFOUND if database does not exist.
func (s *DatabaseService) PatchDatabase(ctx context.Context, patch *api.DatabasePatch) (*api.Database, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, FormatError(err)
	}
	defer tx.Rollback()

	database, err := s.patchDatabase(ctx, tx, patch)
	if err != nil {
		return nil, FormatError(err)
	}

	if err := tx.Commit(); err != nil {
		return nil, FormatError(err)
	}

	return database, nil
}

// createDatabase creates a new database.
func (s *DatabaseService) createDatabase(ctx context.Context, tx *sql.Tx, create *api.DatabaseCreate) (*api.Database, error) {
	// Insert row into database.
	row, err := tx.QueryContext(ctx, `
		INSERT INTO db (
			creator_id,
			updater_id,
			instance_id,
			project_id,
			name,
			character_set,
			collation,
			sync_status,
			last_successful_sync_ts
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, 'OK', (strftime('%s', 'now')))
		RETURNING id, creator_id, created_ts, updater_id, updated_ts, instance_id, project_id, name, character_set, collation, sync_status, last_successful_sync_ts
	`,
		create.CreatorId,
		create.CreatorId,
		create.InstanceId,
		create.ProjectId,
		create.Name,
		create.CharacterSet,
		create.Collation,
	)

	if err != nil {
		return nil, FormatError(err)
	}
	defer row.Close()

	row.Next()
	var database api.Database
	if err := row.Scan(
		&database.ID,
		&database.CreatorId,
		&database.CreatedTs,
		&database.UpdaterId,
		&database.UpdatedTs,
		&database.InstanceId,
		&database.ProjectId,
		&database.Name,
		&database.CharacterSet,
		&database.Collation,
		&database.SyncStatus,
		&database.LastSuccessfulSyncTs,
	); err != nil {
		return nil, FormatError(err)
	}

	return &database, nil
}

func (s *DatabaseService) findDatabaseList(ctx context.Context, tx *Tx, find *api.DatabaseFind) (_ []*api.Database, err error) {
	// Build WHERE clause.
	where, args := []string{"1 = 1"}, []interface{}{}
	if v := find.ID; v != nil {
		where, args = append(where, "id = ?"), append(args, *v)
	}
	if v := find.InstanceId; v != nil {
		where, args = append(where, "instance_id = ?"), append(args, *v)
	}
	if v := find.ProjectId; v != nil {
		where, args = append(where, "project_id = ?"), append(args, *v)
	}
	if v := find.Name; v != nil {
		where, args = append(where, "name = ?"), append(args, *v)
	}
	if !find.IncludeAllDatabase {
		where = append(where, "name != '"+api.ALL_DATABASE_NAME+"'")
	}

	rows, err := tx.QueryContext(ctx, `
		SELECT 
		    id,
		    creator_id,
		    created_ts,
		    updater_id,
		    updated_ts,
			instance_id,
			project_id,
		    name,
			character_set,
			collation,
		    sync_status,
			last_successful_sync_ts
		FROM db
		WHERE `+strings.Join(where, " AND "),
		args...,
	)
	if err != nil {
		return nil, FormatError(err)
	}
	defer rows.Close()

	// Iterate over result set and deserialize rows into list.
	list := make([]*api.Database, 0)
	for rows.Next() {
		var database api.Database
		if err := rows.Scan(
			&database.ID,
			&database.CreatorId,
			&database.CreatedTs,
			&database.UpdaterId,
			&database.UpdatedTs,
			&database.InstanceId,
			&database.ProjectId,
			&database.Name,
			&database.CharacterSet,
			&database.Collation,
			&database.SyncStatus,
			&database.LastSuccessfulSyncTs,
		); err != nil {
			return nil, FormatError(err)
		}

		list = append(list, &database)
	}
	if err := rows.Err(); err != nil {
		return nil, FormatError(err)
	}

	return list, nil
}

// patchDatabase updates a database by ID. Returns the new state of the database after update.
func (s *DatabaseService) patchDatabase(ctx context.Context, tx *Tx, patch *api.DatabasePatch) (*api.Database, error) {
	// Build UPDATE clause.
	set, args := []string{"updater_id = ?"}, []interface{}{patch.UpdaterId}
	if v := patch.ProjectId; v != nil {
		set, args = append(set, "project_id = ?"), append(args, *v)
	}
	if v := patch.SyncStatus; v != nil {
		set, args = append(set, "sync_status = ?"), append(args, api.SyncStatus(*v))
	}
	if v := patch.LastSuccessfulSyncTs; v != nil {
		set, args = append(set, "last_successful_sync_ts = ?"), append(args, *v)
	}

	args = append(args, patch.ID)

	// Execute update query with RETURNING.
	row, err := tx.QueryContext(ctx, `
		UPDATE db
		SET `+strings.Join(set, ", ")+`
		WHERE id = ?
		RETURNING id, creator_id, created_ts, updater_id, updated_ts, instance_id, project_id, name, character_set, collation, sync_status, last_successful_sync_ts
	`,
		args...,
	)
	if err != nil {
		return nil, FormatError(err)
	}
	defer row.Close()

	if row.Next() {
		var database api.Database
		if err := row.Scan(
			&database.ID,
			&database.CreatorId,
			&database.CreatedTs,
			&database.UpdaterId,
			&database.UpdatedTs,
			&database.InstanceId,
			&database.ProjectId,
			&database.Name,
			&database.CharacterSet,
			&database.Collation,
			&database.SyncStatus,
			&database.LastSuccessfulSyncTs,
		); err != nil {
			return nil, FormatError(err)
		}
		return &database, nil
	}

	return nil, &bytebase.Error{Code: bytebase.ENOTFOUND, Message: fmt.Sprintf("database ID not found: %d", patch.ID)}
}
