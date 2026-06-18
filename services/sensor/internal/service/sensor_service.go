package service

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/19parwiz/agripro-core/services/sensor/internal/config"
	"github.com/19parwiz/agripro-core/services/sensor/internal/model"
	"github.com/19parwiz/agripro-core/services/sensor/internal/repository"
	apperrors "github.com/19parwiz/agripro-core/shared/errors"
)

// SensorService handles ingestion, archiving, and history queries.
type SensorService struct {
	devices  repository.DeviceRepository
	readings repository.ReadingRepository
	archives repository.ArchiveRepository

	snapshotURL       string
	ingestDeviceID    string
	ingestionEnabled  bool
	ingestionInterval time.Duration
	httpClient        *http.Client
}

func NewSensorService(
	cfg config.Config,
	devices repository.DeviceRepository,
	readings repository.ReadingRepository,
	archives repository.ArchiveRepository,
) *SensorService {
	return &SensorService{
		devices:           devices,
		readings:          readings,
		archives:          archives,
		snapshotURL:       cfg.SnapshotURL,
		ingestDeviceID:    cfg.IngestDeviceID,
		ingestionEnabled:  cfg.IngestionEnabled,
		ingestionInterval: cfg.IngestionInterval,
		httpClient:        &http.Client{Timeout: 30 * time.Second},
	}
}

// HistoryPoint is one chart bucket returned to the mobile app.
type HistoryPoint struct {
	Time         time.Time `json:"time"`
	AvgValue     float64   `json:"avg_value"`
	MinValue     float64   `json:"min_value"`
	MaxValue     float64   `json:"max_value"`
	ReadingCount int       `json:"reading_count"`
}

// StartSchedulers runs ingestion and archive background jobs until ctx is cancelled.
func (s *SensorService) StartSchedulers(ctx context.Context) {
	if s.ingestionEnabled && s.snapshotURL != "" {
		go s.runIngestionScheduler(ctx)
	}

	go s.runHourlyArchiveScheduler(ctx)
	go s.runDailyArchiveScheduler(ctx)
}

