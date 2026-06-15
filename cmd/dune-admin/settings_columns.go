package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
)

// settings_columns.go stores the global (non-per-server) settings as typed
// rows across nine domain tables instead of the legacy app_settings.config_json
// blob. The appConfig struct and its json/yaml tags are unchanged; only storage
// moves to columns. The old app_settings.config_json column is kept but no
// longer read once migrated (see migrateSettingsColumns).

// dbExecer / dbRowQueryer are satisfied by both *sql.DB and *sql.Tx, so the
// column writers/readers work inside a migration transaction or standalone.
type dbExecer interface {
	Exec(query string, args ...any) (sql.Result, error)
}
type dbRowQueryer interface {
	QueryRow(query string, args ...any) *sql.Row
}

const settingsColumnsSchema = `
CREATE TABLE IF NOT EXISTS settings_connection (
	id INTEGER PRIMARY KEY CHECK (id = 1),
	ssh_host TEXT NOT NULL DEFAULT '', ssh_user TEXT NOT NULL DEFAULT '',
	ssh_key TEXT NOT NULL DEFAULT '', ssh_mode TEXT NOT NULL DEFAULT '',
	ssh_extra_opts TEXT NOT NULL DEFAULT '', auto_discover INTEGER NOT NULL DEFAULT 0,
	db_host TEXT NOT NULL DEFAULT '', db_port INTEGER NOT NULL DEFAULT 0,
	db_user TEXT NOT NULL DEFAULT '', db_pass TEXT NOT NULL DEFAULT '',
	db_name TEXT NOT NULL DEFAULT '', db_schema TEXT NOT NULL DEFAULT '',
	control TEXT NOT NULL DEFAULT '', control_namespace TEXT NOT NULL DEFAULT '',
	docker_gameserver TEXT NOT NULL DEFAULT '', docker_broker_game TEXT NOT NULL DEFAULT '',
	docker_broker_admin TEXT NOT NULL DEFAULT '', docker_db TEXT NOT NULL DEFAULT '',
	cmd_start TEXT NOT NULL DEFAULT '', cmd_stop TEXT NOT NULL DEFAULT '',
	cmd_restart TEXT NOT NULL DEFAULT '', cmd_status TEXT NOT NULL DEFAULT ''
);
CREATE TABLE IF NOT EXISTS settings_broker (
	id INTEGER PRIMARY KEY CHECK (id = 1),
	broker_game_addr TEXT NOT NULL DEFAULT '', broker_admin_addr TEXT NOT NULL DEFAULT '',
	broker_tls INTEGER NOT NULL DEFAULT 0, broker_user TEXT NOT NULL DEFAULT '',
	broker_pass TEXT NOT NULL DEFAULT '', broker_jwt_secret TEXT NOT NULL DEFAULT '',
	broker_exec_prefix TEXT NOT NULL DEFAULT ''
);
CREATE TABLE IF NOT EXISTS settings_amp (
	id INTEGER PRIMARY KEY CHECK (id = 1),
	amp_instance TEXT NOT NULL DEFAULT '', amp_container TEXT NOT NULL DEFAULT '',
	amp_user TEXT NOT NULL DEFAULT '', amp_log_path TEXT NOT NULL DEFAULT '',
	amp_use_container INTEGER, amp_container_runtime TEXT NOT NULL DEFAULT '',
	amp_data_root TEXT NOT NULL DEFAULT '', amp_api_user TEXT NOT NULL DEFAULT '',
	amp_api_pass TEXT NOT NULL DEFAULT '', amp_api_port INTEGER NOT NULL DEFAULT 0,
	amp_pg_bin TEXT NOT NULL DEFAULT '', amp_pg_lib TEXT NOT NULL DEFAULT '',
	amp_backup_dir TEXT NOT NULL DEFAULT '', director_url TEXT NOT NULL DEFAULT ''
);
CREATE TABLE IF NOT EXISTS settings_market_bot (
	id INTEGER PRIMARY KEY CHECK (id = 1),
	market_bot_enabled INTEGER, cache_db TEXT NOT NULL DEFAULT '',
	item_data TEXT NOT NULL DEFAULT '', state TEXT NOT NULL DEFAULT '',
	buy_interval TEXT NOT NULL DEFAULT '', list_interval TEXT NOT NULL DEFAULT '',
	buy_threshold REAL NOT NULL DEFAULT 0, max_buys INTEGER NOT NULL DEFAULT 0,
	remote_url TEXT NOT NULL DEFAULT '', remote_token TEXT NOT NULL DEFAULT ''
);
CREATE TABLE IF NOT EXISTS settings_discord (
	id INTEGER PRIMARY KEY CHECK (id = 1),
	bot_enabled INTEGER, bot_token TEXT NOT NULL DEFAULT '',
	guild_id TEXT NOT NULL DEFAULT '', roles_viewer TEXT NOT NULL DEFAULT '',
	roles_economy TEXT NOT NULL DEFAULT '', roles_admin TEXT NOT NULL DEFAULT '',
	announce_channel_id TEXT NOT NULL DEFAULT '', status_enabled INTEGER,
	status_channel_id TEXT NOT NULL DEFAULT '', status_interval_seconds INTEGER NOT NULL DEFAULT 0
);
CREATE TABLE IF NOT EXISTS settings_auth (
	id INTEGER PRIMARY KEY CHECK (id = 1),
	auth_enabled INTEGER, auth_local_username TEXT NOT NULL DEFAULT '',
	auth_local_password_hash TEXT NOT NULL DEFAULT '', auth_discord_enabled INTEGER,
	auth_discord_client_id TEXT NOT NULL DEFAULT '', auth_discord_client_secret TEXT NOT NULL DEFAULT '',
	auth_discord_redirect_url TEXT NOT NULL DEFAULT '', auth_owner_discord_ids TEXT NOT NULL DEFAULT '',
	auth_owner_role_ids TEXT NOT NULL DEFAULT '', auth_session_ttl_hours INTEGER NOT NULL DEFAULT 0,
	auth_guest_enabled INTEGER, auth_cookie_samesite TEXT NOT NULL DEFAULT ''
);
CREATE TABLE IF NOT EXISTS settings_battlepass (
	id INTEGER PRIMARY KEY CHECK (id = 1),
	enabled INTEGER, award_past INTEGER, auto_grant INTEGER,
	poll_seconds INTEGER NOT NULL DEFAULT 0, scan_pace_ms INTEGER NOT NULL DEFAULT 0,
	scan_start_delay_ms INTEGER NOT NULL DEFAULT 0
);
CREATE TABLE IF NOT EXISTS settings_welcome (
	id INTEGER PRIMARY KEY CHECK (id = 1),
	enabled INTEGER, scan_interval_secs INTEGER NOT NULL DEFAULT 0,
	active_version TEXT NOT NULL DEFAULT ''
);
CREATE TABLE IF NOT EXISTS settings_misc (
	id INTEGER PRIMARY KEY CHECK (id = 1),
	listen_addr TEXT NOT NULL DEFAULT '', scrip_currency INTEGER NOT NULL DEFAULT 0,
	backup_dir TEXT NOT NULL DEFAULT '', server_ini_dir TEXT NOT NULL DEFAULT '',
	default_ini_dir TEXT NOT NULL DEFAULT '', events_enabled INTEGER,
	default_server TEXT NOT NULL DEFAULT '', default_server_name TEXT NOT NULL DEFAULT ''
);`

