package validate

import "testing"

func TestSessionID(t *testing.T) {
	if err := SessionID(""); err == nil {
		t.Fatal("expected error for empty")
	}
	if err := SessionID("not-uuid"); err == nil {
		t.Fatal("expected error for invalid uuid")
	}
	if err := SessionID("a19b1da6-465a-406b-bdf9-7ee15f56068f"); err != nil {
		t.Fatalf("valid uuid: %v", err)
	}
}

func TestRecipient(t *testing.T) {
	cases := []struct {
		in    string
		valid bool
	}{
		{"6281234567890", true},
		{"+6281234567890", true},
		{"120363123456789012@g.us", true},
		{"08xxx", false},
		{"", false},
	}
	for _, c := range cases {
		err := Recipient(c.in)
		if c.valid && err != nil {
			t.Errorf("%q: %v", c.in, err)
		}
		if !c.valid && err == nil {
			t.Errorf("%q: expected error", c.in)
		}
	}
}

func TestMessageType(t *testing.T) {
	for _, tpe := range []string{"text", "image", "document", ""} {
		if _, err := MessageType(tpe); err != nil {
			t.Errorf("%q: %v", tpe, err)
		}
	}
	if _, err := MessageType("video"); err == nil {
		t.Fatal("video should fail")
	}
}

func TestHTTPURL(t *testing.T) {
	if err := HTTPURL("https://example.com/a.pdf"); err != nil {
		t.Fatal(err)
	}
	if err := HTTPURL("ftp://x.com"); err == nil {
		t.Fatal("ftp should fail")
	}
}

func TestPagination(t *testing.T) {
	l, o, err := Pagination(0, 0)
	if err != nil || l != DefaultListLimit || o != 0 {
		t.Fatalf("defaults: %d %d %v", l, o, err)
	}
	if _, _, err := Pagination(500, 0); err == nil {
		t.Fatal("limit too high")
	}
}