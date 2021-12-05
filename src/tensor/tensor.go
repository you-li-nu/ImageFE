package tensor

import (
	"fmt"
	"image"
	"image/color"
	"math"
)

type Tensor struct {
	W, H       int
	R, G, B, A []uint32
}

func NewTensor(img image.Image) Tensor {
	bounds := img.Bounds()
	W, H := bounds.Max.X, bounds.Max.Y
	var R, G, B, A []uint32

	for x := 0; x < H; x++ {
		for y := 0; y < W; y++ {
			c1 := color.RGBAModel.Convert(img.At(x, y)).(color.RGBA)
			R = append(R, uint32(c1.R))
			G = append(G, uint32(c1.G))
			B = append(B, uint32(c1.B))
			A = append(A, uint32(c1.A))
		}
	}

	return Tensor{
		W: W,
		H: H,
		R: R,
		G: G,
		B: B,
		A: A,
	}
}

func (t *Tensor) GetAString() string { //youl
	ret := ""
	for i := 0; i != t.H; i++ {
		for j := 0; j != t.W; j++ {
			ret += fmt.Sprintf("%3d", (t.A[i*t.W+j])) + " "
		}
		ret += "\n"
	}
	return ret
}

func GetZeroTensor(w, h int) Tensor {
	W, H := w, h
	var R, G, B, A []uint32

	for x := 0; x < H; x++ {
		for y := 0; y < W; y++ {
			R = append(R, uint32(0))
			G = append(G, uint32(0))
			B = append(B, uint32(0))
			A = append(A, uint32(0))
		}
	}

	return Tensor{
		W: W,
		H: H,
		R: R,
		G: G,
		B: B,
		A: A,
	}
}

func (t *Tensor) GetRow(x int, c string) []uint32 {
	switch c {
	case "R":
		return t.R[x*t.W : x*t.W+t.W]
	case "G":
		return t.G[x*t.W : x*t.W+t.W]
	case "B":
		return t.B[x*t.W : x*t.W+t.W]
	case "A":
		return t.A[x*t.W : x*t.W+t.W]
	}
	return nil
}

func makeAndCopy(s []uint32, d *[]uint32) {
	l := len(s)
	tmp := make([]uint32, l)
	copy(tmp, s)
	*d = append(*d, tmp...)
}

func (t *Tensor) PaddingCenter(dist int, filler uint32) Tensor {
	var currR, currG, currB, currA []uint32
	top := make([]uint32, t.W+2*dist)
	for i := 0; i < t.W+2*dist; i++ {
		top[i] = filler
	}
	side := make([]uint32, dist)
	for i := 0; i < dist; i++ {
		side[i] = filler
	}

	for i := 0; i < dist; i++ {
		makeAndCopy(top, &currR)
		makeAndCopy(top, &currG)
		makeAndCopy(top, &currB)
		makeAndCopy(top, &currA)
	}
	for i := 0; i < t.H; i++ {
		makeAndCopy(side, &currR)
		makeAndCopy(side, &currG)
		makeAndCopy(side, &currB)
		makeAndCopy(side, &currA)
		makeAndCopy(t.R[i*t.W:i*t.W+t.W], &currR)
		makeAndCopy(t.G[i*t.W:i*t.W+t.W], &currG)
		makeAndCopy(t.B[i*t.W:i*t.W+t.W], &currB)
		makeAndCopy(t.A[i*t.W:i*t.W+t.W], &currA)
		makeAndCopy(side, &currR)
		makeAndCopy(side, &currG)
		makeAndCopy(side, &currB)
		makeAndCopy(side, &currA)
	}
	for i := 0; i < dist; i++ {
		makeAndCopy(top, &currR)
		makeAndCopy(top, &currG)
		makeAndCopy(top, &currB)
		makeAndCopy(top, &currA)
	}
	return Tensor{
		W: t.W + 2*dist,
		H: t.H + 2*dist,
		R: currR,
		G: currG,
		B: currB,
		A: currA,
	}
}

