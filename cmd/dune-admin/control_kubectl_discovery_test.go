package main

import (
	"testing"
)

// TestKubectlControl_VmHostIP verifies that vmHostIP resolves from the struct's
// sshHost field (not the stale process-wide global), covering the multi-server
// boot regression in issue #234.
func TestKubectlControl_VmHostIP(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		sshHost  string
		wantHost string
	}{
		{"bare ip", "172.30.0.100", "172.30.0.100"},
		{"ip:port", "172.30.0.100:22", "172.30.0.100"},
		{"hostname only", "myserver.local", "myserver.local"},
		{"hostname:port", "myserver.local:2222", "myserver.local"},
		{"empty → empty", "", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &kubectlControl{sshHost: tt.sshHost}
			got := c.vmHostIP()
			if got != tt.wantHost {
				t.Errorf("vmHostIP() = %q, want %q", got, tt.wantHost)
			}
		})
	}
}

// TestKubectlControl_VmHostIP_Override verifies that a non-empty hostOverride
// takes precedence over sshHost, allowing operators to manually specify the
// host used in auto-discovered Web Interface URLs.
func TestKubectlControl_VmHostIP_Override(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name         string
		sshHost      string
		hostOverride string
		wantHost     string
	}{
		{"override wins over sshHost", "172.30.0.100", "10.0.0.50", "10.0.0.50"},
		{"override with port stripped", "172.30.0.100", "10.0.0.50:9090", "10.0.0.50"},
		{"no override → sshHost used", "172.30.0.100", "", "172.30.0.100"},
		{"both empty → empty", "", "", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &kubectlControl{sshHost: tt.sshHost, hostOverride: tt.hostOverride}
			got := c.vmHostIP()
			if got != tt.wantHost {
				t.Errorf("vmHostIP() = %q, want %q", got, tt.wantHost)
			}
		})
	}
}

// TestWebInterfaceURL_HostRewrite exercises the webInterfaceURL helper that
// rewrites the CRD-reported host with the operator's VM IP.
func TestWebInterfaceURL_HostRewrite(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		vmHost string
		addr   string
		want   string
	}{
		{"rewrite host with port", "172.30.0.100", "10.0.0.1:30881", "http://172.30.0.100:30881/"},
		{"rewrite host with port 2", "172.30.0.100", "10.0.0.1:18888", "http://172.30.0.100:18888/"},
		{"vmHost empty → keep addr host", "", "10.0.0.1:30881", "http://10.0.0.1:30881/"},
		{"empty addr → skip", "172.30.0.100", "", ""},
		{"addr no port", "172.30.0.100", "10.0.0.1", "http://172.30.0.100/"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := webInterfaceURL(tt.vmHost, tt.addr)
			if got != tt.want {
				t.Errorf("webInterfaceURL(%q, %q) = %q, want %q", tt.vmHost, tt.addr, got, tt.want)
			}
		})
	}
}

// TestNewControlPlane_KubectlPassesSSHHost verifies that newControlPlane wires
// cfg.SSHHost into the kubectlControl so vmHostIP returns the right host even
// after the multi-server boot blanks the process-wide sshHost global.
func TestNewControlPlane_KubectlPassesSSHHost(t *testing.T) {
	t.Parallel()
	cfg := appConfig{
		SSHHost:          "172.30.0.100",
		ControlNamespace: "funcom-seabass-mybg",
	}
	cp := newControlPlane("kubectl", cfg)
	kc, ok := cp.(*kubectlControl)
	if !ok {
		t.Fatalf("expected *kubectlControl, got %T", cp)
	}
	if kc.sshHost != "172.30.0.100" {
		t.Errorf("kubectlControl.sshHost = %q, want %q", kc.sshHost, "172.30.0.100")
	}
	// vmHostIP must use the struct field, not the global sshHost (which is "").
	sshHost = "" // simulate the state after multi-server boot
	if got := kc.vmHostIP(); got != "172.30.0.100" {
		t.Errorf("vmHostIP() after global cleared = %q, want %q", got, "172.30.0.100")
	}
}
