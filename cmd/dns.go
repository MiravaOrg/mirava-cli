package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"strings"

	"github.com/MiravaOrg/mirava-core"
	"github.com/spf13/cobra"
)

// addDNSFlag registers the persistent --dns flag on the root command so
// every subcommand inherits it. With --dns, each test subcommand first
// probes the mirror's hostname through DNS resolvers and reports
// per-resolver IPs, RTTs and reachability, then runs the normal test.
// Diverging answers across resolvers suggest DNS poisoning, filtering or
// broken geo-DNS.
//
// The flag takes an optional value (NoOptDefVal keeps bare --dns working):
//
//	--dns            probe the default set (system, Cloudflare, Google, Quad9)
//	--dns=1.1.1.1    probe only that resolver (custom UDP resolver)
//
// Cobra cannot bind an optional-value flag with the space form, so argv is
// preprocessed (see RewriteDNSArgs) to turn `--dns 1.1.1.1` into
// `--dns=1.1.1.1` before cobra parses it.
func addDNSFlag(root *cobra.Command) {
	root.PersistentFlags().String("dns", "",
		"Probe the mirror's DNS resolution before the test; optional value is a custom resolver address (e.g. 1.1.1.1)")
	root.PersistentFlags().Lookup("dns").NoOptDefVal = "default"
}

// RewriteDNSArgs rewrites `--dns <addr>` argv pairs into the `--dns=<addr>`
// form cobra can parse. The token after a bare --dns is taken as the flag's
// value only when it looks like a resolver address, so mirror URLs and
// package names are never swallowed.
//
// It is applied from main() before rootCmd.Execute(); cobra offers no hook
// for this (SetArgs just stores the slice).
func RewriteDNSArgs(args []string) []string {
	out := make([]string, 0, len(args))
	for i := 0; i < len(args); i++ {
		a := args[i]
		if a == "--dns" && i+1 < len(args) && looksLikeResolverAddr(args[i+1]) {
			out = append(out, "--dns="+args[i+1])
			i++
			continue
		}
		out = append(out, a)
	}
	return out
}

// looksLikeResolverAddr reports whether s could be a DNS resolver address:
// an IP (optionally with port) or a bare hostname. URLs (http://...) and
// anything with a path separator are excluded so mirror arguments are never
// mistaken for resolver addresses.
func looksLikeResolverAddr(s string) bool {
	if s == "" || strings.HasPrefix(s, "-") {
		return false
	}
	if strings.Contains(s, "://") || strings.Contains(s, "/") {
		return false
	}
	if host, _, err := net.SplitHostPort(s); err == nil {
		s = host
	}
	return net.ParseIP(s) != nil || !strings.Contains(s, " ")
}

// dnsRequested reports whether --dns was set, and returns the flag's value
// ("default" for bare --dns, otherwise the resolver address).
func dnsRequested(cmd *cobra.Command) (string, bool) {
	v, err := cmd.Flags().GetString("dns")
	if err != nil || v == "" {
		return "", false
	}
	return v, true
}

// runDNSProbe resolves the mirror's hostname through DNS resolvers and
// prints one line per provider plus a divergence warning when resolvers
// disagree. With resolverAddr set (from --dns=<addr>) it probes only that
// resolver via mirava.CustomDNS; otherwise the default public set runs.
//
// The custom provider is passed as a direct variadic argument: Go converts
// it to core's unexported dns.Provider parameter type at the call site,
// which a []dnsProvider slice cannot do (interface slices are not
// assignable across module boundaries).
func runDNSProbe(ctx context.Context, m *mirava.Mirava, mirrorURL, resolverAddr string) {
	var (
		res *dnsProbeResult
		err error
	)
	if resolverAddr != "" && resolverAddr != "default" {
		res, err = probeWith(ctx, m, mirrorURL, resolverAddr)
	} else {
		res, err = probeWith(ctx, m, mirrorURL)
	}
	if err != nil {
		fmt.Printf("DNS probe: error: %v\n", err)
		return
	}

	fmt.Printf("DNS probe for %s (port %s):\n", res.Host, res.Port)
	for _, p := range res.Providers {
		if p.Error != "" {
			fmt.Printf("  %-24s error: %s\n", p.Provider, p.Error)
			continue
		}
		line := fmt.Sprintf("  %-24s %s  rtt=%s", p.Provider, strings.Join(p.IPs, ","), p.ResolveRTT)
		if len(p.IPs) > 0 {
			if p.Reachable {
				line += fmt.Sprintf("  connect=%s ok", p.ConnectRTT)
			} else {
				line += "  connect=failed"
			}
		}
		fmt.Println(line)
	}
	if res.Divergence != "" {
		fmt.Printf("  ⚠ %s\n", res.Divergence)
	}
	fmt.Println()
}

// dnsProbeResult is the CLI-side view of core's probe result. The concrete
// type lives in mirava-core/internal/dns, which this module cannot import,
// so the raw result is round-tripped through JSON and decoded here.
type dnsProbeResult struct {
	Host       string
	Port       string
	Providers  []dnsProviderResult
	Divergence string
}

type dnsProviderResult struct {
	Provider   string   `json:"provider"`
	IPs        []string `json:"ips"`
	ResolveRTT string   `json:"resolve_rtt"`
	ConnectRTT string   `json:"connect_rtt,omitempty"`
	Reachable  bool     `json:"reachable"`
	Error      string   `json:"error,omitempty"`
}

// probeWith calls Mirava.DNSProbe and adapts its result. The custom
// resolver (when set) is passed through as a direct variadic argument so Go
// converts it to core's dns.Provider at the call site; interface slices
// cannot be converted across module boundaries, but individual values can.
func probeWith(ctx context.Context, m *mirava.Mirava, mirrorURL string, custom ...any) (*dnsProbeResult, error) {
	var (
		raw any
		err error
	)
	switch len(custom) {
	case 0:
		raw, err = m.DNSProbe(ctx, mirrorURL)
	case 1:
		// The concrete type is dns.UDP (unexported package). A direct
		// variadic call converts it; assert via the exported constructor to
		// rebuild the value at the call site instead of storing it in any.
		addr, _ := custom[0].(string)
		raw, err = m.DNSProbe(ctx, mirrorURL, mirava.CustomDNS(addr))
	}
	if err != nil {
		return nil, err
	}

	// The concrete ProbeResult type is internal; marshal and re-decode into
	// the CLI-side view. All fields carry JSON tags in core, so this is
	// lossless.
	b, err := json.Marshal(raw)
	if err != nil {
		return nil, err
	}
	var res dnsProbeResult
	if err := json.Unmarshal(b, &res); err != nil {
		return nil, err
	}
	return &res, nil
}
