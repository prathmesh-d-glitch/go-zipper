package worker

type Result struct {
	TaskID     string
	Output     []byte
	Err        error
	BytesIn    int64
	BytesOut   int64
	DurationMs int64
	WorkerID   int
}

func (r *Result) CompressionRatio() float64 {
	if r.BytesIn == 0 {
		return 0
	}
	return float64(r.BytesOut) / float64(r.BytesIn)
}

func (r *Result) IsSuccess() bool {
	return r.Err == nil
}