// RunIngestionOnce fetches the VPS snapshot and stores readings for target devices.
func (s *SensorService) RunIngestionOnce(ctx context.Context) error {
	if s.snapshotURL == "" {
		return fmt.Errorf("snapshot url is not configured")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.snapshotURL, nil)
	if err != nil {
		return fmt.Errorf("create snapshot request: %w", err)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("fetch snapshot: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("fetch snapshot: status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return fmt.Errorf("read snapshot body: %w", err)
	}

	parsed, recordedAt, err := parseSnapshot(body)
	if err != nil {
		return err
	}

	deviceIDs, err := s.targetDeviceIDs(ctx)
	if err != nil {
		return err
	}
	if len(deviceIDs) == 0 {
		slog.Info("sensor ingestion skipped: no target devices")
		return nil
	}

	now := time.Now().UTC()
	var batch []model.Reading
	for _, deviceID := range deviceIDs {
		for _, item := range parsed {
			batch = append(batch, model.Reading{
				DeviceID:   deviceID,
				SensorType: item.SensorType,
				Value:      item.Value,
				Unit:       item.Unit,
				RecordedAt: recordedAt,
				CreatedAt:  now,
			})
		}
	}

	if err := s.readings.CreateBatch(ctx, batch); err != nil {
		return fmt.Errorf("store snapshot readings: %w", err)
	}

	slog.Info("sensor ingestion complete",
		"devices", len(deviceIDs),
		"readings", len(batch),
		"recorded_at", recordedAt,
	)

	return nil
}

func (s *SensorService) HourlyHistory(
	ctx context.Context,
	userID, deviceID, sensorType string,
	from, to time.Time,
) ([]HistoryPoint, error) {
	if err := s.validateHistoryRequest(ctx, userID, deviceID, sensorType, from, to); err != nil {
		return nil, err
	}

	archives, err := s.archives.ListHourly(ctx, deviceID, sensorType, from, to)
	if err != nil {
		return nil, apperrors.Wrap(err, http.StatusInternalServerError, "could not load hourly history")
	}
	if len(archives) > 0 {
		return toHourlyPoints(archives), nil
	}

	return s.hourlyFromReadings(ctx, deviceID, sensorType, from, to)
}

func (s *SensorService) DailyHistory(
	ctx context.Context,
	userID, deviceID, sensorType string,
	from, to time.Time,
) ([]HistoryPoint, error) {
	if err := s.validateHistoryRequest(ctx, userID, deviceID, sensorType, from, to); err != nil {
		return nil, err
	}

	archives, err := s.archives.ListDaily(ctx, deviceID, sensorType, from, to)
	if err != nil {
		return nil, apperrors.Wrap(err, http.StatusInternalServerError, "could not load daily history")
	}
	if len(archives) > 0 {
		return toDailyPoints(archives), nil
	}

	return s.dailyFromReadings(ctx, deviceID, sensorType, from, to)
}

func (s *SensorService) ArchivePreviousHour(ctx context.Context) error {
	bucketStart := time.Now().UTC().Truncate(time.Hour).Add(-time.Hour)
	return s.archiveHour(ctx, bucketStart)
}

func (s *SensorService) ArchivePreviousDay(ctx context.Context) error {
	yesterday := time.Now().UTC().AddDate(0, 0, -1)
	return s.archiveDay(ctx, yesterday)
}

func (s *SensorService) targetDeviceIDs(ctx context.Context) ([]string, error) {
	if s.ingestDeviceID != "" {
		return []string{s.ingestDeviceID}, nil
	}

	ids, err := s.devices.ListIDs(ctx)
	if err != nil {
		return nil, fmt.Errorf("list devices for ingestion: %w", err)
	}

	return ids, nil
}

func (s *SensorService) validateHistoryRequest(
	ctx context.Context,
	userID, deviceID, sensorType string,
	from, to time.Time,
) error {
	if !isValidSensorType(sensorType) {
		return apperrors.ErrBadRequest
	}
	if !to.After(from) {
		return apperrors.ErrBadRequest
	}

	owned, err := s.devices.BelongsToUser(ctx, deviceID, userID)
	if err != nil {
		return apperrors.Wrap(err, http.StatusInternalServerError, "could not verify device access")
	}
	if !owned {
		return apperrors.ErrForbidden
	}

	return nil
}

func (s *SensorService) hourlyFromReadings(
	ctx context.Context,
	deviceID, sensorType string,
	from, to time.Time,
) ([]HistoryPoint, error) {
	start := from.Truncate(time.Hour)
	var points []HistoryPoint

	for bucket := start; bucket.Before(to); bucket = bucket.Add(time.Hour) {
		archive, err := s.readings.AggregateForHour(ctx, deviceID, sensorType, bucket)
		if err != nil {
			return nil, apperrors.Wrap(err, http.StatusInternalServerError, "could not aggregate hourly history")
		}
		if archive == nil {
			continue
		}

		points = append(points, HistoryPoint{
			Time:         archive.BucketStart,
			AvgValue:     archive.AvgValue,
			MinValue:     archive.MinValue,
			MaxValue:     archive.MaxValue,
			ReadingCount: archive.ReadingCount,
		})
	}

	return points, nil
}

func (s *SensorService) dailyFromReadings(
	ctx context.Context,
	deviceID, sensorType string,
	from, to time.Time,
) ([]HistoryPoint, error) {
	start := time.Date(from.Year(), from.Month(), from.Day(), 0, 0, 0, 0, time.UTC)
	endDay := time.Date(to.Year(), to.Month(), to.Day(), 0, 0, 0, 0, time.UTC)

	var points []HistoryPoint
	for day := start; day.Before(endDay); day = day.AddDate(0, 0, 1) {
		archive, err := s.readings.AggregateForDay(ctx, deviceID, sensorType, day)
		if err != nil {
			return nil, apperrors.Wrap(err, http.StatusInternalServerError, "could not aggregate daily history")
		}
		if archive == nil {
			continue
		}

		points = append(points, HistoryPoint{
			Time:         archive.BucketDate,
			AvgValue:     archive.AvgValue,
			MinValue:     archive.MinValue,
			MaxValue:     archive.MaxValue,
			ReadingCount: archive.ReadingCount,
		})
	}

	return points, nil
}

func (s *SensorService) archiveHour(ctx context.Context, bucketStart time.Time) error {
	deviceIDs, err := s.devices.ListIDs(ctx)
	if err != nil {
		return fmt.Errorf("list devices for hourly archive: %w", err)
	}

	now := time.Now().UTC()
	for _, deviceID := range deviceIDs {
		for _, sensorType := range allSensorTypes() {
			archive, err := s.readings.AggregateForHour(ctx, deviceID, sensorType, bucketStart)
			if err != nil {
				return err
			}
			if archive == nil {
				continue
			}

			archive.CreatedAt = now
			if err := s.archives.UpsertHourly(ctx, archive); err != nil {
				return err
			}
		}
	}

	return nil
}

func (s *SensorService) archiveDay(ctx context.Context, bucketDate time.Time) error {
	deviceIDs, err := s.devices.ListIDs(ctx)
	if err != nil {
		return fmt.Errorf("list devices for daily archive: %w", err)
	}

	day := time.Date(bucketDate.Year(), bucketDate.Month(), bucketDate.Day(), 0, 0, 0, 0, time.UTC)
	now := time.Now().UTC()

	for _, deviceID := range deviceIDs {
		for _, sensorType := range allSensorTypes() {
			archive, err := s.readings.AggregateForDay(ctx, deviceID, sensorType, day)
			if err != nil {
				return err
			}
			if archive == nil {
				continue
			}

			archive.CreatedAt = now
			if err := s.archives.UpsertDaily(ctx, archive); err != nil {
				return err
			}
		}
	}

	return nil
}

func (s *SensorService) runIngestionScheduler(ctx context.Context) {
	s.runJob(ctx, "ingestion", s.ingestionInterval, s.RunIngestionOnce)
}

func (s *SensorService) runHourlyArchiveScheduler(ctx context.Context) {
	for {
		next := nextHourlyArchiveRun(time.Now().UTC())
		if !waitUntil(ctx, next) {
			return
		}

		if err := s.ArchivePreviousHour(ctx); err != nil {
			slog.Error("hourly archive failed", "error", err)
		}
	}
}

func (s *SensorService) runDailyArchiveScheduler(ctx context.Context) {
	for {
		next := nextDailyArchiveRun(time.Now().UTC())
		if !waitUntil(ctx, next) {
			return
		}

		if err := s.ArchivePreviousDay(ctx); err != nil {
			slog.Error("daily archive failed", "error", err)
		}
	}
}

func (s *SensorService) runJob(ctx context.Context, name string, interval time.Duration, job func(context.Context) error) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	if err := job(ctx); err != nil {
		slog.Error("sensor job failed", "job", name, "error", err)
	}

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := job(ctx); err != nil {
				slog.Error("sensor job failed", "job", name, "error", err)
			}
		}
	}
}

