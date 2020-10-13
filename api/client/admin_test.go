package client

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	proto "github.com/textileio/powergate/proto/admin/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestCreate(t *testing.T) {
	t.Run("WithoutAdminToken", func(t *testing.T) {
		a, done := setupAdmin(t, "")
		defer done()

		resp, err := a.CreateInstance(ctx)
		require.NoError(t, err)
		require.NotEmpty(t, resp.Id)
		require.NotEmpty(t, resp.Token)
	})

	t.Run("WithAdminToken", func(t *testing.T) {
		authToken := uuid.New().String()
		a, done := setupAdmin(t, authToken)
		defer done()

		t.Run("UnauthorizedEmpty", func(t *testing.T) {
			resp, err := a.CreateInstance(ctx)
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
				resp, err := a.CreateInstance(ctx)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.PermissionDenied, st.Code())
				require.Nil(t, resp)
			}
		})
		t.Run("Authorized", func(t *testing.T) {
			ctx := context.WithValue(ctx, AdminKey, authToken)
			resp, err := a.CreateInstance(ctx)
			require.NoError(t, err)
			require.NotEmpty(t, resp.Id)
			require.NotEmpty(t, resp.Token)
		})
	})
}

func setupAdmin(t *testing.T, adminAuthToken string) (*Admin, func()) {
	defConfig := defaultServerConfig(t)
	if adminAuthToken != "" {
		defConfig.FFSAdminToken = adminAuthToken
	}
	serverDone := setupServer(t, defConfig)
	conn, done := setupConnection(t)
	return &Admin{client: proto.NewPowergateAdminServiceClient(conn)}, func() {
		done()
		serverDone()
	}
}
