// Package trogonobjectid provides type-safe, prefixed object IDs such as "user_abc-123".
// It is the Go counterpart of the Elixir package trogon_object_id and uses the same
// string forms, so IDs read the same in both languages.
//
// # Defining a type
//
// Each ID type is an ID with its own marker type, so the compiler rejects a UserID
// where an OrderID is expected:
//
//	type invoiceType struct{}
//
//	func (invoiceType) Definition() trogonobjectid.Definition {
//		return trogonobjectid.Definition{
//			ObjectType:    "inv",
//			StorageFormat: trogonobjectid.FormatDropPrefix,
//			JSONFormat:    trogonobjectid.FormatDropPrefix,
//			Validate:      trogonobjectid.ValidateUUID,
//			Generate:      trogonobjectid.GenerateUUIDv7,
//		}
//	}
//
//	type InvoiceID = trogonobjectid.ID[invoiceType]
//
// # Emitted forms
//
// An ID has a full form, "inv_0199e3a1-...", and a bare form, "0199e3a1-...".
//
//	Method                          Emits             Controlled by
//	------                          -----             -------------
//	String, fmt verbs               Full              Always full
//	LogValue (log/slog)             Full              Always full
//	MarshalText, UnmarshalText      Full              Always full
//	MarshalJSON, UnmarshalJSON      Full or bare      Definition.JSONFormat
//	Storage, FromStorage            Full or bare      Definition.StorageFormat
//	Value, Scan (database/sql)      Full or bare      Definition.StorageFormat
//
// Logs always carry the full form, so a log line tells you which kind of ID it holds even
// when the database and the wire carry bare UUIDs. Decoding accepts only the form that the
// matching encoder emits, so every encoding round trips.
//
// # Unions
//
// A Union holds one ID of any of several member types and picks the member type by prefix.
// It differs from the Elixir UnionObjectId in two ways. When prefixes overlap, such as
// "acme_" and "acme_admin_", the longest prefix wins. JSON always uses the full form,
// even for a member that drops the prefix, because without the prefix the union cannot
// tell its member types apart when it decodes.
//
// # Zero value
//
// The zero ID means "absent". It encodes as SQL NULL and JSON null, String returns "",
// and IsZero reports true, so the json "omitzero" tag option skips it.
package trogonobjectid
