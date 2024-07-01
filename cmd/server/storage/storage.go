package storage

type MemStorage struct {
	GaugeMetrics   GaugeMetrics
	CounterMetrics CounterMetrics
}

type GaugeMetrics map[string]float64

func (gm GaugeMetrics) Save(name string, value *float64) {
	gm[name] = *value
}

type CounterMetrics map[string]int64

func (cm CounterMetrics) Save(name string, value *int64) {
	cm[name] += *value
}
