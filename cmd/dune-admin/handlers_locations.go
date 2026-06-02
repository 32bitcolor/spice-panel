package main

import (
	"fmt"
	"net/http"
	"strings"
)

// resolveLocation looks up a named location from the editable store (or the
// compile-time cheatLocations fallback when the store is unavailable).
// Returns an error suitable for a 400 response if the name is unknown.
func resolveLocation(name string) (teleportLocation, error) {
	if globalLocationStore != nil {
		locs, err := globalLocationStore.list()
		if err == nil {
			for _, l := range locs {
				if l.Name == name {
					return l, nil
				}
			}
			return teleportLocation{}, fmt.Errorf("unknown location: %s", name)
		}
	}
	// Fallback to compile-time seeds.
	for _, l := range cheatLocations {
		if l.Name == name {
			return l, nil
		}
	}
	return teleportLocation{}, fmt.Errorf("unknown location: %s", name)
}

// globalLocationStore holds the open SQLite location store. Set once in
// main.go alongside globalSessionDB; nil when the store failed to open.
var globalLocationStore *locationStore

// @Summary List all saved teleport/spawn locations
// @Tags locations
// @Produce json
// @Success 200 {array} teleportLocation
// @Failure 503 {object} map[string]string
// @Router /api/v1/locations [get]
// GET /api/v1/locations
func handleListLocations(w http.ResponseWriter, _ *http.Request) {
	if globalLocationStore == nil {
		jsonErr(w, fmt.Errorf("location store not available"), http.StatusServiceUnavailable)
		return
	}
	locs, err := globalLocationStore.list()
	if err != nil {
		jsonErr(w, fmt.Errorf("list locations: %w", err), http.StatusInternalServerError)
		return
	}
	jsonOK(w, locs)
}

// @Summary Add or update a named teleport/spawn location
// @Tags locations
// @Accept json
// @Produce json
// @Param body body object true "name, x, y, z"
// @Success 200 {array} teleportLocation
// @Failure 400 {object} map[string]string
// @Failure 503 {object} map[string]string
// @Router /api/v1/locations [post]
// POST /api/v1/locations
func handleUpsertLocation(w http.ResponseWriter, r *http.Request) {
	if globalLocationStore == nil {
		jsonErr(w, fmt.Errorf("location store not available"), http.StatusServiceUnavailable)
		return
	}
	var req struct {
		Name string  `json:"name"`
		X    float64 `json:"x"`
		Y    float64 `json:"y"`
		Z    float64 `json:"z"`
	}
	if err := decode(r, &req); err != nil {
		jsonErr(w, err, http.StatusBadRequest)
		return
	}
	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		jsonErr(w, fmt.Errorf("name required"), http.StatusBadRequest)
		return
	}
	if err := globalLocationStore.upsert(req.Name, req.X, req.Y, req.Z); err != nil {
		jsonErr(w, err, http.StatusInternalServerError)
		return
	}
	locs, err := globalLocationStore.list()
	if err != nil {
		jsonErr(w, fmt.Errorf("list after upsert: %w", err), http.StatusInternalServerError)
		return
	}
	jsonOK(w, locs)
}

// @Summary Rename an existing location
// @Tags locations
// @Accept json
// @Produce json
// @Param body body object true "old_name, new_name"
// @Success 200 {array} teleportLocation
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 503 {object} map[string]string
// @Router /api/v1/locations [put]
// PUT /api/v1/locations
func handleRenameLocation(w http.ResponseWriter, r *http.Request) {
	if globalLocationStore == nil {
		jsonErr(w, fmt.Errorf("location store not available"), http.StatusServiceUnavailable)
		return
	}
	var req struct {
		OldName string `json:"old_name"`
		NewName string `json:"new_name"`
	}
	if err := decode(r, &req); err != nil {
		jsonErr(w, err, http.StatusBadRequest)
		return
	}
	req.OldName = strings.TrimSpace(req.OldName)
	req.NewName = strings.TrimSpace(req.NewName)
	if req.OldName == "" || req.NewName == "" {
		jsonErr(w, fmt.Errorf("old_name and new_name required"), http.StatusBadRequest)
		return
	}
	if err := globalLocationStore.rename(req.OldName, req.NewName); err != nil {
		if strings.Contains(err.Error(), "not found") {
			jsonErr(w, err, http.StatusNotFound)
			return
		}
		jsonErr(w, err, http.StatusInternalServerError)
		return
	}
	locs, err := globalLocationStore.list()
	if err != nil {
		jsonErr(w, fmt.Errorf("list after rename: %w", err), http.StatusInternalServerError)
		return
	}
	jsonOK(w, locs)
}

// @Summary Delete a named location
// @Tags locations
// @Accept json
// @Produce json
// @Param body body object true "name"
// @Success 200 {array} teleportLocation
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 503 {object} map[string]string
// @Router /api/v1/locations [delete]
// DELETE /api/v1/locations
func handleDeleteLocation(w http.ResponseWriter, r *http.Request) {
	if globalLocationStore == nil {
		jsonErr(w, fmt.Errorf("location store not available"), http.StatusServiceUnavailable)
		return
	}
	var req struct {
		Name string `json:"name"`
	}
	if err := decode(r, &req); err != nil {
		jsonErr(w, err, http.StatusBadRequest)
		return
	}
	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		jsonErr(w, fmt.Errorf("name required"), http.StatusBadRequest)
		return
	}
	if err := globalLocationStore.delete(req.Name); err != nil {
		if strings.Contains(err.Error(), "not found") {
			jsonErr(w, err, http.StatusNotFound)
			return
		}
		jsonErr(w, err, http.StatusInternalServerError)
		return
	}
	locs, err := globalLocationStore.list()
	if err != nil {
		jsonErr(w, fmt.Errorf("list after delete: %w", err), http.StatusInternalServerError)
		return
	}
	jsonOK(w, locs)
}
