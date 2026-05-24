package links

import (
	"database/sql"
	"testing"
	"time"

	"urlshortener/pkg/response"

	"github.com/stretchr/testify/assert"
)

func TestComputeExpiresAt_NoExpiry(t *testing.T) {
	result, err := computeExpiresAt(0, "")
	assert.NoError(t, err)
	assert.False(t, result.Valid)
}

func TestComputeExpiresAt_Minutes(t *testing.T) {
	result, err := computeExpiresAt(30, "minutes")
	assert.NoError(t, err)
	assert.True(t, result.Valid)
	assert.WithinDuration(t, time.Now().Add(30*time.Minute), result.Time, time.Second)
}

func TestComputeExpiresAt_Hours(t *testing.T) {
	result, err := computeExpiresAt(2, "hours")
	assert.NoError(t, err)
	assert.True(t, result.Valid)
	assert.WithinDuration(t, time.Now().Add(2*time.Hour), result.Time, time.Second)
}

func TestComputeExpiresAt_Days(t *testing.T) {
	result, err := computeExpiresAt(7, "days")
	assert.NoError(t, err)
	assert.True(t, result.Valid)
	assert.WithinDuration(t, time.Now().AddDate(0, 0, 7), result.Time, time.Second)
}

func TestComputeExpiresAt_InvalidUnit(t *testing.T) {
	_, err := computeExpiresAt(5, "years")
	assert.Error(t, err)
	var appErr *response.AppError
	assert.ErrorAs(t, err, &appErr)
	assert.Equal(t, 400, appErr.Code)
}

func TestAggregateClickRows_Empty(t *testing.T) {
	browsers, devices, topCountries, clicksPerDay := aggregateClickRows([]clickRow{})
	assert.Empty(t, browsers)
	assert.Empty(t, devices)
	assert.Empty(t, topCountries)
	assert.Empty(t, clicksPerDay)
}

func TestAggregateClickRows_SingleRow(t *testing.T) {
	now := time.Now()
	rows := []clickRow{
		{
			Country:    sql.NullString{String: "US", Valid: true},
			Device:     sql.NullString{String: "desktop", Valid: true},
			Browser:    sql.NullString{String: "Chrome", Valid: true},
			ClickDate:  now,
			ClickCount: 10,
		},
	}

	browsers, devices, topCountries, clicksPerDay := aggregateClickRows(rows)

	assert.Equal(t, int64(10), browsers["Chrome"])
	assert.Equal(t, int64(10), devices["desktop"])
	assert.Len(t, topCountries, 1)
	assert.Equal(t, "US", topCountries[0].Country)
	assert.Len(t, clicksPerDay, 1)
}

func TestAggregateClickRows_NullBrowserDevice(t *testing.T) {
	now := time.Now()
	rows := []clickRow{
		{
			Country:    sql.NullString{String: "US", Valid: true},
			ClickDate:  now,
			ClickCount: 5,
		},
	}

	browsers, devices, _, _ := aggregateClickRows(rows)
	assert.Empty(t, browsers)
	assert.Empty(t, devices)
}

func TestAggregateClickRows_UnknownCountry(t *testing.T) {
	now := time.Now()
	rows := []clickRow{
		{
			ClickDate:  now,
			ClickCount: 5,
		},
	}

	_, _, topCountries, _ := aggregateClickRows(rows)
	assert.Len(t, topCountries, 1)
	assert.Equal(t, "Unknown", topCountries[0].Country)
	assert.Equal(t, int64(5), topCountries[0].Count)
}

func TestBuildStatsResponse(t *testing.T) {
	now := time.Now()
	rows := []clickRow{
		{
			Country:    sql.NullString{String: "US", Valid: true},
			Device:     sql.NullString{String: "mobile", Valid: true},
			Browser:    sql.NullString{String: "Safari", Valid: true},
			ClickDate:  now,
			ClickCount: 15,
		},
	}

	resp := buildStatsResponse(100, 80, rows)
	assert.Equal(t, int64(100), resp.TotalClicks)
	assert.Equal(t, int64(80), resp.UniqueClicks)
	assert.Equal(t, int64(15), resp.Browsers["Safari"])
	assert.Equal(t, int64(15), resp.Devices["mobile"])
}
