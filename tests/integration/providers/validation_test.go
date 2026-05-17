package providers_test

import (
	"mime/multipart"
	"strings"
	"testing"
	"time"

	validation "fiber-starter/app/Providers/Validation"
	validationContracts "fiber-starter/app/Providers/Validation/Contracts"
	"fiber-starter/configs"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidationFactory_MakeValidateAndExposeState(t *testing.T) {
	factory, err := validation.RegisterValidation(&configs.Config{})
	require.NoError(t, err)

	type payload struct {
		Email string `json:"email" validate:"required,email"`
		Name  string `json:"name" validate:"required,min=2"`
	}

	v := factory.Make(&payload{
		Email: "not-an-email",
		Name:  "A",
	}, nil, nil, nil)

	err = v.Validate()
	require.Error(t, err)

	var validationErrors validator.ValidationErrors
	require.ErrorAs(t, err, &validationErrors)
	assert.True(t, v.Fails())
	assert.Contains(t, v.Failed(), "email")
	assert.Contains(t, v.Failed(), "name")
	assert.NotEmpty(t, v.Errors().All())
	assert.True(t, v.Errors().Has("email"))
	assert.False(t, v.Errors().Has("missing"))
	assert.Nil(t, v.Validated())
}

func TestValidationFactory_ExtendSometimesAndValidated(t *testing.T) {
	factory, err := validation.RegisterValidation(&configs.Config{})
	require.NoError(t, err)

	err = factory.Extend("starts_with_x", func(fl validator.FieldLevel) bool {
		return strings.HasPrefix(fl.Field().String(), "x")
	}, "must start with x")
	require.NoError(t, err)

	v := factory.Make(map[string]any{
		"code": "x-123",
	}, map[string]string{
		"code": "starts_with_x",
	}, nil, nil).
		Sometimes("code", "min=2", func(data any) bool {
			_ = data
			return true
		})

	require.NoError(t, v.Validate())
	assert.Equal(t, "x-123", v.Validated()["code"])
	assert.Empty(t, v.Failed())

	afterCalls := 0
	after := factory.Make(map[string]any{
		"name": "fiber",
	}, nil, nil, nil).
		After(func(v validationContracts.Validator) {
			afterCalls++
			assert.NotNil(t, v)
		})
	require.NoError(t, after.Validate())
	assert.Equal(t, 1, afterCalls)
}

func TestValidationFactory_CustomRulesAndFileValidators(t *testing.T) {
	factory, err := validation.RegisterValidation(&configs.Config{})
	require.NoError(t, err)

	type testCase struct {
		name    string
		rule    string
		value   any
		wantErr bool
	}

	now := time.Now().UTC().Truncate(time.Second)
	later := now.Add(time.Hour)
	earlier := now.Add(-time.Hour)
	fh := &multipart.FileHeader{
		Filename: "avatar.png",
		Size:     1024,
		Header:   map[string][]string{"Content-Type": {"image/png"}},
	}

	cases := []testCase{
		{name: "phone valid", rule: "phone", value: "+12345678901"},
		{name: "phone invalid", rule: "phone", value: "1234", wantErr: true},
		{name: "mobile valid", rule: "mobile", value: "13800138000"},
		{name: "mobile invalid", rule: "mobile", value: "12345678901", wantErr: true},
		{name: "positive int valid", rule: "positive_int", value: 1},
		{name: "positive int invalid", rule: "positive_int", value: 0, wantErr: true},
		{name: "positive float valid", rule: "positive", value: 1.5},
		{name: "positive float invalid", rule: "positive", value: 0.0, wantErr: true},
		{name: "price valid", rule: "price", value: 99.99},
		{name: "price invalid", rule: "price", value: -1.0, wantErr: true},
		{name: "array slice valid", rule: "array", value: []string{"one", "two"}},
		{name: "array map allowed keys", rule: "array=one two", value: map[string]any{"one": 1, "two": 2}},
		{name: "array invalid type", rule: "array", value: "nope", wantErr: true},
		{name: "list slice valid", rule: "list", value: []string{"a", "b"}},
		{name: "list map valid", rule: "list", value: map[string]any{"0": "a", "1": "b"}},
		{name: "list invalid", rule: "list", value: map[string]any{"0": "a", "2": "b"}, wantErr: true},
		{name: "required array keys valid", rule: "required_array_keys=foo bar", value: map[string]any{"foo": 1, "bar": 2}},
		{name: "required array keys invalid", rule: "required_array_keys=foo bar", value: map[string]any{"foo": 1}, wantErr: true},
		{name: "alpha dash valid", rule: "alpha_dash", value: "alpha_dash-123"},
		{name: "alpha dash invalid", rule: "alpha_dash", value: "bad value!", wantErr: true},
		{name: "accepted bool", rule: "accepted", value: true},
		{name: "accepted string", rule: "accepted", value: "yes"},
		{name: "accepted int", rule: "accepted", value: 1},
		{name: "accepted invalid", rule: "accepted", value: false, wantErr: true},
		{name: "declined bool", rule: "declined", value: false},
		{name: "declined string", rule: "declined", value: "no"},
		{name: "declined invalid", rule: "declined", value: true, wantErr: true},
		{name: "date format valid", rule: "date_format=2006-01-02", value: "2026-05-06"},
		{name: "date format invalid", rule: "date_format=2006-01-02", value: "06/05/2026", wantErr: true},
		{name: "date equals valid", rule: "date_equals=2026-05-06", value: "2026-05-06"},
		{name: "date equals invalid", rule: "date_equals=2026-05-06", value: "2026-05-07", wantErr: true},
		{name: "after valid", rule: "after=2026-05-06", value: "2026-05-07"},
		{name: "after invalid", rule: "after=2026-05-06", value: "2026-05-05", wantErr: true},
		{name: "after or equal valid", rule: "after_or_equal=2026-05-06", value: "2026-05-06"},
		{name: "before valid", rule: "before=2026-05-06", value: "2026-05-05"},
		{name: "before or equal valid", rule: "before_or_equal=2026-05-06", value: "2026-05-06"},
		{name: "date valid", rule: "date", value: "2026-05-06"},
		{name: "date pointer valid", rule: "date", value: &now},
		{name: "date invalid", rule: "date", value: "not-a-date", wantErr: true},
		{name: "after now", rule: "after=now", value: later},
		{name: "before now", rule: "before=now", value: earlier},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			v := factory.Make(map[string]any{"value": tc.value}, map[string]string{"value": tc.rule}, nil, nil)
			err := v.Validate()
			if tc.wantErr {
				require.Error(t, err)
				assert.True(t, v.Fails())
				return
			}
			require.NoError(t, err)
			assert.False(t, v.Fails())
		})
	}

	t.Run("uploaded file valid", func(t *testing.T) {
		type payload struct {
			File *multipart.FileHeader `json:"file" validate:"uploaded_file"`
		}

		v := factory.Make(&payload{File: fh}, nil, nil, nil)
		require.NoError(t, v.Validate())
	})

	t.Run("uploaded file invalid", func(t *testing.T) {
		type payload struct {
			File *multipart.FileHeader `json:"file" validate:"uploaded_file"`
		}

		v := factory.Make(&payload{}, nil, nil, nil)
		require.Error(t, v.Validate())
	})

	t.Run("max bytes invalid", func(t *testing.T) {
		type payload struct {
			File *multipart.FileHeader `json:"file" validate:"max_bytes=1"`
		}

		v := factory.Make(&payload{File: fh}, nil, nil, nil)
		require.Error(t, v.Validate())
	})

	t.Run("mime types exact", func(t *testing.T) {
		type payload struct {
			File *multipart.FileHeader `json:"file" validate:"mime_types=image/png"`
		}

		v := factory.Make(&payload{File: fh}, nil, nil, nil)
		require.NoError(t, v.Validate())
	})

	t.Run("mime types wildcard", func(t *testing.T) {
		type payload struct {
			File *multipart.FileHeader `json:"file" validate:"mime_types=image/*"`
		}

		v := factory.Make(&payload{File: fh}, nil, nil, nil)
		require.NoError(t, v.Validate())
	})

	t.Run("mime types invalid", func(t *testing.T) {
		type payload struct {
			File *multipart.FileHeader `json:"file" validate:"mime_types=text/plain"`
		}

		v := factory.Make(&payload{File: fh}, nil, nil, nil)
		require.Error(t, v.Validate())
	})

	t.Run("mime types missing header", func(t *testing.T) {
		type payload struct {
			File *multipart.FileHeader `json:"file" validate:"mime_types=image/png"`
		}

		v := factory.Make(&payload{File: &multipart.FileHeader{Filename: "empty"}}, nil, nil, nil)
		require.Error(t, v.Validate())
	})
}
