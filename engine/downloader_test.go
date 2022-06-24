package engine_test

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/UserProblem/reposcanner/engine"
)

func TestCloneRepositoryInvalidUrl(t *testing.T) {
	if tmpDir, err := os.MkdirTemp("", "reposcanner"); err != nil {
		t.Fatalf("Could not create temporary directory.")
	} else {
		engine.DeleteTmpDirectory(tmpDir)
		if err := engine.CloneRepository("not a url", "main", tmpDir); err == nil {
			t.Errorf("Expected failure, but error not returned.")
		}
	}
}

func TestCloneRepositoryUrlNotFound(t *testing.T) {
	if tmpDir, err := os.MkdirTemp("", "reposcanner"); err != nil {
		t.Fatalf("Could not create temporary directory.")
	} else {
		engine.DeleteTmpDirectory(tmpDir)
		if err := engine.CloneRepository("http://not.a/url", "main", tmpDir); err == nil {
			t.Errorf("Expected failure, but error not returned.")
		}
	}
}

func TestCloneRepository(t *testing.T) {
	if tmpDir, err := os.MkdirTemp("", "reposcanner"); err != nil {
		t.Fatalf("Could not create temporary directory.")
	} else {
		defer engine.DeleteTmpDirectory(tmpDir)
		if err := engine.CloneRepository("https://github.com/UserProblem/testdata.git", "master", tmpDir); err != nil {
			t.Fatalf("Clone repository failed: %v", err.Error())
		} else {
			expectedFiles := []string{
				"src.env",
				"subsrc1.c",
				"subsrc2.c",
				"subsubsrc1.go",
				"subsrc2.py",
			}

			fileList := make([]string, 0)
			err := filepath.WalkDir(tmpDir, func(path string, d fs.DirEntry, e error) error {
				if e == nil {
					if d.IsDir() {
						if d.Name() == ".git" {
							return filepath.SkipDir
						}
						return nil
					}

					fileList = append(fileList, d.Name())
					return nil
				}
				return e
			})

			if err != nil {
				t.Fatalf("Could not process the checkout directory: %v", err.Error())
			}

			if len(fileList) != len(expectedFiles) {
				t.Errorf("Expected %v files, but only got %v.\n", len(expectedFiles), len(fileList))
			}

			for i, fname := range fileList {
				if expectedFiles[i] != fname {
					t.Errorf("Expected '%v' but got '%v' instead.\n", expectedFiles[i], fname)
				}
			}
		}
	}
}

func TestCloneRepositoryDifferentBranch(t *testing.T) {
	if tmpDir, err := os.MkdirTemp("", "reposcanner"); err != nil {
		t.Fatalf("Could not create temporary directory.")
	} else {
		defer engine.DeleteTmpDirectory(tmpDir)
		if err := engine.CloneRepository("https://github.com/UserProblem/testdata.git", "dev1", tmpDir); err != nil {
			t.Fatalf("Clone repository failed: %v", err.Error())
		} else {
			found := false
			filepath.WalkDir(tmpDir, func(path string, d fs.DirEntry, e error) error {
				if e == nil {
					if d.IsDir() {
						if d.Name() == ".git" {
							return filepath.SkipDir
						}
						return nil
					}

					if d.Name() == "dev1.only" {
						found = true
						return fmt.Errorf("done")
					}
					return nil
				}
				return e
			})

			if !found {
				t.Errorf("Did not find required file dev1.only from the dev1.")
			}
		}
	}
}
