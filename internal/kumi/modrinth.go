package kumi

import (
	"database/sql"
	"errors"
	"fmt"
	"time"
)

func (s *Service) CloneModrinthProfile(request ModrinthCloneRequest) (*ActionResult, error) {
	result := newResult()

	if request.DBPath == "" {
		return nil, errors.New("database path is required")
	}
	if request.SourcePath == "" || request.NewPath == "" {
		return nil, errors.New("source and target profile identifiers are required")
	}
	if request.NewName == "" {
		request.NewName = request.NewPath
	}

	db, err := sql.Open("sqlite3", request.DBPath)
	if err != nil {
		result.Error(fmt.Sprintf("failed to open database: %v", err))
		result.Success = false
		return result, nil
	}
	defer db.Close()

	var srcCount int
	if err := db.QueryRow("SELECT COUNT(1) FROM profiles WHERE path = ?", request.SourcePath).Scan(&srcCount); err != nil {
		result.Error(fmt.Sprintf("failed to query source profile: %v", err))
		result.Success = false
		return result, nil
	}
	if srcCount < 1 {
		result.Error(fmt.Sprintf("source profile '%s' not found", request.SourcePath))
		result.Success = false
		return result, nil
	}

	var newCount int
	if err := db.QueryRow("SELECT COUNT(1) FROM profiles WHERE path = ?", request.NewPath).Scan(&newCount); err != nil {
		result.Error(fmt.Sprintf("failed to query target profile: %v", err))
		result.Success = false
		return result, nil
	}
	if newCount > 0 {
		result.Error(fmt.Sprintf("target profile '%s' already exists", request.NewPath))
		result.Success = false
		return result, nil
	}

	cloneSQL := `INSERT INTO profiles (
  path, install_stage, name, icon_path,
  game_version, mod_loader, mod_loader_version,
  groups, linked_project_id, linked_version_id, locked,
  created, modified, last_played,
  submitted_time_played, recent_time_played,
  override_java_path, override_extra_launch_args, override_custom_env_vars,
  override_mc_memory_max, override_mc_force_fullscreen,
  override_mc_game_resolution_x, override_mc_game_resolution_y,
  override_hook_pre_launch, override_hook_wrapper, override_hook_post_exit,
  protocol_version, launcher_feature_version
)
SELECT
  ?, install_stage, ?, icon_path,
  game_version, mod_loader, mod_loader_version,
  groups, linked_project_id, linked_version_id, locked,
  created, modified, last_played,
  submitted_time_played, recent_time_played,
  override_java_path, override_extra_launch_args, override_custom_env_vars,
  override_mc_memory_max, override_mc_force_fullscreen,
  override_mc_game_resolution_x, override_mc_game_resolution_y,
  override_hook_pre_launch, override_hook_wrapper, override_hook_post_exit,
  protocol_version, launcher_feature_version
FROM profiles WHERE path = ?`

	if _, err := db.Exec(cloneSQL, request.NewPath, request.NewName, request.SourcePath); err != nil {
		result.Error(fmt.Sprintf("failed to clone profile: %v", err))
		result.Success = false
		return result, nil
	}

	nowUnix := time.Now().UTC().Unix()
	updateSQL := `UPDATE profiles
SET game_version = ?,
    mod_loader = ?,
    mod_loader_version = ?,
    modified = ?,
    last_played = ?,
    recent_time_played = ?
WHERE path = ?`

	lastPlayed := sql.NullInt64{Valid: false}
	if !request.ResetLastPlayed {
		row := db.QueryRow("SELECT last_played FROM profiles WHERE path = ?", request.NewPath)
		var lp sql.NullInt64
		if err := row.Scan(&lp); err == nil && lp.Valid {
			lastPlayed = lp
		}
	}

	playCount := 0
	if !request.ResetPlayCounters {
		row := db.QueryRow("SELECT recent_time_played FROM profiles WHERE path = ?", request.NewPath)
		var pc sql.NullInt64
		if err := row.Scan(&pc); err == nil && pc.Valid {
			playCount = int(pc.Int64)
		}
	}

	if _, err := db.Exec(updateSQL,
		request.GameVersion,
		request.ModLoader,
		request.ModLoaderVersion,
		nowUnix,
		lastPlayed,
		playCount,
		request.NewPath,
	); err != nil {
		result.Error(fmt.Sprintf("failed to update profile overrides: %v", err))
		result.Success = false
		return result, nil
	}

	verifySQL := `SELECT path, name, game_version, mod_loader, mod_loader_version FROM profiles WHERE path = ?`
	row := db.QueryRow(verifySQL, request.NewPath)
	var path, name, gameVersion, modLoader, modLoaderVersion string
	if err := row.Scan(&path, &name, &gameVersion, &modLoader, &modLoaderVersion); err == nil {
		result.Info(fmt.Sprintf("Cloned '%s' to '%s' (version %s - %s %s)", request.SourcePath, request.NewPath, gameVersion, modLoader, modLoaderVersion))
	}

	result.Success = true
	return result, nil
}
