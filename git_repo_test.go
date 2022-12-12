package gt

import "testing"

func TestLocal(t *testing.T) {
	r, err := New("", "master", ".")
	if err != nil {
		t.Errorf("New Error")
	}
	r.Open()
	t.Logf("Latest SHA: %s", r.LatestSHA(7))
	if r.RevisionExists("non-existent") {
		t.Errorf("It should not exist!")
	}
	if r.RevisionExists("master") {
		t.Logf("Everything works as expected!")
	}
}
