package service

import (
	"testing"

	"github.com/google/uuid"
)

// ── CreateRoutine validators ──────────────────────────────────────────────────

func TestValidateCreateRoutine_Valid(t *testing.T) {
	req := CreateRoutineRequest{
		Title:       "Morning Run",
		PeriodType:  "daily",
		TargetCount: 1,
	}
	errs := validateCreateRoutine(req)
	if len(errs) != 0 {
		t.Fatalf("expected no errors, got %v", errs)
	}
}

func TestValidateCreateRoutine_MissingTitle(t *testing.T) {
	req := CreateRoutineRequest{
		Title:       "",
		PeriodType:  "daily",
		TargetCount: 1,
	}
	errs := validateCreateRoutine(req)
	if len(errs) == 0 {
		t.Fatal("expected validation error for empty title")
	}
	if errs[0].Field != "title" {
		t.Fatalf("expected field=title, got %s", errs[0].Field)
	}
}

func TestValidateCreateRoutine_InvalidPeriodType(t *testing.T) {
	req := CreateRoutineRequest{
		Title:       "Run",
		PeriodType:  "hourly",
		TargetCount: 1,
	}
	errs := validateCreateRoutine(req)
	if len(errs) == 0 {
		t.Fatal("expected validation error for invalid period_type")
	}
	if errs[0].Field != "period_type" {
		t.Fatalf("expected field=period_type, got %s", errs[0].Field)
	}
}

func TestValidateCreateRoutine_TargetCountZero(t *testing.T) {
	req := CreateRoutineRequest{
		Title:       "Run",
		PeriodType:  "weekly",
		TargetCount: 0,
	}
	errs := validateCreateRoutine(req)
	if len(errs) == 0 {
		t.Fatal("expected validation error for target_count=0")
	}
	if errs[0].Field != "target_count" {
		t.Fatalf("expected field=target_count, got %s", errs[0].Field)
	}
}

func TestValidateCreateRoutine_AllPeriodTypes(t *testing.T) {
	for _, pt := range []string{"daily", "weekly", "monthly"} {
		req := CreateRoutineRequest{Title: "T", PeriodType: pt, TargetCount: 1}
		errs := validateCreateRoutine(req)
		if len(errs) != 0 {
			t.Fatalf("period_type=%s: expected no errors, got %v", pt, errs)
		}
	}
}

func TestValidateCreateRoutine_TitleTooLong(t *testing.T) {
	req := CreateRoutineRequest{
		Title:       string(make([]byte, 101)),
		PeriodType:  "daily",
		TargetCount: 1,
	}
	errs := validateCreateRoutine(req)
	if len(errs) == 0 {
		t.Fatal("expected validation error for title > 100 chars")
	}
}

// ── UpdateRoutine validators ──────────────────────────────────────────────────

func TestValidateUpdateRoutine_AllNil_Valid(t *testing.T) {
	req := UpdateRoutineRequest{}
	errs := validateUpdateRoutine(req)
	if len(errs) != 0 {
		t.Fatalf("expected no errors for all-nil update, got %v", errs)
	}
}

func TestValidateUpdateRoutine_EmptyTitle(t *testing.T) {
	title := ""
	req := UpdateRoutineRequest{Title: &title}
	errs := validateUpdateRoutine(req)
	if len(errs) == 0 {
		t.Fatal("expected validation error for empty title")
	}
	if errs[0].Field != "title" {
		t.Fatalf("expected field=title, got %s", errs[0].Field)
	}
}

func TestValidateUpdateRoutine_TitleTooLong(t *testing.T) {
	title := string(make([]byte, 101))
	req := UpdateRoutineRequest{Title: &title}
	errs := validateUpdateRoutine(req)
	if len(errs) == 0 {
		t.Fatal("expected validation error for title > 100 chars")
	}
}

func TestValidateUpdateRoutine_InvalidPeriodType(t *testing.T) {
	pt := "hourly"
	req := UpdateRoutineRequest{PeriodType: &pt}
	errs := validateUpdateRoutine(req)
	if len(errs) == 0 {
		t.Fatal("expected validation error for invalid period_type")
	}
	if errs[0].Field != "period_type" {
		t.Fatalf("expected field=period_type, got %s", errs[0].Field)
	}
}

func TestValidateUpdateRoutine_ValidPeriodTypes(t *testing.T) {
	for _, pt := range []string{"daily", "weekly", "monthly"} {
		p := pt
		req := UpdateRoutineRequest{PeriodType: &p}
		errs := validateUpdateRoutine(req)
		if len(errs) != 0 {
			t.Fatalf("period_type=%s: expected no errors, got %v", pt, errs)
		}
	}
}

func TestValidateUpdateRoutine_TargetCountZero(t *testing.T) {
	tc := 0
	req := UpdateRoutineRequest{TargetCount: &tc}
	errs := validateUpdateRoutine(req)
	if len(errs) == 0 {
		t.Fatal("expected validation error for target_count=0")
	}
	if errs[0].Field != "target_count" {
		t.Fatalf("expected field=target_count, got %s", errs[0].Field)
	}
}

func TestValidateUpdateRoutine_TargetCountValid(t *testing.T) {
	tc := 3
	req := UpdateRoutineRequest{TargetCount: &tc}
	errs := validateUpdateRoutine(req)
	if len(errs) != 0 {
		t.Fatalf("expected no errors for valid target_count, got %v", errs)
	}
}

// ── ReorderRoutine validators ─────────────────────────────────────────────────

func TestValidateReorderRoutine_EmptyOrder(t *testing.T) {
	req := ReorderRoutineRequest{Order: []ReorderRoutineItem{}}
	errs := validateReorderRoutine(req)
	if len(errs) == 0 {
		t.Fatal("expected validation error for empty order")
	}
	if errs[0].Field != "order" {
		t.Fatalf("expected field=order, got %s", errs[0].Field)
	}
}

func TestValidateReorderRoutine_NegativeSortOrder(t *testing.T) {
	req := ReorderRoutineRequest{
		Order: []ReorderRoutineItem{
			{ID: uuid.New(), SortOrder: -1},
		},
	}
	errs := validateReorderRoutine(req)
	if len(errs) == 0 {
		t.Fatal("expected validation error for sort_order < 0")
	}
}

func TestValidateReorderRoutine_Valid(t *testing.T) {
	req := ReorderRoutineRequest{
		Order: []ReorderRoutineItem{
			{ID: uuid.New(), SortOrder: 0},
			{ID: uuid.New(), SortOrder: 1},
			{ID: uuid.New(), SortOrder: 2},
		},
	}
	errs := validateReorderRoutine(req)
	if len(errs) != 0 {
		t.Fatalf("expected no errors for valid reorder, got %v", errs)
	}
}

// Compile-time checks
var _ = NewRoutineService
var _ = uuid.UUID{}