func (t *Tensor) PaddingTopLeft(dist int, filler uint32) Tensor {
	var currR, currG, currB, currA []uint32
	var side, top []uint32
	side = make([]uint32, t.W%dist)
	if t.H%dist == 0 {
		top = make([]uint32, 0)
	} else {
		top = make([]uint32, t.W+t.W%dist)
	}

	for i := 0; i < len(top); i++ {
		top[i] = filler
	}

	for i := 0; i < len(side); i++ {
		side[i] = filler
	}

	for i := 0; i < t.H; i++ {
		makeAndCopy(t.R[i*t.W:i*t.W+t.W], &currR)
		makeAndCopy(t.G[i*t.W:i*t.W+t.W], &currG)
		makeAndCopy(t.B[i*t.W:i*t.W+t.W], &currB)
		makeAndCopy(t.A[i*t.W:i*t.W+t.W], &currA)
		makeAndCopy(side, &currR)
		makeAndCopy(side, &currG)
		makeAndCopy(side, &currB)
		makeAndCopy(side, &currA)
	}
	makeAndCopy(top, &currR)
	makeAndCopy(top, &currG)
	makeAndCopy(top, &currB)
	makeAndCopy(top, &currA)
	return Tensor{
		W: t.W + t.W%dist,
		H: t.H + t.H%dist,
		R: currR,
		G: currG,
		B: currB,
		A: currA,
	}
}

func (t *Tensor) CroppingTopLeft(dist int, filler uint32) Tensor {
	var currR, currG, currB, currA []uint32
	for i := 0; i < t.H-t.H%dist; i++ {
		makeAndCopy(t.R[i*t.W:i*t.W+t.W-t.W%dist], &currR)
		makeAndCopy(t.G[i*t.W:i*t.W+t.W-t.W%dist], &currG)
		makeAndCopy(t.B[i*t.W:i*t.W+t.W-t.W%dist], &currB)
		makeAndCopy(t.A[i*t.W:i*t.W+t.W-t.W%dist], &currA)
	}
	return Tensor{
		W: t.W - t.W%dist,
		H: t.H - t.H%dist,
		R: currR,
		G: currG,
		B: currB,
		A: currA,
	}
}

func (t *Tensor) UnPaddingCenter(dist int) Tensor {
	var currR, currG, currB, currA []uint32
	new_W := t.W - 2*dist
	for i := dist; i < t.H+dist; i++ {
		makeAndCopy(t.R[i*t.W+dist:i*t.W+new_W+dist], &currR)
		makeAndCopy(t.G[i*t.W+dist:i*t.W+new_W+dist], &currG)
		makeAndCopy(t.B[i*t.W+dist:i*t.W+new_W+dist], &currB)
		makeAndCopy(t.A[i*t.W+dist:i*t.W+new_W+dist], &currA)
	}
	return Tensor{
		W: t.W - 2*dist,
		H: t.H - 2*dist,
		R: currR,
		G: currG,
		B: currB,
		A: currA,
	}
}

func (t *Tensor) GetXY(idx int) (int, int) {
	var x, y int
	x = int(math.Ceil(float64(idx)/float64(t.W))) - 1
	if idx%t.W == 0 {
		x += 1
		y = 0
	} else {
		y = idx % t.W
	}
	return x, y
}

func (t *Tensor) GetIdx(x, y int) int {
	return x*t.W + y
}

func (t *Tensor) IsNonEdgeCenter(dist int, idx int) bool {
	x, y := t.GetXY(idx)
	if x < 0+dist || x > t.H-1-dist || y < 0+dist || y > t.W-1-dist {
		return false
	} else {
		return true
	}
}

func (t *Tensor) IsTopLeft(dist int, idx int) bool {
	x, y := t.GetXY(idx)
	if x%dist == 0 && y%dist == 0 {
		return true
	} else {
		return false
	}
}

