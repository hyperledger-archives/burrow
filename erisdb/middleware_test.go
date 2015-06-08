package erisdb

import (
	"fmt"
	"testing"
	"reflect"
	ep "github.com/androlo/blockchain_rpc/erisdb/pipe"
)

// Test empty query.
func TestEmptyQuery(t *testing.T) {
	arr, err := _parseQuery("")
	if err != nil{
		t.Error(err)
	}
	if arr != nil{
		t.Error(fmt.Errorf("Array should be nil"))
	}
}

// Test no colon separated filter.
func TestQueryNoColonSeparator(t *testing.T) {
	_, err := _parseQuery("test")
	if err == nil{
		t.Fatal("Should detect missing colon.")
	}
}

// Test no colon separated filter and proper filters mixed.
func TestQueryNoColonSeparatorMulti(t *testing.T) {
	_, err := _parseQuery("test + test1:24 + test2")
	if err == nil {
		t.Fatal("Should detect missing colon.")
	}
	// fmt.Println("Error: " + err.Error())
}

// Test how it handles a query with op and value empty.
func TestQueryOpEmptyValueEmpty(t *testing.T) {
	arr, err := _parseQuery("test:")
	if err != nil{
		t.Fatal("Caused an error: " + err.Error())
	}
	if arr == nil{
		t.Fatal(fmt.Errorf("Array should not be nil"))
	}
	if len(arr) != 1 {
		t.Fatal(fmt.Errorf("Array should have one element in it."))
	}
	if !reflect.DeepEqual(arr[0], &ep.FilterData{"test", "==", ""}) {
		t.Error("Error: Does not match.")
	}
}

// Test how it handles a query with an empty op but a proper value.
func TestQueryOpEmptyValue(t *testing.T) {
	arr, err := _parseQuery("test:val")
	if err != nil{
		t.Fatal("Caused an error: " + err.Error())
	}
	if arr == nil{
		t.Fatal(fmt.Errorf("Array should not be nil"))
	}
	if len(arr) != 1 {
		t.Fatal(fmt.Errorf("Array should have one element in it."))
	}
	if !reflect.DeepEqual(arr[0], &ep.FilterData{"test", "==", "val"}) {
		t.Error("Error: Does not match.")
	}
}

// Test the '>' operator.
func TestQueryGT(t *testing.T) {
	arr, err := _parseQuery("test:>33")
	if err != nil{
		t.Fatal("Caused an error: " + err.Error())
	}
	if arr == nil{
		t.Fatal(fmt.Errorf("Array should not be nil"))
	}
	if len(arr) != 1 {
		t.Fatal(fmt.Errorf("Array should have one element in it."))
	}
	if !reflect.DeepEqual(arr[0], &ep.FilterData{"test", ">", "33"}) {
		t.Error("Error: Does not match.")
	}
}

// Test the '<' operator.
func TestQueryLT(t *testing.T) {
	arr, err := _parseQuery("test:<33")
	if err != nil{
		t.Fatal("Caused an error: " + err.Error())
	}
	if arr == nil{
		t.Fatal(fmt.Errorf("Array should not be nil"))
	}
	if len(arr) != 1 {
		t.Fatal(fmt.Errorf("Array should have one element in it."))
	}
	if !reflect.DeepEqual(arr[0], &ep.FilterData{"test", "<", "33"}) {
		t.Error("Error: Does not match.")
	}
}

// Test the '>=' operator.
func TestQueryGTEQ(t *testing.T) {
	arr, err := _parseQuery("test:>=33")
	if err != nil{
		t.Fatal("Caused an error: " + err.Error())
	}
	if arr == nil{
		t.Fatal(fmt.Errorf("Array should not be nil"))
	}
	if len(arr) != 1 {
		t.Fatal(fmt.Errorf("Array should have one element in it."))
	}
	if !reflect.DeepEqual(arr[0], &ep.FilterData{"test", ">=", "33"}) {
		fmt.Printf("Data: %v\n", arr[0])
		t.Error("Error: Does not match.")
	}
}

