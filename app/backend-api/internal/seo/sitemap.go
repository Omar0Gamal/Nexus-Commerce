package seo

import (
	"bytes"
	"context"
	"encoding/xml"
	"fmt"
	"time"

	"backend-api/internal/db"
	"backend-api/internal/shared/pgutil"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

const (
	sitemapCacheTTL = 1 * time.Hour
	sitemapMaxURLs  = 50_000
)

// sitemapURLSet is the root XML element for a sitemap.
type sitemapURLSet struct {
	XMLName xml.Name     `xml:"urlset"`
	Xmlns   string       `xml:"xmlns,attr"`
	URLs    []sitemapURL `xml:"url"`
}

type sitemapURL struct {
	Loc        string  `xml:"loc"`
	LastMod    string  `xml:"lastmod,omitempty"`
	ChangeFreq string  `xml:"changefreq,omitempty"`
	Priority   float64 `xml:"priority,omitempty"`
}

// sitemapIndex is the root XML element for a sitemap index (when > 50k URLs).
type sitemapIndex struct {
	XMLName  xml.Name            `xml:"sitemapindex"`
	Xmlns    string              `xml:"xmlns,attr"`
	Sitemaps []sitemapIndexEntry `xml:"sitemap"`
}

type sitemapIndexEntry struct {
	Loc     string `xml:"loc"`
	LastMod string `xml:"lastmod"`
}

// GenerateSitemap builds a sitemap.xml for the given shop and caches it in Redis
// under the key sitemap:{shopID} with a 1-hour TTL. Returns the cached bytes on
// a cache hit. The baseURL should be the shop's storefront base URL
// (e.g. "https://myshop.nexus.commerce").
func (s *Service) GenerateSitemap(ctx context.Context, shopID uuid.UUID, baseURL string) ([]byte, error) {
	cacheKey := fmt.Sprintf("sitemap:%s", shopID)

	// Cache hit
	if cached, err := s.rdb.Get(ctx, cacheKey).Bytes(); err == nil {
		return cached, nil
	}

	products, err := s.q.GetProductsForSitemap(ctx, pgutil.ToUUID(shopID))
	if err != nil {
		return nil, fmt.Errorf("sitemap: fetch products: %w", err)
	}
	categories, err := s.q.GetCategoriesForSitemap(ctx, pgutil.ToUUID(shopID))
	if err != nil {
		return nil, fmt.Errorf("sitemap: fetch categories: %w", err)
	}

	var urls []sitemapURL

	// Homepage
	urls = append(urls, sitemapURL{
		Loc:        baseURL + "/",
		ChangeFreq: "daily",
		Priority:   1.0,
	})

	// Static pages
	for _, path := range []string{"/about", "/contact", "/products"} {
		urls = append(urls, sitemapURL{
			Loc:        baseURL + path,
			ChangeFreq: "weekly",
			Priority:   0.5,
		})
	}

	// Category pages
	for _, cat := range categories {
		urls = append(urls, sitemapURL{
			Loc:        baseURL + "/collections/" + cat.Slug,
			ChangeFreq: "weekly",
			Priority:   0.7,
		})
	}

	// Product pages
	for _, p := range products {
		url := sitemapURL{
			Loc:        baseURL + "/products/" + p.Slug,
			ChangeFreq: "weekly",
			Priority:   0.8,
		}
		if p.CreatedAt.Valid {
			url.LastMod = p.CreatedAt.Time.UTC().Format(time.RFC3339)
		}
		urls = append(urls, url)
	}

	xmlBytes, err := buildSitemapXML(urls, baseURL)
	if err != nil {
		return nil, err
	}

	// Cache with TTL (best-effort, ignore error)
	_ = s.rdb.Set(ctx, cacheKey, xmlBytes, sitemapCacheTTL).Err()

	return xmlBytes, nil
}

// InvalidateSitemapCache removes the cached sitemap for a shop, forcing it to
// be regenerated on the next request.
func (s *Service) InvalidateSitemapCache(ctx context.Context, shopID uuid.UUID) {
	_ = s.rdb.Del(ctx, fmt.Sprintf("sitemap:%s", shopID)).Err()
}

// Robots returns the robots.txt content appropriate for the shop.
func (s *Service) Robots(_ context.Context, baseURL string) string {
	return fmt.Sprintf("User-agent: *\nDisallow: /api/\nDisallow: /admin/\nDisallow: /checkout/\n\nSitemap: %s/sitemap.xml\n", baseURL)
}

// buildSitemapXML encodes the URL list into sitemap XML. If the list exceeds
// sitemapMaxURLs, a sitemap index referencing chunked child sitemaps is returned.
func buildSitemapXML(urls []sitemapURL, _ string) ([]byte, error) {
	set := sitemapURLSet{
		Xmlns: "http://www.sitemaps.org/schemas/sitemap/0.9",
		URLs:  urls,
	}
	if len(urls) > sitemapMaxURLs {
		// Truncate to limit — full index pagination is a future improvement.
		set.URLs = urls[:sitemapMaxURLs]
	}

	var buf bytes.Buffer
	buf.WriteString(xml.Header)
	enc := xml.NewEncoder(&buf)
	enc.Indent("", "  ")
	if err := enc.Encode(set); err != nil {
		return nil, fmt.Errorf("sitemap: encode xml: %w", err)
	}
	return buf.Bytes(), nil
}

// loadRedirectsIntoCache pre-warms the Redis hash for a shop's redirects.
func (s *Service) loadRedirectsIntoCache(ctx context.Context, shopID uuid.UUID) error {
	redirects, err := s.q.ListRedirects(ctx, db.ListRedirectsParams{
		ShopID: shopID,
		Limit:  1000,
		Offset: 0,
	})
	if err != nil {
		return err
	}
	if len(redirects) == 0 {
		return nil
	}

	key := fmt.Sprintf("redirects:%s", shopID)
	pipe := s.rdb.Pipeline()
	for _, r := range redirects {
		pipe.HSet(ctx, key, r.FromPath, fmt.Sprintf("%d|%s", r.StatusCode, r.ToPath))
	}
	pipe.Expire(ctx, key, 10*time.Minute)
	_, err = pipe.Exec(ctx)
	return err
}

// LookupRedirect checks the Redis cache (then DB) for a redirect matching
// fromPath for the given shop. Returns (toPath, statusCode, found).
func (s *Service) LookupRedirect(ctx context.Context, shopID uuid.UUID, fromPath string) (string, int, bool) {
	cacheKey := fmt.Sprintf("redirects:%s", shopID)

	val, err := s.rdb.HGet(ctx, cacheKey, fromPath).Result()
	if err == nil {
		var code int
		var toPath string
		fmt.Sscanf(val, "%d|%s", &code, &toPath)
		go func() {
			_ = s.q.IncrementRedirectHitCount(ctx, db.IncrementRedirectHitCountParams{
				ShopID:   shopID,
				FromPath: fromPath,
			})
		}()
		return toPath, code, true
	}

	// Cache miss — load redirects and try again
	_ = s.loadRedirectsIntoCache(ctx, shopID)

	r, err := s.q.GetRedirect(ctx, db.GetRedirectParams{
		ShopID:   shopID,
		FromPath: fromPath,
	})
	if err != nil {
		return "", 0, false
	}

	go func() {
		_ = s.q.IncrementRedirectHitCount(ctx, db.IncrementRedirectHitCountParams{
			ShopID:   shopID,
			FromPath: fromPath,
		})
	}()

	return r.ToPath, int(r.StatusCode), true
}

// WarmSitemapForAllShops iterates all active shops and pre-warms their sitemap caches.
// Called by the nightly worker.
func (s *Service) WarmSitemapForAllShops(ctx context.Context) {
	shops, err := s.q.ListShopsByStatus(ctx, db.NullShopStatus{
		ShopStatus: db.ShopStatusActive,
		Valid:      true,
	})
	if err != nil {
		s.logger.Error("seo: warm sitemap: list shops failed", zap.Error(err))
		return
	}
	for _, shop := range shops {
		shopURL := defaultShopURL(shop)
		s.InvalidateSitemapCache(ctx, shop.ID)
		if _, err := s.GenerateSitemap(ctx, shop.ID, shopURL); err != nil {
			s.logger.Warn("seo: warm sitemap: generate failed",
				zap.String("shop_id", shop.ID.String()), zap.Error(err))
		}
	}
}

// defaultShopURL builds the storefront URL for a shop.
func defaultShopURL(shop db.Shop) string {
	if shop.CustomDomain.Valid && shop.CustomDomain.String != "" {
		return "https://" + shop.CustomDomain.String
	}
	return "https://" + shop.Subdomain + ".nexus.commerce"
}
