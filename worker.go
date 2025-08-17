package main

import (
	"context"
	"encoding/xml"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

type rss struct {
	Channel struct {
		Items []struct {
			Title       string `xml:"title"`
			Description string `xml:"description"`
			PubDate     string `xml:"pubDate"`
		} `xml:"item"`
	} `xml:"channel"`
}

type atom struct {
	Entries []struct {
		Title   string `xml:"title"`
		Content string `xml:"content"`
		Updated string `xml:"updated"`
	} `xml:"entry"`
}

func StartFeedWorker(ctx context.Context, feeds *FeedService, posts *PostService, interval time.Duration) {
	t := time.NewTicker(interval)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			processFeeds(ctx, feeds, posts)
		}
	}
}

func processFeeds(ctx context.Context, feeds *FeedService, posts *PostService) {
	flist, err := feeds.List(ctx)
	if err != nil {
		log.Printf("feed list error: %v", err)
		return
	}
	for _, f := range flist {
		if err := fetchAndIngest(ctx, f.URL, posts); err != nil {
			log.Printf("feed fetch error for %s: %v", f.URL, err)
		}
	}
}

func fetchAndIngest(ctx context.Context, url string, posts *PostService) error {
	client := &http.Client{Timeout: 15 * time.Second}
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	resp, err := client.Do(req)
	if err != nil { return err }
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil
	}
	b, err := io.ReadAll(resp.Body)
	if err != nil { return err }
	content := string(b)
	if strings.Contains(content, "<rss") || strings.Contains(content, "<channel") {
		var r rss
		if err := xml.Unmarshal(b, &r); err != nil { return err }
		for _, it := range r.Channel.Items {
			title := strings.TrimSpace(it.Title)
			desc := strings.TrimSpace(stripHTML(it.Description))
			if title == "" && desc == "" { continue }
			_, _ = posts.Create(ctx, nonEmpty(title, desc), firstNonEmpty(desc, title), strPtr(url), nil)
		}
		return nil
	}
	if strings.Contains(content, "<feed") && strings.Contains(content, "<entry") {
		var a atom
		if err := xml.Unmarshal(b, &a); err != nil { return err }
		for _, e := range a.Entries {
			title := strings.TrimSpace(e.Title)
			cnt := strings.TrimSpace(stripHTML(e.Content))
			if title == "" && cnt == "" { continue }
			_, _ = posts.Create(ctx, nonEmpty(title, cnt), firstNonEmpty(cnt, title), strPtr(url), nil)
		}
	}
	return nil
}

func stripHTML(s string) string {
	in := false
	var b strings.Builder
	for _, r := range s {
		switch r {
		case '<': in = true
		case '>': in = false
		default:
			if !in { b.WriteRune(r) }
		}
	}
	return b.String()
}

func nonEmpty(a, b string) string { if a != "" { return a }; return b }
func firstNonEmpty(a, b string) string { if a != "" { return a }; return b }
func strPtr(s string) *string { return &s }