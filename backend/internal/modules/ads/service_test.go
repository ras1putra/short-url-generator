package ads

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/shopspring/decimal"

	"urlshortener/internal/modules/ads/dto"
	"urlshortener/internal/repository"
	"urlshortener/internal/testutil"
	"urlshortener/pkg/response"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func appErrCode(err error) int {
	var appErr *response.AppError
	if errors.As(err, &appErr) {
		return appErr.Code
	}
	return -1
}

func newTestAdService(t *testing.T) (*AdService, *repository.Queries) {
	_ = zap.ReplaceGlobals(zap.NewNop())
	db, queries := testutil.SetupTestDB(t)
	return NewAdService(db, queries), queries
}

func createAdUser(t *testing.T, queries *repository.Queries, ctx context.Context) repository.User {
	user, err := queries.CreateUser(ctx, repository.CreateUserParams{
		Name:     "Advertiser",
		Email:    "adv-" + uuid.New().String() + "@example.com",
		Password: sql.NullString{String: "hashed", Valid: true},
		Role:     "advertiser",
	})
	require.NoError(t, err)

	err = queries.CreateWallet(ctx, repository.CreateWalletParams{
		UserID:  user.ID,
		Balance: "1000.00",
	})
	require.NoError(t, err)

	return user
}

func createSampleAd(t *testing.T, queries *repository.Queries, ctx context.Context, userID uuid.UUID, status string) repository.Ad {
	ad, err := queries.CreateAd(ctx, repository.CreateAdParams{
		AdvertiserID:    userID,
		Title:           "Test Ad",
		Description:     sql.NullString{String: "Description", Valid: true},
		ImageUrl:        "https://example.com/image.jpg",
		TargetUrl:       "https://example.com",
		Category:        "regular",
		TotalBudget:     "100.00",
		RemainingBudget: "80.00",
		Cpm:             "1.00",
		AdType:          "BANNER",
	})
	require.NoError(t, err)
	if status != "" && status != "active" {
		err = queries.UpdateAdStatus(ctx, repository.UpdateAdStatusParams{ID: ad.ID, Status: status})
		require.NoError(t, err)
		ad.Status = status
	}
	return ad
}

func TestCreate_Success(t *testing.T) {
	svc, queries := newTestAdService(t)
	ctx := context.Background()

	user := createAdUser(t, queries, ctx)

	req := dto.CreateAdRequest{
		Title:       "New Ad",
		Description: "Description",
		ImageURL:    "https://example.com/image.jpg",
		TargetURL:   "https://example.com",
		Category:    "regular",
		TotalBudget: 100.00,
		AdType:      "BANNER",
	}

	resp, err := svc.Create(ctx, user.ID, req)
	require.NoError(t, err)
	assert.NotEmpty(t, resp.ID)
	assert.Equal(t, "New Ad", resp.Title)
	assert.True(t, decimal.NewFromFloat(100.00).Equal(resp.TotalBudget))
	assert.True(t, decimal.NewFromFloat(100.00).Equal(resp.RemainingBudget))
	assert.True(t, decimal.NewFromFloat(1.00).Equal(resp.CPM))
	assert.Equal(t, "active", resp.Status)
}

func TestGetByID_Success(t *testing.T) {
	svc, queries := newTestAdService(t)
	ctx := context.Background()

	user := createAdUser(t, queries, ctx)
	ad := createSampleAd(t, queries, ctx, user.ID, "active")

	resp, err := svc.GetByID(ctx, ad.ID, user.ID)
	require.NoError(t, err)
	assert.Equal(t, ad.ID.String(), resp.ID)
	assert.Equal(t, "Test Ad", resp.Title)
}

func TestGetByID_NotFound(t *testing.T) {
	svc, _ := newTestAdService(t)
	ctx := context.Background()

	resp, err := svc.GetByID(ctx, uuid.New(), uuid.New())
	require.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, 404, appErrCode(err))
}

func TestGetByID_Forbidden(t *testing.T) {
	svc, queries := newTestAdService(t)
	ctx := context.Background()

	owner := createAdUser(t, queries, ctx)
	other := createAdUser(t, queries, ctx)
	ad := createSampleAd(t, queries, ctx, owner.ID, "active")

	resp, err := svc.GetByID(ctx, ad.ID, other.ID)
	require.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, 403, appErrCode(err))
}

func TestListByAdvertiser_Success(t *testing.T) {
	svc, queries := newTestAdService(t)
	ctx := context.Background()

	user := createAdUser(t, queries, ctx)
	createSampleAd(t, queries, ctx, user.ID, "active")
	createSampleAd(t, queries, ctx, user.ID, "paused")

	resp, err := svc.ListByAdvertiser(ctx, user.ID, 1, 10, "", "created_at", "desc")
	require.NoError(t, err)
	assert.Len(t, resp.Campaigns, 2)
}

func TestListByAdvertiser_Empty(t *testing.T) {
	svc, queries := newTestAdService(t)
	ctx := context.Background()

	user := createAdUser(t, queries, ctx)

	resp, err := svc.ListByAdvertiser(ctx, user.ID, 1, 10, "", "created_at", "desc")
	require.NoError(t, err)
	assert.Empty(t, resp.Campaigns)
}

func TestUpdate_Success(t *testing.T) {
	svc, queries := newTestAdService(t)
	ctx := context.Background()

	user := createAdUser(t, queries, ctx)
	ad := createSampleAd(t, queries, ctx, user.ID, "active")

	title := "Updated Ad"
	status := "paused"
	req := dto.UpdateAdRequest{Title: &title, Status: &status}

	resp, err := svc.Update(ctx, ad.ID, user.ID, req)
	require.NoError(t, err)
	assert.Equal(t, "Updated Ad", resp.Title)
	assert.Equal(t, "paused", resp.Status)
}

func TestUpdate_NotFound(t *testing.T) {
	svc, _ := newTestAdService(t)
	ctx := context.Background()

	req := dto.UpdateAdRequest{}
	resp, err := svc.Update(ctx, uuid.New(), uuid.New(), req)
	require.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, 404, appErrCode(err))
}

func TestDelete_Success(t *testing.T) {
	svc, queries := newTestAdService(t)
	ctx := context.Background()

	user := createAdUser(t, queries, ctx)
	ad := createSampleAd(t, queries, ctx, user.ID, "active")

	err := svc.Delete(ctx, ad.ID, user.ID)
	assert.NoError(t, err)

	updated, err := queries.GetAdByID(ctx, ad.ID)
	require.NoError(t, err)
	assert.Equal(t, "deleted", updated.Status)
}

func TestGetStats_Success(t *testing.T) {
	svc, queries := newTestAdService(t)
	ctx := context.Background()

	user := createAdUser(t, queries, ctx)
	ad := createSampleAd(t, queries, ctx, user.ID, "active")

	resp, err := svc.GetStats(ctx, ad.ID, user.ID)
	require.NoError(t, err)
	assert.Equal(t, ad.ID.String(), resp.AdID)
	assert.Equal(t, int64(0), resp.Impressions)
	assert.Equal(t, int64(0), resp.Clicks)
	assert.Equal(t, int64(0), resp.Completions)
}
