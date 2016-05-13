package edges

import (
	"image"
	"math"
)

func estimateThresholds(lap [][]float32, rect image.Rectangle) (low, high float32) {
	var k, count int
	var hist [256]int
	var vmax, vmin, scale, x float32

	// build a histogram of the laplacian image
	vmin = float32(math.Abs(float64(lap[rect.Min.X+20][rect.Min.X+20])))
	vmax = vmin
	for row := rect.Min.Y; row < rect.Max.Y; row++ {
		for col := rect.Min.X; col < rect.Max.X; col++ {
			x = lap[col][row]
			if vmin > x {
				vmin = x
			}
			if vmax < x {
				vmax = x
			}
		}
	}

	scale = 256.0 / (vmax - vmin + 1)

	for row := rect.Min.Y; row < rect.Max.Y; row++ {
		for col := rect.Min.X; col < rect.Max.X; col++ {
			x = lap[col][row]
			k = int((x - vmin) * scale)
			hist[k] += 1
		}
	}
	// the high threshold should be > 80 or 90% of the pixels
	k = 255
	total := ((rect.Max.Y - 1) * (rect.Max.X - 1))
	max := int(Ratio * float32(total))
	count = hist[k]
	for count < max {
		k--
		if k < 0 {
			break
		}
		count += hist[k]
	}

	high = (float32(k) / scale) + vmin
	low = high / 2
	return
}

// return true if it marked something
func (d *Detector) markConnected(col, row, level int, low, high float32) uint8 {
	var notChainEnd uint8
	rect := d.workRect

	// stop if you go off the edge of the image
	if row >= rect.Max.Y || row < rect.Min.Y || col >= rect.Max.X || col < rect.Min.X {
		return 0
	}
	// stop if the point has already been visited
	if d.res[col][row] != 0 {
		return 0
	}
	// stop when you hit an image boundary
	if d.buffer[col][row] == 0.0 {
		return 0
	}

	if d.buffer[col][row] > low {
		d.res[col][row] = 1
	} else {
		d.res[col][row] = 255
	}

	notChainEnd |= d.markConnected(col+1, row, level+1, low, high)
	notChainEnd |= d.markConnected(col-1, row, level+1, low, high)
	notChainEnd |= d.markConnected(col+1, row+1, level+1, low, high)
	notChainEnd |= d.markConnected(col, row+1, level+1, low, high)
	notChainEnd |= d.markConnected(col-1, row+1, level+1, low, high)
	notChainEnd |= d.markConnected(col-1, row-1, level+1, low, high)
	notChainEnd |= d.markConnected(col, row-1, level+1, low, high)
	notChainEnd |= d.markConnected(col+1, row-1, level+1, low, high)

	if notChainEnd != 0 && level > 0 {
		// do some contour thinning
		if d.thinningFactor > 0 && (level%d.thinningFactor != 0) {
			// delete this point
			d.res[col][row] = 255
		}
	}
	return 1
}

func (d *Detector) thresholdEdges() {
	d.res = d.i2d()
	rect := d.workRect
	low, high := estimateThresholds(d.buffer, rect)

	if d.doHysteresis == false {
		low = high
	}

	for row := rect.Min.Y; row < rect.Max.Y; row++ {
		for col := rect.Min.X; col < rect.Max.X; col++ {
			// only check a contour if it is above the high threshold
			if d.buffer[col][row] > high {
				d.markConnected(col, row, 0, low, high)
			}
		}
	}

	// erase all points which were 255
	for row := rect.Min.Y; row < rect.Max.Y; row++ {
		for col := rect.Min.X; col < rect.Max.X; col++ {
			if d.res[col][row] == 255 {
				d.res[col][row] = 0
			}
		}
	}
}
