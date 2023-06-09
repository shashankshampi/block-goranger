package genericUtilities

import (
	"bufio"
	"bytes"
	"io"
	"os"
	"sync"
)

var lat_longs []string
var gurgoan_lat_longs []string
var bangaloreLatLongs [] string
var search_texts []string
var onceLatLong sync.Once
var onceSearchText sync.Once
var onceGurgoanLatLong sync.Once
var onceBangaloreLatLong sync.Once

func ParseFile(filename string) ([]string, string, error) {

	var lines []string
	// Opening a file
	file, err := os.Open(filename)
	defer file.Close()

	if err != nil {
		return lines, "Not able to open the file " + filename, err
	}

	// Start reading from the file with a reader.
	reader := bufio.NewReader(file)

	for {
		var buffer bytes.Buffer

		var l []byte
		var isPrefix bool
		for {
			l, isPrefix, err = reader.ReadLine()
			buffer.Write(l)

			if !isPrefix {
				break
			}

			if err != nil {
				break
			}
		}

		if err == io.EOF {
			break
		}

		// converting buffer to string
		line := buffer.String()
		if len(line) > 0 {
			lines = append(lines, line)
		}
	}

	return lines, "Successfully read and parsed", nil
}

func ReadFile(filename string) (string,error) {

	var final string = ""
	// Opening a file
	file, err := os.Open(filename)
	defer file.Close()

	if err != nil {
		return final, err
	}

	// Start reading from the file with a reader.
	reader := bufio.NewReader(file)

	for {
		var buffer bytes.Buffer

		var l []byte
		var isPrefix bool
		for {
			l, isPrefix, err = reader.ReadLine()
			buffer.Write(l)

			if !isPrefix {
				break
			}

			if err != nil {
				break
			}
		}

		if err == io.EOF {
			break
		}

		// converting buffer to string
		line := buffer.String()
		final = final + line
	}

	return final, nil
}

func ReadCSV(filename string) ([]string, string, error) {

	var lines []string
	// Opening a file
	file, err := os.Open(filename)
	defer file.Close()

	if err != nil {
		return lines, "Not able to open the file " + filename, err
	}

	// Start reading from the file with a reader.
	reader := bufio.NewReader(file)

	for {
		var buffer bytes.Buffer

		var l []byte
		var isPrefix bool
		for {
			l, isPrefix, err = reader.ReadLine()
			buffer.Write(l)

			if !isPrefix {
				break
			}

			if err != nil {
				break
			}
		}

		if err == io.EOF {
			break
		}

		// converting buffer to string
		line := buffer.String()
		lines = append(lines, line)
	}

	return lines, "Successfully read and parsed", nil
}




