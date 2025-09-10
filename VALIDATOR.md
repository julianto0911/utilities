# Validator Documentation

The Validator package provides a comprehensive validation system for Go structs using struct tags. It supports validation for multiple data types including strings, integers, and floats with various validation rules.

## Overview

The Validator uses reflection to inspect struct fields and apply validation rules defined in struct tags. It supports validation for the following data types:
- `string`
- `int`
- `int32`
- `int64`
- `float64`

## Interface

```go
type Validator interface {
    Validate(item any) error
}
```

## Usage

### Basic Usage

```go
package main

import (
    "fmt"
    "your-package/utilities"
)

type User struct {
    Name     string `json:"name" validate:"required;min=3;max=50"`
    Email    string `json:"email" validate:"required;min=5;max=100"`
    Age      int    `json:"age" validate:"required;min=18;max=120"`
    Score    float64 `json:"score" validate:"min=0;max=100"`
}

func main() {
    validator := utilities.NewValidator()
    
    user := User{
        Name:  "John",
        Email: "john@example.com",
        Age:   25,
        Score: 85.5,
    }
    
    if err := validator.Validate(user); err != nil {
        fmt.Printf("Validation error: %v\n", err)
    } else {
        fmt.Println("Validation passed!")
    }
}
```

## Validation Rules

### String Validation Rules

| Rule | Syntax | Description | Example |
|------|--------|-------------|---------|
| **required** | `validate:"required"` | Field must have a non-empty value | `validate:"required"` |
| **min** | `validate:"min=[length]"` | String length must be at least the specified value | `validate:"min=3"` |
| **max** | `validate:"max=[length]"` | String length must be at most the specified value | `validate:"max=50"` |
| **length** | `validate:"length=[length]"` | String length must be exactly the specified value | `validate:"length=10"` |
| **range** | `validate:"range[val1,val2]"` | String value must be one of the specified values | `validate:"range[admin,user,guest]"` |
| **optx** | `validate:"optx=[length]"` | If field has a value, length must be at least the specified value | `validate:"optx=5"` |
| **opty** | `validate:"opty=[length]"` | If field has a value, length must be at most the specified value | `validate:"opty=20"` |

### Numeric Validation Rules (int, int32, int64, float64)

| Rule | Syntax | Description | Example |
|------|--------|-------------|---------|
| **required** | `validate:"required"` | Field must not be zero | `validate:"required"` |
| **min** | `validate:"min=[value]"` | Value must be at least the specified number | `validate:"min=18"` |
| **max** | `validate:"max=[value]"` | Value must be at most the specified number | `validate:"max=120"` |
| **range** | `validate:"range[val1,val2]"` | Value must be one of the specified values | `validate:"range[1,2,3,4,5]"` |
| **optx** | `validate:"optx=[value]"` | If field has a non-zero value, it must be at least the specified value | `validate:"optx=10"` |
| **opty** | `validate:"opty=[value]"` | If field has a non-zero value, it must be at most the specified value | `validate:"opty=100"` |

## Examples

### String Validation Examples

```go
type UserProfile struct {
    Username    string `json:"username" validate:"required;min=3;max=20"`
    Email       string `json:"email" validate:"required;min=5;max=100"`
    Bio         string `json:"bio" validate:"optx=10;opty=500"`        // Optional, but if provided, 10-500 chars
    Status      string `json:"status" validate:"range[active,inactive,pending]"`
    Phone       string `json:"phone" validate:"length=10"`             // Exactly 10 characters
    Description string `json:"description" validate:"max=1000"`        // Max 1000 characters
}
```

### Numeric Validation Examples

```go
type Product struct {
    ID          int     `json:"id" validate:"required;min=1"`
    Price       float64 `json:"price" validate:"required;min=0.01;max=9999.99"`
    Quantity    int32   `json:"quantity" validate:"required;min=0;max=10000"`
    Rating      float64 `json:"rating" validate:"min=0;max=5"`         // Optional rating
    CategoryID  int64   `json:"category_id" validate:"range[1,2,3,4,5]"` // Must be one of these IDs
    Discount    float64 `json:"discount" validate:"optx=0.01;opty=0.5"`  // Optional, but if provided, 0.01-0.5
}
```

### Multiple Rules Example

```go
type RegistrationForm struct {
    FirstName   string  `json:"first_name" validate:"required;min=2;max=50"`
    LastName    string  `json:"last_name" validate:"required;min=2;max=50"`
    Email       string  `json:"email" validate:"required;min=5;max=100"`
    Password    string  `json:"password" validate:"required;min=8;max=128"`
    Age         int     `json:"age" validate:"required;min=18;max=120"`
    Country     string  `json:"country" validate:"range[US,CA,UK,AU,DE,FR]"`
    ZipCode     string  `json:"zip_code" validate:"length=5"`           // US ZIP code
    Phone       string  `json:"phone" validate:"optx=10;opty=15"`      // Optional phone
    Score       float64 `json:"score" validate:"min=0;max=100"`        // Optional score
}
```

## Error Messages

The validator provides descriptive error messages for validation failures:

### String Validation Errors
- `"field {name} must be filled"` - Required field is empty
- `"field {name} must have at least {n} character(s)"` - Min length not met
- `"total characters for field {name} must be less or same than {n} character(s)"` - Max length exceeded
- `"field {name} must have {n} character(s)"` - Exact length not met
- `"field {name} value must in [{values}]"` - Value not in allowed range
- `"when have value, field {name} must have at least {n} character(s)"` - Optional min not met
- `"when have value, total characters for field {name} must be less or same than {n} character(s)"` - Optional max exceeded

### Numeric Validation Errors
- `"field {name} must not zero"` - Required field is zero
- `"field {name} must not less than {n}"` - Min value not met
- `"field {name} must not greater than {n}"` - Max value exceeded
- `"field {name} value must in [{values}]"` - Value not in allowed range
- `"when have value, field {name} must have at least {n} character(s)"` - Optional min not met
- `"when have value, total characters for field {name} must be less or same than {n} character(s)"` - Optional max exceeded

## Testing Support

The package includes a `MockValidator` for testing purposes:

```go
type MockValidator struct {
    mock.Mock
}

func (m *MockValidator) Validate(item any) error {
    args := m.Called(item)
    return args.Error(0)
}
```

### Testing Example

```go
func TestUserService(t *testing.T) {
    mockValidator := &MockValidator{}
    
    // Setup mock expectations
    mockValidator.On("Validate", mock.Anything).Return(nil)
    
    // Use mock in your tests
    userService := NewUserService(mockValidator)
    
    // Your test logic here
    
    mockValidator.AssertExpectations(t)
}
```

## Best Practices

1. **Use descriptive field names** in your struct tags for better error messages
2. **Combine rules with semicolons** to apply multiple validations: `validate:"required;min=3;max=50"`
3. **Use appropriate data types** for your validation needs
4. **Test your validation rules** thoroughly with both valid and invalid data
5. **Consider using optional rules** (`optx`, `opty`) for fields that are not required but have constraints when provided

## Limitations

1. **Struct tags only** - Validation rules must be defined in struct tags
2. **No nested struct validation** - Only validates top-level struct fields
3. **No custom validation functions** - Limited to predefined validation rules
4. **No cross-field validation** - Cannot validate relationships between fields

## Dependencies

- `github.com/stretchr/testify/mock` - For testing support
- Standard Go packages: `errors`, `fmt`, `reflect`, `strconv`, `strings`

## Performance Considerations

- Uses reflection which has some performance overhead
- Consider caching validation results for frequently validated structs
- For high-performance scenarios, consider implementing custom validation logic
