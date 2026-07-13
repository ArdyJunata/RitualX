package service

import (
	"testing"

	"github.com/google/uuid"
)

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

// Compile-time check: RoutineService satisfies expected constructor signature.
var _ = NewRoutineService
var _ = uuid.UUID{}
