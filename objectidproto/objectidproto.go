// Package objectidproto builds a trogonobjectid.Definition from the
// trogon.object_id.v1alpha1.enum_value annotation on a protobuf enum value:
//
//	enum ObjectType {
//	  OBJECT_TYPE_UNSPECIFIED = 0;
//	  OBJECT_TYPE_TICKET = 1 [(trogon.object_id.v1alpha1.enum_value).object_type = "tkt"];
//	}
//
// The object type and separator then live in one place for every language that uses
// the proto. Keep the result in a package-level variable, because Definition methods
// run on every call:
//
//	var ticketDefinition = func() trogonobjectid.Definition {
//		def := objectidproto.MustDefinition(acmev1.ObjectType_OBJECT_TYPE_TICKET)
//		def.StorageFormat = trogonobjectid.FormatDropPrefix
//		def.Validate = trogonobjectid.ValidateUUID
//		return def
//	}()
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
// value's annotation. An absent separator gives trogonobjectid.DefaultSeparator.
// Set the other fields on the result.
func Definition(value protoreflect.Enum) (trogonobjectid.Definition, error) {
	desc := value.Descriptor().Values().ByNumber(value.Number())
	if desc == nil {
		return trogonobjectid.Definition{}, fmt.Errorf("%w: %s has no value %d",
			trogonobjectid.ErrInvalidDefinition, value.Descriptor().FullName(), value.Number())
	}

	opts, _ := desc.Options().(*descriptorpb.EnumValueOptions)
	if opts == nil || !proto.HasExtension(opts, objectidv1alpha1.E_EnumValue) {
		return trogonobjectid.Definition{}, fmt.Errorf("%w: %s has no %s annotation",
			trogonobjectid.ErrInvalidDefinition, desc.FullName(), objectidv1alpha1.E_EnumValue.TypeDescriptor().FullName())
	}

	ext, _ := proto.GetExtension(opts, objectidv1alpha1.E_EnumValue).(*objectidv1alpha1.EnumValueOptions)
	if ext.GetObjectType() == "" {
		return trogonobjectid.Definition{}, fmt.Errorf("%w: %s has an empty object_type",
			trogonobjectid.ErrInvalidDefinition, desc.FullName())
	}

	separator := trogonobjectid.DefaultSeparator
	if ext.Separator != nil {
		separator = ext.GetSeparator()
	}
	return trogonobjectid.Definition{ObjectType: ext.GetObjectType(), Separator: separator}, nil
}

// MustDefinition is like Definition but panics on error. Use it in package-level variables.
func MustDefinition(value protoreflect.Enum) trogonobjectid.Definition {
	def, err := Definition(value)
	if err != nil {
		panic(err)
	}
	return def
}
