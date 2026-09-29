package seo

import (
	"regexp"
	"strings"
	"unicode/utf8"

	"backend-api/internal/db"
	"backend-api/internal/shared/pgutil"
)

const maxSEOScore = int16(100)

var slugOK = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)

func ScoreProduct(r db.GetAllProductsForSEOAuditRow) int16 {
	score := int16(0)

	if r.SeoTitle.Valid {
		l := utf8.RuneCountInString(r.SeoTitle.String)
		if l >= 30 && l <= 60 {
			score += 15
		} else if l > 0 {
			score += 7
		}
	}

	if r.SeoDescription.Valid {
		l := utf8.RuneCountInString(r.SeoDescription.String)
		if l >= 50 && l <= 155 {
			score += 20
		} else if l > 0 {
			score += 10
		}
	}

	if r.Description.Valid && utf8.RuneCountInString(r.Description.String) >= 100 {
		score += 15
	}

	if slugOK.MatchString(strings.ToLower(r.Slug)) {
		score += 10
	}

	if r.ReviewCount > 0 {
		score += 10
	}

	if pgutil.NumericToFloat(r.Price) > 0 {
		score += 10
	}

	if r.CategoryID.Valid {
		score += 10
	}

	tl := utf8.RuneCountInString(r.Title)
	if tl >= 20 && tl <= 80 {
		score += 10
	}

	if score > maxSEOScore {
		score = maxSEOScore
	}
	return score
}
