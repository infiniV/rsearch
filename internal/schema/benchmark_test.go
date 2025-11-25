package schema

import (
	"testing"
)

// BenchmarkResolveField benchmarks basic field resolution
func BenchmarkResolveField(b *testing.B) {
	fields := map[string]Field{
		"productCode": {Type: TypeText, Indexed: true},
		"productName": {Type: TypeText, Indexed: true},
		"region":      {Type: TypeText, Indexed: true},
		"status":      {Type: TypeText, Indexed: true},
		"price":       {Type: TypeFloat, Indexed: true},
	}

	options := SchemaOptions{
		NamingConvention: "none",
		StrictFieldNames: true,
	}

	schema := NewSchema("benchmark", fields, options)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _, err := schema.ResolveField("productCode")
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkResolveFieldWithNaming benchmarks field resolution with naming convention
func BenchmarkResolveFieldWithNaming(b *testing.B) {
	fields := map[string]Field{
		"productCode": {Type: TypeText, Indexed: true},
		"productName": {Type: TypeText, Indexed: true},
		"region":      {Type: TypeText, Indexed: true},
		"status":      {Type: TypeText, Indexed: true},
		"price":       {Type: TypeFloat, Indexed: true},
	}

	options := SchemaOptions{
		NamingConvention: "snake_case",
		StrictFieldNames: false,
	}

	schema := NewSchema("benchmark", fields, options)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _, err := schema.ResolveField("productCode")
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkResolveFieldCaseInsensitive benchmarks case-insensitive field resolution
func BenchmarkResolveFieldCaseInsensitive(b *testing.B) {
	fields := map[string]Field{
		"productCode": {Type: TypeText, Indexed: true},
		"productName": {Type: TypeText, Indexed: true},
		"region":      {Type: TypeText, Indexed: true},
		"status":      {Type: TypeText, Indexed: true},
		"price":       {Type: TypeFloat, Indexed: true},
	}

	options := SchemaOptions{
		NamingConvention: "snake_case",
		StrictFieldNames: false, // Case-insensitive
	}

	schema := NewSchema("benchmark", fields, options)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _, err := schema.ResolveField("PRODUCTCODE")
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkResolveFieldWithAlias benchmarks alias resolution
func BenchmarkResolveFieldWithAlias(b *testing.B) {
	fields := map[string]Field{
		"productCode": {
			Type:    TypeText,
			Indexed: true,
			Aliases: []string{"code", "product_id", "sku"},
		},
	}

	options := SchemaOptions{
		NamingConvention: "snake_case",
		StrictFieldNames: false,
	}

	schema := NewSchema("benchmark", fields, options)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _, err := schema.ResolveField("sku")
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkSchemaValidation benchmarks schema validation
func BenchmarkSchemaValidation(b *testing.B) {
	fields := map[string]Field{
		"productCode": {Type: TypeText, Indexed: true},
		"productName": {Type: TypeText, Indexed: true},
		"region":      {Type: TypeText, Indexed: true},
		"status":      {Type: TypeText, Indexed: true},
		"price":       {Type: TypeFloat, Indexed: true},
		"quantity":    {Type: TypeInteger, Indexed: true},
		"active":      {Type: TypeBoolean},
		"createdAt":   {Type: TypeDateTime},
		"metadata":    {Type: TypeJSON},
	}

	options := SchemaOptions{
		NamingConvention: "snake_case",
		StrictFieldNames: false,
		DefaultField:     "productName",
		EnabledFeatures: EnabledFeatures{
			Fuzzy:     true,
			Proximity: true,
			Regex:     true,
		},
	}

	schema := NewSchema("benchmark", fields, options)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		err := ValidateSchema(schema)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkNewSchema benchmarks schema creation with cache building
func BenchmarkNewSchema(b *testing.B) {
	fields := map[string]Field{
		"productCode": {Type: TypeText, Indexed: true},
		"productName": {Type: TypeText, Indexed: true},
		"region":      {Type: TypeText, Indexed: true},
		"status":      {Type: TypeText, Indexed: true},
		"price":       {Type: TypeFloat, Indexed: true},
		"quantity":    {Type: TypeInteger, Indexed: true},
		"active":      {Type: TypeBoolean},
		"createdAt":   {Type: TypeDateTime},
		"metadata":    {Type: TypeJSON},
	}

	options := SchemaOptions{
		NamingConvention: "snake_case",
		StrictFieldNames: false,
		DefaultField:     "productName",
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = NewSchema("benchmark", fields, options)
	}
}

// BenchmarkResolveFieldLargeSchema benchmarks field resolution with large schema
func BenchmarkResolveFieldLargeSchema(b *testing.B) {
	// Create schema with 100 fields
	fields := make(map[string]Field)
	fieldNames := make([]string, 100)
	for i := 0; i < 100; i++ {
		fieldName := "field" + string(rune('A'+i%26)) + string(rune('0'+i/26))
		fieldNames[i] = fieldName
		fields[fieldName] = Field{
			Type:    TypeText,
			Indexed: i%3 == 0,
			Aliases: []string{"alias" + fieldName},
		}
	}

	options := SchemaOptions{
		NamingConvention: "snake_case",
		StrictFieldNames: false,
	}

	schema := NewSchema("large", fields, options)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		fieldName := fieldNames[i%100]
		_, _, err := schema.ResolveField(fieldName)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkToSnakeCase benchmarks snake_case conversion
func BenchmarkToSnakeCase(b *testing.B) {
	testCases := []string{
		"productCode",
		"productName",
		"HTTPResponseCode",
		"JSONData",
		"userID",
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		input := testCases[i%len(testCases)]
		_ = ToSnakeCase(input)
	}
}

// BenchmarkToCamelCase benchmarks camelCase conversion
func BenchmarkToCamelCase(b *testing.B) {
	testCases := []string{
		"product_code",
		"product_name",
		"http_response_code",
		"json_data",
		"user_id",
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		input := testCases[i%len(testCases)]
		_ = ToCamelCase(input)
	}
}

// BenchmarkRegistryOperations benchmarks schema registry operations
func BenchmarkRegistryOperations(b *testing.B) {
	fields := map[string]Field{
		"field1": {Type: TypeText},
		"field2": {Type: TypeInteger},
	}
	options := SchemaOptions{}

	b.Run("Register", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			b.StopTimer()
			registry := NewRegistry()
			schema := NewSchema("test", fields, options)
			b.StartTimer()

			err := registry.Register(schema)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("Get", func(b *testing.B) {
		registry := NewRegistry()
		schema := NewSchema("test", fields, options)
		err := registry.Register(schema)
		if err != nil {
			b.Fatal(err)
		}

		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			_, err := registry.Get("test")
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("Exists", func(b *testing.B) {
		registry := NewRegistry()
		schema := NewSchema("test", fields, options)
		err := registry.Register(schema)
		if err != nil {
			b.Fatal(err)
		}

		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			_ = registry.Exists("test")
		}
	})
}
