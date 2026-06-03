package models

import "testing"

func TestRegisterInputNameHelpers(t *testing.T) {
	input := RegisterInput{
		Name:  "  Ada Lovelace  ",
		Email: "ada@example.com",
	}

	if !input.HasName() {
		t.Fatal("expected input to have a name")
	}
	firstName, lastName := input.NameParts()
	if firstName != "Ada" || lastName != "Lovelace" {
		t.Fatalf("unexpected name parts: first=%q last=%q", firstName, lastName)
	}
	if input.DisplayName() != "Ada Lovelace" {
		t.Fatalf("unexpected display name: %q", input.DisplayName())
	}
}

func TestRegisterInputNamePartsPreferExplicitFields(t *testing.T) {
	input := RegisterInput{
		Name:      "Ignored Name",
		FirstName: " Grace ",
		LastName:  " Hopper ",
	}

	firstName, lastName := input.NameParts()
	if firstName != "Grace" || lastName != "Hopper" {
		t.Fatalf("unexpected name parts: first=%q last=%q", firstName, lastName)
	}
}

func TestRegisterInputHasNameRejectsBlankValues(t *testing.T) {
	input := RegisterInput{
		Name:      " ",
		FirstName: "\t",
		LastName:  "\n",
	}
	if input.HasName() {
		t.Fatal("expected blank name values to be rejected")
	}
	if input.DisplayName() != "" {
		t.Fatalf("expected empty display name, got %q", input.DisplayName())
	}
}
