# TrogonObjectID

**TrogonObjectID gives Go services type-safe, prefixed object IDs such as `user_abc-123`.** It is the Go counterpart of the Elixir package [`trogon_object_id`](https://github.com/straw-hat-team/beam-monorepo/tree/main/apps/trogon_object_id) and uses the same string forms.

**Each ID type is a distinct Go type, so the compiler rejects a `UserID` where an `OrderID` is expected.** A raw string ID loses its meaning when it leaves the function that made it.

**The prefix stays in logs even when the database and the wire carry bare UUIDs.** `String`, `fmt` and `log/slog` always emit the full form. JSON and `database/sql` follow the format that you configure.

## Getting Started

```bash
go get github.com/TrogonStack/trogonobjectid
```

### Define an ID type

```go
type invoiceType struct{}

func (invoiceType) Definition() trogonobjectid.Definition {
	return trogonobjectid.Definition{
		ObjectType:    "inv",
		StorageFormat: trogonobjectid.FormatDropPrefix,
		JSONFormat:    trogonobjectid.FormatDropPrefix,
		Validate:      trogonobjectid.ValidateUUID,
		Generate:      trogonobjectid.GenerateUUIDv7,
	}
}

type InvoiceID = trogonobjectid.ID[invoiceType]
```

### Use it

```go
id, err := trogonobjectid.Generate[invoiceType]()
id, err = trogonobjectid.New[invoiceType]("0199e3a1-7c2b-7d4e-9f10-2a3b4c5d6e7f")
id, err = trogonobjectid.Parse[invoiceType]("inv_0199e3a1-7c2b-7d4e-9f10-2a3b4c5d6e7f")

id.String()  // "inv_0199e3a1-..."  (fmt and slog use this form too)
id.Bare()    // "0199e3a1-..."
id.Storage() // "0199e3a1-..."      (StorageFormat)

slog.Info("charged", "invoice_id", id) // invoice_id=inv_0199e3a1-...
```

### Unions

A `Union` holds one ID of any of several types and picks the type by prefix:

```go
type principalUnion struct{}

func (principalUnion) Members() []trogonobjectid.MemberType {
	return []trogonobjectid.MemberType{
		trogonobjectid.MemberOf[userType](),
		trogonobjectid.MemberOf[systemType](),
	}
}

type PrincipalID = trogonobjectid.Union[principalUnion]

p, err := trogonobjectid.ParseUnion[principalUnion]("system_cron")
switch m := p.Member().(type) {
case SystemID:
	// m.Bare() == "cron"
}
```

A union always encodes the full form, in JSON, SQL and Spanner too, because the prefix is the only thing that tells the member types apart. When prefixes overlap, such as `acme_` and `acme_admin_`, the longest prefix wins.

### Read the prefix from a proto annotation

Package `objectidproto` reads `object_type` and `separator` from the `trogon.object_id.v1alpha1.enum_value` annotation on an enum value, so every language that uses the proto shares one prefix:

```go
var ticketDefinition = func() trogonobjectid.Definition {
	def := objectidproto.MustDefinition(acmev1.ObjectType_OBJECT_TYPE_TICKET)
	def.Validate = trogonobjectid.ValidateUUID
	return def
}()

type ticketType struct{}

func (ticketType) Definition() trogonobjectid.Definition { return ticketDefinition }
```

Keep the result in a package-level variable, because `Definition` methods run on every call.

## Options

| Field | Default | Effect |
| --- | --- | --- |
| `ObjectType` | required | Prefix without the separator, for example `"user"` |
| `Separator` | `"_"` | Goes between the object type and the bare value |
| `StorageFormat` | `FormatFull` | Form for `Storage`, `FromStorage`, `Value` and `Scan` |
| `JSONFormat` | `FormatFull` | Form for `MarshalJSON` and `UnmarshalJSON` |
| `Validate` | none | Checks the bare value; use `ValidateUUID`, `ValidateInteger` or your own function. `ValidateUUID` accepts only the lowercase hyphenated form |
| `Generate` | none | Makes a bare value for `Generate`; use `GenerateUUIDv7`, `GenerateUUIDv4` or your own function |

## Emitted forms

| Method | Emits | Controlled by |
| --- | --- | --- |
| `String`, `fmt` verbs | Full | Always full |
| `LogValue` (`log/slog`) | Full | Always full |
| `MarshalText`, `UnmarshalText` | Full | Always full |
| `MarshalJSON`, `UnmarshalJSON` | Full or bare | `JSONFormat` |
| `Storage`, `FromStorage` | Full or bare | `StorageFormat` |
| `Value`, `Scan` (`database/sql`) | Full or bare | `StorageFormat` |
| `EncodeSpanner`, `DecodeSpanner` (Cloud Spanner) | Full or bare; unions always full | `StorageFormat` |

`LogValue` exists because `slog.JSONHandler` otherwise calls `MarshalJSON`, which logs the bare value when `JSONFormat` is `FormatDropPrefix`. `MarshalText` always uses the full form because JSON map keys and flag values must parse back without other context.

Decoders accept only the form that the matching encoder emits, so each encoding round trips. The zero ID means "absent": it encodes as SQL `NULL`, Spanner `NULL` and JSON `null`, and the JSON `omitzero` tag option skips it.

The Spanner methods implement `spanner.Encoder` and `spanner.Decoder` by method set, so the package does not import the Spanner client.