// initSettingsColumnsSchema creates the nine settings_* tables. Idempotent.
func initSettingsColumnsSchema(db *sql.DB) error {
	if _, err := db.Exec(settingsColumnsSchema); err != nil {
		return fmt.Errorf("init settings columns schema: %w", err)
	}
	return nil
}

func b2i(b bool) int {
	if b {
		return 1
	}
	return 0
}

// saveSettingsColumns upserts cfg across the nine settings_* tables (id=1 each).
// Caller is expected to pass globalSettingsOnly(cfg); per-server/connection
// fields therefore persist as their zero values.
func saveSettingsColumns(db dbExecer, cfg appConfig) error {
	stmts := []struct {
		sql  string
		args []any
	}{
		{`INSERT INTO settings_connection (id, ssh_host, ssh_user, ssh_key, ssh_mode, ssh_extra_opts,
			auto_discover, db_host, db_port, db_user, db_pass, db_name, db_schema, control,
			control_namespace, docker_gameserver, docker_broker_game, docker_broker_admin, docker_db,
			cmd_start, cmd_stop, cmd_restart, cmd_status)
			VALUES (1,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)
			ON CONFLICT(id) DO UPDATE SET ssh_host=excluded.ssh_host, ssh_user=excluded.ssh_user,
			ssh_key=excluded.ssh_key, ssh_mode=excluded.ssh_mode, ssh_extra_opts=excluded.ssh_extra_opts,
			auto_discover=excluded.auto_discover, db_host=excluded.db_host, db_port=excluded.db_port,
			db_user=excluded.db_user, db_pass=excluded.db_pass, db_name=excluded.db_name,
			db_schema=excluded.db_schema, control=excluded.control, control_namespace=excluded.control_namespace,
			docker_gameserver=excluded.docker_gameserver, docker_broker_game=excluded.docker_broker_game,
			docker_broker_admin=excluded.docker_broker_admin, docker_db=excluded.docker_db,
			cmd_start=excluded.cmd_start, cmd_stop=excluded.cmd_stop, cmd_restart=excluded.cmd_restart,
			cmd_status=excluded.cmd_status`,
			[]any{cfg.SSHHost, cfg.SSHUser, cfg.SSHKey, cfg.SSHMode, cfg.SSHExtraOpts, b2i(cfg.AutoDiscover),
				cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPass, cfg.DBName, cfg.DBSchema, cfg.Control,
				cfg.ControlNamespace, cfg.DockerGameserver, cfg.DockerBrokerGame, cfg.DockerBrokerAdmin,
				cfg.DockerDB, cfg.CmdStart, cfg.CmdStop, cfg.CmdRestart, cfg.CmdStatus}},

		{`INSERT INTO settings_broker (id, broker_game_addr, broker_admin_addr, broker_tls, broker_user,
			broker_pass, broker_jwt_secret, broker_exec_prefix)
			VALUES (1,?,?,?,?,?,?,?)
			ON CONFLICT(id) DO UPDATE SET broker_game_addr=excluded.broker_game_addr,
			broker_admin_addr=excluded.broker_admin_addr, broker_tls=excluded.broker_tls,
			broker_user=excluded.broker_user, broker_pass=excluded.broker_pass,
			broker_jwt_secret=excluded.broker_jwt_secret, broker_exec_prefix=excluded.broker_exec_prefix`,
			[]any{cfg.BrokerGameAddr, cfg.BrokerAdminAddr, b2i(cfg.BrokerTLS), cfg.BrokerUser,
				cfg.BrokerPass, cfg.BrokerJWTSecret, cfg.BrokerExecPrefix}},

		{`INSERT INTO settings_amp (id, amp_instance, amp_container, amp_user, amp_log_path,
			amp_use_container, amp_container_runtime, amp_data_root, amp_api_user, amp_api_pass,
			amp_api_port, amp_pg_bin, amp_pg_lib, amp_backup_dir, director_url)
			VALUES (1,?,?,?,?,?,?,?,?,?,?,?,?,?,?)
			ON CONFLICT(id) DO UPDATE SET amp_instance=excluded.amp_instance,
			amp_container=excluded.amp_container, amp_user=excluded.amp_user, amp_log_path=excluded.amp_log_path,
			amp_use_container=excluded.amp_use_container, amp_container_runtime=excluded.amp_container_runtime,
			amp_data_root=excluded.amp_data_root, amp_api_user=excluded.amp_api_user,
			amp_api_pass=excluded.amp_api_pass, amp_api_port=excluded.amp_api_port, amp_pg_bin=excluded.amp_pg_bin,
			amp_pg_lib=excluded.amp_pg_lib, amp_backup_dir=excluded.amp_backup_dir, director_url=excluded.director_url`,
			[]any{cfg.AmpInstance, cfg.AmpContainer, cfg.AmpUser, cfg.AmpLogPath,
				boolPtrToNullInt(cfg.AmpUseContainer), cfg.AmpContainerRuntime, cfg.AmpDataRoot, cfg.AmpAPIUser,
				cfg.AmpAPIPass, cfg.AmpAPIPort, cfg.AmpPgBin, cfg.AmpPgLib, cfg.AmpBackupDir, cfg.DirectorURL}},

		{`INSERT INTO settings_market_bot (id, market_bot_enabled, cache_db, item_data, state,
			buy_interval, list_interval, buy_threshold, max_buys, remote_url, remote_token)
			VALUES (1,?,?,?,?,?,?,?,?,?,?)
			ON CONFLICT(id) DO UPDATE SET market_bot_enabled=excluded.market_bot_enabled,
			cache_db=excluded.cache_db, item_data=excluded.item_data, state=excluded.state,
			buy_interval=excluded.buy_interval, list_interval=excluded.list_interval,
			buy_threshold=excluded.buy_threshold, max_buys=excluded.max_buys, remote_url=excluded.remote_url,
			remote_token=excluded.remote_token`,
			[]any{boolPtrToNullInt(cfg.MarketBotEnabled), cfg.MarketBotCacheDB, cfg.MarketBotItemData,
				cfg.MarketBotState, cfg.MarketBotBuyInt, cfg.MarketBotListInt, cfg.MarketBotThresh,
				cfg.MarketBotMaxBuys, cfg.MarketBotRemoteURL, cfg.MarketBotRemoteToken}},

		{`INSERT INTO settings_discord (id, bot_enabled, bot_token, guild_id, roles_viewer, roles_economy,
			roles_admin, announce_channel_id, status_enabled, status_channel_id, status_interval_seconds)
			VALUES (1,?,?,?,?,?,?,?,?,?,?)
			ON CONFLICT(id) DO UPDATE SET bot_enabled=excluded.bot_enabled, bot_token=excluded.bot_token,
			guild_id=excluded.guild_id, roles_viewer=excluded.roles_viewer, roles_economy=excluded.roles_economy,
			roles_admin=excluded.roles_admin, announce_channel_id=excluded.announce_channel_id,
			status_enabled=excluded.status_enabled, status_channel_id=excluded.status_channel_id,
			status_interval_seconds=excluded.status_interval_seconds`,
			[]any{boolPtrToNullInt(cfg.DiscordBotEnabled), cfg.DiscordBotToken, cfg.DiscordGuildID,
				cfg.DiscordRolesViewer, cfg.DiscordRolesEconomy, cfg.DiscordRolesAdmin,
				cfg.DiscordAnnounceChannelID, boolPtrToNullInt(cfg.DiscordStatusEnabled),
				cfg.DiscordStatusChannelID, cfg.DiscordStatusIntervalSeconds}},

		{`INSERT INTO settings_auth (id, auth_enabled, auth_local_username, auth_local_password_hash,
			auth_discord_enabled, auth_discord_client_id, auth_discord_client_secret, auth_discord_redirect_url,
			auth_owner_discord_ids, auth_owner_role_ids, auth_session_ttl_hours, auth_guest_enabled,
			auth_cookie_samesite)
			VALUES (1,?,?,?,?,?,?,?,?,?,?,?,?)
			ON CONFLICT(id) DO UPDATE SET auth_enabled=excluded.auth_enabled,
			auth_local_username=excluded.auth_local_username,
			auth_local_password_hash=excluded.auth_local_password_hash,
			auth_discord_enabled=excluded.auth_discord_enabled,
			auth_discord_client_id=excluded.auth_discord_client_id,
			auth_discord_client_secret=excluded.auth_discord_client_secret,
			auth_discord_redirect_url=excluded.auth_discord_redirect_url,
			auth_owner_discord_ids=excluded.auth_owner_discord_ids,
			auth_owner_role_ids=excluded.auth_owner_role_ids,
			auth_session_ttl_hours=excluded.auth_session_ttl_hours,
			auth_guest_enabled=excluded.auth_guest_enabled,
			auth_cookie_samesite=excluded.auth_cookie_samesite`,
			[]any{boolPtrToNullInt(cfg.AuthEnabled), cfg.AuthLocalUsername, cfg.AuthLocalPasswordHash,
				boolPtrToNullInt(cfg.AuthDiscordEnabled), cfg.AuthDiscordClientID, cfg.AuthDiscordClientSecret,
				cfg.AuthDiscordRedirectURL, cfg.AuthOwnerDiscordIDs, cfg.AuthOwnerRoleIDs,
				cfg.AuthSessionTTLHours, boolPtrToNullInt(cfg.AuthGuestEnabled), cfg.AuthCookieSameSite}},

		{`INSERT INTO settings_battlepass (id, enabled, award_past, auto_grant, poll_seconds, scan_pace_ms,
			scan_start_delay_ms)
			VALUES (1,?,?,?,?,?,?)
			ON CONFLICT(id) DO UPDATE SET enabled=excluded.enabled, award_past=excluded.award_past,
			auto_grant=excluded.auto_grant, poll_seconds=excluded.poll_seconds, scan_pace_ms=excluded.scan_pace_ms,
			scan_start_delay_ms=excluded.scan_start_delay_ms`,
			[]any{boolPtrToNullInt(cfg.BattlepassEnabled), boolPtrToNullInt(cfg.BattlepassAwardPast),
				boolPtrToNullInt(cfg.BattlepassAutoGrant), cfg.BattlepassPollSeconds, cfg.BattlepassScanPaceMs,
				cfg.BattlepassScanStartDelayMs}},

		{`INSERT INTO settings_welcome (id, enabled, scan_interval_secs, active_version)
			VALUES (1,?,?,?)
			ON CONFLICT(id) DO UPDATE SET enabled=excluded.enabled,
			scan_interval_secs=excluded.scan_interval_secs, active_version=excluded.active_version`,
			[]any{boolPtrToNullInt(cfg.WelcomePackageEnabled), cfg.WelcomePackageScanSecs,
				cfg.WelcomePackageActiveVersion}},

		{`INSERT INTO settings_misc (id, listen_addr, scrip_currency, backup_dir, server_ini_dir,
			default_ini_dir, events_enabled, default_server, default_server_name)
			VALUES (1,?,?,?,?,?,?,?,?)
			ON CONFLICT(id) DO UPDATE SET listen_addr=excluded.listen_addr,
			scrip_currency=excluded.scrip_currency, backup_dir=excluded.backup_dir,
			server_ini_dir=excluded.server_ini_dir, default_ini_dir=excluded.default_ini_dir,
			events_enabled=excluded.events_enabled, default_server=excluded.default_server,
			default_server_name=excluded.default_server_name`,
			[]any{cfg.ListenAddr, cfg.ScripCurrency, cfg.BackupDir, cfg.ServerIniDir, cfg.DefaultIniDir,
				boolPtrToNullInt(cfg.EventsEnabled), cfg.DefaultServer, cfg.DefaultServerName}},
	}
	for _, st := range stmts {
		if _, err := db.Exec(st.sql, st.args...); err != nil {
			return fmt.Errorf("save settings columns: %w", err)
		}
	}
	return nil
}

