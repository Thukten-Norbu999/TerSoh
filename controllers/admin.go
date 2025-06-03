package controllers

import (
	"net/http"
	"tersoh-backend/config"
	"tersoh-backend/utils"
)

type DailyLogin struct {
	Date  string `json:"date"`
	Count int64  `json:"count"`
}

type Retention struct {
	CohortDate string  `json:"cohort_date"`
	Day1       float32 `json:"day_1"`
	Day7       float32 `json:"day_7"`
}

func ComputeAnalytics(w http.ResponseWriter, r *http.Request) {
	var dailies []DailyLogin
	config.DB.Raw(
		"SELECT to_char(logged_at::date,'YYYY-MM-DD') AS date, count(*) AS count " +
			"FROM login_events " +
			"WHERE logged_at >= NOW() - INTERVAL '14 days' " +
			"GROUP BY date ORDER BY date").
		Scan(&dailies)

	type cohortSz struct {
		CohortDate string
		Size       int64
	}
	var cohorts []cohortSz
	config.DB.Raw(
		"SELECT to_char(created_at::date,'YYYY-MM-DD') AS cohort_date, count(*) AS size " +
			"FROM users " +
			"WHERE created_at >= NOW() - INTERVAL '14 days' " +
			"GROUP BY cohort_date ORDER BY cohort_date").
		Scan(&cohorts)

	var ret []Retention
	for _, c := range cohorts {
		var day1, day7 int64
		config.DB.Raw(
			"SELECT count(distinct uid) FROM login_events "+
				"WHERE uid IN (SELECT uid FROM users WHERE created_at::date = ?) "+
				"AND logged_at::date = (?::date + INTERVAL '1 day')::date",
			c.CohortDate, c.CohortDate).
			Scan(&day1)
		config.DB.Raw(
			"SELECT count(distinct uid) FROM login_events "+
				"WHERE uid IN (SELECT uid FROM users WHERE created_at::date = ?) "+
				"AND logged_at::date = (?::date + INTERVAL '7 day')::date",
			c.CohortDate, c.CohortDate).
			Scan(&day7)
		ret = append(ret, Retention{CohortDate: c.CohortDate, Day1: float32(day1) / float32(c.Size), Day7: float32(day7) / float32(c.Size)})
	}

	utils.RespondJSON(w, http.StatusOK, map[string]interface{}{"daily_logins": dailies, "retention": ret})
}
