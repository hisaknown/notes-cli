package notes

import (
	"github.com/rhysd/go-fakeio"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"
)

func testNewConfigForNewCmd(subdir string) *Config {
	cwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	return &Config{
		GitPath:  "git",
		HomePath: filepath.Join(cwd, "testdata", "new", subdir),
	}
}

func TestNewCmdNewNote(t *testing.T) {
	cfg := testNewConfigForNewCmd("empty")
	fake := fakeio.Stdout().Stdin("this\nis\ntest").CloseStdin()
	defer fake.Restore()

	cmd := &NewCmd{
		Config:   cfg,
		Category: "cat",
		Filename: "test",
		Tags:     "foo, bar",
	}

	if err := cmd.Do(); err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(filepath.Join(cfg.HomePath, "cat"))

	p := filepath.Join(cfg.HomePath, "cat", "test.md")
	n, err := LoadNote(p, cfg)
	if err != nil {
		t.Fatal(err)
	}

	if n.Category != "cat" {
		t.Error(n.Category)
	}

	if !reflect.DeepEqual(n.Tags, []string{"foo", "bar"}) {
		t.Error("Tags are not correct", n.Tags)
	}

	if n.Title != "test" {
		t.Error(n.Title)
	}

	if n.Created.After(time.Now()) {
		t.Error("Created date invalid", n.Created.Format(time.RFC3339))
	}

	f, err := os.Open(p)
	if err != nil {
		t.Fatal("File was not created", err)
	}
	defer f.Close()

	b, err := ioutil.ReadAll(f)
	if err != nil {
		panic(err)
	}
	s := string(b)

	if !strings.Contains(s, "this\nis\ntest") {
		t.Fatal("Note body is not correct:", s)
	}

	if cfg.GitPath != "" {
		dotgit := filepath.Join(cfg.HomePath, ".git")
		if s, err := os.Stat(dotgit); err != nil || !s.IsDir() {
			t.Fatal(".git directory was not created. `git init` did not run:", err)
		}
		defer os.RemoveAll(dotgit)
	}

	stdout, err := fake.String()
	if err != nil {
		panic(err)
	}
	stdout = strings.TrimSuffix(stdout, "\n")
	if stdout != p {
		t.Error("Output is not path to the file:", stdout)
	}

	// Second note

	cmd = &NewCmd{
		Config:   cfg,
		Category: "cat",
		Filename: "test2",
		Tags:     "foo, bar",
		NoInline: true,
	}
	if err := cmd.Do(); err != nil {
		t.Fatal(err)
	}

	n, err = LoadNote(filepath.Join(cfg.HomePath, "cat", "test2.md"), cfg)
	if err != nil {
		t.Fatal(err)
	}

	if n.Title != "test2" {
		t.Fatal(n.Title)
	}

	// Check .git is still there and does not raise an error
	if cfg.GitPath != "" {
		dotgit := filepath.Join(cfg.HomePath, ".git")
		if s, err := os.Stat(dotgit); err != nil || !s.IsDir() {
			t.Fatal(".git directory was not created. `git init` did not run:", err)
		}
	}
}

func TestNewCmdNewNoteWithNoInlineInput(t *testing.T) {
	fake := fakeio.Stdout()
	defer fake.Restore()

	cfg := testNewConfigForNewCmd("empty")

	cmd := &NewCmd{
		Config:   cfg,
		Category: "cat",
		Filename: "test3",
		Tags:     "foo, bar",
		NoInline: true,
	}

	if err := cmd.Do(); err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(filepath.Join(cfg.HomePath, "cat"))

	p := filepath.Join(cfg.HomePath, "cat", "test3.md")
	n, err := LoadNote(p, cfg)
	if err != nil {
		t.Fatal(err)
	}

	if n.Category != "cat" {
		t.Error(n.Category)
	}

	if !reflect.DeepEqual(n.Tags, []string{"foo", "bar"}) {
		t.Error("Tags are not correct", n.Tags)
	}

	if n.Title != "test3" {
		t.Error(n.Title)
	}

	if n.Created.After(time.Now()) {
		t.Error("Created date invalid", n.Created.Format(time.RFC3339))
	}

	stdout, err := fake.String()
	if err != nil {
		panic(err)
	}
	stdout = strings.TrimSuffix(stdout, "\n")
	if stdout != p {
		t.Error("Output is not path to the file:", stdout)
	}
}

func TestNewCmdNoteAlreadyExists(t *testing.T) {
	cfg := testNewConfigForNewCmd("fail")
	cmd := &NewCmd{
		Config:   cfg,
		Category: "cat",
		Filename: "already-exists",
		Tags:     "",
		NoInline: true,
	}

	err := cmd.Do()
	if err == nil {
		t.Fatal("No error occurred")
	}
	if !strings.Contains(err.Error(), "already exists") {
		t.Fatal("Unexpected error:", err)
	}
}

func TestNewCmdNoteInvalidInput(t *testing.T) {
	cfg := testNewConfigForNewCmd("fail")
	cmd := &NewCmd{
		Config:   cfg,
		Category: "", // Empty category is not permitted
		Filename: "test",
		Tags:     "",
		NoInline: true,
	}

	err := cmd.Do()
	if err == nil {
		t.Fatal("No error occurred")
	}
	if !strings.Contains(err.Error(), "Invalid category as directory name") {
		t.Fatal("Unexpected error:", err)
	}
}
