package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/jbowens/postcards2dianne"
)

const (
	envLobAPIKey        = "LOB_API_KEY"
	defaultPostcardSize = "6x11"
	defaultConfigFile   = ".postcards2dianne.json"
)

type linesFlag []string

func (l *linesFlag) String() string {
	return strings.Join(*l, "\n")
}

func (l *linesFlag) Set(value string) error {
	*l = append(*l, value)
	return nil
}

func main() {
	var dry bool
	var lines, fonts linesFlag
	var toAlias, message string
	flag.StringVar(&toAlias, "to", "dianne", "the address alias of the recipient")
	flag.StringVar(&message, "message", "", "the message to appear on the back of the postcard (required)")
	flag.Var(&lines, "lines", "the lines to appear on the front of the postcard. provide multiple times for multiple lines (required)")
	flag.Var(&fonts, "font", "the preferred font to use")
	flag.BoolVar(&dry, "dryrun", false, "set to render the postcard but not send it")
	flag.Parse()
	message = strings.TrimSpace(message)
	if message == "" || len(lines) == 0 || toAlias == "" {
		flag.Usage()
		return
	}

	lobAPIKey := os.Getenv(envLobAPIKey)
	if lobAPIKey == "" {
		fatalf("%s environment variable unset", envLobAPIKey)
	}

	usr, err := user.Current()
	if err != nil {
		fatalf("error: %s", err)
	}
	configFile := filepath.Join(usr.HomeDir, defaultConfigFile)

	lob, err := postcards2dianne.NewLobClient(lobAPIKey, configFile)
	if err != nil {
		fatalf("error: %s", err)
	}

	// Build the postcard.
	p := postcards2dianne.New(defaultPostcardSize, lines, message)
	if len(fonts) > 0 {
		ok := p.SetFontPreferences(fonts...)
		if !ok {
			fatalf("Unable to find font(s): %s", strings.Join(fonts, ", "))
		}
	}

	// If it's just a dry run, render the image to a temporary file.
	if dry {
		const postcardPreviewFile = "/tmp/postcard-preview.png"
		pngBytes, err := p.Render()
		if err != nil {
			fatalf("error: %s", err)
		}
		err = ioutil.WriteFile(postcardPreviewFile, pngBytes, 0644)
		if err != nil {
			fatalf("error: %s", err)
		}
		fmt.Printf("Wrote postcard preview to %s\n", postcardPreviewFile)
		return
	}

	// Render it and send it to the Lob API for mailing.
	confirmation, err := lob.Send(p, toAlias)
	if err != nil {
		fatalf("error: %s", err)
	}

	// Print the confirmation details.
	fmt.Println(confirmation.URL)
	fmt.Println(confirmation.ExpectedDelivery)
}

func fatalf(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
