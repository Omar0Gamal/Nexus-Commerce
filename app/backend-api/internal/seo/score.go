package seo

import (
	"regexp"
	"strings"
	"unicode/utf8"

	"backend-api/internal/db"
	"backend-api/internal/shared/pgutil"
)

// maxSEOScore is the maximum achievable SEO score.
const maxSEOScore = int16(100)

// slugOK matches a well-formed URL slug: only a-z, 0-9, single hyphens, no leading/trailing hyphens.
var slugOK = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)

// ScoreProduct computes an SEO score (0–100) for a product.
//
// Scoring criteria:
//
//	+15  seo_title is present and 30–60 characters
//	+20  seo_description is present and 50–155 characters
//	+15  description length >= 100 characters
//	+10  slug is well-formed (a-z, 0-9, single hyphens only)
//	+10  review_count > 0
//	+10  price > 0
//	+10  category is assigned (category_id is non-null)
//	+10  title is 20–80 characters
//	────
//	100  maximum
func ScoreProduct(r db.GetAllProductsForSEOAuditRow) int16 {
	score := int16(0)

	// +15 SEO title quality
	if r.SeoTitle.Valid {
		l := utf8.RuneCountInString(r.SeoTitle.String)
		if l >= 30 && l <= 60 {
			score += 15
		} else if l > 0 {
			score += 7 // partial credit: present but not ideal length
		}
	}

	// +20 SEO description quality
	if r.SeoDescription.Valid {
		l := utf8.RuneCountInString(r.SeoDescription.String)
		if l >= 50 && l <= 155 {
			score += 20
		} else if l > 0 {
			score += 10 // partial credit
		}
	}

	// +15 description length
	if r.Description.Valid && utf8.RuneCountInString(r.Description.String) >= 100 {
		score += 15
	}

	// +10 well-formed slug
	if slugOK.MatchString(strings.ToLower(r.Slug)) {
		score += 10
	}

	// +10 has at least one review
	if r.ReviewCount > 0 {
		score += 10
	}

	// +10 price > 0
	if pgutil.NumericToFloat(r.Price) > 0 {
		score += 10
	}

	// +10 category assigned
	if r.CategoryID.Valid {
		score += 10
	}

	// +10 title length 20–80 chars
	tl := utf8.RuneCountInString(r.Title)
	if tl >= 20 && tl <= 80 {
		score += 10
	}

	if score > maxSEOScore {
		score = maxSEOScore
	}
	return score
}
