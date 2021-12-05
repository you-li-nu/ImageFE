package kernel

import "fmt"

// https://en.wikipedia.org/wiki/Kernel_(image_processing)

func Gaussian_blur(dist int) ([]uint32, uint32, error) {
	if dist == 3 {
		mask := []uint32{1, 2, 1, 2, 4, 2, 1, 2, 1}
		var divisor uint32 = 16
		return mask, divisor, nil
	}
	if dist == 5 {
		mask := []uint32{1, 4, 6, 4, 1, 4, 16, 24, 16, 4, 6, 24, 36, 24, 6, 4, 16, 24, 16, 4, 1, 4, 6, 4, 1}
		var divisor uint32 = 256
		return mask, divisor, nil
	}
	return nil, 0, fmt.Errorf("unsupported dist: %d. should be 3 or 5", dist)
}

func Box_blur(dist int) ([]uint32, uint32, error) {
	if dist == 3 {
		mask := []uint32{1, 1, 1, 1, 1, 1, 1, 1, 1}
		var divisor uint32 = 9
		return mask, divisor, nil
	}
	if dist == 5 {
		mask := []uint32{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1}
		var divisor uint32 = 25
		return mask, divisor, nil
	}
	return nil, 0, fmt.Errorf("unsupported dist: %d. should be 3 or 5", dist)
}

func Downsize_partial(dist int, idx int) ([]uint32, uint32, error) {
	mask := make([]uint32, dist*dist)
	if idx <= dist*dist-1 {
		mask[idx] = 1
		var divisor uint32 = 1
		return mask, divisor, nil
	} else {
		return nil, 0, fmt.Errorf("index %d out of range: 0 - %d", idx, dist*dist-1)
	}
}

func Identity(dist int) ([]uint32, uint32, error) {
	mask := make([]uint32, dist*dist)
	mask[(dist*dist-1)/2] = 1
	var divisor uint32 = 1
	return mask, divisor, nil
}
