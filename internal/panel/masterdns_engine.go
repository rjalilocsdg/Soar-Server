package panel

import (
	"context"
	"net"
	"time"

	masterdns "masterdnsvpn-go/server"
)

func (m *Manager) masterDNSFor(rt *accountRuntime) (*masterdns.Server, error) {
	rt.masterMu.Lock()
	defer rt.masterMu.Unlock()
	if rt.ctx.Err() != nil {
		return nil, ErrDenied
	}
	if rt.master != nil {
		return rt.master, nil
	}
	seed, err := m.store.Seed(rt.client.ID)
	if err != nil {
		return nil, err
	}
	domains := make([]string, 0, 2*len(rt.zones))
	for _, zone := range rt.zones {
		domains = append(domains, "master."+zone, "mdns-1."+zone)
	}
	engine, err := masterdns.New(rt.ctx, masterdns.Config{
		EncryptionMethod: rt.client.encryptionMethod(), Domain: domains[0], Domains: domains, Key: masterDNSKey(seed),
		MaxActiveSessions: rt.client.MasterDNSSessionLimit, MaxStreamsPerSession: cottenStreamLimit(rt.client.StreamLimit),
		Allowed: func() bool {
			c, err := m.store.Get(rt.client.ID)
			return err == nil && c.Status(time.Now()) == "active" && c.hasEngine("masterdns") && c.UsageEpoch == rt.client.UsageEpoch && c.PublicKey == rt.client.PublicKey && c.encryptionMethod() == rt.client.encryptionMethod()
		},
		DialContext: func(ctx context.Context, network, address string) (net.Conn, error) {
			return m.dialForEngine(rt, ctx, network, address, upstreamMasterDNS)
		},
		ExchangeDNS: func(ctx context.Context, query []byte) ([]byte, error) {
			return m.exchangeEngineDNS(rt, ctx, query, upstreamMasterDNS)
		},
	})
	if err != nil {
		m.diagnostics.masterStartErrors.Add(1)
		m.diagnostics.lastEngineIssueUnixMS.Store(time.Now().UnixMilli())
		return nil, err
	}
	m.diagnostics.masterStarts.Add(1)
	m.diagnostics.masterRunning.Add(1)
	rt.master = engine
	return engine, nil
}
