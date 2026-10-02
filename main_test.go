package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestAction(t *testing.T) {
	for _, tc := range []struct {
		name, aggregate, dryRun, output string
		creates                         int
	}{
		{"individual", "false", "false", "41,42", 2},
		{"aggregate", "true", "false", "41", 1},
		{"dry-run", "false", "true", "", 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var titles, bodies []string
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/feed" && r.Header.Get("Authorization") != "Bearer test-token" {
					t.Error("GitHub request lacks token authentication")
				}
				switch {
				case r.URL.Path == "/feed":
					w.Header().Set("Content-Type", "application/rss+xml")
					fmt.Fprint(w, `<rss version="2.0"><channel><title>Test feed</title><link>https://example.com</link><description>Test</description><item><title>Existing</title></item><item><title>First</title><description>&lt;p&gt;First body&lt;/p&gt;</description></item><item><title>Second</title><description>Second body</description></item></channel></rss>`)
				case r.URL.Path == "/repos/owner/repo/issues" && r.Method == http.MethodGet:
					fmt.Fprint(w, `[{"number":1,"title":"[RSS] Existing"}]`)
				case r.URL.Path == "/repos/owner/repo/issues" && r.Method == http.MethodPost:
					var issue struct {
						Title, Body string
						Labels      []string
					}
					if err := json.NewDecoder(r.Body).Decode(&issue); err != nil {
						t.Error(err)
						w.WriteHeader(http.StatusBadRequest)
						return
					}
					if strings.Join(issue.Labels, ",") != "news" {
						t.Errorf("unexpected issue labels: %v", issue.Labels)
					}
					titles = append(titles, issue.Title)
					bodies = append(bodies, issue.Body)
					w.WriteHeader(http.StatusCreated)
					fmt.Fprintf(w, `{"number":%d}`, 40+len(titles))
				default:
					t.Errorf("unexpected request: %s %s", r.Method, r.URL)
					w.WriteHeader(http.StatusNotFound)
				}
			}))
			defer server.Close()
			serverURL, _ := url.Parse(server.URL)
			transport := http.DefaultTransport
			http.DefaultTransport = roundTripFunc(func(r *http.Request) (*http.Response, error) {
				r = r.Clone(r.Context())
				r.URL.Scheme, r.URL.Host = serverURL.Scheme, serverURL.Host
				return transport.RoundTrip(r)
			})
			defer func() { http.DefaultTransport = transport }()
			output := filepath.Join(t.TempDir(), "output")
			t.Setenv("GITHUB_OUTPUT", output)
			t.Setenv("GITHUB_REPOSITORY", "owner/repo")
			t.Setenv("INPUT_FEED", server.URL+"/feed")
			t.Setenv("INPUT_REPO-TOKEN", "test-token")
			t.Setenv("INPUT_PREFIX", "[RSS]")
			t.Setenv("INPUT_AGGREGATE", tc.aggregate)
			t.Setenv("INPUT_DRY-RUN", tc.dryRun)
			t.Setenv("INPUT_LABELS", "news")
			t.Setenv("INPUT_LASTTIME", "720h")
			t.Setenv("INPUT_CHARACTERLIMIT", "")
			t.Setenv("INPUT_TITLEFILTER", "")
			t.Setenv("INPUT_CONTENTFILTER", "")
			main()
			if len(titles) != tc.creates {
				t.Fatalf("created %d issues, want %d", len(titles), tc.creates)
			}
			if tc.creates == 2 && strings.Join(titles, ",") != "[RSS] First,[RSS] Second" {
				t.Fatalf("unexpected titles: %v", titles)
			}
			if tc.creates > 0 && (!strings.Contains(strings.Join(bodies, ""), "First body") || !strings.Contains(strings.Join(bodies, ""), "Second body")) {
				t.Fatalf("missing feed content: %v", bodies)
			}
			data, err := os.ReadFile(output)
			if err != nil {
				t.Fatalf("GITHUB_OUTPUT was not written: %v", err)
			}
			if !strings.Contains(string(data), "issues<<") || !strings.Contains(string(data), "\n"+tc.output+"\n") {
				t.Fatalf("unexpected output file: %q", data)
			}
		})
	}
}
