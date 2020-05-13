package quikirenderer

// Modeled from the html renderer at
// https://github.com/yuin/goldmark/blob/master/renderer/html/html.go
//
// Copyright (c) 2020 Mitchell Cooper
// Copyright (c) 2019 Yusuke Inuzuka
//
// See LICENSE

import "github.com/yuin/goldmark/renderer"

// A Config struct has configurations for the quiki markup renderer.
type Config struct {
	Writer          Writer
	HardWraps       bool
	Unsafe          bool
	PartialPage     bool
	TableOfContents bool
	PageTitle       string
	AbsolutePrefix  string
}

// NewConfig returns a new Config with defaults.
func NewConfig() Config {
	return Config{
		Writer:          DefaultWriter,
		HardWraps:       false,
		Unsafe:          false,
		PartialPage:     false,
		TableOfContents: false,
		PageTitle:       "",
		AbsolutePrefix:  "",
	}
}

// SetOption implements renderer.NodeRenderer.SetOption.
func (c *Config) SetOption(name renderer.OptionName, value interface{}) {
	switch name {
	case optHardWraps:
		c.HardWraps = value.(bool)
	case optUnsafe:
		c.Unsafe = value.(bool)
	case optPartialPage:
		c.PartialPage = value.(bool)
	case optTableOfContents:
		c.TableOfContents = value.(bool)
	case optPageTitle:
		c.PageTitle = value.(string)
	case optAbsolutePrefix:
		c.AbsolutePrefix = value.(string)
	case optTextWriter:
		c.Writer = value.(Writer)
	}
}

// An Option interface sets options for the quiki markup renderer.
type Option interface {
	SetQuikiOption(*Config)
}

// TextWriter is an option name used in WithWriter.
const optTextWriter renderer.OptionName = "Writer"

type withWriter struct {
	value Writer
}

func (o *withWriter) SetConfig(c *renderer.Config) {
	c.Options[optTextWriter] = o.value
}

func (o *withWriter) SetQuikiOption(c *Config) {
	c.Writer = o.value
}

// WithWriter is a functional option that allow you to set the given writer to
// the renderer.
func WithWriter(writer Writer) interface {
	renderer.Option
	Option
} {
	return &withWriter{writer}
}

// HardWraps is an option name used in WithHardWraps.
const optHardWraps renderer.OptionName = "HardWraps"

type withHardWraps struct {
}

func (o *withHardWraps) SetConfig(c *renderer.Config) {
	c.Options[optHardWraps] = true
}

func (o *withHardWraps) SetQuikiOption(c *Config) {
	c.HardWraps = true
}

// WithHardWraps is a functional option that indicates whether softline breaks
// should be rendered as '<br>'.
func WithHardWraps() interface {
	renderer.Option
	Option
} {
	return &withHardWraps{}
}

// Unsafe is an option name used in WithUnsafe.
const optUnsafe renderer.OptionName = "Unsafe"

type withUnsafe struct {
}

func (o *withUnsafe) SetConfig(c *renderer.Config) {
	c.Options[optUnsafe] = true
}

func (o *withUnsafe) SetQuikiOption(c *Config) {
	c.Unsafe = true
}

// WithUnsafe is a functional option that renders dangerous contents
// (raw htmls and potentially dangerous links) as it is.
func WithUnsafe() interface {
	renderer.Option
	Option
} {
	return &withUnsafe{}
}

// PartialPage is an option name used in WithPartialPage.
const optPartialPage renderer.OptionName = "PartialPage"

type withPartialPage struct {
}

func (o *withPartialPage) SetConfig(c *renderer.Config) {
	c.Options[optPartialPage] = true
}

func (o *withPartialPage) SetQuikiOption(c *Config) {
	c.PartialPage = true
}

// WithPartialPage is a functional option that renders the Markdown
// as a portion of a page, without including quiki `@page` variables.
func WithPartialPage() interface {
	renderer.Option
	Option
} {
	return &withPartialPage{}
}

// TableOfContents is an option name used in WithTableOfContents.
const optTableOfContents renderer.OptionName = "TableOfContents"

type withTableOfContents struct {
}

func (o *withTableOfContents) SetConfig(c *renderer.Config) {
	c.Options[optTableOfContents] = true
}

func (o *withTableOfContents) SetQuikiOption(c *Config) {
	c.TableOfContents = true
}

// WithTableOfContents is a functional option that renders a table
// of contents on the page.
func WithTableOfContents() interface {
	renderer.Option
	Option
} {
	return &withTableOfContents{}
}

// PageTitle is an option name used in WithPageTitle.
const optPageTitle renderer.OptionName = "PageTitle"

type withAbsolutePrefix struct {
	pfx string
}

func (o *withAbsolutePrefix) SetConfig(c *renderer.Config) {
	c.Options[optPageTitle] = o.pfx
}

func (o *withAbsolutePrefix) SetQuikiOption(c *Config) {
	c.AbsolutePrefix = o.pfx
}

// WithAbsolutePrefix is a functional option that specifies the
// absolute prefix.
func WithAbsolutePrefix(pfx string) interface {
	renderer.Option
	Option
} {
	return &withAbsolutePrefix{pfx: pfx}
}

// AbsolutePrefix is an option name used in WithAbsolutePrefix.
const optAbsolutePrefix renderer.OptionName = "AbsolutePrefix"

type withPageTitle struct {
	title string
}

func (o *withPageTitle) SetConfig(c *renderer.Config) {
	c.Options[optPageTitle] = o.title
}

func (o *withPageTitle) SetQuikiOption(c *Config) {
	c.PageTitle = o.title
}

// WithPageTitle is a functional option that renders the `@page.title`
// variable to the provided text.
func WithPageTitle(title string) interface {
	renderer.Option
	Option
} {
	return &withPageTitle{title: title}
}
