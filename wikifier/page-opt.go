package wikifier

type PageOpts struct {
	Root struct {
		Image string
	}
	Image struct {
		SizeMethod string
		Calc       func(file string, width, height int, page *Page, override bool) (w, h, bigW, bigH int, fullSize bool)
		Sizer      func(file string, width, height int, page *Page) (path string)
	}
}
