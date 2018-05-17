package main

import (
	"encoding/csv"
	"errors"
	"flag"
	"fmt"
	"image/color"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	sm "github.com/flopp/go-staticmaps"
	"github.com/fogleman/gg"
	"github.com/golang/geo/s2"
)

// Error constants
var (
	ErrLatLong           = errors.New("LatLong is incorrect")
	ErrLatLongOutOfRange = errors.New("LatLong is out of range")
)

func main() {

	var limit int
	var filename, mode string

	flag.StringVar(&mode, "mode", "plot", "a string var")
	flag.StringVar(&filename, "file", "", "a string var")
	flag.IntVar(&limit, "limit", 0, "an int var")

	flag.Parse()

	fmt.Println(fmt.Sprintf("Input: %s, mode: %s, limit: %d", filename, mode, limit))
	ctx, rowCount, err := markLocations(limit, filename, mode)
	if err != nil {
		terminate(err)
	}

	img, err := ctx.Render()
	if err != nil {
		terminate(err)
	}

	baseName := filepath.Base(filename)
	baseName = strings.TrimSuffix(baseName, filepath.Ext(baseName))
	outFileName := fmt.Sprintf("images/img-%v-%d-%d.png", baseName, rowCount, time.Now().Unix())
	if err := gg.SavePNG(outFileName, img); err != nil {
		terminate(err)
	}

	fmt.Println("Generated: ", outFileName)
}

func markLocations(limit int, filename, mode string) (*sm.Context, int, error) {
	ctx := sm.NewContext()
	ctx.SetSize(1000, 800)

	filePath, err := filepath.Abs(filename)
	if err != nil {
		return ctx, 0, err
	}

	file, err := os.Open(filePath)
	if err != nil {
		return ctx, 0, err
	}

	defer file.Close()

	rowCount := -1
	if file != nil {
		reader := csv.NewReader(file)

		for {
			record, err := reader.Read()
			if err == io.EOF {
				break
			}

			rowCount++
			if rowCount == 0 {
				continue
			} else if limit == 0 || rowCount < limit {
				x1, y1, err := getLatLong(record[9])
				if err != nil {
					continue
				}

				x2, y2, err := getLatLong(record[12])
				if err != nil {
					continue
				}

				ctx.AddMarker(sm.NewMarker(s2.LatLngFromDegrees(x1, y1), color.RGBA{0x00, 0xff, 0x00, 0xff}, 4.0)) //source
				ctx.AddMarker(sm.NewMarker(s2.LatLngFromDegrees(x2, y2), color.RGBA{0xff, 0, 0, 0xff}, 4.0))       //destination

				if mode == "line" {
					var pos []s2.LatLng
					pos = append(pos, s2.LatLngFromDegrees(x1, y1))
					pos = append(pos, s2.LatLngFromDegrees(x2, y2))

					ctx.AddPath(sm.NewPath(pos, color.RGBA{0x00, 0x00, 0x00, 0xff}, 1.0))
				}
			}
		}
	}

	return ctx, rowCount, nil
}

func getLatLong(latlong string) (float64, float64, error) {
	var err error

	if latlong == "" || latlong == "," || latlong == "-999,-999" {
		return 0, 0, ErrLatLong
	}

	xy := strings.Split(latlong, ",")
	if len(xy) != 2 {
		return 0, 0, ErrLatLong
	}

	x, err := strconv.ParseFloat(strings.TrimSpace(xy[0]), 64)
	if err != nil {
		return 0, 0, err
	}

	y, err := strconv.ParseFloat(strings.TrimSpace(xy[1]), 64)
	if err != nil {
		return 0, 0, err
	}

	if y < 95.1125 || y > 112.874722 || x > 5.792222 || x < -9.128333 {
		return 0, 0, ErrLatLongOutOfRange
	}

	return x, y, err
}

func terminate(err error) {
	// ~ won't expand if we use `file=~/some-file`, use ``-file ~/some-file` instead
	fmt.Println("Usage: go run app.go file <filename> mode plot|line limit 0|N")
	if err != nil {
		fmt.Println("Error: ", err.Error())
	}

	os.Exit(0)
}