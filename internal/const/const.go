package constants

const (
	MetricTypeGauge   = "gauge"
	MetricTypeCounter = "counter"

	// почему то используя в Exec получаю ошибку: syntax error at or near "$1" (SQLSTATE 42601)
	// pgDB.Exec("CREATE TABLE IF NOT EXISTS $1 (id VARCHAR(250) PRIMARY KEY, type VARCHAR(250) NOT NULL, value DOUBLE PRECISION, delta INTEGER)", constants.TableName)
	TableName = "metrics"
)
