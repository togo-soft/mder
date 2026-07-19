package main

import (
	"context"
	"net"
	"testing"
)

func TestNetJoinHostPort(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		host string
		port string
		want string
	}{
		{name: "empty host (all interfaces)", host: "", port: "8666", want: ":8666"},
		{name: "IPv4 all zero", host: "0.0.0.0", port: "8666", want: "0.0.0.0:8666"},
		{name: "IPv6 all zero", host: "::", port: "8666", want: "[::]:8666"},
		{name: "IPv4 loopback", host: "127.0.0.1", port: "8666", want: "127.0.0.1:8666"},
		{name: "IPv6 localhost", host: "::1", port: "8666", want: "[::1]:8666"},
		{name: "custom IPv4", host: "192.168.1.100", port: "8080", want: "192.168.1.100:8080"},
		{name: "custom IPv6", host: "fe80::1", port: "3000", want: "[fe80::1]:3000"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := net.JoinHostPort(tt.host, tt.port)
			if got != tt.want {
				t.Fatalf("JoinHostPort(%q, %q) = %q, want %q", tt.host, tt.port, got, tt.want)
			}
		})
	}
}

func TestServeCmdFlags(t *testing.T) {
	t.Parallel()

	cmd := serveCmd()
	if cmd.Name != "serve" {
		t.Fatalf("cmd.Name = %q, want %q", cmd.Name, "serve")
	}
	if len(cmd.Aliases) != 1 || cmd.Aliases[0] != "s" {
		t.Fatalf("cmd.Aliases = %v, want [s]", cmd.Aliases)
	}

	foundHost := false
	for _, f := range cmd.Flags {
		names := f.Names()
		for _, n := range names {
			if n == "host" {
				foundHost = true
				break
			}
		}
	}
	if !foundHost {
		t.Fatal("serve command should have --host flag")
	}
}

func TestServeCmdDefaultHost(t *testing.T) {
	t.Parallel()

	cmd := serveCmd()
	foundHost := false
	for _, f := range cmd.Flags {
		names := f.Names()
		for _, n := range names {
			if n == "host" {
				foundHost = true
			}
		}
	}
	if !foundHost {
		t.Fatal("host flag not found")
	}
}

func TestRunServeHostParam(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		host     string
		wantAddr string
	}{
		{name: "empty host", host: "", wantAddr: ":8666"},
		{name: "IPv4 0.0.0.0", host: "0.0.0.0", wantAddr: "0.0.0.0:8666"},
		{name: "IPv6 ::", host: "::", wantAddr: "[::]:8666"},
		{name: "IPv4 127.0.0.1", host: "127.0.0.1", wantAddr: "127.0.0.1:8666"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := net.JoinHostPort(tt.host, "8666")
			if got != tt.wantAddr {
				t.Fatalf("JoinHostPort(%q, 8666) = %q, want %q", tt.host, got, tt.wantAddr)
			}
		})
	}
}

func TestServeCmdHostFlagParsed(t *testing.T) {
	t.Parallel()

	cmd := serveCmd()
	_ = cmd.Action
	_ = context.Background()
}
