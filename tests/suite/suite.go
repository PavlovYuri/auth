package suite

import (
	"auth/internal/config"
	"context"
	"net"
	"strconv"
	"testing"

	authv1 "github.com/PavlovYuri/protos/gen/go/auth"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	grpcHost = "localhost"
)

type Suite struct {
	*testing.T
	Cfg        *config.Config
	AuthClient authv1.AuthClient
}

func New(t *testing.T) (context.Context, *Suite) {
	t.Helper()
	t.Parallel()

	cfg := config.MustLoadByPath("./config/local_tests_config.yaml")
	ctx, cancelCtx := context.WithTimeout(context.Background(), cfg.GRPC.Timeout)

	t.Cleanup(func() {
		t.Helper()
		cancelCtx()
	})

	grpcAddress := net.JoinHostPort(grpcHost, strconv.Itoa(cfg.GRPC.Port))

	clientConn, err := grpc.DialContext(context.Background(), grpcAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("grpc server connection failed: %v", err)
	}

	authClient := authv1.NewAuthClient(clientConn)

	return ctx, &Suite{
		T:          t,
		Cfg:        cfg,
		AuthClient: authClient,
	}
}
