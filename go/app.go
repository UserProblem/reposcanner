package swagger

import "log"

type App struct {
	Router    HttpRouter
	RepoStore RepoStore
	ScanStore ScanStore
}

func (a *App) Initialize() {
	a.Router = a.NewRouter()
	a.ClearStores()
}

func (a *App) ClearStores() {
	var err error
	
	var rs *RepoStore
	if rs, err = NewRepoStore(); err != nil {
		log.Fatal("Cannot initialize repository data store.\n")
	}
	a.RepoStore = *rs

	var ss *ScanStore
	if ss, err = NewScanStore(); err != nil {
		log.Fatal("Cannot initialize scan data store.\n")
	}
	a.ScanStore = *ss
}
