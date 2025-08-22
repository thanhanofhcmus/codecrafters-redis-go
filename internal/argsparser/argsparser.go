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
- support different fieldName and argName
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
	isVariadic      bool
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

type optionMetadata struct {
	fieldName string
	argName   string
	kind      reflect.Kind
	attribute attribute
}

type optionOrEnumMember struct {
	option     *optionMetadata
	enumMember *enumMemberMetadata
}

type structMetadata struct {
	positions                []positionMetadata
	requiredPositionArgsSize int
	options                  []optionMetadata
	enums                    []enumMetadata

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

	value := reflect.ValueOf(&result)
	value = value.Elem()
	idx := 1

	if err := parsePositions(args, smd, value, &idx); err != nil {
		return result, err
	}

	if err := parseOptionsAndEnums(args, smd, value, &idx); err != nil {
		return result, err
	}

	return result, nil
}

func parsePositions(args []string, smd structMetadata, value reflect.Value, idx *int) error {
	posIdx := 0

	length := len(args)
	positions := smd.positions
	posLength := len(smd.positions)

	hasVariadic := posLength >= 0 && positions[posLength-1].attribute.isVariadic

	// handle required position arguments
	if *idx+smd.requiredPositionArgsSize > length {
		return fmt.Errorf("not enough arguments")
	}

	for *idx < length && posIdx < smd.requiredPositionArgsSize {
		if err := setFieldValue(value, positions[posIdx].fieldName, args[*idx]); err != nil {
			return fmt.Errorf("set required position argument %d failed: %w", posIdx, err)
		}
		*idx += 1
		posIdx += 1
	}

	// handle variadic arguments
	if hasVariadic {
		if err := setFieldArrayValue(value, positions[posLength-1].fieldName, args[smd.requiredPositionArgsSize+1:]); err != nil {
			return err
		}
		// we have already gone through the whole array
		*idx = length
		posIdx = posLength
	}

	// handle optional position arguments
	for *idx < length && posIdx < posLength {
		if err := setFieldValue(value, smd.positions[posIdx].fieldName, args[*idx]); err != nil {
			return fmt.Errorf("set optional position argument %d failed: %w", posIdx, err)
		}
		*idx += 1
		posIdx += 1
	}

	// handle default value for optional position arguments
	for posIdx < posLength {
		if pmd := positions[posIdx]; pmd.attribute.rawDefault != "" {
			if err := setFieldValue(value, pmd.fieldName, pmd.attribute.rawDefault); err != nil {
				return fmt.Errorf("set default value for optional position argument %d failed: %w", posIdx, err)
			}
		}
		posIdx += 1
	}

	return nil
}

func parseOptionsAndEnums(args []string, smd structMetadata, value reflect.Value, idx *int) error {
	length := len(args)
	doneFields := map[string]bool{}
	for *idx < length {
		name := strings.ToUpper(args[*idx])
		*idx += 1

		// TODO: handle enum can only appears one alternative (e.g. "SET a a NX XX")

		if doneFields[name] {
			return fmt.Errorf("argument `%s` appears more than one time", name)
		}

		metadata, exists := smd.argKeys[name]
		if !exists {
			return fmt.Errorf("invalid argument `%s`", name)
		}

		if err := processOptionsAndEnums(args, name, metadata, value, idx); err != nil {
			return fmt.Errorf("process `%s` failed: %w", name, err)
		}

		doneFields[name] = true
	}
	return nil
}

func processOptionsAndEnums(args []string, key string, md optionOrEnumMember, value reflect.Value, idx *int) error {
	if md.option != nil {
		return processOption(args, *md.option, value, idx)
	} else if md.enumMember != nil {
		return processEnumMember(args, key, *md.enumMember, value, idx)
	} else {
		return fmt.Errorf("not an option or enum argument")
	}
}

