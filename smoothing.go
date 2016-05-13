package edges

func (d *Detector) applyVerticalISEF(A, B [][]float32) {
	var b1 float32 = (1.0 - d.smoothingFactor) / (1.0 + d.smoothingFactor)
	b2 := d.smoothingFactor * b1
	x := d.buffer
	y := d.smoothed
	rect := d.workRect

	// compute boundary conditions
	for col := rect.Min.X; col < rect.Max.X; col++ {
		// boundary exists for 1st and last column
		A[col][rect.Min.Y] = b1 * x[col][rect.Min.Y]
		B[col][rect.Max.Y-1] = b2 * x[col][rect.Max.Y-1]
	}
	// compute causal component
	for row := rect.Min.Y + 1; row < rect.Max.Y; row++ {
		for col := rect.Min.X; col < rect.Max.X; col++ {
			A[col][row] = b1*x[col][row] + SmoothingFactor*A[col][row-1]
		}
	}
	// compute anti-causal component
	for row := rect.Max.Y - 2; row >= rect.Min.Y; row-- {
		for col := rect.Min.X; col < rect.Max.X; col++ {
			B[col][row] = b2*x[col][row] + SmoothingFactor*B[col][row+1]
		}
	}
	// boundary case for computing output of first filter
	for col := rect.Min.X; col < rect.Max.X-1; col++ {
		y[col][rect.Max.Y-1] = A[col][rect.Max.Y-1]
	}
	/*
	 * now compute the output of the first filter and store in y
	 * this is the sum of the causal and anti-causal components
	 */
	for row := rect.Min.Y; row < rect.Max.Y-2; row++ {
		for col := rect.Min.X; col < rect.Max.X-1; col++ {
			y[col][row] = A[col][row] + B[col][row+1]
		}
	}
}

func (d *Detector) applyHorizontalISEF(A, B [][]float32) {
	var b1 float32 = (1.0 - d.smoothingFactor) / (1.0 + d.smoothingFactor)
	b2 := d.smoothingFactor * b1
	x := d.smoothed
	y := d.smoothed
	rect := d.workRect

	// compute boundary conditions
	for row := rect.Min.Y; row < rect.Max.Y; row++ {
		A[rect.Min.X][row] = b1 * x[rect.Min.X][row]
		B[rect.Max.X-1][row] = b2 * x[rect.Max.X-1][row]
	}
	// compute causal component
	for col := rect.Min.X + 1; col < rect.Max.X; col++ {
		for row := rect.Min.Y; row < rect.Max.Y; row++ {
			A[col][row] = b1*x[col][row] + SmoothingFactor*A[col-1][row]
		}
	}
	// compute anti-causal component
	for col := rect.Max.X - 2; col >= rect.Min.X; col-- {
		for row := rect.Min.Y; row < rect.Max.Y; row++ {
			B[col][row] = b2*x[col][row] + SmoothingFactor*B[col+1][row]
		}
	}
	// boundary case for computing output of first filter
	for row := rect.Min.Y; row < rect.Max.Y; row++ {
		y[rect.Max.X-1][row] = A[rect.Max.X-1][row]
	}
	/*
	 * now compute the output of the second filter and store in y
	 * this is the sum of the causal and anti-causal components
	 */
	for row := rect.Min.Y; row < rect.Max.Y; row++ {
		for col := rect.Min.X; col < rect.Max.X-1; col++ {
			y[col][row] = A[col][row] + B[col+1][row]
		}
	}
}

func (d *Detector) computeISEF() {
	d.smoothed = d.f2d()
	// store causal component
	A := d.f2d()
	// store anti-causal component
	B := d.f2d()
	// first apply the filter in the vertical direction (to the rows)
	d.applyVerticalISEF(A, B)
	/*
	 * now apply the filter in the horizontal direction (to the columns)
	 * and apply this filter to the results of the previous one
	 */
	d.applyHorizontalISEF(A, B)
}
