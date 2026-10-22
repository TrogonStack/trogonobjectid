package objectidproto_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/dynamicpb"

	objectidv1alpha1 "github.com/TrogonStack/trogonproto/gen/trogon/object_id/v1alpha1"

	"github.com/TrogonStack/trogonobjectid"
	"github.com/TrogonStack/trogonobjectid/objectidproto"
)

const (
	valueUnspecified protoreflect.EnumNumber = iota
	valueTicket
	valueOrder
	valueEmptyType
)

// objectTypeEnum builds an ObjectType enum at runtime with the object_id annotation,
// so the tests need no generated protos.
func objectTypeEnum(t *testing.T) protoreflect.EnumType {
	t.Helper()

	annotated := func(opts *objectidv1alpha1.EnumValueOptions) *descriptorpb.EnumValueOptions {
		o := &descriptorpb.EnumValueOptions{}
		proto.SetExtension(o, objectidv1alpha1.E_EnumValue, opts)
		return o
	}

	file := &descriptorpb.FileDescriptorProto{
		Name:    proto.String("acme/type/v1/object_type.proto"),
		Package: proto.String("acme.type.v1"),
		Syntax:  proto.String("proto3"),
		EnumType: []*descriptorpb.EnumDescriptorProto{{
			Name: proto.String("ObjectType"),
			Value: []*descriptorpb.EnumValueDescriptorProto{
				{Name: proto.String("OBJECT_TYPE_UNSPECIFIED"), Number: proto.Int32(int32(valueUnspecified))},
				{
					Name:    proto.String("OBJECT_TYPE_TICKET"),
					Number:  proto.Int32(int32(valueTicket)),
					Options: annotated(&objectidv1alpha1.EnumValueOptions{ObjectType: "tkt"}),
				},
				{
					Name:    proto.String("OBJECT_TYPE_ORDER"),
					Number:  proto.Int32(int32(valueOrder)),
					Options: annotated(&objectidv1alpha1.EnumValueOptions{ObjectType: "order", Separator: proto.String("#")}),
				},
				{
					Name:    proto.String("OBJECT_TYPE_EMPTY"),
					Number:  proto.Int32(int32(valueEmptyType)),
					Options: annotated(&objectidv1alpha1.EnumValueOptions{}),
				},
			},
		}},
	}

	fd, err := protodesc.NewFile(file, nil)
	require.NoError(t, err)
	return dynamicpb.NewEnumType(fd.Enums().Get(0))
}

func TestDefinition(t *testing.T) {
	enum := objectTypeEnum(t)

	t.Run("reads the object type and uses the default separator", func(t *testing.T) {
		def, err := objectidproto.Definition(enum.New(valueTicket))
		require.NoError(t, err)
		assert.Equal(t, trogonobjectid.Definition{ObjectType: "tkt", Separator: "_"}, def)
	})

	t.Run("reads the separator", func(t *testing.T) {
		def, err := objectidproto.Definition(enum.New(valueOrder))
		require.NoError(t, err)
		assert.Equal(t, trogonobjectid.Definition{ObjectType: "order", Separator: "#"}, def)
	})

	tests := []struct {
		name   string
		number protoreflect.EnumNumber
	}{
		{"value without the annotation", valueUnspecified},
		{"empty object type", valueEmptyType},
		{"unknown enum number", 99},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := objectidproto.Definition(enum.New(tt.number))
			require.ErrorIs(t, err, trogonobjectid.ErrInvalidDefinition)
		})
	}

	t.Run("MustDefinition panics on error", func(t *testing.T) {
		assert.Panics(t, func() { objectidproto.MustDefinition(enum.New(valueUnspecified)) })
	})
}

var ticketEnum protoreflect.EnumType

type ticketType struct{}

func (ticketType) Definition() trogonobjectid.Definition {
	def := objectidproto.MustDefinition(ticketEnum.New(valueTicket))
	def.Validate = trogonobjectid.ValidateInteger
	return def
}

func TestDefinitionWithID(t *testing.T) {
	ticketEnum = objectTypeEnum(t)

	id, err := trogonobjectid.Parse[ticketType]("tkt_42")
	require.NoError(t, err)
	assert.Equal(t, "42", id.Bare())

	_, err = trogonobjectid.Parse[ticketType]("tkt_abc")
	require.ErrorIs(t, err, trogonobjectid.ErrInvalidInteger)
}

func TestDefinitionOptions(t *testing.T) {
	enum := objectTypeEnum(t)

	def, err := objectidproto.Definition(enum.New(valueTicket),
		objectidproto.WithStorageFormat(trogonobjectid.FormatDropPrefix),
		objectidproto.WithJSONFormat(trogonobjectid.FormatDropPrefix),
		objectidproto.WithNormalize(trogonobjectid.NormalizeUUID),
		objectidproto.WithValidate(trogonobjectid.ValidateUUID),
		objectidproto.WithGenerate(trogonobjectid.GenerateUUIDv7),
	)
	require.NoError(t, err)

	assert.Equal(t, trogonobjectid.ObjectType("tkt"), def.ObjectType)
	assert.Equal(t, trogonobjectid.FormatDropPrefix, def.StorageFormat)
	assert.Equal(t, trogonobjectid.FormatDropPrefix, def.JSONFormat)
	normalized, err := def.Normalize("0199E3A1-7C2B-7D4E-9F10-2A3B4C5D6E7F")
	require.NoError(t, err)
	assert.Equal(t, "0199e3a1-7c2b-7d4e-9f10-2a3b4c5d6e7f", normalized)
	require.ErrorIs(t, def.Validate("nope"), trogonobjectid.ErrInvalidUUID)
	generated, err := def.Generate()
	require.NoError(t, err)
	require.NoError(t, trogonobjectid.ValidateUUID(generated))

	t.Run("options do not change the annotation fields", func(t *testing.T) {
		def, err := objectidproto.Definition(enum.New(valueOrder), objectidproto.WithStorageFormat(trogonobjectid.FormatDropPrefix))
		require.NoError(t, err)
		assert.Equal(t, trogonobjectid.ObjectType("order"), def.ObjectType)
		assert.Equal(t, trogonobjectid.Separator("#"), def.Separator)
	})
}
