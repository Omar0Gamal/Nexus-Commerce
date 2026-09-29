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

type sitemapIndex struct {
	XMLName  xml.Name            `xml:"sitemapindex"`
	Xmlns    string              `xml:"xmlns,attr"`
	Sitemaps []sitemapIndexEntry `xml:"sitemap"`
}

type sitemapIndexEntry struct {
	Loc     string `xml:"loc"`
	LastMod string `xml:"lastmod"`
}

func (s *Service) GenerateSitemap(ctx context.Context, shopID uuid.UUID, baseURL string) ([]byte, error) {
	cacheKey := fmt.Sprintf("sitemap:%s", shopID)

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

	urls = append(urls, sitemapURL{
		Loc:        baseURL + "/",
		ChangeFreq: "daily",
		Priority:   1.0,
	})

	for _, path := range []string{"/about", "/contact", "/products"} {
		urls = append(urls, sitemapURL{
			Loc:        baseURL + path,
			ChangeFreq: "weekly",
			Priority:   0.5,
		})
	}

	for _, cat := range categories {
		urls = append(urls, sitemapURL{
			Loc:        baseURL + "/collections/" + cat.Slug,
			ChangeFreq: "weekly",
			Priority:   0.7,
		})
	}

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

	_ = s.rdb.Set(ctx, cacheKey, xmlBytes, sitemapCacheTTL).Err()

	return xmlBytes, nil
}

func (s *Service) InvalidateSitemapCache(ctx context.Context, shopID uuid.UUID) {
	_ = s.rdb.Del(ctx, fmt.Sprintf("sitemap:%s", shopID)).Err()
}

func (s *Service) Robots(_ context.Context, baseURL string) string {
	return fmt.Sprintf("User-agent: *\nDisallow: /api/\nDisallow: /admin/\nDisallow: /checkout/\n\nSitemap: %s/sitemap.xml\n", baseURL)
}

func buildSitemapXML(urls []sitemapURL, _ string) ([]byte, error) {
	set := sitemapURLSet{
		Xmlns: "http://www.sitemaps.org/schemas/sitemap/0.9",
		URLs:  urls,
	}
	if len(urls) > sitemapMaxURLs {
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

func (s *Service) LookupRedirect(ctx context.Context, shopID uuid.UUID, fromPath string) (string, int, bool) {
	cacheKey := fmt.Sprintf("redirects:%s", shopID)

	cachedEntry, err := s.rdb.HGet(ctx, cacheKey, fromPath).Result()
	if err == nil {
		var code int
		var toPath string
		fmt.Sscanf(cachedEntry, "%d|%s", &code, &toPath)
		go func() {
			_ = s.q.IncrementRedirectHitCount(ctx, db.IncrementRedirectHitCountParams{
				ShopID:   shopID,
				FromPath: fromPath,
			})
		}()
		return toPath, code, true
	}

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

func defaultShopURL(shop db.Shop) string {
	if shop.CustomDomain.Valid && shop.CustomDomain.String != "" {
		return "https://" + shop.CustomDomain.String
	}
	return "https://" + shop.Subdomain + ".nexus.commerce"
}