func nextHourlyArchiveRun(now time.Time) time.Time {
	run := time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), 10, 0, 0, time.UTC)
	if !now.Before(run) {
		run = run.Add(time.Hour)
	}
	return run
}

func nextDailyArchiveRun(now time.Time) time.Time {
	run := time.Date(now.Year(), now.Month(), now.Day(), 0, 5, 0, 0, time.UTC)
	if !now.Before(run) {
		run = run.Add(24 * time.Hour)
	}
	return run
}

func waitUntil(ctx context.Context, at time.Time) bool {
	timer := time.NewTimer(time.Until(at))
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}

func isValidSensorType(sensorType string) bool {
	return strings.TrimSpace(sensorType) != "" && sensorType == strings.ToLower(sensorType) &&
		containsSensorType(sensorType)
}

func containsSensorType(sensorType string) bool {
	for _, item := range allSensorTypes() {
		if item == sensorType {
			return true
		}
	}
	return false
}

func allSensorTypes() []string {
	return []string{
		model.SensorTemperature,
		model.SensorHumidity,
		model.SensorPH,
		model.SensorEC,
		model.SensorCO2,
		model.SensorTDS,
	}
}

func toHourlyPoints(archives []model.HourlyArchive) []HistoryPoint {
	points := make([]HistoryPoint, 0, len(archives))
	for _, archive := range archives {
		points = append(points, HistoryPoint{
			Time:         archive.BucketStart,
			AvgValue:     archive.AvgValue,
			MinValue:     archive.MinValue,
			MaxValue:     archive.MaxValue,
			ReadingCount: archive.ReadingCount,
		})
	}
	return points
}

func toDailyPoints(archives []model.DailyArchive) []HistoryPoint {
	points := make([]HistoryPoint, 0, len(archives))
	for _, archive := range archives {
		points = append(points, HistoryPoint{
			Time:         archive.BucketDate,
			AvgValue:     archive.AvgValue,
			MinValue:     archive.MinValue,
			MaxValue:     archive.MaxValue,
			ReadingCount: archive.ReadingCount,
		})
	}
	return points
}
