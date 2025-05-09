package utilities

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

/*
validate string rules
required ("validate:required") , must have value
min ("validate:min=[length]") , value typed length must be at least of value
max ("validate:max=[length]") , value typed length must same or less length than value
length ("validate:length=[length]") , value typed length must be in exact length value
range ("validate:range[val1,val2]"), value typed must be in range of value declared
optx ("validate:optx=[length]") , if have value, length must be at least of value
opty ("validate:opty=[length]") , if have value, length must be same or less than value
*/

func NewValidator() Validator {
	return validator{}
}

type Validator interface {
	Validate(item any) error
}

type validator struct{}

func (c validator) Validate(item any) error {
	val := reflect.ValueOf(item)
	for i := range val.NumField() {
		var err error
		fieldType := val.Type().Field(i)
		rule := fieldType.Tag.Get("validate")
		rules := strings.Split(rule, ";")

		field := val.Field(i)
		value := field.Interface()

		//get name of variable
		name := fieldType.Tag.Get("json")
		//get value of variable
		//check type of validation
		switch field.Type().Kind() {
		case reflect.String:
			err = c.validateString(rules, name, value)
		case reflect.Int:
			err = c.validateInt(rules, name, value)
		case reflect.Int32:
			err = c.validateInt32(rules, name, value)
		case reflect.Int64:
			err = c.validateInt64(rules, name, value)
		case reflect.Float64:
			err = c.validateFloat64(rules, name, value)
		}

		if err != nil {
			return err
		}
	}

	return nil
}

func (c validator) validateFloat64(rules []string, name string, v any) error {
	value, _ := v.(float64)
	for _, rule := range rules {
		if err := float64MinValue(rule, name, value); err != nil {
			return err
		}
		if err := float64MaxValue(rule, name, value); err != nil {
			return err
		}
		if err := float64Range(rule, name, value); err != nil {
			return err
		}
		if err := float64Required(rule, name, value); err != nil {
			return err
		}
		if err := float64OptX(rule, name, value); err != nil {
			return err
		}
		if err := float64OptY(rule, name, value); err != nil {
			return err
		}
	}

	return nil
}

func (c validator) validateInt(rules []string, name string, v any) error {
	value, _ := v.(int)
	for _, rule := range rules {
		if err := intMinValue(rule, name, value); err != nil {
			return err
		}
		if err := intMaxValue(rule, name, value); err != nil {
			return err
		}
		if err := intRange(rule, name, value); err != nil {
			return err
		}
		if err := intRequired(rule, name, value); err != nil {
			return err
		}
		if err := intOptX(rule, name, value); err != nil {
			return err
		}
		if err := intOptY(rule, name, value); err != nil {
			return err
		}
	}

	return nil
}

func (c validator) validateInt32(rules []string, name string, v any) error {
	value, _ := v.(int32)
	for _, rule := range rules {
		if err := int32MinValue(rule, name, value); err != nil {
			return err
		}
		if err := int32MaxValue(rule, name, value); err != nil {
			return err
		}
		if err := int32Range(rule, name, value); err != nil {
			return err
		}
		if err := int32Required(rule, name, value); err != nil {
			return err
		}
		if err := int32OptX(rule, name, value); err != nil {
			return err
		}
		if err := int32OptY(rule, name, value); err != nil {
			return err
		}
	}

	return nil
}

func (c validator) validateInt64(rules []string, name string, v any) error {
	value, _ := v.(int64)
	for _, rule := range rules {
		if err := int64MinValue(rule, name, value); err != nil {
			return err
		}
		if err := int64MaxValue(rule, name, value); err != nil {
			return err
		}
		if err := int64Range(rule, name, value); err != nil {
			return err
		}
		if err := int64Required(rule, name, value); err != nil {
			return err
		}
		if err := int64OptX(rule, name, value); err != nil {
			return err
		}
		if err := int64OptY(rule, name, value); err != nil {
			return err
		}
	}

	return nil
}

