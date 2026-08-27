package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"
	"text/template"
)

var (
	templateFile  = flag.String("t", "", "path to the template file")
	templateText  = flag.String("i", "", "template text")
	outputFile    = flag.String("o", "", "path to the output file")
	keyValuePairs []string
)

func main() {
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "fills a text template with key=value pairs and writes the result to an output file")
		fmt.Fprintln(os.Stderr, "exactly one template source must be provided (-i or -t)")
		fmt.Fprintln(os.Stderr, "template syntax: https://pkg.go.dev/text/template")
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, "usage: tmplfill [options] key=value ...")
		fmt.Fprintln(os.Stderr, "options:")
		flag.PrintDefaults()
	}
	flag.Parse()
	keyValuePairs = flag.Args()

	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func run() error {
	if (*templateFile == "") == (*templateText == "") {
		return errors.New("provide exactly one template source (-t or -i)")
	}

	if *outputFile == "" {
		return errors.New("output file (-o) is required")
	}

	var tmplContent string
	if *templateFile != "" {
		data, err := os.ReadFile(*templateFile)
		if err != nil {
			return fmt.Errorf("reading template file: %w", err)
		}
		tmplContent = string(data)
	} else {
		tmplContent = *templateText
	}

	data := make(map[string]string)
	for _, pair := range keyValuePairs {
		name, value, found := strings.Cut(pair, "=")
		if !found {
			return fmt.Errorf("invalid name/value pair: %q", pair)
		}
		if name == "" {
			return fmt.Errorf("invalid name in pair: %q", pair)
		}
		data[name] = value
	}

	tmpl, err := template.New("tmplfill").Option("missingkey=error").Parse(tmplContent)
	if err != nil {
		return fmt.Errorf("parsing template: %w", err)
	}

	out, err := os.Create(*outputFile)
	if err != nil {
		return fmt.Errorf("creating output file: %w", err)
	}
	defer out.Close()

	if err := tmpl.Execute(out, data); err != nil {
		return fmt.Errorf("executing template: %w", err)
	}

	return nil
}
