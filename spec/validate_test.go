package spec

import (
	"testing"

	assert "github.com/stretchr/testify/require"
)

func TestValidator_Simple(t *testing.T) {
	schema := Schema{
		Type: "string",
	}
	v, err := GetValidatorForOpenAPI3Schema(&schema, nil, false)
	assert.NoError(t, err)
	assert.NoError(t, v.Validate("hello"))
	assert.Error(t, v.Validate(nil))
	assert.Error(t, v.Validate(123))
}

func TestValidator_Nullable(t *testing.T) {
	schema := Schema{
		Type:     "string",
		Nullable: true,
	}
	v, err := GetValidatorForOpenAPI3Schema(&schema, nil, false)
	assert.NoError(t, err)
	assert.NoError(t, v.Validate("hello"))
	assert.NoError(t, v.Validate(nil))
	assert.Error(t, v.Validate(123))
}

func TestValidator_Reference(t *testing.T) {
	fooSchema := Schema{
		Type: "string",
	}
	components := Components{
		Schemas: map[string]*Schema{
			"foo": &fooSchema,
		},
	}
	componentsForValidation := GetComponentsForValidation(&components)
	fooRefSchema := Schema{
		Ref: "#/components/schemas/foo",
	}
	v, err := GetValidatorForOpenAPI3Schema(&fooRefSchema, componentsForValidation, false)
	assert.NoError(t, err)
	assert.NoError(t, v.Validate("hello"))
	assert.Error(t, v.Validate(123))
}

func TestValidator_StrictAdditionalProperties_RejectsUnknownNested(t *testing.T) {
	// A nested object without explicit additionalProperties should
	// reject unknown properties in strict mode.
	schema := Schema{
		AdditionalPropertiesAllowed: false,
		Type:                        TypeObject,
		Properties: map[string]*Schema{
			"phases": {
				Type: TypeArray,
				Items: &Schema{
					Type: TypeObject,
					// AdditionalPropertiesAllowed defaults to true when
					// not specified in the spec, but strict mode should
					// override this for objects.
					AdditionalPropertiesAllowed: true,
					Properties: map[string]*Schema{
						"currency": {Type: TypeString},
					},
				},
			},
		},
	}
	v, err := GetValidatorForOpenAPI3Schema(&schema, nil, true)
	assert.NoError(t, err)

	// Valid data should pass
	goodData := map[string]interface{}{
		"phases": []interface{}{
			map[string]interface{}{
				"currency": "usd",
			},
		},
	}
	assert.NoError(t, v.Validate(goodData))

	// Unknown nested property should be rejected
	badData := map[string]interface{}{
		"phases": []interface{}{
			map[string]interface{}{
				"currency":   "usd",
				"iterations": 1,
			},
		},
	}
	assert.Error(t, v.Validate(badData))
}

func TestValidator_StrictAdditionalProperties_AllowsMetadata(t *testing.T) {
	// Objects with explicit additionalProperties schema (like metadata)
	// should still allow additional properties even in strict mode.
	schema := Schema{
		AdditionalPropertiesAllowed: false,
		Type:                        TypeObject,
		Properties: map[string]*Schema{
			"metadata": {
				Type:                        TypeObject,
				AdditionalPropertiesAllowed: true,
				AdditionalProperties: &Schema{
					Type: TypeString,
				},
			},
		},
	}
	v, err := GetValidatorForOpenAPI3Schema(&schema, nil, true)
	assert.NoError(t, err)

	// Arbitrary keys in metadata should be allowed
	goodData := map[string]interface{}{
		"metadata": map[string]interface{}{
			"custom_key": "custom_value",
		},
	}
	assert.NoError(t, v.Validate(goodData))
}

func TestValidator_NonStrictMode_AllowsUnknownNested(t *testing.T) {
	// In non-strict mode (the default), nested objects without explicit
	// additionalProperties should allow unknown properties (old behavior).
	schema := Schema{
		Type:                        TypeObject,
		AdditionalPropertiesAllowed: true,
		Properties: map[string]*Schema{
			"name": {Type: TypeString},
		},
	}
	v, err := GetValidatorForOpenAPI3Schema(&schema, nil, false)
	assert.NoError(t, err)

	data := map[string]interface{}{
		"name":    "test",
		"unknown": "value",
	}
	assert.NoError(t, v.Validate(data))
}
