package translator

// NewSQLOutput creates a TranslatorOutput for SQL databases.
func NewSQLOutput(whereClause string, params []interface{}, types []string) *TranslatorOutput {
	return &TranslatorOutput{
		Type:           "sql",
		WhereClause:    whereClause,
		Parameters:     params,
		ParameterTypes: types,
	}
}

// NewMongoDBOutput creates a TranslatorOutput for MongoDB.
func NewMongoDBOutput(filter interface{}) *TranslatorOutput {
	return &TranslatorOutput{
		Type:   "mongodb",
		Filter: filter,
	}
}

// NewElasticsearchOutput creates a TranslatorOutput for Elasticsearch.
func NewElasticsearchOutput(query interface{}) *TranslatorOutput {
	return &TranslatorOutput{
		Type:   "elasticsearch",
		Filter: query,
	}
}
