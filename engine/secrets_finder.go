package engine

import (
	"bufio"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/UserProblem/reposcanner/models"
)

type SecretFinder struct {
	findings []*models.FindingsInfo
	reportCh chan *models.FindingsInfo
	ruleDefs map[string]RuleDefinition
}

type RuleDefinition struct {
	Id          string
	Description string
	Severity    string
}

func (a *SecretFinder) Initialize() {
	a.findings = make([]*models.FindingsInfo, 0)
	a.reportCh = make(chan *models.FindingsInfo)

	a.ruleDefs = make(map[string]RuleDefinition)
	a.ruleDefs["private_key"] = RuleDefinition{
		Id:          "G001",
		Description: "Hard-coded secret - private key",
		Severity:    "HIGH",
	}
	a.ruleDefs["public_key"] = RuleDefinition{
		Id:          "G002",
		Description: "Hard-coded secret - public key",
		Severity:    "HIGH",
	}
}

func (a *SecretFinder) FindSecrets(basepath string) []*models.FindingsInfo {
	walkDone := make(chan bool)

	go func() {
		if err := filepath.WalkDir(basepath, a.WalkDirHandler); err != nil {
			log.Printf("Error traversing repository tree: %s", err.Error())
		}
		walkDone <- true
	}()

	for running := true; running; {
		select {
		case fi := <-a.reportCh:
			fi.Location.Path = strings.TrimPrefix(fi.Location.Path, basepath)
			a.findings = append(a.findings, fi)
		case <-walkDone:
			running = false
		}
	}

	return a.findings
}

func (a *SecretFinder) WalkDirHandler(path string, d fs.DirEntry, err error) error {
	if err != nil {
		if d == nil {
			return err
		} else {
			log.Printf("Skipping directory %v due to error.", path)
			return filepath.SkipDir
		}
	}

	if d.IsDir() {
		// skip .git subdirectory
		if d.Name() == ".git" {
			return filepath.SkipDir
		}

		// don't need to do anything for other subdirectories
		return nil
	}

	if errr := a.ScanFile(path); errr != nil {
		log.Printf("Error when scanning %v: %v", path, errr.Error())
	}

	return nil
}

func (a *SecretFinder) ScanFile(path string) error {
	if file, err := os.Open(path); err != nil {
		return err
	} else {
		defer file.Close()
		scanner := bufio.NewScanner(file)
		scanner.Split(bufio.ScanLines)

		// Regular expression attempts to match the keywords 'private_key' or 'public_key', followed
		// by an optional ':', '=', or ':=', then followed by a token not starting with ',' or ';'.
		//
		// e.g. private_key := "lkasjdlkajsdlkajsdp"
		re := regexp.MustCompile(`((?:private_key)|(?:public_key))['"]?\s*(?:(?::=)|(?:[:=])|(?:\s))\s*(?:([^;,:={}\s]+)|$)`)
		prevLine := ""
		prevLineCnt := 0
		for lineCnt := 1; scanner.Scan(); lineCnt++ {
			line := scanner.Text()
			groups := re.FindAllStringSubmatch(prevLine+line, -1)
			if len(groups) > 0 {
				for _, group := range groups {
					if group[2] == "" {
						// prefix found, but token not found
						// buffer it and see if the token appears later
						if prevLine == "" {
							prevLineCnt = lineCnt
						}
						prevLine = prevLine + line
					} else {
						// full match
						matchType := group[1]

						var fl models.FileLocation
						if prevLine != "" {
							fl.Begin = &models.LineLocation{Line: int32(prevLineCnt)}
							fl.End = &models.LineLocation{Line: int32(lineCnt)}
							prevLine = ""
						} else {
							fl.Begin = &models.LineLocation{Line: int32(lineCnt)}
						}

						a.reportCh <- &models.FindingsInfo{
							Type_:  "sast",
							RuleId: a.ruleDefs[matchType].Id,
							Location: &models.FindingsLocation{
								Path:      path,
								Positions: &fl,
							},
							Metadata: &models.FindingsMetadata{
								Description: a.ruleDefs[matchType].Description,
								Severity:    a.ruleDefs[matchType].Severity,
							},
						}
					}
				}
			} else {
				prevLine = ""
			}
		}
	}

	return nil
}
