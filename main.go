package main

import (
	"bytes"
	"encoding/xml"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"strconv"
)

const (
	Header = `<?xml version="1.0" encoding="UTF-8"?>` + "\n"
)

type XmlReport struct {
	XMLName xml.Name
	XML     string `xml:",innerxml"`
	XMLNS   string `xml:"xmlns,attr"`
	//	Time      float64      `xml:"time,attr"`
	//	Tests     uint64       `xml:"tests,attr"`
	//	Failures  uint64       `xml:"failures,attr"`
	XSI       string       `xml:"xmlns:xsi,attr"`
	SCHEMA    string       `xml:"xsi:schemalocation,attr"`
	XMLBuffer bytes.Buffer `xml:"-"`
}

var usage = `Usage: xml-merger [options] [files]
Options:
  -o  Merged report filename`

func main() {
	flag.Usage = func() {
		fmt.Println(usage)
	}
	outputFileName := flag.String("o", "", "merged report filename")
	flag.Parse()
	files := flag.Args()
	printReport := *outputFileName == ""
	if len(files) == 0 {
		flag.Usage()
		return
	}

	var mergedReport XmlReport
	startedReading := false
	fileCount := 0

	for _, fileName := range files {
		var report XmlReport
		in, err := ioutil.ReadFile(fileName)

		if err != nil {
			panic(err)
		}

		err = xml.Unmarshal(in, &report)

		if err != nil {
			panic(err)
		}

		if report.XMLName.Local == "testsuite" {
			panic(errors.New("Reports with a root <testsuite> are not supported"))
		}

		if startedReading && report.XMLNS != mergedReport.XMLNS {
			panic(errors.New("All reports must have the same <testsuites> name"))
		}

		startedReading = true
		fileCount++
		mergedReport.XMLName = xml.Name{Local: "Audit"}
		mergedReport.XMLNS = "http://xmlns.oracle.com/oracleas/schema/dbserver_audittrail-11_2.xsd"
		mergedReport.XSI = "http://www.w3.org/2001/XMLSchema-instance"
		mergedReport.SCHEMA = "http://xmlns.oracle.com/oracleas/schema/dbserver_audittrail-11_2.xsd"
		mergedReport.XMLBuffer.WriteString(report.XML)
	}

	mergedReport.XML = mergedReport.XMLBuffer.String()
	//mergedOutput = "<?xml version="1.0" encoding="UTF-8"?>"
	mergedOutput, _ := xml.MarshalIndent(&mergedReport, "", "  ")
	mergedOutput = []byte(xml.Header + string(mergedOutput))

	if printReport {

		fmt.Println(string(mergedOutput))
	} else {

		err := ioutil.WriteFile(*outputFileName, mergedOutput, 0644)

		if err != nil {
			panic(err)
		}

		fmt.Println("Merged " + strconv.Itoa(fileCount) + " reports to " + *outputFileName)
	}
}

