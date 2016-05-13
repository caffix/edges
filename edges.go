package edges

import (
	"image"
	"image/color"
	"image/draw"
)

/*
 * the four dials / knobs you can play with
 * to control the results from this algorithm
 */
const (
	WindowSize      = 7
	Ratio           = 0.80
	SmoothingFactor = 0.94
	ThinningFactor  = 0
)

type Detector struct {
	original                                image.Image
	gray                                    *image.Gray
	graySize, bufSize                       image.Point
	workRect                                image.Rectangle
	buffer, smoothed                        [][]float32
	bli, res                                [][]uint8
	bgColor, fgColor                        color.Color
	doHysteresis                            bool
	outlineSize, windowSize, thinningFactor int
	ratio, smoothingFactor                  float32
}

func (d *Detector) i2d() (slice [][]uint8) {
	slice = make([][]uint8, d.bufSize.X)
	for i := range slice {
		slice[i] = make([]uint8, d.bufSize.Y)
	}
	return
}

func (d *Detector) f2d() (slice [][]float32) {
	slice = make([][]float32, d.bufSize.X)
	for i := range slice {
		slice[i] = make([]float32, d.bufSize.Y)
	}
	return
}

// compute the binary Laplacian image from the band-limited Laplacian of the input image
func (d *Detector) computeBLI() {
	d.bli = d.i2d()
	rect := d.workRect
	/*
	 * the bli is computed by taking the difference between the smoothed image
	 * and the original image. In Shen and Castan's paper, this is shown to
	 * approximate the band-limited Laplacian of the image. The bli is then
	 * made by setting all values in the bli to 1 where the Laplacian is
	 * positive and 0 otherwise.
	 */
	for row := rect.Min.Y; row < rect.Max.Y; row++ {
		for col := rect.Min.X; col < rect.Max.X; col++ {
			if d.smoothed[col][row]-d.buffer[col][row] > 0.0 {
				d.bli[col][row] = 1
			}
		}
	}
}

func (d *Detector) inputGray() {
	// convert the original image to grayscale
	d.gray = image.NewGray(d.original.Bounds())
	d.graySize = d.original.Bounds().Size()
	draw.Draw(d.gray, d.original.Bounds(), d.original, image.ZP, draw.Src)
	// calculate various sizes used throughout the implementation
	d.calcSizes()
	// convert the grayscale image to a 2D floating point slice
	d.buffer = d.f2d()
	rect1 := d.gray.Bounds()
	rect2 := d.workRect
	for y1, y2 := rect1.Min.Y, rect2.Min.Y; y1 < rect1.Max.Y; y1, y2 = y1+1, y2+1 {
		for x1, x2 := rect1.Min.X, rect2.Min.X; x1 < rect1.Max.X; x1, x2 = x1+1, x2+1 {
			r, _, _, _ := d.gray.At(x1, y1).RGBA()
			d.buffer[x2][y2] = float32(r)
		}
	}
}

func (d *Detector) outputGray() *image.Gray {
	// create the grayscale image to be returned
	edges := image.NewGray(d.gray.Bounds())
	// transfer the final 2D slice to the grayscale image
	rect1 := edges.Bounds()
	rect2 := d.workRect
	for y1, y2 := rect1.Min.Y, rect2.Min.Y; y1 < rect1.Max.Y; y1, y2 = y1+1, y2+1 {
		for x1, x2 := rect1.Min.X, rect2.Min.X; x1 < rect1.Max.X; x1, x2 = x1+1, x2+1 {
			if d.res[x2][y2] > 0 {
				edges.Set(x1, y1, d.fgColor)
			} else {
				edges.Set(x1, y1, d.bgColor)
			}
		}
	}
	return edges
}

func (d *Detector) calcSizes() {
	d.outlineSize = (d.windowSize/2 + 1) * 2

	d.bufSize.X = d.graySize.X + (d.outlineSize * 2)
	d.bufSize.Y = d.graySize.Y + (d.outlineSize * 2)

	min := image.Point{X: 0 + d.outlineSize, Y: 0 + d.outlineSize}
	max := image.Point{X: (d.bufSize.X - d.outlineSize) + 1, Y: (d.bufSize.Y - d.outlineSize) + 1}

	d.workRect = image.Rectangle{Min: min, Max: max}
}

func (d *Detector) ShenCastan() *image.Gray {
	d.inputGray()
	// smooth input image using recursively implemented ISEF filter
	d.computeISEF()
	// compute binary Laplacian image from smoothed image
	d.computeBLI()
	// perform edge detection using bli and gradient thresholding
	d.locateZeroCrossings()
	// perform hysteresis to remove false positives
	d.thresholdEdges()

	return d.outputGray()
}

func (d *Detector) SetThinningFactor(factor int) {
	d.thinningFactor = factor
}

func (d *Detector) SetSmoothingFactor(factor float32) {
	d.smoothingFactor = factor
}

func (d *Detector) DoHysteresis(b bool) {
	d.doHysteresis = b
}

func (d *Detector) SetRatio(r float32) {
	d.ratio = r
}

func (d *Detector) SetWindowSize(size int) {
	if size >= 3 && (size%2 != 0) {
		d.windowSize = size
		d.outlineSize = (size / 2) + 1
	}
}

func (d *Detector) SetForegroundColor(c color.Color) {
	d.fgColor = c
}

func (d *Detector) SetBackgroundColor(c color.Color) {
	d.bgColor = c
}

func NewEdgeDetector(img image.Image) *Detector {
	return &Detector{
		doHysteresis:    true,
		thinningFactor:  ThinningFactor,
		windowSize:      WindowSize,
		ratio:           Ratio,
		smoothingFactor: SmoothingFactor,
		original:        img,
		bgColor:         image.White.C,
		fgColor:         image.Black.C,
	}
}
