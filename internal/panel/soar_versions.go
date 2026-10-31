package panel

import (
	"context"
	"errors"
	"net"

	"github.com/getlantern/soar"
	soarlegacy "soar-v023"
)

// Each version has its own pinned package. Both use the same account identity,
// virtual DNS listener and guarded/metered dialer supplied by runAccount.
func serveSoarVersion(version string, ctx context.Context, pc net.PacketConn, cfg soar.ServerConfig) error {
	switch version {
	case LegacySoarVersion:
		return soarlegacy.Serve(ctx, pc, soarlegacy.ServerConfig(cfg))
	case DefaultSoarVersion:
		return soar.Serve(ctx, pc, cfg)
	default:
		return errors.New("unsupported Soar server version")
	}
}