// Test the '<=' operator.
func TestQueryLTEQ(t *testing.T) {
	arr, err := _parseQuery("test:<=33")
	if err != nil{
		t.Fatal("Caused an error: " + err.Error())
	}
	if arr == nil{
		t.Fatal(fmt.Errorf("Array should not be nil"))
	}
	if len(arr) != 1 {
		t.Fatal(fmt.Errorf("Array should have one element in it."))
	}
	if !reflect.DeepEqual(arr[0], &ep.FilterData{"test", "<=", "33"}) {
		t.Error("Error: Does not match.")
	}
}

// Test the '==' operator.
func TestQueryEQ(t *testing.T) {
	arr, err := _parseQuery("test:==33")
	if err != nil{
		t.Fatal("Caused an error: " + err.Error())
	}
	if arr == nil{
		t.Fatal(fmt.Errorf("Array should not be nil"))
	}
	if len(arr) != 1 {
		t.Fatal(fmt.Errorf("Array should have one element in it."))
	}
	if !reflect.DeepEqual(arr[0], &ep.FilterData{"test", "==", "33"}) {
		t.Error("Error: Does not match.")
	}
}

// Test the '!=' operator.
func TestQueryEQ(t *testing.T) {
	arr, err := _parseQuery("test:!=null")
	if err != nil{
		t.Fatal("Caused an error: " + err.Error())
	}
	if arr == nil{
		t.Fatal(fmt.Errorf("Array should not be nil"))
	}
	if len(arr) != 1 {
		t.Fatal(fmt.Errorf("Array should have one element in it."))
	}
	if !reflect.DeepEqual(arr[0], &ep.FilterData{"test", "!=", "null"}) {
		t.Error("Error: Does not match.")
	}
}

// Test a working range query.
func TestRangeQuery(t *testing.T) {
	arr, err := _parseQuery("test:4..66")
	if err != nil{
		t.Fatal("Caused an error: " + err.Error())
	}
	if arr == nil{
		t.Fatal(fmt.Errorf("Array should not be nil"))
	}
	if len(arr) != 2 {
		t.Fatal(fmt.Errorf("Array should have one element in it."))
	}
	fmt.Printf("%v\n", arr[0])
	fmt.Printf("%v\n", arr[1])
	
	if !reflect.DeepEqual(arr[0], &ep.FilterData{"test", ">=", "4"}) {
		
		t.Error("Error: Does not match.")
	}
	if !reflect.DeepEqual(arr[1], &ep.FilterData{"test", "<=", "66"}) {
		t.Error("Error: Does not match.")
	}
}

// Test a working range-query with wildcards.
func TestRangeQueryWildcards(t *testing.T) {
	arr, err := _parseQuery("test:*..*")
	if err != nil{
		t.Fatal("Caused an error: " + err.Error())
	}
	if arr == nil{
		t.Fatal(fmt.Errorf("Array should not be nil"))
	}
	if len(arr) != 2 {
		t.Fatal(fmt.Errorf("Array should have one element in it."))
	}
	fmt.Printf("%v\n", arr[0])
	fmt.Printf("%v\n", arr[1])
	
	if !reflect.DeepEqual(arr[0], &ep.FilterData{"test", ">=", "min"}) {
		
		t.Error("Error: Does not match.")
	}
	if !reflect.DeepEqual(arr[1], &ep.FilterData{"test", "<=", "max"}) {
		t.Error("Error: Does not match.")
	}
}

// Test a range query with no upper bounds term.
func TestRangeQueryBotchedMax(t *testing.T) {
	_, err := _parseQuery("test:5..")
	if err == nil{
		t.Fatal("Malformed range-query passed")
	}
	fmt.Println(err)
}

// Test a range query with no lower bounds term.
func TestRangeQueryBotchedMin(t *testing.T) {
	_, err := _parseQuery("test:..5")
	if err == nil{
		t.Fatal("Malformed range-query passed")
	}
	fmt.Println(err)
}