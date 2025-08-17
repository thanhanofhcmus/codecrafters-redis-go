package argsparser

import (
	"fmt"
	"reflect"
	"slices"
	"strconv"
	"strings"
)

/*
TODO:
- handle sub command
- handle variadic
- add unimplemented tag for the key present in REDIS but not handle by us now
- support different fieldName and argName
- add default tag
*/

var (
	// TODO: move to use thread safe map
	parsedCache map[reflect.Type]structMetadata
)

func init() {
	parsedCache = map[reflect.Type]structMetadata{}
}

type fieldType = string

var (
	fieldTypePosition  fieldType = "pos"
	fieldTypeOption    fieldType = "opt"
	fieldTypeEnum      fieldType = "enum"
	fieldTypeEnumValue fieldType = "enum-value"
	fieldTypeEnumKey   fieldType = "enum-key"

	fieldTypeAuto = "auto"
)

type attribute struct {
	isOptional      bool
	isUnimplemented bool
	rawDefault      string
}

type positionMetadata struct {
	fieldName string
	kind      reflect.Kind
	attribute attribute
	position  int
}

type enumMemberMetadata struct {
	fieldName string
	argName   string
	kind      reflect.Kind
	attribute attribute
	parent    *enumMetadata
}

type enumMetadata struct {
	fieldName string
	attribute attribute
	// keys are arg name
	enumMembers       map[string]enumMemberMetadata
	storeKeyFieldName string
}

type optionMetadatada struct {
	fieldName string
	argName   string
	kind      reflect.Kind
	attribute attribute
}

type optionOrEnumMember struct {
	option     *optionMetadatada
	enumMember *enumMemberMetadata
}

type structMetadata struct {
	positions []positionMetadata
	options   []optionMetadatada
	enums     []enumMetadata

	argKeys map[string]optionOrEnumMember
}

func Parse[T any](args []string) (T, error) {
	var result T
	var smd structMetadata

	if cached, exists := parsedCache[reflect.TypeFor[T]()]; exists {
		smd = cached
	} else {
		var err error
		smd, err = extractTag[T]()
		if err != nil {
			return result, err
		}
		parsedCache[reflect.TypeFor[T]()] = smd
	}

	length := len(args)
	value := reflect.ValueOf(&result)
	value = value.Elem()
	idx := 1

	// handle position arguments
	posArgLen := len(smd.positions)
	if idx+posArgLen > length {
		return result, fmt.Errorf("not enough arguments")
	}
	for i := range posArgLen {
		if err := setFieldValue(value, smd.positions[i].fieldName, args[idx+i]); err != nil {
			return result, fmt.Errorf("set position argument %d failed: %w", i+1, err)
		}
	}
	idx += posArgLen

	// handle the mix of options and enums
	doneFields := map[string]bool{}
	for idx < length {
		name := strings.ToUpper(args[idx])
		idx += 1

		// TODO: handle enum can only appears one alternative (e.g. "SET a a NX XX")

		if doneFields[name] {
			return result, fmt.Errorf("argument `%s` appears more than one time", name)
		}

		metadata, exists := smd.argKeys[name]
		if !exists {
			return result, fmt.Errorf("invalid argument `%s`", name)
		}

		if err := processOptionsAndEnums(args, metadata, value, &idx); err != nil {
			return result, fmt.Errorf("process `%s` failed: %w", name, err)
		}

		doneFields[name] = true
	}

	return result, nil
}

func processOptionsAndEnums(args []string, md optionOrEnumMember, value reflect.Value, idx *int) error {
	if md.option != nil {
		raw := ""
		if md.option.kind != reflect.Bool {
			if len(args) == *idx {
				return fmt.Errorf("no value provided")
			}
			raw = args[*idx]
			*idx += 1
		}
		if err := setFieldValue(value, md.option.fieldName, raw); err != nil {
			return err
		}
	} else if md.enumMember != nil {
		raw := ""
		if md.enumMember.kind != reflect.Bool {
			if len(args) == *idx {
				return fmt.Errorf("no value provided")
			}
			raw = args[*idx]
			*idx += 1
		}
		if err := setEnumMemberValue(value, *md.enumMember, raw); err != nil {
			return err
		}
	} else {
		return fmt.Errorf("not an option or enum argument")
	}
	return nil
}

