// Package clerrors provides centralized error definitions for the pplx application.
// All sentinel errors and custom error types are defined here to provide a single
// source of truth and enable consistent error handling across the codebase.
package clerrors

import "errors"

// Configuration errors relate to config file loading, initialization, and validation.
var (
	// ErrNoConfigFound is returned when no config file is found in the config directory.
	ErrNoConfigFound = errors.New("no config file found in ~/.config/pplx/")

	// ErrPathIsDirectory is returned when a file path points to a directory instead of a file.
	ErrPathIsDirectory = errors.New("path is a directory, not a file")

	// ErrConfigFileExists is returned when attempting to create a config file that already exists.
	ErrConfigFileExists = errors.New("configuration file already exists")

	// ErrValidationFailed is returned when configuration validation fails.
	ErrValidationFailed = errors.New("validation failed")

	// ErrUnknownSection is returned when an unknown configuration section is referenced.
	ErrUnknownSection = errors.New("unknown section")
)

// Profile errors relate to profile management operations.
var (
	// ErrProfileNameEmpty is returned when attempting to create a profile with an empty name.
	ErrProfileNameEmpty = errors.New("profile name cannot be empty")

	// ErrProfileNameReserved is returned when attempting to create a profile with a reserved name.
	ErrProfileNameReserved = errors.New("cannot create profile with reserved name 'default'")

	// ErrProfileNotFound is returned when a requested profile does not exist.
	ErrProfileNotFound = errors.New("profile not found")

	// ErrProfileAlreadyExists is returned when attempting to create a profile that already exists.
	ErrProfileAlreadyExists = errors.New("profile already exists")

	// ErrDeleteDefaultProfile is returned when attempting to delete the default profile.
	ErrDeleteDefaultProfile = errors.New("cannot delete default profile")

	// ErrUpdateDefaultProfile is returned when attempting to update the default profile directly.
	ErrUpdateDefaultProfile = errors.New("cannot update default profile directly")

	// ErrImportReservedName is returned when attempting to import a profile with a reserved name.
	ErrImportReservedName = errors.New("cannot import profile with reserved name 'default'")
)

// Template errors relate to configuration template operations.
var (
	// ErrTemplateNotFound is returned when a requested template does not exist.
	ErrTemplateNotFound = errors.New("template not found")

	// ErrTemplateInvalid is returned when a template file cannot be parsed or is malformed.
	ErrTemplateInvalid = errors.New("template is invalid")
)

// Metadata errors relate to configuration option metadata operations.
var (
	// ErrOptionNotFound is returned when a configuration option is not found.
	ErrOptionNotFound = errors.New("option not found")

	// ErrUnsupportedFormat is returned when an unsupported output format is requested.
	ErrUnsupportedFormat = errors.New("unsupported format")
)

// Chat errors relate to chat parameter validation and API interaction.
var (
	// ErrInvalidSearchRecency is returned when an invalid search recency value is provided.
	ErrInvalidSearchRecency = errors.New("invalid search-recency value")

	// ErrConflictingResponseFormats is returned when both JSON schema and regex formats are specified.
	ErrConflictingResponseFormats = errors.New("cannot use both JSON schema and regex response formats")

	// ErrResponseFormatNotSupported is returned when response formats are used with non-sonar models.
	ErrResponseFormatNotSupported = errors.New(
		"response formats (JSON schema and regex) are only supported by sonar models")

	// ErrInvalidSearchMode is returned when an invalid search mode is provided.
	ErrInvalidSearchMode = errors.New("invalid search mode")

	// ErrInvalidSearchContextSize is returned when an invalid search context size is provided.
	ErrInvalidSearchContextSize = errors.New("invalid search context size")

	// ErrInvalidSearchAfterDate is returned when search-after-date has an invalid format.
	ErrInvalidSearchAfterDate = errors.New("invalid search-after-date format")

	// ErrInvalidSearchBeforeDate is returned when search-before-date has an invalid format.
	ErrInvalidSearchBeforeDate = errors.New("invalid search-before-date format")

	// ErrInvalidLastUpdatedAfter is returned when last-updated-after has an invalid format.
	ErrInvalidLastUpdatedAfter = errors.New("invalid last-updated-after format")

	// ErrInvalidLastUpdatedBefore is returned when last-updated-before has an invalid format.
	ErrInvalidLastUpdatedBefore = errors.New("invalid last-updated-before format")

	// ErrInvalidReasoningEffort is returned when an invalid reasoning effort level is provided.
	ErrInvalidReasoningEffort = errors.New("invalid reasoning effort")
)

// Command errors relate to CLI command execution and parameter validation.
var (
	// ErrInvalidLogLevel is returned when an invalid log level is specified.
	ErrInvalidLogLevel = errors.New("invalid log level")

	// ErrInvalidLogFormat is returned when an invalid log format is specified.
	ErrInvalidLogFormat = errors.New("invalid log format")

	// ErrFailedToReadInput is returned when reading user input from stdin fails.
	ErrFailedToReadInput = errors.New("failed to read input")

	// ErrFailedToReadAPIKey is returned when reading an API key from user input fails.
	ErrFailedToReadAPIKey = errors.New("failed to read API key")

	// ErrUnsupportedShell is returned when shell completion is requested for an unsupported shell.
	ErrUnsupportedShell = errors.New("unsupported shell")

	// ErrNoShellEnv is returned when the SHELL environment variable is not set.
	ErrNoShellEnv = errors.New("SHELL environment variable not set")
)
