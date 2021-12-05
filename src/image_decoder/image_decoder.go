package image_decoder

import (
	"fmt"
	"image"
	"os"
)

func GetImg(filename string) (image.Image, error) {
	infile, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("error in opening file")
	}
	defer infile.Close()

	src, _, err := image.Decode(infile)
	if err != nil {
		return nil, fmt.Errorf("error in decoding file")
	}
	return src, nil
}

// extract R/G/B vectors of the x-th row of the image
func RowRGB(img image.Image, x int) (r_arr, g_arr, b_arr []uint32) {
	for y := 0; y < img.Bounds().Max.Y; y++ {
		r, g, b, _ := img.At(x, y).RGBA()
		r_arr = append(r_arr, r)
		g_arr = append(g_arr, g)
		b_arr = append(b_arr, b)
	}
	return r_arr, g_arr, b_arr
}
