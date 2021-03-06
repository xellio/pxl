## PXL
pxl is a command-line tool for converting files to pxl images. In some cases this can save some bytes when transfering the data or just retuns a spacy image.

[![go report card](https://goreportcard.com/badge/github.com/xellio/pxl "go report card")](https://goreportcard.com/report/github.com/xellio/pxl)
[![Build Status](https://travis-ci.org/xellio/pxl.svg?branch=master)](https://travis-ci.org/xellio/pxl)
[![MIT license](http://img.shields.io/badge/license-MIT-brightgreen.svg)](http://opensource.org/licenses/MIT)
[![GoDoc](https://godoc.org/github.com/xellio/pxl?status.svg)](https://godoc.org/github.com/xellio/pxl)

## Usage
### encoding:
```
./pxl -e example/xellio.jpg
```
<img src="./example/xellio.jpg" width="163" alt="original file"> --> ![pxl image](./example/xellio.jpg.pxl?raw=true "pxl image")
### decoding:
```
./pxl -d xellio.jpg.pxl
```
![pxl image](./example/xellio.jpg.pxl?raw=true "pxl image") --> <img src="./example/xellio.jpg" width="163" alt="original file">

## More usefull example
Assume we have a ~20MB-logfile like this:
```
for i in {1..100000}
do
	echo '127.0.0.1 - - [20/Aug/2017:13:08:24 +0200] "GET / HTTP/1.1" 403 189 "-" "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Ubuntu Chromium/60.0.3112.78 Chrome/60.0.3112.78 Safari/537.36"' >> ./logfile.log
done
```
Running
```
./pxl -e logfile.log
```
will create a ~100KB file, containing all the information from logfile.log

## Better than any other compression?
No, probably not - But it looks good

## Warning
Encoding too lage files can freeze your computer (depending on your hardware)
