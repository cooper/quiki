package adminifier

// working with multiple branches is disabled until go-git supports it officially
// see: https://github.com/cooper/quiki/issues/156

// func handleSwitchBranch(wr *wikiRequest) {
// 	branchName := strings.TrimPrefix(wr.r.URL.Path, wr.wikiRoot+"func/switch-branch/")
// 	if branchName == "" {
// 		wr.err = errors.New("no branch selected")
// 		return
// 	}

// 	// bad branch name
// 	if !wiki.ValidBranchName(branchName) {
// 		wr.err = errors.New("invalid branch name: " + branchName)
// 		return
// 	}

// 	// fetch the branch
// 	_, wr.err = wr.wi.Branch(branchName)
// 	if wr.err != nil {
// 		return
// 	}

// 	// set branch
// 	sessMgr.Put(wr.r.Context(), "branch", branchName)

// 	// TODO: when this request is submitted by JS, the UI can just reload
// 	// the current frame so the user stays on the same page, just in new branch

// 	// redirect back to dashboard
// 	http.Redirect(wr.w, wr.r, wr.wikiRoot+"dashboard", http.StatusTemporaryRedirect)
// }

// func handleCreateBranch(wr *wikiRequest) {

// 	// TODO: need a different version of parsePost that returns JSON errors
// 	if !parsePost(wr.w, wr.r, "branch") {
// 		return
// 	}

// 	// bad branch name
// 	branchName := wr.r.Form.Get("branch")
// 	if !wiki.ValidBranchName(branchName) {
// 		wr.err = errors.New("invalid branch name: " + branchName)
// 		return
// 	}

// 	// create or switch branches
// 	_, err := wr.wi.NewBranch(branchName)
// 	if err != nil {
// 		wr.err = err
// 		return
// 	}
// 	sessMgr.Put(wr.r.Context(), "branch", branchName)

// 	// redirect back to dashboard
// 	http.Redirect(wr.w, wr.r, wr.wikiRoot+"dashboard", http.StatusTemporaryRedirect)
// }

// // possibly switch wiki branches
// func switchUserWiki(wr *wikiRequest, wi *webserver.WikiInfo) {
// 	userWiki := wi
// 	branchName := sessMgr.GetString(wr.r.Context(), "branch")
// 	if branchName != "" {
// 		branchWiki, err := wi.Branch(branchName)
// 		if err != nil {
// 			wr.err = err
// 			return
// 		}
// 		userWiki = wi.Copy(branchWiki)
// 	}
// 	wr.wi = userWiki
// }

// func handleSwitchBranchFrame(wr *wikiRequest) {
// 	branches, err := wr.wi.BranchNames()
// 	if err != nil {
// 		wr.err = err
// 		return
// 	}
// 	wr.dot = struct {
// 		Branches []string
// 		wikiTemplate
// 	}{
// 		Branches:     branches,
// 		wikiTemplate: getGenericTemplate(wr),
// 	}
// }