func isSimpleKind(kind reflect.Kind) bool {
	switch kind {
	case reflect.Bool,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64,
		reflect.String:
		return true
	}
	return false
}

func setSimpleValue(value reflect.Value, raw string) error {
	switch kind := value.Type().Kind(); kind {

	// TODO: refined parse number method
	// TODO: validate with value.Can***
	case reflect.Bool:
		// bool is always set to true without checking raw
		// since to get here, the key must be present on the arg
		value.SetBool(true)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		v, err := strconv.ParseInt(raw, 10, 64)
		if err != nil {
			return err
		}
		value.SetInt(v)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		v, err := strconv.ParseUint(raw, 10, 64)
		if err != nil {
			return err
		}
		value.SetUint(v)
	case reflect.Float32, reflect.Float64:
		v, err := strconv.ParseFloat(raw, 64)
		if err != nil {
			return err
		}
		value.SetFloat(v)
	case reflect.String:
		value.SetString(raw)
	default:
		return fmt.Errorf("invalid simple type %s", kind)
	}

	return nil
}

func setFieldValue(value reflect.Value, fieldName string, raw string) error {
	// TODO: validate field exists
	fieldValue := value.FieldByName(fieldName)
	if err := setSimpleValue(fieldValue, raw); err != nil {
		return fmt.Errorf("set value for option field %s: %w", fieldName, err)
	}
	return nil
}

func setEnumMemberValue(value reflect.Value, emmd enumMemberMetadata, raw string) error {
	// TODO: handle error
	parentFieldName := emmd.parent.fieldName
	fieldName := emmd.fieldName

	fieldValue := value.FieldByName(parentFieldName).FieldByName(fieldName)
	if err := setSimpleValue(fieldValue, raw); err != nil {
		return fmt.Errorf("set value for option enum %s.%s: %w", parentFieldName, fieldName, err)
	}

	return nil
}

func extractTag[T any]() (structMetadata, error) {
	t := reflect.TypeFor[T]()

	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}

	if t.Kind() != reflect.Struct {
		return structMetadata{}, fmt.Errorf("can only extract metadata from structs")
	}

	result := structMetadata{}

	for i := range t.NumField() {
		field := t.Field(i)
		err := extractFieldTag(t.Field(i), &result)
		if err != nil {
			return structMetadata{}, fmt.Errorf("extract field `%s` metadata failed: %w", field.Name, err)
		}
	}

	// TODO: validate attributes

	argKeys := map[string]optionOrEnumMember{}
	for _, omd := range result.options {
		name := omd.argName
		if _, exists := argKeys[name]; exists {
			return structMetadata{}, fmt.Errorf("option or enum member with name `%s` already exists", name)
		}
		argKeys[name] = optionOrEnumMember{option: &omd}
	}
	for _, emd := range result.enums {
		for name, emmd := range emd.enumMembers {
			if _, exists := argKeys[name]; exists {
				return structMetadata{}, fmt.Errorf("option or enum member with name `%s` already exists", name)
			}
			argKeys[name] = optionOrEnumMember{enumMember: &emmd}
		}
	}
	result.argKeys = argKeys

	seenPos := map[int]bool{}
	maxPos := -1
	slices.SortFunc(result.positions, func(a positionMetadata, b positionMetadata) int {
		return a.position - b.position
	})
	for _, pmd := range result.positions {
		if seenPos[pmd.position] {
			return structMetadata{}, fmt.Errorf("position %d appears more than one time", pmd.position)
		}
		maxPos = max(maxPos, pmd.position)
		seenPos[pmd.position] = true
	}
	if maxPos != len(seenPos) {
		return structMetadata{}, fmt.Errorf("max position %d is not them same as the number of position arguments %d", maxPos, len(seenPos))
	}

	return result, nil
}

