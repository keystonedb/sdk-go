package proto

type PropertyDefinition struct {
	DataType     Property_Type
	ExtendedType Property_ExtendedType
	Options      []Property_Option
}
