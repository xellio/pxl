package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"pxl"
	"time"
)

func main() {

	encode := flag.Bool("e", false, "encode flag")
	decode := flag.Bool("d", false, "decode flag")
	filePath := flag.String("f", "", "the file to convert")
	outputFileName := flag.String("o", "", "the filename of your output")

	flag.Parse()

	if !*encode && !*decode {
		fmt.Println("No action ('-d' for decode or '-e' for encode) specified")
		os.Exit(0)
	}

	if len(*filePath) <= 0 {
		fmt.Println("No input file specified ('-f=/path/to/file')")
		os.Exit(0)
	}

	if *encode {
		startEncodingTime := time.Now()
		dat, err := ioutil.ReadFile(*filePath)
		if err != nil {
			fmt.Println(err)
			os.Exit(0)
		}

		if len(*outputFileName) <= 0 {
			*outputFileName = "out.png"
		}

		_, err = pxl.Encode(dat, *outputFileName)
		if err != nil {
			fmt.Println(err)
			os.Exit(0)
		}
		/*
			//imgFile, err := ioutil.ReadFile("out.png")
			var b bytes.Buffer
			w := gzip.NewWriter(&b)
			w.Write(dat)
			w.Close()
			err = ioutil.WriteFile("out.gz", b.Bytes(), 0666)
		*/
		elapsedEncodingTime := time.Since(startEncodingTime)

		fmt.Printf("Time for encoding: %s\n", elapsedEncodingTime)
	}

	if *decode {

		startDecodingTime := time.Now()
		decodedOutput, err := pxl.Decode(*filePath)
		if err != nil {
			fmt.Println(err)
			os.Exit(0)
		}
		elapsedDecodingTime := time.Since(startDecodingTime)
		fmt.Printf("Time for decoding: %s\n", elapsedDecodingTime)
		fmt.Println(string(decodedOutput))
	}
}
