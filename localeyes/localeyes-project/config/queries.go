package config

import (
	"fmt"
	"strings"
)

func InsertQuery(tableName string, columns []string) string {
	colNames := strings.Join(columns, ", ")
	placeholders := strings.Repeat("?, ", len(columns))
	placeholders = strings.TrimSuffix(placeholders, ", ")
	query := fmt.Sprintf(Insert, tableName, colNames, placeholders)
	return query
}

func SelectQuery(tableName, condition1, condition2 string, columns []string) string {
	colNames := strings.Join(columns, ", ")
	var query string
	if condition1 == "" && condition2 == "" {
		query = fmt.Sprintf(Select, colNames, tableName)
	}
	if condition1 != "" && condition2 == "" {
		query = fmt.Sprintf(SelectWithCondition, colNames, tableName, condition1)
	}
	if condition1 != "" && condition2 != "" {
		query = fmt.Sprintf(SelectWith2Condition, colNames, tableName, condition1, condition2)
	}
	fmt.Println(query)
	return query
}

func SelectQueryWithValue(tableName string, columns []string) string {
	colNames := strings.Join(columns, ", ")
	query := fmt.Sprintf(SelectWithValues, colNames, tableName)
	return query
}

func CountQuery(tableName string) string {
	query := fmt.Sprintf(Count, tableName)
	return query
}

func DeleteQuery(tableName, condition1, condition2 string) string {
	if condition2 == "" {
		query := fmt.Sprintf(Delete, tableName, condition1)
		return query
	}
	query := fmt.Sprintf(DeleteWith2Condition, tableName, condition1, condition2)
	return query
}

func UpdateQuery(tableName, condition1, condition2 string, columns []string) string {
	setClause := make([]string, len(columns))
	for i, col := range columns {
		setClause[i] = fmt.Sprintf("%s = ?", col)
	}
	setClauseStr := strings.Join(setClause, ", ")
	if condition2 == "" {
		query := fmt.Sprintf(Update, tableName, setClauseStr, condition1)
		return query
	}
	query := fmt.Sprintf(UpdateWith2Condition, tableName, setClauseStr, condition1, condition2)
	return query
}

func UpdateQueryWithValue(tableName, condition1, condition2 string, columns string) string {
	if condition2 == "" {
		query := fmt.Sprintf(Update, tableName, columns, condition1)
		return query
	}
	query := fmt.Sprintf(UpdateWith2Condition, tableName, columns, condition1, condition2)
	return query
}
