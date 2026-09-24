package srv

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestAdminIdentityRequiresExplicitAllowlist(t *testing.T) {
	for _, tc := range []struct {
		name, id, email string
		allow           []string
		want            bool
	}{
		{"owner", "uid", "owner@example.test", []string{"owner@example.test"}, true},
		{"case and whitespace", "uid", " OWNER@example.test ", []string{" owner@example.test "}, true},
		{"shared visitor", "uid", "customer@example.test", []string{"owner@example.test"}, false},
		{"identity without email", "uid", "", []string{"owner@example.test"}, false},
		{"email without identity", "", "owner@example.test", []string{"owner@example.test"}, false},
		{"empty configuration", "uid", "owner@example.test", nil, false},
		{"empty entry", "uid", "", []string{""}, false},
		{"suffix is not a match", "uid", "owner@example.test.evil", []string{"owner@example.test"}, false},
		{"no loopback identity exception", "local-admin", "local@localhost", []string{"owner@example.test"}, false},
		{"explicit local launcher", "local-admin", "local@localhost", []string{"local@localhost"}, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s := &Server{AdminEmails: tc.allow}
			r := httptest.NewRequest("GET", "/", nil)
			r.Header.Set("X-ExeDev-UserID", tc.id)
			r.Header.Set("X-ExeDev-Email", tc.email)
			if got := s.isAdmin(r); got != tc.want {
				t.Fatalf("isAdmin = %v, want %v", got, tc.want)
			}
		})
	}
	s := &Server{AdminEmails: []string{"owner@example.test"}}
	r := httptest.NewRequest("GET", "/", nil)
	r.Header.Set("X-ExeDev-UserID", "uid")
	r.Header.Set("X-ExeDev-Email", "owner@example.test")
	r.Header.Add("X-ExeDev-Email", "customer@example.test")
	if s.isAdmin(r) {
		t.Fatal("ambiguous identity headers granted admin")
	}
}

func TestSharedIdentityCannotBypassProjectOrAdminGates(t *testing.T) {
	s, ts, cleanup := testServer(t)
	defer cleanup()
	seedClient(t, s, "protected", "Protected", "client-password")
	resp := apiRequestAdmin(t, ts, "POST", "/api/projects", map[string]string{
		"name": "Protected", "client_slug": "protected", "project_slug": "book", "start_date": "2026-09-24",
	})
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("create project: %d", resp.StatusCode)
	}
	var created map[string]any
	decodeJSON(t, resp, &created)
	pid := int64(created["ID"].(float64))

	for path, want := range map[string]int{
		"/admin/":                               http.StatusForbidden,
		"/api/admin/passes":                     http.StatusForbidden,
		"/api/projects/" + itoa(pid) + "/tasks": http.StatusUnauthorized,
	} {
		r := httptest.NewRequest("GET", path, nil)
		r.Header.Set("X-ExeDev-UserID", "shared-user")
		r.Header.Set("X-ExeDev-Email", "customer@example.test")
		w := httptest.NewRecorder()
		s.Handler().ServeHTTP(w, r)
		if w.Code != want {
			t.Errorf("%s: got %d want %d", path, w.Code, want)
		}
	}
	r := httptest.NewRequest("POST", "/", nil)
	r.Header.Set("X-ExeDev-UserID", "shared-user")
	r.Header.Set("X-ExeDev-Email", "customer@example.test")
	w := httptest.NewRecorder()
	if _, admin, ok := s.requirePassAccess(w, r, pid); ok || admin {
		t.Fatal("shared identity bypassed factory pass gate")
	}
}