func (c validator) validateString(rules []string, name string, v any) error {
	value := v.(string)
	for _, rule := range rules {
		if err := strRequired(rule, name, value); err != nil {
			return err
		}
		if err := strMinLength(rule, name, value); err != nil {
			return err
		}
		if err := strMaxLength(rule, name, value); err != nil {
			return err
		}
		if err := strLength(rule, name, value); err != nil {
			return err
		}
		if err := strRange(rule, name, value); err != nil {
			return err
		}
		if err := strOptX(rule, name, value); err != nil {
			return err
		}
		if err := strOptY(rule, name, value); err != nil {
			return err
		}
	}

	return nil
}

func strOptX(rule, name, value string) error {
	if value == "" {
		return nil
	}

	if !strings.Contains(rule, "optx") {
		return nil
	}

	r := strings.Split(rule, "=")
	if len(r) < 2 {
		return fmt.Errorf("length invalid rule definition, name : " + name)
	}

	limit, err := strconv.Atoi(strings.TrimSpace(r[1]))
	if err != nil {
		return fmt.Errorf("invalid rule:(%v) %w", name, err)
	}

	if len(value) < limit {
		return fmt.Errorf("when have value, field %v must have at least %v character(s)", name, limit)
	}

	return nil
}

func strOptY(rule, name, value string) error {
	if value == "" {
		return nil
	}

	if !strings.Contains(rule, "opty") {
		return nil
	}

	r := strings.Split(rule, "=")
	if len(r) < 2 {
		return fmt.Errorf("length invalid rule definition, name : " + name)
	}

	limit, err := strconv.Atoi(strings.TrimSpace(r[1]))
	if err != nil {
		return fmt.Errorf("invalid rule:(%v) %w", name, err)
	}

	if len(value) > limit {
		return fmt.Errorf("when have value, total characters for field %v must be less or same than %v character(s)", name, limit)
	}

	return nil
}

