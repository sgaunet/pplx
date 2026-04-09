package cmd

import (
	"fmt"
	"io"
	"os"
	"strconv"

	huh "charm.land/huh/v2"
	"github.com/sgaunet/pplx/pkg/config"
	"github.com/spf13/cobra"
)

// ProfileEditor handles interactive profile editing.
type ProfileEditor struct {
	input      io.Reader
	output     io.Writer
	accessible bool
	profile    *config.Profile
	data       *config.ConfigData
}

// NewProfileEditor creates a new profile editor.
func NewProfileEditor(profile *config.Profile, data *config.ConfigData) *ProfileEditor {
	return &ProfileEditor{
		input:      os.Stdin,
		output:     os.Stdout,
		accessible: os.Getenv("ACCESSIBLE") != "",
		profile:    profile,
		data:       data,
	}
}

// Run starts the interactive editing.
//
// The editor presents the current model, temperature, and stream settings and
// allows the user to change them. Leaving a field empty keeps the existing setting.
func (pe *ProfileEditor) Run() error {
	// Prepare current values for display
	currentModel := ""
	if pe.profile.Defaults.Model != nil {
		currentModel = *pe.profile.Defaults.Model
	}

	currentTemp := ""
	if pe.profile.Defaults.Temperature != nil {
		currentTemp = strconv.FormatFloat(*pe.profile.Defaults.Temperature, 'f', -1, 64)
	}

	stream := false
	if pe.profile.Output.Stream != nil {
		stream = *pe.profile.Output.Stream
	}

	var model, tempStr string

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Model").
				Description(fmt.Sprintf("Current: %s. Leave empty to keep.", currentModel)).
				Placeholder(currentModel).
				Value(&model),
			huh.NewInput().
				Title("Temperature").
				Description(fmt.Sprintf("Current: %s. Leave empty to keep.", currentTemp)).
				Placeholder(currentTemp).
				Validate(validateOptionalFloat(minTemperature, maxTemperature)).
				Value(&tempStr),
			huh.NewConfirm().
				Title("Enable Streaming?").
				Affirmative("Yes").
				Negative("No").
				Value(&stream),
		).Title(fmt.Sprintf("Editing profile '%s'", pe.profile.Name)),
	).
		WithAccessible(pe.accessible).
		WithInput(pe.input).
		WithOutput(pe.output).
		WithTheme(huh.ThemeFunc(huh.ThemeCatppuccin))

	if err := form.Run(); err != nil {
		return fmt.Errorf("profile editor form failed: %w", err)
	}

	// Apply non-empty values back to profile
	if model != "" {
		pe.profile.Defaults.Model = &model
	}
	if tempStr != "" {
		parsed, err := strconv.ParseFloat(tempStr, 64)
		if err != nil {
			return fmt.Errorf("invalid temperature %q: %w", tempStr, err)
		}
		pe.profile.Defaults.Temperature = &parsed
	}
	pe.profile.Output.Stream = &stream

	return nil
}

// configProfileEditCmd opens an interactive editor for a profile.
var configProfileEditCmd = &cobra.Command{
	Use:   "edit <name>",
	Short: "Interactively edit a profile",
	Long: `Open an interactive editor to modify a profile's settings.

Press Enter without typing to keep the current value.

Examples:
  pplx config profile edit research
  pplx config profile edit myprofile`,
	Args: cobra.ExactArgs(1),
	RunE: func(_ *cobra.Command, args []string) error {
		name := args[0]

		data, err := loadConfigData(configFilePath)
		if err != nil {
			return err
		}

		pm := config.NewProfileManager(data)
		profile, err := pm.LoadProfile(name)
		if err != nil {
			return fmt.Errorf("failed to load profile %q: %w", name, err)
		}

		editor := NewProfileEditor(profile, data)
		if err := editor.Run(); err != nil {
			return fmt.Errorf("profile edit failed: %w", err)
		}

		// Persist the updated profile back into the config data map.
		if err := pm.UpdateProfile(name, profile); err != nil {
			return fmt.Errorf("failed to update profile %q: %w", name, err)
		}

		if err := saveConfigData(data); err != nil {
			return err
		}

		fmt.Printf("Profile '%s' updated successfully\n", name)
		return nil
	},
}
