package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"testing"
	"time"

	constants "github.com/dglazkoff/go-metrics/internal/const"
	"github.com/dglazkoff/go-metrics/internal/logger"
	"github.com/dglazkoff/go-metrics/internal/models"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"
)

func TestReadMetrics(t *testing.T) {
	err := logger.Initialize()
	require.NoError(t, err)

	t.Run("successful read of metrics", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		assert.NoError(t, err)
		defer db.Close()

		rows := sqlmock.NewRows([]string{"id", "type", "value", "delta"}).
			AddRow("1", "gauge", 10.5, nil).
			AddRow("2", "counter", nil, 15)

		mock.ExpectQuery("SELECT id, type, value, delta from metrics").
			WillReturnRows(rows)

		storage := New(db, RetryIntervals)

		metrics, err := storage.ReadMetrics(context.Background())

		assert.NoError(t, err)
		assert.Len(t, metrics, 2)
		assert.Equal(t, "1", metrics[0].ID)
		assert.Equal(t, "gauge", metrics[0].MType)
		assert.Equal(t, 10.5, *metrics[0].Value)
		assert.Nil(t, metrics[0].Delta)
		assert.Equal(t, "2", metrics[1].ID)
		assert.Equal(t, "counter", metrics[1].MType)
		assert.Nil(t, metrics[1].Value)
		assert.Equal(t, int64(15), *metrics[1].Delta)

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database query error", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		assert.NoError(t, err)
		defer db.Close()

		mock.ExpectQuery("SELECT id, type, value, delta from metrics").
			WillReturnError(fmt.Errorf("query error"))

		storage := New(db, RetryIntervals)

		metrics, err := storage.ReadMetrics(context.Background())

		assert.Error(t, err)
		assert.Nil(t, metrics)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("row scanning error", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		assert.NoError(t, err)
		defer db.Close()

		rows := sqlmock.NewRows([]string{"id", "type", "value", "delta"}).
			AddRow("invalid", "invalid", "not-a-float", "not-an-int")

		mock.ExpectQuery("SELECT id, type, value, delta from metrics").
			WillReturnRows(rows)

		storage := New(db, RetryIntervals)

		metrics, err := storage.ReadMetrics(context.Background())

		assert.NoError(t, err)
		assert.Empty(t, metrics)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestReadMetric(t *testing.T) {
	err := logger.Initialize()
	require.NoError(t, err)

	t.Run("successful read of metrics", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		assert.NoError(t, err)
		defer db.Close()

		rowsGauge := sqlmock.NewRows([]string{"id", "type", "value", "delta"}).
			AddRow("1", "gauge", 10.5, nil)

		rowsCounter := sqlmock.NewRows([]string{"id", "type", "value", "delta"}).
			AddRow("2", "counter", nil, 15)

		mock.ExpectQuery("SELECT (.+) from metrics (.+)").
			WithArgs("1").
			WillReturnRows(rowsGauge)

		mock.ExpectQuery("SELECT (.+) from metrics (.+)").
			WithArgs("2").
			WillReturnRows(rowsCounter)

		storage := New(db, RetryIntervals)

		metricGauge, err := storage.ReadMetric(context.Background(), "1")
		assert.NoError(t, err)
		metricCounter, err := storage.ReadMetric(context.Background(), "2")

		assert.NoError(t, err)
		assert.Equal(t, "1", metricGauge.ID)
		assert.Equal(t, "gauge", metricGauge.MType)
		assert.Equal(t, 10.5, *metricGauge.Value)
		assert.Nil(t, metricGauge.Delta)

		assert.Equal(t, "2", metricCounter.ID)
		assert.Equal(t, "counter", metricCounter.MType)
		assert.Nil(t, metricCounter.Value)
		assert.Equal(t, int64(15), *metricCounter.Delta)

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database query error", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		assert.NoError(t, err)
		defer db.Close()

		mock.ExpectQuery("SELECT (.+) from metrics (.+)").
			WithArgs("1").
			WillReturnError(fmt.Errorf("query error"))

		storage := New(db, RetryIntervals)

		metric, err := storage.ReadMetric(context.Background(), "1")

		assert.Error(t, err)
		assert.Equal(t, models.Metrics{}, metric)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("row scanning error", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		assert.NoError(t, err)
		defer db.Close()

		rows := sqlmock.NewRows([]string{"id", "type", "value", "delta"}).
			AddRow("invalid", "invalid", "not-a-float", "not-an-int")

		mock.ExpectQuery("SELECT (.+) from metrics (.+)").
			WithArgs("1").
			WillReturnRows(rows)

		storage := New(db, RetryIntervals)

		metric, err := storage.ReadMetric(context.Background(), "1")

		assert.Error(t, err)
		assert.Equal(t, "error while reading metric", err.Error())
		assert.Equal(t, models.Metrics{}, metric)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("metric not found", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		assert.NoError(t, err)
		defer db.Close()

		mock.ExpectQuery("SELECT (.+) from metrics (.+)").
			WithArgs("nonexistent-id").
			WillReturnError(sql.ErrNoRows)

		storage := New(db, RetryIntervals)

		metric, err := storage.ReadMetric(context.Background(), "nonexistent-id")

		assert.Error(t, err)
		assert.Equal(t, "error while reading metric", err.Error())
		assert.Equal(t, models.Metrics{}, metric)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestDbQueryRow(t *testing.T) {
	err := logger.Initialize()
	require.NoError(t, err)

	retryIntervals := []time.Duration{50 * time.Millisecond, 100 * time.Millisecond}

	t.Run("successful query on first attempt", func(t *testing.T) {
		db, _, err := sqlmock.New()
		assert.NoError(t, err)
		defer db.Close()

		query := func() (*sql.Row, error) {
			return &sql.Row{}, nil
		}

		storage := New(db, retryIntervals)
		row, err := storage.dbQueryRow(query)

		assert.NoError(t, err)
		assert.NotNil(t, row)
	})

	t.Run("connection error with retries", func(t *testing.T) {
		attempt := 0

		query := func() (*sql.Row, error) {
			attempt++
			if attempt < 2 {
				return nil, &pgconn.PgError{Code: pgerrcode.ConnectionException}
			}
			return &sql.Row{}, nil
		}

		storage := New(nil, retryIntervals)
		row, err := storage.dbQueryRow(query)

		assert.NoError(t, err)
		assert.NotNil(t, row)
		assert.Equal(t, 2, attempt)
	})

	t.Run("max retries exceeded", func(t *testing.T) {
		attempt := 0

		query := func() (*sql.Row, error) {
			attempt++
			return nil, &pgconn.PgError{Code: pgerrcode.ConnectionException}
		}

		storage := New(nil, retryIntervals)
		row, err := storage.dbQueryRow(query)

		assert.Error(t, err)
		assert.Nil(t, row)
		assert.Equal(t, "no connection to database", err.Error())
		assert.Equal(t, len(retryIntervals), attempt)
	})

	t.Run("non-retryable error", func(t *testing.T) {
		query := func() (*sql.Row, error) {
			return nil, errors.New("non-retryable error")
		}

		storage := New(nil, retryIntervals)
		row, err := storage.dbQueryRow(query)

		assert.Error(t, err)
		assert.Equal(t, "non-retryable error", err.Error())
		assert.Nil(t, row)
	})
}

func TestDbExecute(t *testing.T) {
	err := logger.Initialize()
	require.NoError(t, err)

	retryIntervals := []time.Duration{50 * time.Millisecond, 100 * time.Millisecond}
	storage := New(nil, retryIntervals)

	t.Run("successful execution on first attempt", func(t *testing.T) {
		exec := func() (sql.Result, error) {
			return sqlmock.NewResult(1, 1), nil
		}

		res, err := storage.dbExecute(exec)

		assert.NoError(t, err)
		assert.NotNil(t, res)
	})

	t.Run("connection error with retries", func(t *testing.T) {
		attempt := 0

		exec := func() (sql.Result, error) {
			attempt++
			if attempt < 2 {
				return nil, &pgconn.PgError{Code: pgerrcode.ConnectionException}
			}
			return sqlmock.NewResult(1, 1), nil
		}

		res, err := storage.dbExecute(exec)

		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, 2, attempt)
	})

	t.Run("max retries exceeded", func(t *testing.T) {
		attempt := 0

		exec := func() (sql.Result, error) {
			attempt++
			return nil, &pgconn.PgError{Code: pgerrcode.ConnectionException}
		}

		res, err := storage.dbExecute(exec)

		assert.Error(t, err)
		assert.Nil(t, res)
		assert.Equal(t, "no connection to database", err.Error())
		assert.Equal(t, len(retryIntervals), attempt)
	})

	t.Run("non-retryable error", func(t *testing.T) {
		exec := func() (sql.Result, error) {
			return nil, errors.New("non-retryable error")
		}

		res, err := storage.dbExecute(exec)

		assert.Error(t, err)
		assert.Equal(t, "non-retryable error", err.Error())
		assert.Nil(t, res)
	})
}

func TestBootstrap(t *testing.T) {
	err := logger.Initialize()
	require.NoError(t, err)

	t.Run("successful table creation", func(t *testing.T) {
		mockDB, mock, err := sqlmock.New()
		assert.NoError(t, err)
		defer mockDB.Close()

		storage := New(mockDB, RetryIntervals)

		mock.ExpectExec("CREATE TABLE IF NOT EXISTS metrics").
			WillReturnResult(sqlmock.NewResult(1, 1))

		err = Bootstrap(storage)

		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("table creation error", func(t *testing.T) {
		mockDB, mock, err := sqlmock.New()
		assert.NoError(t, err)
		defer mockDB.Close()

		storage := New(mockDB, RetryIntervals)

		mock.ExpectExec("CREATE TABLE IF NOT EXISTS metrics").
			WillReturnError(errors.New("failed to create table"))

		err = Bootstrap(storage)

		assert.Error(t, err)
		assert.EqualError(t, err, "failed to create table")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestUpdateMetric(t *testing.T) {
	ctx := context.Background()

	err := logger.Initialize()
	require.NoError(t, err)

	t.Run("insert or update gauge metric", func(t *testing.T) {
		mockDB, mock, err := sqlmock.New()
		assert.NoError(t, err)
		defer mockDB.Close()

		value := 42.0

		storage := New(mockDB, RetryIntervals)

		metric := models.Metrics{
			ID:    "gauge_metric_1",
			MType: constants.MetricTypeGauge,
			Value: &value,
		}

		mock.ExpectExec("INSERT INTO metrics").
			WithArgs(metric.ID, metric.MType, metric.Value, sql.NullInt64{}).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err = storage.UpdateMetric(ctx, metric)

		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("insert new counter metric", func(t *testing.T) {
		mockDB, mock, err := sqlmock.New()
		assert.NoError(t, err)
		defer mockDB.Close()

		value := int64(5)
		storage := New(mockDB, RetryIntervals)

		metric := models.Metrics{
			ID:    "counter_metric_1",
			MType: constants.MetricTypeCounter,
			Delta: &value,
		}

		mock.ExpectQuery("SELECT (.+) from metrics (.+)").
			WithArgs("counter_metric_1").
			WillReturnError(fmt.Errorf("query error"))
		mock.ExpectExec("INSERT INTO metrics").
			WithArgs(metric.ID, metric.MType, sql.NullFloat64{}, metric.Delta).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err = storage.UpdateMetric(ctx, metric)

		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("update existing counter metric", func(t *testing.T) {
		mockDB, mock, err := sqlmock.New()
		assert.NoError(t, err)
		defer mockDB.Close()

		valueExisting := int64(10)
		valueNew := int64(5)
		storage := New(mockDB, RetryIntervals)

		metric := models.Metrics{
			ID:    "counter_metric_1",
			MType: constants.MetricTypeCounter,
			Delta: &valueNew,
		}

		rowsCounter := sqlmock.NewRows([]string{"id", "type", "value", "delta"}).
			AddRow("counter_metric_1", "counter", nil, valueExisting)

		mock.ExpectQuery("SELECT (.+) from metrics (.+)").
			WithArgs("counter_metric_1").
			WillReturnRows(rowsCounter)
		mock.ExpectExec("UPDATE metrics SET delta").
			WithArgs(int64(15), metric.ID).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err = storage.UpdateMetric(ctx, metric)

		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("unknown metric type", func(t *testing.T) {
		mockDB, _, err := sqlmock.New()
		assert.NoError(t, err)
		defer mockDB.Close()

		storage := New(mockDB, RetryIntervals)

		metric := models.Metrics{
			ID:    "unknown_metric",
			MType: "unknown_type",
		}

		err = storage.UpdateMetric(ctx, metric)

		assert.Error(t, err)
		assert.Equal(t, "unknown metric type unknown_type", err.Error())
	})
}
