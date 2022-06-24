package engine_test

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/UserProblem/reposcanner/engine"
)

func TestFindSecretsHandlesInvalidDirectory(t *testing.T) {
	var sf engine.SecretFinder
	sf.Initialize()

	findings := sf.FindSecrets("not a real path")

	if len(findings) != 0 {
		t.Errorf("Unexpected findings from non-existent directory.")
	}
}

func TestFindSecretsHandlesDirectorTree(t *testing.T) {
	var sf engine.SecretFinder
	sf.Initialize()

	var checkoutDir string
	if tmpDir, err := os.MkdirTemp("", "reposcanner"); err != nil {
		t.Fatalf("Could not create temporary directory.")
	} else {
		checkoutDir = tmpDir
		defer engine.DeleteTmpDirectory(checkoutDir)
	}

	if err := engine.CloneRepository("https://github.com/UserProblem/testdata.git", "master", checkoutDir); err != nil {
		t.Fatalf("Clone repository failed: %v", err.Error())
	}

	findings := sf.FindSecrets(checkoutDir)

	if len(findings) == 0 {
		t.Errorf("No findings returned.")
	}

	for _, fi := range findings {
		b, _ := json.Marshal(fi)
		t.Logf("%v\n", string(b))
	}
}
