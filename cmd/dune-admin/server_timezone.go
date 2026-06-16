package main

// serverTimezone returns the effective IANA timezone name for the given server
// using the fallback chain:
//
//  1. servers.timezone (server-level setting) — wins when non-empty
//  2. scheduled_restart config timezone — backward-compat for existing data
//  3. "" — caller passes to restartLocation which resolves to time.Local
func serverTimezone(serverID int) string {
	if globalStore != nil {
		var tz string
		row := globalStore.QueryRow(`SELECT timezone FROM servers WHERE id=?`, serverID)
		if err := row.Scan(&tz); err == nil && tz != "" {
			return tz
		}
	}
	return getScheduledRestartConfig(serverID).Timezone
}