// optRow tolerates a missing row: an absent settings_* row leaves the struct's
// zero values, matching the legacy blob's "field unset" semantics.
func optRow(err error) error {
	if errors.Is(err, sql.ErrNoRows) {
		return nil
	}
	return err
}

// loadSettingsColumns reads the global settings from the nine settings_* tables.
// ok=false on first boot (no settings_misc row yet — nothing persisted).
func loadSettingsColumns(db dbRowQueryer) (appConfig, bool, error) {
	cfg, ok, err := loadMiscSettings(db)
	if err != nil || !ok {
		return appConfig{}, ok, err
	}
	loaders := []func(dbRowQueryer, *appConfig) error{
		loadConnectionSettings, loadBrokerSettings, loadAmpSettings, loadMarketBotSettings,
		loadDiscordSettings, loadAuthSettings, loadBattlepassSettings, loadWelcomeSettings,
	}
	for _, load := range loaders {
		if err := load(db, &cfg); err != nil {
			return appConfig{}, false, err
		}
	}
	return cfg, true, nil
}

// loadMiscSettings reads settings_misc, the canonical presence marker: every
// saveSettingsColumns writes it, so its absence means "nothing persisted yet".
func loadMiscSettings(db dbRowQueryer) (appConfig, bool, error) {
	var cfg appConfig
	var eventsEnabled sql.NullInt64
	err := db.QueryRow(`SELECT listen_addr, scrip_currency, backup_dir, server_ini_dir, default_ini_dir,
		events_enabled, default_server, default_server_name FROM settings_misc WHERE id = 1`).Scan(
		&cfg.ListenAddr, &cfg.ScripCurrency, &cfg.BackupDir, &cfg.ServerIniDir, &cfg.DefaultIniDir,
		&eventsEnabled, &cfg.DefaultServer, &cfg.DefaultServerName)
	if errors.Is(err, sql.ErrNoRows) {
		return appConfig{}, false, nil
	}
	if err != nil {
		return appConfig{}, false, fmt.Errorf("load settings_misc: %w", err)
	}
	cfg.EventsEnabled = nullIntToBoolPtr(eventsEnabled)
	return cfg, true, nil
}

