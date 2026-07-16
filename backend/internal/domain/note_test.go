package domain

import "testing"

func TestCategory_Valid(t *testing.T) {
	t.Parallel()

	valid := []Category{
		CategoryIdea,
		CategoryDate,
		CategoryGift,
		CategoryMovie,
		CategoryTravel,
		CategoryThought,
		CategoryOther,
	}
	for _, c := range valid {
		t.Run(string(c), func(t *testing.T) {
			t.Parallel()
			if !c.Valid() {
				t.Errorf("Category(%q).Valid() = false, want true", c)
			}
		})
	}

	invalid := []Category{
		"",
		"invalid",
		"IDEA",
		"ideas",
		Category("gift "),
	}
	for _, c := range invalid {
		t.Run("invalid_"+string(c), func(t *testing.T) {
			t.Parallel()
			if c.Valid() {
				t.Errorf("Category(%q).Valid() = true, want false", c)
			}
		})
	}
}
