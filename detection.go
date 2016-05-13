package edges

// finds zero-crossings in Laplacian (bli)
func isCandidateEdge(bli [][]uint8, smoothed [][]float32, col, row int) (result bool) {
	result = false
	/*
	 * test for zero-crossings of Laplacian, then make sure that zero-crossing
	 * sign correspondence principle is satisfied (i.e. a positive z-c must
	 * have a positive 1st derivative where positive z-c means the 2nd derivative
	 * goes from positive to negative as we pass through the step edge
	 */
	if bli[col][row] == 1 && bli[col][row+1] == 0 {
		// positive z-c
		if smoothed[col][row+1]-smoothed[col][row-1] > 0 {
			result = true
		}
	} else if bli[col][row] == 1 && bli[col+1][row] == 0 {
		// positive z-c
		if smoothed[col+1][row]-smoothed[col-1][row] > 0 {
			result = true
		}
	} else if bli[col][row] == 1 && bli[col][row-1] == 0 {
		// negative z-c
		if smoothed[col][row+1]-smoothed[col][row-1] < 0 {
			result = true
		}
	} else if bli[col][row] == 1 && bli[col-1][row] == 0 {
		// negative z-c
		if smoothed[col+1][row]-smoothed[col-1][row] < 0 {
			result = true
		}
	}
	return
}

func (d *Detector) computeAdaptiveGradient(col, row int) (grad float32) {
	var numOn, numOff int
	var sumOn, sumOff, avgOn, avgOff float32

	halfWin := d.windowSize / 2
	for i := -d.windowSize / 2; i <= halfWin; i++ {
		for j := -d.windowSize / 2; j <= halfWin; j++ {
			if d.bli[col+j][row+i] != 0 {
				sumOn += d.smoothed[col+j][row+i]
				numOn++
			} else {
				sumOff += d.smoothed[col+j][row+i]
				numOff++
			}
		}
	}
	if sumOff > 0 {
		avgOff = sumOff / float32(numOff)
	}
	if sumOn > 0 {
		avgOn = sumOn / float32(numOn)
	}
	return avgOff - avgOn
}

func (d *Detector) locateZeroCrossings() {
	rect := d.workRect

	for row := rect.Min.Y; row < rect.Max.Y; row++ {
		for col := rect.Min.X; col < rect.Max.X; col++ {
			// check if pixel is a zero-crossing of the Laplacian
			if isCandidateEdge(d.bli, d.smoothed, col, row) {
				// now do gradient thresholding
				grad := d.computeAdaptiveGradient(col, row)
				d.buffer[col][row] = grad
			} else {
				d.buffer[col][row] = 0.0
			}
		}
	}
}