func loadConnectionSettings(db dbRowQueryer, cfg *appConfig) error {
	var autoDiscover int
	err := db.QueryRow(`SELECT ssh_host, ssh_user, ssh_key, ssh_mode, ssh_extra_opts, auto_discover,
		db_host, db_port, db_user, db_pass, db_name, db_schema, control, control_namespace,
		docker_gameserver, docker_broker_game, docker_broker_admin, docker_db,
		cmd_start, cmd_stop, cmd_restart, cmd_status FROM settings_connection WHERE id = 1`).Scan(
		&cfg.SSHHost, &cfg.SSHUser, &cfg.SSHKey, &cfg.SSHMode, &cfg.SSHExtraOpts, &autoDiscover,
		&cfg.DBHost, &cfg.DBPort, &cfg.DBUser, &cfg.DBPass, &cfg.DBName, &cfg.DBSchema, &cfg.Control,
		&cfg.ControlNamespace, &cfg.DockerGameserver, &cfg.DockerBrokerGame, &cfg.DockerBrokerAdmin,
		&cfg.DockerDB, &cfg.CmdStart, &cfg.CmdStop, &cfg.CmdRestart, &cfg.CmdStatus)
	cfg.AutoDiscover = autoDiscover != 0
	return optRow(err)
}

func loadBrokerSettings(db dbRowQueryer, cfg *appConfig) error {
	var brokerTLS int
	err := db.QueryRow(`SELECT broker_game_addr, broker_admin_addr, broker_tls, broker_user,
		broker_pass, broker_jwt_secret, broker_exec_prefix FROM settings_broker WHERE id = 1`).Scan(
		&cfg.BrokerGameAddr, &cfg.BrokerAdminAddr, &brokerTLS, &cfg.BrokerUser, &cfg.BrokerPass,
		&cfg.BrokerJWTSecret, &cfg.BrokerExecPrefix)
	cfg.BrokerTLS = brokerTLS != 0
	return optRow(err)
}

