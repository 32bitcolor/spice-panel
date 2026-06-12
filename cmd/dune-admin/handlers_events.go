package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
)

// ── CRUD handlers ─────────────────────────────────────────────────────────────

func handleListEvents(w http.ResponseWriter, r *http.Request) {
	if globalEventStore == nil {
		jsonErr(w, fmt.Errorf("events store not available"), http.StatusServiceUnavailable)
		return
	}
	events, err := globalEventStore.list()
	if err != nil {
		log.Printf("handleListEvents: %v", err)
		jsonErr(w, fmt.Errorf("internal error"), http.StatusInternalServerError)
		return
	}
	jsonOK(w, events)
}

func handleCreateEvent(w http.ResponseWriter, r *http.Request) {
	if globalEventStore == nil {
		jsonErr(w, fmt.Errorf("events store not available"), http.StatusServiceUnavailable)
		return
	}
	var req eventDefinition
	if err := decode(r, &req); err != nil {
		jsonErr(w, fmt.Errorf("invalid request body"), http.StatusBadRequest)
		return
	}
	if req.Name == "" {
		jsonErr(w, fmt.Errorf("name is required"), http.StatusBadRequest)
		return
	}
	if !isValidEventType(req.Type) {
		jsonErr(w, fmt.Errorf("invalid type %q: must be %q or %q", req.Type, eventTypeZoneRace, eventTypeMilestone), http.StatusBadRequest)
		return
	}
	created, err := globalEventStore.create(req)
	if err != nil {
		log.Printf("handleCreateEvent: %v", err)
		jsonErr(w, fmt.Errorf("internal error"), http.StatusInternalServerError)
		return
	}
	jsonOK(w, created)
}

func handleUpdateEvent(w http.ResponseWriter, r *http.Request) {
	if globalEventStore == nil {
		jsonErr(w, fmt.Errorf("events store not available"), http.StatusServiceUnavailable)
		return
	}
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		jsonErr(w, fmt.Errorf("invalid id"), http.StatusBadRequest)
		return
	}
	var req eventDefinition
	if err := decode(r, &req); err != nil {
		jsonErr(w, fmt.Errorf("invalid request body"), http.StatusBadRequest)
		return
	}
	req.ID = id
	if req.Name == "" {
		jsonErr(w, fmt.Errorf("name is required"), http.StatusBadRequest)
		return
	}
	if !isValidEventType(req.Type) {
		jsonErr(w, fmt.Errorf("invalid type %q", req.Type), http.StatusBadRequest)
		return
	}
	updated, err := globalEventStore.update(req)
	if err != nil {
		if errors.Is(err, errNotFound) {
			jsonErr(w, fmt.Errorf("event not found"), http.StatusNotFound)
			return
		}
		log.Printf("handleUpdateEvent: %v", err)
		jsonErr(w, fmt.Errorf("internal error"), http.StatusInternalServerError)
		return
	}
	jsonOK(w, updated)
}

func handleDeleteEvent(w http.ResponseWriter, r *http.Request) {
	if globalEventStore == nil {
		jsonErr(w, fmt.Errorf("events store not available"), http.StatusServiceUnavailable)
		return
	}
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		jsonErr(w, fmt.Errorf("invalid id"), http.StatusBadRequest)
		return
	}
	if err := globalEventStore.delete(id); err != nil {
		if errors.Is(err, errNotFound) {
			jsonErr(w, fmt.Errorf("event not found"), http.StatusNotFound)
			return
		}
		log.Printf("handleDeleteEvent: %v", err)
		jsonErr(w, fmt.Errorf("internal error"), http.StatusInternalServerError)
		return
	}
	jsonOK(w, map[string]bool{"ok": true})
}

// ── enable / disable ──────────────────────────────────────────────────────────

