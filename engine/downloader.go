package engine

import (
	"fmt"
	"log"
	"os"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

func CloneRepository(url, branch string, checkoutDir string) error {
	_, err := git.PlainClone(checkoutDir, false, &git.CloneOptions{
		URL:           url,
		ReferenceName: plumbing.NewBranchReferenceName(branch),
	})

	if err != nil {
		return fmt.Errorf("failed to clone url %v: %v", url, err.Error())
	}
	return nil
}

func DeleteTmpDirectory(path string) {
	if err := os.RemoveAll(path); err != nil {
		log.Printf("Failed to delete tmp directory %v: %v", path, err.Error())
	}
}