func loadAmpSettings(db dbRowQueryer, cfg *appConfig) error {
	var ampUseContainer sql.NullInt64
	err := db.QueryRow(`SELECT amp_instance, amp_container, amp_user, amp_log_path, amp_use_container,
		amp_container_runtime, amp_data_root, amp_api_user, amp_api_pass, amp_api_port, amp_pg_bin,
		amp_pg_lib, amp_backup_dir, director_url FROM settings_amp WHERE id = 1`).Scan(
		&cfg.AmpInstance, &cfg.AmpContainer, &cfg.AmpUser, &cfg.AmpLogPath, &ampUseContainer,
		&cfg.AmpContainerRuntime, &cfg.AmpDataRoot, &cfg.AmpAPIUser, &cfg.AmpAPIPass, &cfg.AmpAPIPort,
		&cfg.AmpPgBin, &cfg.AmpPgLib, &cfg.AmpBackupDir, &cfg.DirectorURL)
	cfg.AmpUseContainer = nullIntToBoolPtr(ampUseContainer)
	return optRow(err)
}

func loadMarketBotSettings(db dbRowQueryer, cfg *appConfig) error {
	var marketBotEnabled sql.NullInt64
	err := db.QueryRow(`SELECT market_bot_enabled, cache_db, item_data, state, buy_interval,
		list_interval, buy_threshold, max_buys, remote_url, remote_token FROM settings_market_bot
		WHERE id = 1`).Scan(
		&marketBotEnabled, &cfg.MarketBotCacheDB, &cfg.MarketBotItemData, &cfg.MarketBotState,
		&cfg.MarketBotBuyInt, &cfg.MarketBotListInt, &cfg.MarketBotThresh, &cfg.MarketBotMaxBuys,
		&cfg.MarketBotRemoteURL, &cfg.MarketBotRemoteToken)
	cfg.MarketBotEnabled = nullIntToBoolPtr(marketBotEnabled)
	return optRow(err)
}

