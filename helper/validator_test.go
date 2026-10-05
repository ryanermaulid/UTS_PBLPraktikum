package helper

import (
	"testing"
)

type sampleReq struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

func TestValidateStruct_OK(t *testing.T) {
	err := ValidateStruct(&sampleReq{Email: "a@b.com", Password: "password123"})
	if err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}

func TestValidateStruct_Required(t *testing.T) {
	err := ValidateStruct(&sampleReq{})
	ae, ok := AsAppError(err)
	if !ok {
		t.Fatalf("expected AppError, got %T (%v)", err, err)
	}
	if ae.Status != 422 {
		t.Fatalf("status = %d, want 422", ae.Status)
	}
	if len(ae.Fields) == 0 {
		t.Fatal("expected fields map populated")
	}
}

func TestValidateStruct_FieldNames(t *testing.T) {
	err := ValidateStruct(&sampleReq{Email: "bukan-email", Password: "x"})
	ae, ok := AsAppError(err)
	if !ok {
		t.Fatalf("expected AppError, got %v", err)
	}
	if _, ok := ae.Fields["email"]; !ok {
		t.Errorf("expected field email in errors, got %v", ae.Fields)
	}
	if _, ok := ae.Fields["password"]; !ok {
		t.Errorf("expected field password in errors, got %v", ae.Fields)
	}
}

func TestValidateStruct_Nil(t *testing.T) {
	if err := ValidateStruct(nil); err == nil {
		t.Fatal("expected error untuk input nil")
	}
}
