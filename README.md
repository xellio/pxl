## PXL
pxl is a command-line tool for converting files to pxl images. In some cases this can save some bytes when transfering the data or just retuns a spacy image.

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
	echo '127.0.0.1 - - [22/Aug/2017:13:08:24 +0200] "GET / HTTP/1.1" 403 189 "-" "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Ubuntu Chromium/60.0.3112.78 Chrome/60.0.3112.78 Safari/537.36"' >> ./logfile.log
done
```
Running
```
./pxl -e example/xellio.jpg
```
will create a ~100KB file, containing all the information from logfile.log

## Installation
go get github.com/xellio/pxl
