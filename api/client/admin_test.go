package client

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/textileio/powergate/v2/api/client/admin"
	adminPb "github.com/textileio/powergate/v2/api/gen/powergate/admin/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	ctx = context.Background()
)

func TestCreate(t *testing.T) {
	t.Run("WithoutAdminToken", func(t *testing.T) {
		a, done := setupAdmin(t, "")
		defer done()

		resp, err := a.Users.Create(ctx)
		require.NoError(t, err)
		require.NotEmpty(t, resp.User.Id)
		require.NotEmpty(t, resp.User.Token)
	})

	t.Run("WithAdminToken", func(t *testing.T) {
		authToken := uuid.New().String()
		a, done := setupAdmin(t, authToken)
		defer done()

		t.Run("UnauthorizedEmpty", func(t *testing.T) {
			resp, err := a.Users.Create(ctx)
			require.Error(t, err)
			require.Nil(t, resp)
		})

		t.Run("UnauthorizedWrong", func(t *testing.T) {
			wrongAuths := []string{
				"",      // Empty
				"wrong", // Non-empty
			}
			for _, auth := range wrongAuths {
				ctx := context.WithValue(ctx, AdminKey, auth)
				resp, err := a.Users.Create(ctx)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.PermissionDenied, st.Code())
				require.Nil(t, resp)
			}
		})
		t.Run("Authorized", func(t *testing.T) {
			ctx := context.WithValue(ctx, AdminKey, authToken)
			resp, err := a.Users.Create(ctx)
			require.NoError(t, err)
			require.NotEmpty(t, resp.User.Id)
			require.NotEmpty(t, resp.User.Token)
		})
	})
}

func setupAdmin(t *testing.T, adminAuthToken string) (*admin.Admin, func()) {
	defConfig := defaultServerConfig(t)
	if adminAuthToken != "" {
		defConfig.FFSAdminToken = adminAuthToken
	}
	serverDone := setupServer(t, defConfig)
	conn, done := setupConnection(t, defConfig.GrpcHostAddress)
	return admin.NewAdmin(adminPb.NewAdminServiceClient(conn)), func() {
		done()
		serverDone()
	}
}