func loadDiscordSettings(db dbRowQueryer, cfg *appConfig) error {
	var discordBotEnabled, discordStatusEnabled sql.NullInt64
	err := db.QueryRow(`SELECT bot_enabled, bot_token, guild_id, roles_viewer, roles_economy,
		roles_admin, announce_channel_id, status_enabled, status_channel_id, status_interval_seconds
		FROM settings_discord WHERE id = 1`).Scan(
		&discordBotEnabled, &cfg.DiscordBotToken, &cfg.DiscordGuildID, &cfg.DiscordRolesViewer,
		&cfg.DiscordRolesEconomy, &cfg.DiscordRolesAdmin, &cfg.DiscordAnnounceChannelID,
		&discordStatusEnabled, &cfg.DiscordStatusChannelID, &cfg.DiscordStatusIntervalSeconds)
	cfg.DiscordBotEnabled = nullIntToBoolPtr(discordBotEnabled)
	cfg.DiscordStatusEnabled = nullIntToBoolPtr(discordStatusEnabled)
	return optRow(err)
}

func loadAuthSettings(db dbRowQueryer, cfg *appConfig) error {
	var authEnabled, authDiscordEnabled, authGuestEnabled sql.NullInt64
	err := db.QueryRow(`SELECT auth_enabled, auth_local_username, auth_local_password_hash,
		auth_discord_enabled, auth_discord_client_id, auth_discord_client_secret, auth_discord_redirect_url,
		auth_owner_discord_ids, auth_owner_role_ids, auth_session_ttl_hours, auth_guest_enabled,
		auth_cookie_samesite FROM settings_auth WHERE id = 1`).Scan(
		&authEnabled, &cfg.AuthLocalUsername, &cfg.AuthLocalPasswordHash, &authDiscordEnabled,
		&cfg.AuthDiscordClientID, &cfg.AuthDiscordClientSecret, &cfg.AuthDiscordRedirectURL,
		&cfg.AuthOwnerDiscordIDs, &cfg.AuthOwnerRoleIDs, &cfg.AuthSessionTTLHours, &authGuestEnabled,
		&cfg.AuthCookieSameSite)
	cfg.AuthEnabled = nullIntToBoolPtr(authEnabled)
	cfg.AuthDiscordEnabled = nullIntToBoolPtr(authDiscordEnabled)
	cfg.AuthGuestEnabled = nullIntToBoolPtr(authGuestEnabled)
	return optRow(err)
}