func extractFieldTag(field reflect.StructField, smd *structMetadata) (err error) {
	kind := field.Type.Kind()
	fieldName := field.Name

	fieldType, fieldSnd, attribute, err := parseTag(field.Tag.Get("arg"))
	if err != nil {
		return fmt.Errorf("extract tag for field %s failed: %w", fieldName, err)
	}

	switch fieldType {
	case fieldTypePosition:
		if kind == reflect.Bool || !isSimpleKind(kind) {
			return fmt.Errorf("invalid position argument kind")
		}
		if fieldSnd == "" {
			return fmt.Errorf("invalid pos format")
		}
		position, err := strconv.Atoi(fieldSnd)
		if err != nil {
			return fmt.Errorf("convert position failed %w", err)
		}
		smd.positions = append(smd.positions, positionMetadata{
			fieldName: fieldName,
			kind:      kind,
			attribute: attribute,
			position:  position,
		})
	case fieldTypeEnum:
		emd := enumMetadata{
			fieldName: fieldName,
			attribute: attribute,
		}

		enumMembers, err := extractEnumMember(field.Type, &emd)
		if err != nil {
			return fmt.Errorf("extract enum state failed: %w", err)
		}

		emd.enumMembers = enumMembers
		smd.enums = append(smd.enums, emd)
	case fieldTypeOption, fieldTypeAuto:
		smd.options = append(smd.options, optionMetadatada{
			fieldName: fieldName,
			argName:   fieldName,
			kind:      kind,
			attribute: attribute,
		})
	default:
		return fmt.Errorf("unknown tag field type %s", fieldType)
	}

	return
}

func parseTag(tag string) (fieldType fieldType, fieldSnd string, attribute attribute, err error) {
	// TODO: handle trim after split
	fieldType = fieldTypeAuto

	attrParts := strings.Split(tag, ",")
	if len(attrParts) == 0 {
		return fieldTypeAuto, "", attribute, nil
	}
	firstParts := strings.Split(attrParts[0], ",")
	if len(firstParts) > 0 {
		if firstParts[0] != "" {
			fieldType = firstParts[0]
		}
		if len(firstParts) == 2 {
			fieldSnd = firstParts[1]
		}
		// TODO: strict validation, only allow splits of 2
	}
	for _, attrRaw := range attrParts[1:] {
		parts := strings.Split(attrRaw, ",")
		switch parts[0] {
		case "default":
			// TODO:
			panic("default attribute not handled")
		case "name":
			// TODO:
			panic("name attribute not handled")
		case "optional":
			attribute.isOptional = true
		case "unimplemented":
			attribute.isUnimplemented = true
		default:
			// TODO:
			panic("attribute not handled")
		}
		// TODO: strict validation
	}

	return

}

func extractEnumMember(t reflect.Type, parent *enumMetadata) (map[string]enumMemberMetadata, error) {
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return nil, fmt.Errorf("enum state can only be extracted from structs")
	}

	result := map[string]enumMemberMetadata{}

	for i := range t.NumField() {
		field := t.Field(i)
		fieldKind := field.Type.Kind()

		fieldName := field.Name
		argName := field.Name

		// TODO: validate fieldSnd
		fieldType, _, attribute, err := parseTag(field.Tag.Get("arg"))
		if err != nil {
			return nil, fmt.Errorf("extract tag for field %s failed: %w", fieldName, err)
		}

		switch fieldType {
		case fieldTypeEnumValue, fieldTypeAuto:
			if !isSimpleKind(fieldKind) {
				return nil, fmt.Errorf("invalid kind %s", fieldKind)
			}
			result[argName] = enumMemberMetadata{
				fieldName: fieldName,
				argName:   argName,
				kind:      fieldKind,
				attribute: attribute,
				parent:    parent,
			}
		case fieldTypeEnumKey:
			// TODO: strict validate no other options
			if parent.storeKeyFieldName != "" {
				return nil, fmt.Errorf("cannot have multiple %s tag", fieldTypeEnumKey)
			}
			parent.storeKeyFieldName = fieldName
		default:
			return nil, fmt.Errorf("invalid tag field type %s", fieldType)
		}
	}

	return result, nil
}
