package stock

// Trend represents a linear function about a particular trend a stock has
type Trend struct {
	Hits int		// Number of poitns a maxima/minima has hit the trend
	Slope float64		// Slope of linear function
	Constant float64	// Constant of linear function
	Points []*Point		// Points that hit trend
	Parent1 *Point
	Parent2 *Point
}