func loadBattlepassSettings(db dbRowQueryer, cfg *appConfig) error {
	var bpEnabled, bpAwardPast, bpAutoGrant sql.NullInt64
	err := db.QueryRow(`SELECT enabled, award_past, auto_grant, poll_seconds, scan_pace_ms,
		scan_start_delay_ms FROM settings_battlepass WHERE id = 1`).Scan(
		&bpEnabled, &bpAwardPast, &bpAutoGrant, &cfg.BattlepassPollSeconds, &cfg.BattlepassScanPaceMs,
		&cfg.BattlepassScanStartDelayMs)
	cfg.BattlepassEnabled = nullIntToBoolPtr(bpEnabled)
	cfg.BattlepassAwardPast = nullIntToBoolPtr(bpAwardPast)
	cfg.BattlepassAutoGrant = nullIntToBoolPtr(bpAutoGrant)
	return optRow(err)
}

func loadWelcomeSettings(db dbRowQueryer, cfg *appConfig) error {
	var welcomeEnabled sql.NullInt64
	err := db.QueryRow(`SELECT enabled, scan_interval_secs, active_version FROM settings_welcome
		WHERE id = 1`).Scan(
		&welcomeEnabled, &cfg.WelcomePackageScanSecs, &cfg.WelcomePackageActiveVersion)
	cfg.WelcomePackageEnabled = nullIntToBoolPtr(welcomeEnabled)
	return optRow(err)
}

// migrateSettingsColumns translates the legacy app_settings.config_json blob into
// the typed settings_* tables, once, guarded by the migrated:settings_columns
// marker. After this runs the blob column is never read again.
func migrateSettingsColumns(db *sql.DB) error {
	return runColumnMigrationOnce(db, "migrated:settings_columns", func(tx *sql.Tx) error {
		var blob string
		err := tx.QueryRow(`SELECT config_json FROM app_settings WHERE id = 1`).Scan(&blob)
		if errors.Is(err, sql.ErrNoRows) {
			return nil // no legacy settings to migrate
		}
		if err != nil {
			return fmt.Errorf("read legacy settings blob: %w", err)
		}
		var cfg appConfig
		if err := json.Unmarshal([]byte(blob), &cfg); err != nil {
			return fmt.Errorf("unmarshal legacy settings blob: %w", err)
		}
		return saveSettingsColumns(tx, globalSettingsOnly(cfg))
	})
}
