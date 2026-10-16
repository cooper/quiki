package wikifier

import (
	"os"
	"path/filepath"
	"slices"
)

type modelBlock struct {
	modelName   string
	model       *Page
	includeTags bool // @model.tags - wrap in HTML tags
	*Map
}

func newModelBlock(name string, b *parserBlock) block {
	b.typ = "model"
	return &modelBlock{Map: newMapBlock("", b).(*Map)}
}

func (mb *modelBlock) parse(page *Page) {
	mb.Map.parse(page)

	// remember that the page uses this model.
	// a model is written in quiki markup (name.model) or markdown (name.md)
	name := mb.blockName()
	file := ModelName(name)
	path := pageAbs(filepath.Join(page.Opt.Dir.Model, file))
	if _, err := os.Stat(path); err != nil {
		if mdFile := PageNameExt(name, ".md"); fileExists(filepath.Join(page.Opt.Dir.Model, mdFile)) {
			file = mdFile
			path = pageAbs(filepath.Join(page.Opt.Dir.Model, file))
		}
	}

	// create page
	model := NewPage(path)
	model.name = name
	model.model = true

	// copy wiki opt from this page
	model.Opt = page.Opt

	// assign the underlying Map of the model{} block to @m
	model.Set("m", mb.Map)

	// check if it exists before anything else
	if !model.Exists() {
		mb.warn(mb.openPos, "Model $"+name+"{} does not exist")
		return
	}

	// parse the page
	if err := model.Parse(); err != nil {
		mb.warn(mb.openPos, "Model $"+name+"{} error: "+err.Error())
		return
	}

	// determine whether to include model tags
	mb.includeTags, _ = model.GetBool("model.tags")

	mb.modelName = name
	mb.model = model

	// remember the page uses this
	page.Models[file] = model.modelInfo()
}

func (mb *modelBlock) html(page *Page, mbEl element) {
	mb.Map.html(page, mbEl)

	// my $model      = $block->{model} or return;
	// my $main_block = $model->{wikifier}{main_block} or return;

	// $block->html_base($page); # call hash html.

	// # generate the DOM
	// my $el = $main_block->html($model) or return;

	// if there's nothing here, an error occurred in parse()
	model := mb.model
	mb.model = nil
	if model == nil {
		return
	}

	// generate the DOM
	mainBlock := model.mainBlock()
	mainEl := mainBlock.el()
	mainBlock.html(model, mainEl)

	// # add the main page element to our element.
	// $el->remove_class('main');
	// $el->add_class('model');
	// $el->add_class("model-$$model{model_name}");
	// $el->add_class($model_el->{id});

	// add the main block element to the model{} block element itself,
	// which never has tags
	mbEl.setMeta("noTags", true)
	mbEl.addChild(mainEl)

	// these properties are set just incase @model.tags is true
	// FIXME: need to remove q-main
	mainEl.addClass("model", "model-"+mb.modelName)

	// styles are collected on the page being generated, which inside the
	// model was the model itself. hand them up to the page that uses the model
	// so they make it into its CSS. a style{} at the top level of the model
	// targets the model's main element, so that element has to be rendered
	// for the style to have any effect, regardless of @model.tags.
	for _, style := range model.styles {
		if style.parent == mainEl {
			mb.includeTags = true
		}
	}
	page.styles = append(page.styles, model.styles...)

	// static styles (e.g. the syntax highlighting stylesheet) are shared,
	// so don't repeat them for every model instance
	for _, static := range model.staticStyles {
		if !slices.Contains(page.staticStyles, static) {
			page.staticStyles = append(page.staticStyles, static)
		}
	}
	if model.codeStyles {
		page.codeStyles = true
	}

	// disable tags on the main block element also unless @model.tags
	if !mb.includeTags {
		mainEl.setMeta("noTags", true)
	}
}
