package main

import (
	"os"
	"io"
	"fmt"
	"strings"
	"archive/zip"
	"io/ioutil"
	"bytes"
	"golang.org/x/net/html"
	"golang.org/x/net/html/charset"
	"golang.org/x/text/transform"
	"github.com/aarzilli/sandblast"
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
	name string
	input *zip.File
	target *zip.File
}

type Dataset struct {
	datazip *zip.ReadCloser
	tests []test
}

func (d *Dataset) Close() error {
	return d.datazip.Close()
}

func openDataset(datapath string) *Dataset {
	dataset, err := zip.OpenReader(datapath)
	must(err)
	
	ins := map[string]*zip.File{}
	outs := map[string]*zip.File{}

	for _, file := range dataset.File {
		if file.Name == "index.txt" {
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
		tests = append(tests, test{ name: k, input: in, target: out })
	}
	
	return &Dataset{ datazip: dataset, tests: tests }
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
			return
		}
	}
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

func qaruntest(test test, writein bool) bool {
	in, err := test.input.Open()
	must(err)
	defer in.Close()
	
	body, err := ioutil.ReadAll(in)
	must(err)
	
	e, _, _ := charset.DetermineEncoding(body, "UTF-8")
	r := transform.NewReader(bytes.NewReader(body), e.NewDecoder())
	node, err := html.Parse(r)
	must(err)
	
	_, output, _, _, _, err := sandblast.ExtractEx(node)
	must(err)
	
	output = strings.TrimSpace(output)
	
	tgt, err := test.target.Open()
	must(err)
	defer tgt.Close()
	
	tgtbody, err := ioutil.ReadAll(tgt)
	must(err)
	target := strings.TrimSpace(string(tgtbody))
	
	if target != output {
		fmt.Printf("%s output and target differ\n", test.name)
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
	//TODO:
	// - run on all tests, build new version of the test-set
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