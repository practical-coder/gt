package gt

import (
	"io"
	"os"
	"testing"
)

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

func TestCloneOrPull(t *testing.T) {
	r, err := New("", "master", ".")
	if err != nil {
		t.Errorf("New Error")
	}
	f, err := os.Open("test/keys/gt_id_rsa")
	if err != nil {
		t.Errorf("Open test key error")
	}
	defer f.Close()
	key, err := io.ReadAll(f)
	if err != nil {
		t.Errorf("Reading key error")
	}
	err = r.SetKeys(key)
	if err != nil {
		t.Errorf("SetKeys Error")
	}
	err = r.CloneOrPull()
	if err != nil {
		t.Errorf("CloneOrPull Error")
	}
}