func (t *Tensor) SubTensorCenter(dist int, x, y int) (Tensor, error) { //square matrix, w & h = dist*2 + 1
	if x-dist < 0 || x+dist >= t.H || y-dist < 0 || y+dist >= t.W {
		fmt.Println("dist:", "h:", t.H, "w:", t.W, "x:", x, "y:", y)
		return *t, fmt.Errorf("SubTensor out of bound")
	}
	var currR, currG, currB, currA []uint32
	for i := x - dist; i < x+dist+1; i++ {
		makeAndCopy(t.R[i*t.W+y-dist:i*t.W+y+dist+1], &currR)
		makeAndCopy(t.G[i*t.W+y-dist:i*t.W+y+dist+1], &currG)
		makeAndCopy(t.B[i*t.W+y-dist:i*t.W+y+dist+1], &currB)
		makeAndCopy(t.A[i*t.W+y-dist:i*t.W+y+dist+1], &currA)
	}
	return Tensor{
		W: dist*2 + 1,
		H: dist*2 + 1,
		R: currR,
		G: currG,
		B: currB,
		A: currA,
	}, nil
}

func (t *Tensor) SubTensorTopLeft(l int, x, y int) (Tensor, error) { //square matrix, w & h = dist*2 + 1
	if x+l-1 >= t.H || y+l-1 >= t.W {
		return *t, fmt.Errorf("SubTensor out of bound")
	}
	var currR, currG, currB, currA []uint32
	for i := x; i < x+l; i++ {
		makeAndCopy(t.R[i*t.W+y:i*t.W+y+l], &currR)
		makeAndCopy(t.G[i*t.W+y:i*t.W+y+l], &currG)
		makeAndCopy(t.B[i*t.W+y:i*t.W+y+l], &currB)
		makeAndCopy(t.A[i*t.W+y:i*t.W+y+l], &currA)
	}
	return Tensor{
		W: l,
		H: l,
		R: currR,
		G: currG,
		B: currB,
		A: currA,
	}, nil
}

func (t *Tensor) InnerProd(mask []float32) (color.Color, error) {
	if len(mask) != len(t.R) {
		return nil, fmt.Errorf("unmatched size for inner product")
	}
	var r, g, b float32
	for i := range mask {
		r += float32(t.R[i]) * mask[i]
		g += float32(t.G[i]) * mask[i]
		b += float32(t.B[i]) * mask[i]
		if t.A[i] != t.A[0] {
			return nil, fmt.Errorf("a value mismatch")
		}
	}
	return color.RGBA{uint8(r), uint8(g), uint8(b), uint8(int(t.A[0]))}, nil
}

func (t *Tensor) ToImg() *image.RGBA {
	new_img := image.NewRGBA(image.Rectangle{image.Point{0, 0}, image.Point{t.W, t.H}})
	for i := 0; i < t.H; i++ {
		for j := 0; j < t.W; j++ {
			idx := t.GetIdx(i, j)
			r := t.R[idx]
			g := t.G[idx]
			b := t.B[idx]
			a := t.A[idx]
			c := color.RGBA{uint8(r), uint8(g), uint8(b), uint8(a)}
			new_img.Set(i, j, c)
		}
	}
	return new_img
}

func MergeMultipleTensors(tensors []Tensor, kernel_size int) (Tensor, error) {
	if len(tensors) != kernel_size*kernel_size {
		return Tensor{}, fmt.Errorf("number of tensors %d not equal to number of elements in a kernel %d", len(tensors), kernel_size*kernel_size)
	}
	var mergedR, mergedG, mergedB, mergedA []uint32
	for i := 0; i != tensors[0].H; i++ {
		for ii := 0; ii != kernel_size; ii++ {
			for j := 0; j != tensors[0].W; j++ {
				for jj := 0; jj != kernel_size; jj++ {
					mergedR = append(mergedR, tensors[kernel_size*ii+jj].R[tensors[0].GetIdx(i, j)])
					mergedG = append(mergedG, tensors[kernel_size*ii+jj].G[tensors[0].GetIdx(i, j)])
					mergedB = append(mergedB, tensors[kernel_size*ii+jj].B[tensors[0].GetIdx(i, j)])
					mergedA = append(mergedA, tensors[kernel_size*ii+jj].A[tensors[0].GetIdx(i, j)])
				}
			}
		}
	}
	return Tensor{
		W: tensors[0].W * kernel_size,
		H: tensors[0].H * kernel_size,
		R: mergedR,
		G: mergedG,
		B: mergedB,
		A: mergedA,
	}, nil
}
