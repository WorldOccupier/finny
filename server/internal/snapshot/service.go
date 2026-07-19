package snapshot

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/WorldOccupier/finny/server/internal/database"
	"github.com/WorldOccupier/finny/server/internal/domain"
)

type SnapshotInput struct {
	Assets         []domain.Asset
	FXRate         domain.Decimal
	SpendingLimits []domain.SpendingLimit
	Income         domain.IncomeTotals
}

type Service struct {
	store database.Store
	now   func() time.Time
}

type ValidationError struct{ Err error }

func (e *ValidationError) Error() string { return e.Err.Error() }
func (e *ValidationError) Unwrap() error { return e.Err }

func IsValidationError(err error) bool {
	var validationErr *ValidationError
	return errors.As(err, &validationErr)
}

type PreparedSave struct {
	Dashboard domain.Dashboard
	Save      database.DashboardSnapshotSave
}

func NewService(store database.Store, now func() time.Time) *Service {
	if now == nil {
		now = time.Now
	}
	return &Service{store: store, now: now}
}

func (s *Service) Save(ctx context.Context, input SnapshotInput) (domain.Dashboard, error) {
	prepared, err := s.Prepare(ctx, input)
	if err != nil {
		return domain.Dashboard{}, err
	}
	_, err = s.store.SaveDashboardSnapshot(ctx, prepared.Save)
	if err != nil {
		return domain.Dashboard{}, fmt.Errorf("save dashboard snapshot: %w", err)
	}
	result, err := s.store.LoadDashboard(ctx)
	if err != nil {
		return domain.Dashboard{}, fmt.Errorf("load committed dashboard: %w", err)
	}
	return result, nil
}

func (s *Service) Prepare(ctx context.Context, input SnapshotInput) (PreparedSave, error) {
	current, err := s.store.LoadDashboard(ctx)
	if err != nil {
		return PreparedSave{}, fmt.Errorf("load current dashboard: %w", err)
	}
	err = validateInput(input)
	if err != nil {
		return PreparedSave{}, &ValidationError{Err: err}
	}
	assets, err := resolveAssets(current, input.Assets)
	if err != nil {
		return PreparedSave{}, &ValidationError{Err: err}
	}
	totals, err := CalculateTotals(assets, input.FXRate)
	if err != nil {
		return PreparedSave{}, &ValidationError{Err: err}
	}
	committedAt := s.now().In(domain.LondonLocation())
	if committedAt.IsZero() {
		return PreparedSave{}, &ValidationError{Err: fmt.Errorf("snapshot commit time must be set")}
	}
	for _, previous := range current.History {
		if previous.CommittedAt.Equal(committedAt) {
			return PreparedSave{}, &ValidationError{Err: fmt.Errorf("snapshot commit time must be unique")}
		}
	}

	snapshot := domain.Snapshot{
		ID:          nextSnapshotID(current.History),
		CommittedAt: committedAt,
		FXRate:      input.FXRate,
		Assets:      assets,
		Totals:      totals,
	}
	save := database.DashboardSnapshotSave{
		Assets:           assets,
		Snapshot:         snapshot,
		SpendingLimits:   input.SpendingLimits,
		Income:           input.Income,
		CurrentFXRate:    input.FXRate,
		ExpectedRevision: current.Revision,
		Revision:         current.Revision + 1,
	}
	committed := current
	committed.Revision = save.Revision
	committed.Assets = assets
	committed.CurrentFXRate = input.FXRate
	committed.CurrentTotals = totals
	committed.SpendingLimits = input.SpendingLimits
	committed.Income = input.Income
	committed.History = append(append([]domain.Snapshot(nil), current.History...), snapshot)
	return PreparedSave{Dashboard: committed, Save: save}, nil
}

