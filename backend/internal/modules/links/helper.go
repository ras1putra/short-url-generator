package links

import (
	"database/sql"
	"sort"
	"time"

	"urlshortener/internal/modules/links/dto"
	"urlshortener/internal/repository"
	"urlshortener/pkg/constants"
	"urlshortener/pkg/response"
)

type clickRow struct {
	Country    sql.NullString
	Device     sql.NullString
	Browser    sql.NullString
	ClickDate  time.Time
	ClickCount int64
}

func computeExpiresAt(value int, unit string) (sql.NullTime, error) {
	var expiresAt sql.NullTime
	if value <= 0 {
		return expiresAt, nil
	}
	expiresAt.Valid = true
	switch unit {
	case constants.ExpiresUnitMinutes:
		expiresAt.Time = time.Now().Add(time.Duration(value) * time.Minute)
	case constants.ExpiresUnitHours:
		expiresAt.Time = time.Now().Add(time.Duration(value) * time.Hour)
	case constants.ExpiresUnitDays:
		expiresAt.Time = time.Now().AddDate(0, 0, value)
	default:
		return expiresAt, response.NewAppError(400, "ExpiresUnit must be one of: minutes, hours, days")
	}
	return expiresAt, nil
}

func aggregateClickRows(rows []clickRow) (map[string]int64, map[string]int64, []dto.CountryCount, []dto.DateCount) {
	browsers := make(map[string]int64)
	devices := make(map[string]int64)
	countryMap := make(map[string]int64)
	dateMap := make(map[string]int64)

	for _, row := range rows {
		if row.Browser.Valid && row.Browser.String != "" {
			browsers[row.Browser.String] += row.ClickCount
		}
		if row.Device.Valid && row.Device.String != "" {
			devices[row.Device.String] += row.ClickCount
		}
		if row.Country.Valid && row.Country.String != "" {
			countryMap[row.Country.String] += row.ClickCount
		} else {
			countryMap["Unknown"] += row.ClickCount
		}
		dateStr := row.ClickDate.Format("2006-01-02")
		dateMap[dateStr] += row.ClickCount
	}

	topCountries := make([]dto.CountryCount, 0, len(countryMap))
	for c, cnt := range countryMap {
		topCountries = append(topCountries, dto.CountryCount{Country: c, Count: cnt})
	}
	sort.Slice(topCountries, func(i, j int) bool {
		return topCountries[i].Count > topCountries[j].Count
	})
	if len(topCountries) > constants.MaxTopCountries {
		topCountries = topCountries[:constants.MaxTopCountries]
	}

	clicksPerDay := make([]dto.DateCount, 0, len(dateMap))
	for d, cnt := range dateMap {
		clicksPerDay = append(clicksPerDay, dto.DateCount{Date: d, Count: cnt})
	}
	sort.Slice(clicksPerDay, func(i, j int) bool {
		return clicksPerDay[i].Date > clicksPerDay[j].Date
	})

	return browsers, devices, topCountries, clicksPerDay
}

func clickRowsFromStatsRows(rows []repository.GetStatsBySlugRow) []clickRow {
	result := make([]clickRow, len(rows))
	for i, r := range rows {
		result[i] = clickRow(r)
	}
	return result
}

func clickRowsFromAggRows(rows []repository.GetAggregateStatsByUserRow) []clickRow {
	result := make([]clickRow, len(rows))
	for i, r := range rows {
		result[i] = clickRow(r)
	}
	return result
}

func buildStatsResponse(totalClicks, uniqueClicks int64, clickRows []clickRow) *dto.StatsResponse {
	browsers, devices, topCountries, clicksPerDay := aggregateClickRows(clickRows)
	return &dto.StatsResponse{
		TotalClicks:  totalClicks,
		UniqueClicks: uniqueClicks,
		ClicksPerDay: clicksPerDay,
		TopCountries: topCountries,
		Browsers:     browsers,
		Devices:      devices,
	}
}
