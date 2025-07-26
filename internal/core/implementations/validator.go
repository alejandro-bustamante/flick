package implementations

import (
	"fmt"
	"strings"

	models "github.com/alejandro-bustamante/flick/internal/models"
)

type Validator struct {
	minTitleLength int
}

func NewValidator() *Validator {
	return &Validator{
		minTitleLength: 1,
	}
}

func (v *Validator) Validate(info *models.MediaInfo) error {
	if err := v.ValidateTitle(info.Title); err != nil {
		return err
	}

	if !info.IsMovie && info.Season == 0 {
		return fmt.Errorf("TV show must have season number")
	}

	return nil
}

func (v *Validator) ValidateTitle(title string) error {
	if len(strings.TrimSpace(title)) < v.minTitleLength {
		return fmt.Errorf("title too short: '%s'", title)
	}
	return nil
}