func validateInput(input SnapshotInput) error {
	if input.FXRate.IsZero() || input.FXRate.IsNegative() {
		return fmt.Errorf("snapshot FX rate must be greater than zero")
	}
	if err := input.Income.Validate(); err != nil {
		return fmt.Errorf("validate income: %w", err)
	}
	for _, limit := range input.SpendingLimits {
		if err := limit.Validate(); err != nil {
			return fmt.Errorf("validate spending limit: %w", err)
		}
	}
	seenIDs := make(map[int64]struct{}, len(input.Assets))
	seenNames := make(map[string]struct{}, len(input.Assets))
	for _, asset := range input.Assets {
		if asset.ID < 0 {
			return fmt.Errorf("asset ID must not be negative")
		}
		if _, exists := seenIDs[asset.ID]; exists {
			return fmt.Errorf("duplicate asset ID %d", asset.ID)
		}
		seenIDs[asset.ID] = struct{}{}
		name := strings.TrimSpace(asset.Name)
		if _, exists := seenNames[name]; exists {
			return fmt.Errorf("duplicate asset name %q", name)
		}
		seenNames[name] = struct{}{}
		if err := asset.Validate(); err != nil {
			return fmt.Errorf("validate asset %d: %w", asset.ID, err)
		}
		seenTypes := make(map[domain.ValueType]struct{}, len(asset.Values))
		for _, value := range asset.Values {
			if _, exists := seenTypes[value.Type]; exists {
				return fmt.Errorf("duplicate value type %q for asset %d", value.Type, asset.ID)
			}
			seenTypes[value.Type] = struct{}{}
		}
		seenSelectedTypes := make(map[domain.ValueType]struct{}, len(asset.ValueTypes))
		for _, valueType := range asset.ValueTypes {
			if !valueType.Valid() {
				return fmt.Errorf("unsupported selected value type %q for asset %d", valueType, asset.ID)
			}
			if _, exists := seenSelectedTypes[valueType]; exists {
				return fmt.Errorf("duplicate selected value type %q for asset %d", valueType, asset.ID)
			}
			seenSelectedTypes[valueType] = struct{}{}
		}
	}
	return nil
}

func resolveAssets(current domain.Dashboard, submitted []domain.Asset) ([]domain.Asset, error) {
	previous := make(map[int64]domain.Asset, len(current.Assets))
	for _, asset := range current.Assets {
		previous[asset.ID] = asset
	}
	resolved := make([]domain.Asset, 0, len(submitted))
	for _, asset := range submitted {
		values := make(map[domain.ValueType]domain.AssetValue, len(asset.Values))
		for _, value := range asset.Values {
			values[value.Type] = value
		}
		old, exists := previous[asset.ID]
		selected := make(map[domain.ValueType]struct{}, len(asset.ValueTypes))
		if len(asset.ValueTypes) == 0 {
			for valueType := range values {
				selected[valueType] = struct{}{}
			}
		} else {
			for _, valueType := range asset.ValueTypes {
				selected[valueType] = struct{}{}
			}
		}
		if exists {
			for _, value := range old.Values {
				if _, chosen := selected[value.Type]; chosen {
					if _, present := values[value.Type]; !present {
						values[value.Type] = value
					}
				}
			}
		}
		if len(values) == 0 {
			return nil, fmt.Errorf("asset %d must have at least one value type", asset.ID)
		}
		asset.Values = make([]domain.AssetValue, 0, len(values))
		for _, valueType := range []domain.ValueType{domain.UK_GBP, domain.INDIA_INR} {
			if _, chosen := selected[valueType]; chosen {
				value, present := values[valueType]
				if !present {
					return nil, fmt.Errorf("asset %d is missing selected value type %q", asset.ID, valueType)
				}
				asset.Values = append(asset.Values, value)
			}
		}
		asset.ValueTypes = make([]domain.ValueType, 0, len(asset.Values))
		for _, value := range asset.Values {
			asset.ValueTypes = append(asset.ValueTypes, value.Type)
		}
		resolved = append(resolved, asset)
	}
	return resolved, nil
}

func nextSnapshotID(history []domain.Snapshot) int64 {
	var next int64 = 1
	for _, snapshot := range history {
		if snapshot.ID >= next {
			next = snapshot.ID + 1
		}
	}
	return next
}
