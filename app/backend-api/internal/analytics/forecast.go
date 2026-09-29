package analytics

// Double exponential smoothing (additive trend, no seasonality).
// Fallback when external forecasting engine is not configured.
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

	level := series[0]
	trend := 0.0
	if len(series) >= 2 {
		trend = series[1] - series[0]
	}

	for _, point := range series[1:] {
		prevLevel := level
		level = alpha*point + (1-alpha)*(level+trend)
		trend = beta*(level-prevLevel) + (1-beta)*trend
	}

	forecast := make([]float64, horizon)
	for i := range horizon {
		forecast[i] = level + float64(i+1)*trend
		if forecast[i] < 0 {
			forecast[i] = 0
		}
	}
	return forecast
}
