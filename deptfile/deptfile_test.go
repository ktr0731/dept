package deptfile_test

import (
	"bytes"
	"os"
	"testing"

	"github.com/ktr0731/dept/deptfile"
)

func TestFile(t *testing.T) {
	expected := `[[requirements]]
  name = "github.com/ktr0731/evans"
`
	var buf bytes.Buffer
	f := &deptfile.File{
		Writer: &buf,
	}
	f.Requirements = append(f.Requirements, &deptfile.Requirement{Name: "github.com/ktr0731/evans"})

	err := f.Encode()
	if err != nil {
		t.Fatalf("expected Encode has no errors, but got an error: %s", err)
	}
	if actual := buf.String(); expected != actual {
		t.Errorf("expected %s, but got %s", expected, actual)
	}
}

func TestLoad(t *testing.T) {
	t.Run("Load must return ErrNotFound because deptfile missing", func(t *testing.T) {
		cleanup := changeDeptfileName("test.toml")
		defer cleanup()

		_, err := deptfile.Load()
		if err != deptfile.ErrNotFound {
			t.Fatalf("Load must return ErrNotFound, but got %s", err)
		}
	})

	t.Run("Load must return *File if deptfile found", func(t *testing.T) {
		const testDeptfile = "test.toml"
		cleanup := changeDeptfileName(testDeptfile)
		defer cleanup()

		f, err := os.Create(testDeptfile)
		if err != nil {
			t.Fatalf("deptfile must be open, but got an error: %s", err)
		}
		defer func() {
			f.Close()
			os.Remove(testDeptfile)
		}()

		df := &deptfile.File{
			Writer: f,
		}
		err = df.Encode()
		if err != nil {
			t.Fatalf("deptfile must be encode to a file, but got an error: %s", err)
		}

		_, err = deptfile.Load()
		if err != nil {
			t.Errorf("deptfile must be load by Load, but got an error: %s", err)
		}
	})
}