func processOption(args []string, omd optionMetadata, value reflect.Value, idx *int) error {
	raw := ""
	if omd.kind != reflect.Bool {
		if len(args) == *idx {
			if !omd.attribute.isOptional {
				return fmt.Errorf("no value provided")
			}
			raw = omd.attribute.rawDefault
		} else {
			raw = args[*idx]
			*idx += 1
		}
	}
	if err := setFieldValue(value, omd.fieldName, raw); err != nil {
		return err
	}
	return nil
}

func processEnumMember(args []string, key string, md enumMemberMetadata, value reflect.Value, idx *int) error {
	raw := ""
	if md.kind != reflect.Bool {
		if len(args) == *idx {
			return fmt.Errorf("no value provided")
		}
		raw = args[*idx]
		*idx += 1
	}
	parent := md.parent
	if err := setEnumMemberValue(value, parent.fieldName, md.fieldName, raw); err != nil {
		return err
	}
	if parent.storeKeyFieldName != "" {
		if err := setEnumMemberValue(value, parent.fieldName, parent.storeKeyFieldName, key); err != nil {
			return err
		}
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

func setFieldArrayValue(value reflect.Value, fieldName string, raws []string) error {
	fieldValue := value.FieldByName(fieldName)

	slice := reflect.MakeSlice(fieldValue.Type(), len(raws), len(raws))
	for idx, raw := range raws {
		elem := slice.Index(idx)
		if err := setSimpleValue(elem, raw); err != nil {
			return err
		}
	}
	fieldValue.Set(slice)
	return nil
}

func setFieldValue(value reflect.Value, fieldName string, raw string) error {
	// TODO: validate field exists
	fieldValue := value.FieldByName(fieldName)

	if fieldValue.Kind() == reflect.Pointer {
		ptrValue := reflect.New(fieldValue.Type().Elem())
		fieldValue.Set(ptrValue)
		fieldValue = fieldValue.Elem()
	}

	if err := setSimpleValue(fieldValue, raw); err != nil {
		return fmt.Errorf("set value for field %s: %w", fieldName, err)
	}
	return nil
}

func setEnumMemberValue(value reflect.Value, parentFieldName, fieldName, raw string) error {
	// TODO: handle error
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
	// tag optional or variadic should only be at the last of of pos
	// can only have optional or variadic

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
	requiredCount := 0
	slices.SortFunc(result.positions, func(a positionMetadata, b positionMetadata) int {
		return a.position - b.position
	})
	for _, pmd := range result.positions {
		if seenPos[pmd.position] {
			return structMetadata{}, fmt.Errorf("position %d appears more than one time", pmd.position)
		}
		seenPos[pmd.position] = true
		maxPos = max(maxPos, pmd.position)
		if !pmd.attribute.isOptional && !pmd.attribute.isVariadic {
			requiredCount += 1
		}
	}
	if maxPos != len(seenPos) {
		return structMetadata{}, fmt.Errorf("max position %d is not them same as the number of position arguments %d", maxPos, len(seenPos))
	}
	result.requiredPositionArgsSize = requiredCount

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
		if kind == reflect.Bool {
			return fmt.Errorf("invalid position argument must not be of boolean type")
		}
		if attribute.isVariadic {
			// TODO: check for reflect.Array
			if kind != reflect.Slice {
				return fmt.Errorf("variadic position argument must have the kind array")
			}
		} else {
			elemKind := kind
			if elemKind == reflect.Pointer {
				elemKind = field.Type.Elem().Kind()
			}
			if !isSimpleKind(elemKind) {
				return fmt.Errorf("invalid non variadic position argument kind")
			}
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
		smd.options = append(smd.options, optionMetadata{
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
	firstParts := strings.Split(attrParts[0], ":")
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
		parts := strings.Split(attrRaw, ":")
		switch parts[0] {
		case "default":
			if len(parts) != 2 {
				panic("wrong default attribute format")
			}
			attribute.rawDefault = parts[1]
		case "name":
			// TODO:
			panic("name attribute not handled")
		case "optional":
			attribute.isOptional = true
		case "variadic":
			attribute.isVariadic = true
			// TODO: remove unimplemented since we not get it here
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
