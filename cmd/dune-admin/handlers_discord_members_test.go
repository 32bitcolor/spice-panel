package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bwmarrin/discordgo"
)

func TestHandleSearchDiscordMembersInner(t *testing.T) {
	members := []*discordgo.Member{
		{Nick: "Nicky", User: &discordgo.User{ID: "1", Username: "userone", Avatar: "av1"}},
		{User: &discordgo.User{ID: "2", Username: "usertwo"}},
	}

	t.Run("maps members to rows", func(t *testing.T) {
		w := httptest.NewRecorder()
		handleSearchDiscordMembersInner(w, "guild1", "user", func(_, _ string, _ int) ([]*discordgo.Member, error) {
			return members, nil
		})
		if w.Code != http.StatusOK {
			t.Fatalf("status = %d", w.Code)
		}
		var rows []discordMemberRow
		if err := json.Unmarshal(w.Body.Bytes(), &rows); err != nil {
			t.Fatal(err)
		}
		if len(rows) != 2 {
			t.Fatalf("rows = %d", len(rows))
		}
		if rows[0].ID != "1" || rows[0].Name != "Nicky" || rows[0].Username != "userone" {
			t.Errorf("row[0] = %+v", rows[0])
		}
		if rows[0].Avatar == "" {
			t.Error("avatar URL not built for member with avatar hash")
		}
		if rows[1].Name != "usertwo" {
			t.Errorf("row[1] falls back to username: %+v", rows[1])
		}
		if rows[1].Avatar != "" {
			t.Errorf("avatar should be empty without hash: %+v", rows[1])
		}
	})

	t.Run("search error → 500", func(t *testing.T) {
		w := httptest.NewRecorder()
		handleSearchDiscordMembersInner(w, "guild1", "x", func(_, _ string, _ int) ([]*discordgo.Member, error) {
			return nil, errors.New("boom")
		})
		if w.Code != http.StatusInternalServerError {
			t.Errorf("status = %d", w.Code)
		}
	})
}

func TestHandleSearchDiscordMembers(t *testing.T) {
	t.Run("empty query → 400", func(t *testing.T) {
		oldCfg := loadedConfig
		defer func() { loadedConfig = oldCfg }()
		loadedConfig = appConfig{DiscordBotToken: "tok", DiscordGuildID: "g"}
		w := httptest.NewRecorder()
		handleSearchDiscordMembers(w, httptest.NewRequest(http.MethodGet, "/api/v1/discord/members/search", nil))
		if w.Code != http.StatusBadRequest {
			t.Errorf("status = %d, want 400", w.Code)
		}
	})

	t.Run("discord not configured → 503", func(t *testing.T) {
		oldCfg := loadedConfig
		defer func() { loadedConfig = oldCfg }()
		loadedConfig = appConfig{}
		w := httptest.NewRecorder()
		handleSearchDiscordMembers(w, httptest.NewRequest(http.MethodGet, "/api/v1/discord/members/search?q=ab", nil))
		if w.Code != http.StatusServiceUnavailable {
			t.Errorf("status = %d, want 503", w.Code)
		}
	})
}
