// typoconfig writes the specialised Typst series template that a ProdCal
// build would use for a set of transmittal typography choices, without
// touching the database or the metered API. It exists for
// typesetting/scripts/build-sampler.sh (the customer typography sampler).
//
//	go run ./cmd/typoconfig -out scratch/sampler/classic -pairing classic
//
// prints the template path; the config block goes to stderr with -v.
package main

import (
	"flag"
	"fmt"
	"os"

	"srv.exe.dev/srv"
)

func main() {
	out := flag.String("out", "", "directory to write templates/ into (required)")
	trim := flag.String("trim", "us-trade", "trim name from the spec registry (us-trade = 6 × 9)")
	pairing := flag.String("pairing", "classic", "studio | classic | house | literary")
	size := flag.String("size", "standard", "compact | standard | generous")
	sb := flag.String("break", "space", "space | breve | ornament")
	paras := flag.String("paragraphs", "indented", "indented | block")
	verbose := flag.Bool("v", false, "print the config block to stderr")
	flag.Parse()
	if *out == "" {
		flag.Usage()
		os.Exit(2)
	}
	path, config, err := srv.SamplerTemplate(*out, *trim, *pairing, *size, *sb, *paras)
	if err != nil {
		fmt.Fprintln(os.Stderr, "typoconfig:", err)
		os.Exit(1)
	}
	if *verbose {
		fmt.Fprintln(os.Stderr, config)
	}
	fmt.Println(path)
}
