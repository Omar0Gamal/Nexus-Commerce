package analytics

// HoltWinters applies double exponential smoothing (additive trend, no seasonality)
// to the given historical series and returns a forecast of length horizon.
//
// alpha controls level smoothing (0 < alpha < 1).
// beta  controls trend smoothing  (0 < beta  < 1).
//
// AI seam: when the AI Engine is connected, this function will be replaced by a call
// to POST /v1/forecast/demand on the engine (statsforecast AutoARIMA). The
// GetForecast/GetInventoryForecast service method signatures are unchanged either way.
func HoltWinters(series []float64, horizon int, alpha, beta float64) []float64 {
	if len(series) == 0 || horizon <= 0 {
		return make([]float64, horizon)
	}
	if alpha <= 0 || alpha >= 1 {
		alpha = 0.3
	}
	if beta <= 0 || beta >= 1 {
		beta = 0.1
	}

	// Initialise level and trend from first two observations.
	level := series[0]
	trend := 0.0
	if len(series) >= 2 {
		trend = series[1] - series[0]
	}

	for _, v := range series[1:] {
		prevLevel := level
		level = alpha*v + (1-alpha)*(level+trend)
		trend = beta*(level-prevLevel) + (1-beta)*trend
	}

	out := make([]float64, horizon)
	for i := range horizon {
		out[i] = level + float64(i+1)*trend
		if out[i] < 0 {
			out[i] = 0
		}
	}
	return out
}
