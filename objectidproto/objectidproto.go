// Package objectidproto builds a trogonobjectid.Definition from the
// trogon.object_id.v1alpha1.enum_value annotation on a protobuf enum value:
//
//	enum ObjectType {
//	  OBJECT_TYPE_UNSPECIFIED = 0;
//	  OBJECT_TYPE_TICKET = 1 [(trogon.object_id.v1alpha1.enum_value).object_type = "tkt"];
//	}
//
// The object type and separator then live in one place for every language that uses
// the proto, and the other fields come from a base Definition:
//
//	var ticketDefinition = objectidproto.MustDefinition(acmev1.ObjectType_OBJECT_TYPE_TICKET, trogonobjectid.Definition{
//		StorageFormat: trogonobjectid.FormatDropPrefix,
//		Validate:      trogonobjectid.ValidateUUID,
//	})
//
//	type ticketType struct{}
//
//	func (ticketType) Definition() trogonobjectid.Definition { return ticketDefinition }
package objectidproto

import (
	"fmt"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"

	objectidv1alpha1 "github.com/TrogonStack/trogonproto/gen/trogon/object_id/v1alpha1"

	"github.com/TrogonStack/trogonobjectid"
)

// Definition returns a Definition with ObjectType and Separator read from the enum
// value's annotation, over base. An absent separator gives
// trogonobjectid.DefaultSeparator. base carries the fields the annotation does not, so
// a new Definition field needs no change here.
func Definition(value protoreflect.Enum, base trogonobjectid.Definition) (trogonobjectid.Definition, error) {
	desc := value.Descriptor().Values().ByNumber(value.Number())
	if desc == nil {
		return trogonobjectid.Definition{}, fmt.Errorf("%w: %s has no value %d",
			trogonobjectid.ErrInvalidDefinition, value.Descriptor().FullName(), value.Number())
	}

	valueOpts, _ := desc.Options().(*descriptorpb.EnumValueOptions)
	if valueOpts == nil || !proto.HasExtension(valueOpts, objectidv1alpha1.E_EnumValue) {
		return trogonobjectid.Definition{}, fmt.Errorf("%w: %s has no %s annotation",
			trogonobjectid.ErrInvalidDefinition, desc.FullName(), objectidv1alpha1.E_EnumValue.TypeDescriptor().FullName())
	}

	ext, _ := proto.GetExtension(valueOpts, objectidv1alpha1.E_EnumValue).(*objectidv1alpha1.EnumValueOptions)
	if ext.GetObjectType() == "" {
		return trogonobjectid.Definition{}, fmt.Errorf("%w: %s has an empty object_type",
			trogonobjectid.ErrInvalidDefinition, desc.FullName())
	}

	separator := trogonobjectid.DefaultSeparator
	if ext.Separator != nil {
		separator = trogonobjectid.Separator(ext.GetSeparator())
	}
	base.ObjectType = trogonobjectid.ObjectType(ext.GetObjectType())
	base.Separator = separator
	return base, nil
}

// MustDefinition is like Definition but panics on error. Use it in package-level variables.
func MustDefinition(value protoreflect.Enum, base trogonobjectid.Definition) trogonobjectid.Definition {
	def, err := Definition(value, base)
	if err != nil {
		panic(err)
	}
	return def
}
