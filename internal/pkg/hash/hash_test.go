package hash

import "testing"

func TestDirHash(t *testing.T) {
	// Same path should produce same hash
	h1 := DirHash("/tmp/same")
	h2 := DirHash("/tmp/same")
	if h1 != h2 {
		t.Errorf("DirHash not deterministic: %s != %s", h1, h2)
	}
	if len(h1) != 8 {
		t.Errorf("DirHash length = %d, want 8", len(h1))
	}
	// Different paths should produce different hashes
	h3 := DirHash("/tmp/different")
	if h1 == h3 {
		t.Errorf("DirHash collision: %s == %s", h1, h3)
	}
	// Check absolute path normalization
	h4 := DirHash("/tmp/../tmp/same")
	h5 := DirHash("/tmp/same")
	if h4 != h5 {
		t.Errorf("DirHash not normalized: %s != %s", h4, h5)
	}
}
