package main

import (
	"archive/zip"
	"bytes"
	"fmt"
	"github.com/aarzilli/sandblast"
	"golang.org/x/net/html"
	"golang.org/x/net/html/charset"
	"golang.org/x/text/transform"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"unicode"
)

func usage() {
	fmt.Fprintf(os.Stderr, "./qa run <dataset.zip>\n")
	fmt.Fprintf(os.Stderr, "./qa rebuild <dataset.zip> <out.zip>\n")
	fmt.Fprintf(os.Stderr, "./qa one <dataset.zip> <testname>\n")
	os.Exit(1)
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}

type test struct {
	name   string
	input  *zip.File
	target *zip.File
}

type Dataset struct {
	datazip *zip.ReadCloser
	index   *zip.File
	tests   []test
}

func (d *Dataset) Close() error {
	return d.datazip.Close()
}

func collapseWhitespace(in []rune) []rune {
	var b []rune = make([]rune, len(in))
	d := 0
	spaceSeen := true
	for s := range in {
		if spaceSeen {
			if !unicode.IsSpace(in[s]) {
				spaceSeen = false
				b[d] = in[s]
				d++
			}
		} else {
			if unicode.IsSpace(in[s]) {
				b[d] = ' '
				d++
				spaceSeen = true
			} else {
				b[d] = in[s]
				d++
			}
		}
	}
	return b[:d]
}

func openDataset(datapath string) *Dataset {
	dataset, err := zip.OpenReader(datapath)
	must(err)

	ins := map[string]*zip.File{}
	outs := map[string]*zip.File{}
	var index *zip.File

	for _, file := range dataset.File {
		if file.Name == "index.txt" {
			index = file
			continue
		}
		v := strings.Split(file.Name, ".")
		if len(v) != 2 {
			panic(fmt.Errorf("wrong name in dataset: %s\n", file.Name))
		}

		switch v[1] {
		case "html":
			ins[v[0]] = file
		case "target":
			outs[v[0]] = file
		default:
			panic(fmt.Errorf("wrong name in dataset: %s\n", file.Name))
		}
	}

	tests := make([]test, 0, len(ins))
	for k := range ins {
		in, inok := ins[k]
		out, outok := outs[k]
		if !inok || !outok {
			panic(fmt.Errorf("problem with dataset: %s", k))
		}
		tests = append(tests, test{name: k, input: in, target: out})
	}

	return &Dataset{index: index, datazip: dataset, tests: tests}
}

func qarun(datapath string) {
	dataset := openDataset(datapath)
	defer dataset.Close()

	os.Mkdir("work", 0770)

	count := 0
	for _, test := range dataset.tests {
		fmt.Printf("Processing %s\n", test.name)
		if !qaruntest(test, false) {
			count++
		}
		if count > 10 {
			fmt.Printf("Too many differences\n")
			return
		}
	}
	fmt.Printf("All ok\n")
}

func qaone(datapath string, name string) {
	dataset := openDataset(datapath)
	defer dataset.Close()

	os.Mkdir("work", 0770)

	for _, test := range dataset.tests {
		if test.name == name {
			qaruntest(test, true)
			return
		}
	}
}

func extractTest(test test, writeextract bool) ([]byte, string) {
	in, err := test.input.Open()
	must(err)
	defer in.Close()

	body, err := ioutil.ReadAll(in)
	must(err)

	e, _, _ := charset.DetermineEncoding(body, "UTF-8")
	r := transform.NewReader(bytes.NewReader(body), e.NewDecoder())
	node, err := html.Parse(r)
	must(err)

	_, output, simplified, flattened, cleaned, err := sandblast.ExtractEx(node, 0)
	must(err)

	if writeextract {
		fmt.Printf("SIMPLIFIED:\n%s\n", simplified.DebugString())
		fmt.Printf("FLATTENED:\n%s\n", flattened.DebugString())
		fmt.Printf("CLEANED:\n%s\n", cleaned.DebugString())
	}

	return body, output
}

func qaruntest(test test, writein bool) bool {
	body, output := extractTest(test, writein)

	tgt, err := test.target.Open()
	must(err)
	defer tgt.Close()

	tgtbody, err := ioutil.ReadAll(tgt)
	must(err)
	target := strings.TrimSpace(string(tgtbody))

	a := strings.TrimSpace(string(collapseWhitespace([]rune(target))))
	b := strings.TrimSpace(string(collapseWhitespace([]rune(output))))
	if a != b {
		fmt.Printf("%s output and target differ\n", test.name)
		//fmt.Printf("target: <%s>\noutput: <%s>\n", a, b)
		tgtout, err := os.Create(fmt.Sprintf("work/%s.target", test.name))
		must(err)
		io.WriteString(tgtout, target)
		io.WriteString(tgtout, "\n")
		tgtout.Close()
		outout, err := os.Create(fmt.Sprintf("work/%s.out", test.name))
		must(err)
		io.WriteString(outout, output)
		io.WriteString(outout, "\n")
		outout.Close()

		if writein {
			inout, err := os.Create(fmt.Sprintf("work/%s.html", test.name))
			must(err)
			inout.Write(body)
			inout.Close()
		}

		return false
	}
	return true
}

func qarebuild(datapath, outpath string) {
	dataset := openDataset(datapath)
	defer dataset.Close()

	outw, err := os.Create(outpath)
	must(err)
	defer outw.Close()

	outzip := zip.NewWriter(outw)
	defer outzip.Close()

	copyFile(outzip, "index.txt", dataset.index)

	for _, test := range dataset.tests {
		fmt.Printf("processing %s\n", test.name)
		copyFile(outzip, fmt.Sprintf("%s.html", test.name), test.input)
		_, output := extractTest(test, false)
		w, err := outzip.Create(fmt.Sprintf("%s.target", test.name))
		must(err)
		_, err = io.WriteString(w, output)
		must(err)
	}
}

func copyFile(outzip *zip.Writer, name string, in *zip.File) {
	w, err := outzip.Create(name)
	must(err)
	r, err := in.Open()
	must(err)
	defer r.Close()
	_, err = io.Copy(w, r)
	must(err)
}

func main() {
	if len(os.Args) < 1 {
		usage()
	}
	switch os.Args[1] {
	case "run":
		if len(os.Args) < 3 {
			usage()
		}
		qarun(os.Args[2])
	case "one":
		qaone(os.Args[2], os.Args[3])
	case "rebuild":
		if len(os.Args) < 4 {
			usage()
		}
		qarebuild(os.Args[2], os.Args[3])
	case "help":
		usage()
	default:
		usage()
	}
}
