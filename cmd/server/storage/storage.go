package storage

import "strconv"

type MemStorage struct {
	GaugeMetrics   GaugeMetrics
	CounterMetrics CounterMetrics
}

type GaugeMetrics map[string]float64

func (gm GaugeMetrics) Save(name, value string) error {
	floatValue, err := strconv.ParseFloat(value, 64)

	if err != nil {
		return err
	}

	gm[name] = floatValue

	return nil
}

type CounterMetrics map[string]int64

func (cm CounterMetrics) Save(name, value string) error {
	intValue, err := strconv.ParseInt(value, 10, 64)

	if err != nil {
		return err
	}

	cm[name] += intValue

	return nil
}
