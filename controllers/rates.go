package controllers

import (
    "encoding/json"
    "net/http"

    "tersoh-backend/models"
    "tersoh-backend/internal/utils"
)

type RateAverage struct {
    Currency string  ` + "`json:"currency"`" + `
    Average  float64 ` + "`json:"average"`" + `
}

func OurRates(w http.ResponseWriter, r *http.Request) {
    var cur []string
    utils.DB.Model(&models.Post{}).Distinct("currency").Pluck("currency", &cur)
    var res []RateAverage
    for _, c := range cur {
        var avg float64
        utils.DB.Model(&models.Post{}).
            Where("currency = ?", c).
            Order("created_at desc").
            Limit(5).
            Select("AVG(rate)").Row().Scan(&avg)
        res = append(res, RateAverage{c, avg})
    }
    utils.RespondJSON(w, http.StatusOK, res)
}
