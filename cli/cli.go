package cli

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/cooper/quiki/wikifier"
)

// Config holds the cli configuration
type Config struct {
	Interactive bool
	Wizard      bool
	WikiPath    string
	ForceGen    bool
	JSONOutput  bool
	Reload      bool
	QuikiDir    string
	// server options - only used in full mode
	Bind   string
	Port   string
	Host   string
	Config string
}

// Parser interface for different cli modes
type Parser interface {
	SetupFlags(c *Config)
	HandleCommand(c *Config, args []string) error
}

// ParseFlags sets up flags and returns parsed config and remaining args
func ParseFlags(parser Parser) (*Config, []string) {
	c := &Config{}
	parser.SetupFlags(c)
	flag.Parse()
	InitQuikiDir(c)
	return c, flag.Args()
}

// InitQuikiDir sets up the quiki directory path
func InitQuikiDir(c *Config) {
	if c.QuikiDir == "" && c.Config == "" {
		c.QuikiDir = filepath.Join(os.Getenv("HOME"), "quiki")
	} else if c.QuikiDir == "" && c.Config != "" {
		log.Printf("warning: -config flag is deprecated, use -dir instead")
		c.QuikiDir = filepath.Dir(c.Config)
	}

	c.Config = filepath.Join(c.QuikiDir, "quiki.conf")
}

// RunInteractiveMode reads from stdin and processes a page
func RunInteractiveMode(jsonOutput bool) {
	input, err := io.ReadAll(os.Stdin)
	if err != nil {
		log.Fatal(err)
	}
	page := wikifier.NewPageSource(string(input))
	RunPageAndExit(page, jsonOutput)
}

// RunPageAndExit processes a wikifier page and exits
func RunPageAndExit(page *wikifier.Page, jsonOutput bool) {
	err := page.Parse()
	if err != nil {
		log.Fatal(err)
	}
	if jsonOutput {
		json.NewEncoder(os.Stdout).Encode(page)
		os.Exit(0)
	}
	fmt.Println(page.HTMLAndCSS())
	os.Exit(0)
}