func TestProjectCredentialsRequireAdminEvenBeforeFirstToken(t *testing.T) {
	s, ts, cleanup := testServer(t)
	defer cleanup()
	seedClient(t, s, "protected", "Protected", "client-password")
	resp := apiRequestAdmin(t, ts, "POST", "/api/projects", map[string]string{
		"name": "Protected", "client_slug": "protected", "project_slug": "book", "start_date": "2026-09-24",
	})
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("create project: %d", resp.StatusCode)
	}
	var created map[string]any
	decodeJSON(t, resp, &created)
	pid := itoa(int64(created["ID"].(float64)))
	for _, identity := range []bool{false, true} {
		r := httptest.NewRequest("POST", "/api/projects/"+pid+"/auth", strings.NewReader(`{"password":"unapproved"}`))
		if identity {
			r.Header.Set("X-ExeDev-UserID", "shared-user")
			r.Header.Set("X-ExeDev-Email", "customer@example.test")
		}
		w := httptest.NewRecorder()
		s.Handler().ServeHTTP(w, r)
		if w.Code != http.StatusUnauthorized && w.Code != http.StatusForbidden {
			t.Fatalf("unapproved token creation: %d", w.Code)
		}
		if len(w.Result().Cookies()) != 0 {
			t.Fatal("denied request set a credential cookie")
		}
	}
	var count int
	if err := s.DB.QueryRow("SELECT COUNT(*) FROM auth_tokens WHERE project_id = ?", pid).Scan(&count); err != nil || count != 0 {
		t.Fatalf("unauthorized tokens persisted: count=%d err=%v", count, err)
	}
	resp = apiRequestAdmin(t, ts, "POST", "/api/projects/"+pid+"/auth", map[string]string{"password": "approved"})
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("owner token creation: %d", resp.StatusCode)
	}
}

func TestCustomerCannotMoveOrDuplicateAcrossClients(t *testing.T) {
	s, ts, cleanup := testServer(t)
	defer cleanup()
	seedClient(t, s, "source", "Source", "client-password")
	seedClient(t, s, "target", "Target", "other-password")
	resp := apiRequestAdmin(t, ts, "POST", "/api/projects", map[string]string{
		"name": "Source", "client_slug": "source", "project_slug": "book", "start_date": "2026-09-24",
	})
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("create project: %d", resp.StatusCode)
	}
	var created map[string]any
	decodeJSON(t, resp, &created)
	pid := itoa(int64(created["ID"].(float64)))
	resp = apiRequest(t, ts, "POST", "/api/clients/source/verify", map[string]string{"password": "client-password"})
	cookies := resp.Cookies()
	resp.Body.Close()
	if len(cookies) == 0 {
		t.Fatal("expected customer session")
	}
	for _, method := range []string{"PUT", "POST"} {
		path := "/api/projects/" + pid
		if method == "POST" {
			path += "/duplicate"
		}
		for _, client := range []string{"target", "source"} {
			r := httptest.NewRequest(method, path, strings.NewReader(`{"name":"Copy","client_slug":"`+client+`","project_slug":"copy-`+strings.ToLower(method)+`","start_date":"2026-09-24"}`))
			r.AddCookie(cookies[0])
			w := httptest.NewRecorder()
			s.Handler().ServeHTTP(w, r)
			if client == "target" && w.Code != http.StatusForbidden {
				t.Fatalf("%s cross-client: %d", method, w.Code)
			}
			if client == "source" && (w.Code < 200 || w.Code >= 300) {
				t.Fatalf("%s same-client: %d", method, w.Code)
			}
		}
	}
	var count int
	if err := s.DB.QueryRow("SELECT COUNT(*) FROM projects WHERE client_slug = 'target'").Scan(&count); err != nil || count != 0 {
		t.Fatalf("target tenancy changed: count=%d err=%v", count, err)
	}
}

func TestMultipartRequestBoundedBeforeParsing(t *testing.T) {
	for _, size := range []int{32, 2 << 20} {
		var body bytes.Buffer
		mw := multipart.NewWriter(&body)
		file, err := mw.CreateFormFile("file", "test.docx")
		if err != nil {
			t.Fatal(err)
		}
		_, _ = file.Write(bytes.Repeat([]byte("x"), size))
		_ = mw.Close()
		r := httptest.NewRequest("POST", "/", &body)
		r.Header.Set("Content-Type", mw.FormDataContentType())
		w := httptest.NewRecorder()
		ok := parseUploadForm(w, r, 64)
		if r.MultipartForm != nil {
			defer r.MultipartForm.RemoveAll()
		}
		if size == 32 && !ok {
			t.Fatal("small upload rejected")
		}
		if size > 64 && (ok || w.Code != http.StatusRequestEntityTooLarge) {
			t.Fatalf("oversized upload: accepted=%v status=%d", ok, w.Code)
		}
	}
}