func handleSetEventEnabled(w http.ResponseWriter, r *http.Request) {
	if globalEventStore == nil {
		jsonErr(w, fmt.Errorf("events store not available"), http.StatusServiceUnavailable)
		return
	}
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		jsonErr(w, fmt.Errorf("invalid id"), http.StatusBadRequest)
		return
	}
	var req struct {
		Enabled bool `json:"enabled"`
	}
	if err := decode(r, &req); err != nil {
		jsonErr(w, fmt.Errorf("invalid request body"), http.StatusBadRequest)
		return
	}
	if err := globalEventStore.setEnabled(id, req.Enabled); err != nil {
		if errors.Is(err, errNotFound) {
			jsonErr(w, fmt.Errorf("event not found"), http.StatusNotFound)
			return
		}
		log.Printf("handleSetEventEnabled: %v", err)
		jsonErr(w, fmt.Errorf("internal error"), http.StatusInternalServerError)
		return
	}
	jsonOK(w, map[string]bool{"ok": true})
}

// ── status ────────────────────────────────────────────────────────────────────

func handleGetEventStatus(w http.ResponseWriter, r *http.Request) {
	if globalEventStore == nil {
		jsonErr(w, fmt.Errorf("events store not available"), http.StatusServiceUnavailable)
		return
	}
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		jsonErr(w, fmt.Errorf("invalid id"), http.StatusBadRequest)
		return
	}
	def, err := globalEventStore.get(id)
	if err != nil {
		if errors.Is(err, errNotFound) {
			jsonErr(w, fmt.Errorf("event not found"), http.StatusNotFound)
			return
		}
		log.Printf("handleGetEventStatus: %v", err)
		jsonErr(w, fmt.Errorf("internal error"), http.StatusInternalServerError)
		return
	}
	claims, err := globalEventStore.listClaims(id)
	if err != nil {
		log.Printf("handleGetEventStatus listClaims: %v", err)
		jsonErr(w, fmt.Errorf("internal error"), http.StatusInternalServerError)
		return
	}
	jsonOK(w, map[string]any{
		"event":  def,
		"claims": claims,
	})
}

// ── reset ─────────────────────────────────────────────────────────────────────

func handleResetEvent(w http.ResponseWriter, r *http.Request) {
	if globalEventStore == nil {
		jsonErr(w, fmt.Errorf("events store not available"), http.StatusServiceUnavailable)
		return
	}
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		jsonErr(w, fmt.Errorf("invalid id"), http.StatusBadRequest)
		return
	}
	if _, err := globalEventStore.get(id); err != nil {
		if errors.Is(err, errNotFound) {
			jsonErr(w, fmt.Errorf("event not found"), http.StatusNotFound)
			return
		}
		log.Printf("handleResetEvent get: %v", err)
		jsonErr(w, fmt.Errorf("internal error"), http.StatusInternalServerError)
		return
	}
	if err := globalEventStore.clearClaims(id); err != nil {
		log.Printf("handleResetEvent clearClaims: %v", err)
		jsonErr(w, fmt.Errorf("internal error"), http.StatusInternalServerError)
		return
	}
	jsonOK(w, map[string]bool{"ok": true})
}

// ── validation ────────────────────────────────────────────────────────────────

func isValidEventType(t eventType) bool {
	return t == eventTypeZoneRace || t == eventTypeMilestone
}

// ── config ────────────────────────────────────────────────────────────────────

// eventsConfigPayload is the request/response shape for the events config endpoints.
type eventsConfigPayload struct {
	Enabled *bool `json:"events_enabled"`
}

func eventsConfigFromLoaded() eventsConfigPayload {
	return eventsConfigPayload{Enabled: loadedConfig.EventsEnabled}
}

func handleGetEventsConfig(w http.ResponseWriter, _ *http.Request) {
	jsonOK(w, eventsConfigFromLoaded())
}

func handleSaveEventsConfig(w http.ResponseWriter, r *http.Request) {
	var p eventsConfigPayload
	if err := decode(r, &p); err != nil {
		jsonErr(w, fmt.Errorf("decode: %w", err), http.StatusBadRequest)
		return
	}

	loadedConfig.EventsEnabled = p.Enabled

	if err := writeConfigFile(loadedConfig); err != nil {
		log.Printf("handleSaveEventsConfig: %v", err)
		jsonErr(w, fmt.Errorf("failed to write config"), http.StatusInternalServerError)
		return
	}

	applyEventEngine(loadedConfig)
	jsonOK(w, eventsConfigFromLoaded())
}