func strRange(rule, name string, value string) error {
	if !strings.Contains(rule, "range") {
		return nil
	}

	temp := strings.ReplaceAll(rule, "range", "")
	temp = strings.ReplaceAll(temp, "[", "")
	temp = strings.ReplaceAll(temp, "]", "")

	array := strings.Split(temp, ",")
	found := false
	for _, val := range array {
		if val == value {
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("field %v value must in [%v]", name, temp)
	}

	return nil
}

func strLength(rule, name, value string) error {
	if !strings.Contains(rule, "length") {
		return nil
	}

	r := strings.Split(rule, "=")
	if len(r) < 2 {
		return fmt.Errorf("length invalid rule definition, name : " + name)
	}

	limit, err := strconv.Atoi(strings.TrimSpace(r[1]))
	if err != nil {
		return fmt.Errorf("invalid rule:(%v) %w", name, err)
	}

	if !(len(value) == limit) {
		return fmt.Errorf("field %v must have %v character(s)", name, limit)
	}

	return nil
}

func strRequired(rule, name, value string) error {
	if !strings.Contains(rule, "required") {
		return nil
	}
	if value == "" {
		return fmt.Errorf("field %v must be filled", name)
	}

	return nil
}

func strMinLength(rule, name, value string) error {
	if !strings.Contains(rule, "min") {
		return nil
	}

	r := strings.Split(rule, "=")
	if len(r) < 2 {
		return fmt.Errorf("min-length invalid rule definition, name : " + name)
	}

	limit, err := strconv.Atoi(strings.TrimSpace(r[1]))
	if err != nil {
		return fmt.Errorf("invalid rule:(%v) %w", name, err)
	}

	if len(value) < limit {
		return fmt.Errorf("field %v must have at least %v character(s)", name, limit)
	}

	return nil
}

func strMaxLength(rule, name, value string) error {
	if !strings.Contains(rule, "max") {
		return nil
	}

	r := strings.Split(rule, "=")
	if len(r) < 2 {
		return fmt.Errorf("max-length invalid rule definition, name : " + name)
	}

	limit, err := strconv.Atoi(strings.TrimSpace(r[1]))
	if err != nil {
		return fmt.Errorf("invalid rule:(%v) %w", name, err)
	}

	if len(value) > limit {
		return fmt.Errorf("total characters for field %v must be less or same than %v character(s)", name, limit)
	}

	return nil
}

func int64OptX(rule, name string, value int64) error {
	if value == 0 {
		return nil
	}

	if !strings.Contains(rule, "optx") {
		return nil
	}

	r := strings.Split(rule, "=")
	if len(r) < 2 {
		return fmt.Errorf("length invalid rule definition, name : " + name)
	}

	limit := StringToInt64(strings.TrimSpace(r[1]))
	if limit == 0 {
		return fmt.Errorf("invalid rule:(%v) %w", name, errors.New("cannot define zero in rule"))
	}

	if value < limit {
		return fmt.Errorf("when have value, field %v must have at least %v character(s)", name, limit)
	}

	return nil
}

func int64OptY(rule, name string, value int64) error {
	if value == 0 {
		return nil
	}

	if !strings.Contains(rule, "opty") {
		return nil
	}

	r := strings.Split(rule, "=")
	if len(r) < 2 {
		return fmt.Errorf("length invalid rule definition, name : " + name)
	}

	limit := StringToInt64(strings.TrimSpace(r[1]))
	if limit == 0 {
		return fmt.Errorf("invalid rule:(%v) %w", name, errors.New("cannot define zero in rule"))
	}

	if value > limit {
		return fmt.Errorf("when have value, total characters for field %v must be less or same than %v character(s)", name, limit)
	}

	return nil
}

func int64Required(rule, name string, value int64) error {
	if !strings.Contains(rule, "required") {
		return nil
	}
	if value == 0 {
		return fmt.Errorf("field %v must not zero", name)
	}

	return nil
}
func int64Range(rule, name string, value int64) error {
	if !strings.Contains(rule, "range") {
		return nil
	}

	temp := strings.ReplaceAll(rule, "range", "")
	temp = strings.ReplaceAll(temp, "[", "")
	temp = strings.ReplaceAll(temp, "]", "")

	array := strings.Split(temp, ",")
	found := false
	for _, val := range array {
		t, _ := strconv.ParseInt(val, 10, 64)
		if t == value {
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("field %v value must in [%v]", name, temp)
	}

	return nil
}

func int64MinValue(rule, name string, value int64) error {
	if !strings.Contains(rule, "min") {
		return nil
	}
	r := strings.Split(rule, "=")
	if len(r) < 2 {
		return fmt.Errorf("min-value invalid rule definition, name : " + name)
	}

	limit, err := strconv.ParseInt(strings.TrimSpace(r[1]), 10, 64)
	if err != nil {
		return fmt.Errorf("min-value invalid rule:(%v) %w", name, err)
	}

	if value < limit {
		return fmt.Errorf("field %v must not less than %v", name, limit)
	}

	return nil
}

func int64MaxValue(rule, name string, value int64) error {
	if !strings.Contains(rule, "max") {
		return nil
	}
	r := strings.Split(rule, "=")
	if len(r) < 2 {
		return fmt.Errorf("max-value invalid rule definition, name : " + name)
	}

	limit, err := strconv.ParseInt(strings.TrimSpace(r[1]), 10, 64)
	if err != nil {
		return fmt.Errorf("max-value invalid rule:(%v) %w", name, err)
	}

	if value > limit {
		return fmt.Errorf("field %v must not greater than %v", name, limit)
	}

	return nil
}

func int32OptX(rule, name string, value int32) error {
	if value == 0 {
		return nil
	}

	if !strings.Contains(rule, "optx") {
		return nil
	}

	r := strings.Split(rule, "=")
	if len(r) < 2 {
		return fmt.Errorf("length invalid rule definition, name : " + name)
	}

	limit := StringToInt32(strings.TrimSpace(r[1]))
	if limit == 0 {
		return fmt.Errorf("invalid rule:(%v) %w", name, errors.New("cannot define zero in rule"))
	}

	if value < limit {
		return fmt.Errorf("when have value, field %v must have at least %v character(s)", name, limit)
	}

	return nil
}

func int32OptY(rule, name string, value int32) error {
	if value == 0 {
		return nil
	}

	if !strings.Contains(rule, "opty") {
		return nil
	}

	r := strings.Split(rule, "=")
	if len(r) < 2 {
		return fmt.Errorf("length invalid rule definition, name : " + name)
	}

	limit := StringToInt32(strings.TrimSpace(r[1]))
	if limit == 0 {
		return fmt.Errorf("invalid rule:(%v) %w", name, errors.New("cannot define zero in rule"))
	}

	if value > limit {
		return fmt.Errorf("when have value, total characters for field %v must be less or same than %v character(s)", name, limit)
	}

	return nil
}

func int32Required(rule, name string, value int32) error {
	if !strings.Contains(rule, "required") {
		return nil
	}
	if value == 0 {
		return fmt.Errorf("field %v must not zero", name)
	}

	return nil
}
func int32Range(rule, name string, value int32) error {
	if !strings.Contains(rule, "range") {
		return nil
	}

	temp := strings.ReplaceAll(rule, "range", "")
	temp = strings.ReplaceAll(temp, "[", "")
	temp = strings.ReplaceAll(temp, "]", "")

	array := strings.Split(temp, ",")
	found := false
	for _, val := range array {
		t := StringToInt32(val)
		if t == value {
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("field %v value must in [%v]", name, temp)
	}

	return nil
}

func int32MinValue(rule, name string, value int32) error {
	if !strings.Contains(rule, "min") {
		return nil
	}
	r := strings.Split(rule, "=")
	if len(r) < 2 {
		return fmt.Errorf("min-value invalid rule definition, name : " + name)
	}

	limit := StringToInt32(strings.TrimSpace(r[1]))

	if value < limit {
		return fmt.Errorf("field %v must not less than %v", name, limit)
	}

	return nil
}

func int32MaxValue(rule, name string, value int32) error {
	if !strings.Contains(rule, "max") {
		return nil
	}
	r := strings.Split(rule, "=")
	if len(r) < 2 {
		return fmt.Errorf("max-value invalid rule definition, name : " + name)
	}

	limit := StringToInt32(strings.TrimSpace(r[1]))

	if value > limit {
		return fmt.Errorf("field %v must not greater than %v", name, limit)
	}

	return nil
}

func intOptX(rule, name string, value int) error {
	if value == 0 {
		return nil
	}

	if !strings.Contains(rule, "optx") {
		return nil
	}

	r := strings.Split(rule, "=")
	if len(r) < 2 {
		return fmt.Errorf("length invalid rule definition, name : " + name)
	}

	limit := StringToInt(strings.TrimSpace(r[1]))
	if limit == 0 {
		return fmt.Errorf("invalid rule:(%v) %w", name, errors.New("cannot define zero in rule"))
	}

	if value < limit {
		return fmt.Errorf("when have value, field %v must have at least %v character(s)", name, limit)
	}

	return nil
}

func intOptY(rule, name string, value int) error {
	if value == 0 {
		return nil
	}

	if !strings.Contains(rule, "opty") {
		return nil
	}

	r := strings.Split(rule, "=")
	if len(r) < 2 {
		return fmt.Errorf("length invalid rule definition, name : " + name)
	}

	limit := StringToInt(strings.TrimSpace(r[1]))
	if limit == 0 {
		return fmt.Errorf("invalid rule:(%v) %w", name, errors.New("cannot define zero in rule"))
	}

	if value > limit {
		return fmt.Errorf("when have value, total characters for field %v must be less or same than %v character(s)", name, limit)
	}

	return nil
}

func intRequired(rule, name string, value int) error {
	if !strings.Contains(rule, "required") {
		return nil
	}
	if value == 0 {
		return fmt.Errorf("field %v must not zero", name)
	}

	return nil
}
func intRange(rule, name string, value int) error {
	if !strings.Contains(rule, "range") {
		return nil
	}

	temp := strings.ReplaceAll(rule, "range", "")
	temp = strings.ReplaceAll(temp, "[", "")
	temp = strings.ReplaceAll(temp, "]", "")

	array := strings.Split(temp, ",")
	found := false
	for _, val := range array {
		t, _ := strconv.Atoi(val)
		if t == value {
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("field %v value must in [%v]", name, temp)
	}

	return nil
}

func intMinValue(rule, name string, value int) error {
	if !strings.Contains(rule, "min") {
		return nil
	}
	r := strings.Split(rule, "=")
	if len(r) < 2 {
		return fmt.Errorf("min-value invalid rule definition, name : " + name)
	}

	limit, err := strconv.Atoi(strings.TrimSpace(r[1]))
	if err != nil {
		return fmt.Errorf("min-value invalid rule:(%v) %w", name, err)
	}

	if value < limit {
		return fmt.Errorf("field %v must not less than %v", name, limit)
	}

	return nil
}

func intMaxValue(rule, name string, value int) error {
	if !strings.Contains(rule, "max") {
		return nil
	}
	r := strings.Split(rule, "=")
	if len(r) < 2 {
		return fmt.Errorf("max-value invalid rule definition, name : " + name)
	}

	limit, err := strconv.Atoi(strings.TrimSpace(r[1]))
	if err != nil {
		return fmt.Errorf("max-value invalid rule:(%v) %w", name, err)
	}

	if value > limit {
		return fmt.Errorf("field %v must not greater than %v", name, limit)
	}

	return nil
}

func float64OptX(rule, name string, value float64) error {
	if value == 0 {
		return nil
	}

	if !strings.Contains(rule, "optx") {
		return nil
	}

	r := strings.Split(rule, "=")
	if len(r) < 2 {
		return fmt.Errorf("length invalid rule definition, name : " + name)
	}

	limit := StringToFloat64(strings.TrimSpace(r[1]))
	if limit == 0 {
		return fmt.Errorf("invalid rule:(%v) %w", name, errors.New("cannot define zero in rule"))
	}

	if value < limit {
		return fmt.Errorf("when have value, field %v must have at least %v character(s)", name, limit)
	}

	return nil
}

func float64OptY(rule, name string, value float64) error {
	if value == 0 {
		return nil
	}

	if !strings.Contains(rule, "opty") {
		return nil
	}

	r := strings.Split(rule, "=")
	if len(r) < 2 {
		return fmt.Errorf("length invalid rule definition, name : " + name)
	}

	limit := StringToFloat64(strings.TrimSpace(r[1]))
	if limit == 0 {
		return fmt.Errorf("invalid rule:(%v) %w", name, errors.New("cannot define zero in rule"))
	}

	if value > limit {
		return fmt.Errorf("when have value, total characters for field %v must be less or same than %v character(s)", name, limit)
	}

	return nil
}

func float64Required(rule, name string, value float64) error {
	if !strings.Contains(rule, "required") {
		return nil
	}
	if value == 0 {
		return fmt.Errorf("field %v must not zero", name)
	}

	return nil
}

func float64Range(rule, name string, value float64) error {
	if !strings.Contains(rule, "range") {
		return nil
	}

	temp := strings.ReplaceAll(rule, "range", "")
	temp = strings.ReplaceAll(temp, "[", "")
	temp = strings.ReplaceAll(temp, "]", "")

	array := strings.Split(temp, ",")
	found := false
	for _, val := range array {
		t, _ := strconv.ParseFloat(val, 64)
		if t == value {
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("field %v value must in [%v]", name, temp)
	}

	return nil
}

func float64MinValue(rule, name string, value float64) error {
	if !strings.Contains(rule, "min") {
		return nil
	}
	r := strings.Split(rule, "=")
	if len(r) < 2 {
		return fmt.Errorf("min-value invalid rule definition, name : " + name)
	}

	limit, err := strconv.ParseFloat(r[1], 64)
	if err != nil {
		return fmt.Errorf("min-value invalid rule:(%v) %w", name, err)
	}

	if value < limit {
		return fmt.Errorf("field %v must not less than %v", name, limit)
	}

	return nil
}

func float64MaxValue(rule, name string, value float64) error {
	if !strings.Contains(rule, "max") {
		return nil
	}
	r := strings.Split(rule, "=")
	if len(r) < 2 {
		return fmt.Errorf("max-value invalid rule definition, name : " + name)
	}

	limit, err := strconv.ParseFloat(r[1], 64)
	if err != nil {
		return fmt.Errorf("max-value invalid rule:(%v) %w", name, err)
	}

	if value > limit {
		return fmt.Errorf("field %v must not greater than %v", name, limit)
	}

	return nil
}
