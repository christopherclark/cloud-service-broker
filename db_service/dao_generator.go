// Copyright 2018 the Service Broker Project Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// +build ignore

// This program generates dao.go It can be invoked by running
// go generate
package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"text/template"
	"time"
)

func main() {
	models := []crudModel{
		{
			Type:            "ServiceInstanceDetails",
			PrimaryKeyType:  "string",
			PrimaryKeyField: "id",
			ExampleFields: map[string]interface{}{
				"Name":             "Hello",
				"Location":         "loc",
				"Url":              "https://google.com",
				"OtherDetails":     `{"some":["json","blob","here"]}`,
				"ServiceId":        "123-456-7890",
				"PlanId":           "planid",
				"SpaceGuid":        "0000-0000-0000",
				"OrganizationGuid": "1111-1111-1111",
			},
		},
		{
			Type:            "CloudOperation",
			PrimaryKeyType:  "uint",
			PrimaryKeyField: "id",
			Keys: []fieldList{
				{
					{Type: "string", Column: "service_instance_id"},
				},
			},
			ExampleFields: map[string]interface{}{
				"Name":              "cloud-operation-name",
				"Status":            "DELETED",
				"OperationType":     "Delete",
				"ErrorMessage":      "<empty>",
				"InsertTime":        "1970-01-01T01:01:01Z",
				"StartTime":         "1980-01-01T01:01:01Z",
				"TargetId":          "some-uuid-here",
				"TargetLink":        "https://cloud.google.com/my/target/instance",
				"ServiceId":         "1111-1111-1111",
				"ServiceInstanceId": "2222-2222-2222",
			},
		},
		{
			Type:            "ServiceBindingCredentials",
			PrimaryKeyType:  "uint",
			PrimaryKeyField: "id",
			Keys: []fieldList{
				{
					{Type: "string", Column: "service_instance_id"},
					{Type: "string", Column: "binding_id"},
				},
				{
					{Type: "string", Column: "binding_id"},
				},
			},
			ExampleFields: map[string]interface{}{
				"ServiceId":         "1111-1111-1111",
				"ServiceInstanceId": "2222-2222-2222",
				"BindingId":         "0000-0000-0000",
				"OtherDetails":      `{"some":["json","blob","here"]}`,
			},
		},
		{
			Type:            "ProvisionRequestDetails",
			PrimaryKeyType:  "uint",
			PrimaryKeyField: "id",
			Keys: []fieldList{
				{
					{Type: "string", Column: "service_instance_id"},
				},
			},
			ExampleFields: map[string]interface{}{
				"ServiceInstanceId": "2222-2222-2222",
				"RequestDetails":    `{"some":["json","blob","here"]}`,
			},
		},
		{
			Type:            "PlanDetailsV1",
			PrimaryKeyType:  "string",
			PrimaryKeyField: "id",
			Keys: []fieldList{
				{
					{Type: "string", Column: "service_id"},
					{Type: "string", Column: "name"},
				},
			},
			ExampleFields: map[string]interface{}{
				"ServiceId": "2222-2222-2222",
				"Name":      "service-name",
				"Features":  `{"some":["json","blob","here"]}`,
			},
		},
	}

	for i, model := range models {
		pk := fieldList{{Type: model.PrimaryKeyType, Column: model.PrimaryKeyField}}
		models[i].Keys = append(model.Keys, pk)
	}

	createDao(models)
	createDaoTest(models)
}

func createDao(models []crudModel) {
	f, err := os.Create("dao.go")
	die(err)
	defer f.Close()

	daoTemplate.Execute(f, struct {
		Timestamp time.Time
		Models    []crudModel
	}{
		Timestamp: time.Now(),
		Models:    models,
	})
}

func createDaoTest(models []crudModel) {
	f, err := os.Create("dao_test.go")
	die(err)
	defer f.Close()

	daoTestTemplate.Execute(f, struct {
		Timestamp time.Time
		Models    []crudModel
	}{
		Timestamp: time.Now(),
		Models:    models,
	})
}

func die(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

type crudModel struct {
	Type            string
	PrimaryKeyType  string
	PrimaryKeyField string
	ExampleFields   map[string]interface{}
	Keys            []fieldList
}

type fieldList []crudField

// WhereClause generates a gorm Where clause function and arguments.
func (fl fieldList) WhereClause() string {
	var cols []string

	for _, field := range fl {
		cols = append(cols, field.Column+" = ?")
	}

	colf := strings.Join(cols, " AND ")

	return fmt.Sprintf("Where(%q, %s)", colf, fl.CallParams())
}

// FuncName gets the concatenated function argument descriptor in a
// Spring Data Repository style i.e. <Verb>By(<Field>And)*(<Field>).
func (fl fieldList) FuncName() string {
	out := ""
	for i, field := range fl {
		if i == 0 {
			out += "By"
		} else {
			out += "And"
		}

		out += snakeToProper(field.Column)
	}

	return out
}

// Args creates field type argument pairs to use when defining function params.
func (fl fieldList) Args() string {
	var args []string

	for _, field := range fl {
		arg := fmt.Sprintf("%s %s", snakeToCamel(field.Column), field.Type)
		args = append(args, arg)
	}

	return strings.Join(args, ", ")
}

// CallParams gets the camelCase of the columns in the field list joined by commas
// like you'd need for calling the function generated by Args().
func (fl fieldList) CallParams() string {
	var args []string

	for _, field := range fl {
		args = append(args, snakeToCamel(field.Column))
	}

	return strings.Join(args, ", ")
}

// ExampleArgs is like CallParams, but with the properties on the struct parent.
// e.g. ExampleArgs(foo) would return "foo.ID, foo.PropertyA, foo.PropertyB"
func (fl fieldList) ExampleArgs(parent string) string {
	var exampleArgs []string
	for _, field := range fl {
		fieldName := snakeToProper(field.Column)
		if fieldName == "Id" {
			fieldName = "ID"
		}
		exampleArgs = append(exampleArgs, parent+"."+fieldName)
	}

	return strings.Join(exampleArgs, ", ")
}

type crudField struct {
	Type   string
	Column string
}

func snakeToCamel(in string) string {
	proper := snakeToProper(in)

	return strings.ToLower(proper[0:1]) + proper[1:]
}

func snakeToProper(in string) string {
	out := ""
	for _, word := range strings.Split(in, "_") {
		if len(word) == 0 {
			continue
		}

		out += strings.ToUpper(word[0:1]) + strings.ToLower(word[1:])
	}

	return out
}

func functionNameGen(operation string, objType string, keys ...string) string {
	out := operation + objType
	for i, key := range keys {
		if i == 0 {
			out += "By"
		} else {
			out += "And"
		}

		out += snakeToProper(key)
	}

	return out
}

var daoTemplate = template.Must(template.New("").Funcs(
	template.FuncMap{
		"funcName": functionNameGen,
	},
).Parse(`// Copyright {{ .Timestamp.Year }} the Service Broker Project Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Code generated by go generate; DO NOT EDIT.

package db_service

import (
	"context"

	"github.com/GoogleCloudPlatform/gcp-service-broker/brokerapi/brokers/models"
)

{{- range .Models}}

{{- $type := .Type}}
{{- range $idx, $key := .Keys }}

{{ $fn := (print "Count" $type $key.FuncName)}}
// {{$fn}} gets the count of {{$type}} by its key ({{$key.CallParams}}) in the datastore (0 or 1)
func {{$fn}}(ctx context.Context, {{ $key.Args }}) (int, error) { return defaultDatastore().{{$fn}}(ctx, {{$key.CallParams}}) }
func (ds *SqlDatastore) {{$fn}}(ctx context.Context, {{ $key.Args }}) (int, error) {
	var count int
	err := ds.db.Model(&models.{{$type}}{}).{{ $key.WhereClause }}.Count(&count).Error
	return count, err
}
{{- end }}

// {{funcName "Create" .Type}} creates a new record in the database and assigns it a primary key.
func {{funcName "Create" .Type}}(ctx context.Context, object *models.{{.Type}}) error { return defaultDatastore().{{funcName "Create" .Type}}(ctx, object) }
func (ds *SqlDatastore) Create{{.Type}}(ctx context.Context, object *models.{{.Type}}) error {
	return ds.db.Create(object).Error
}

// {{funcName "Save" .Type}} updates an existing record in the database.
func {{funcName "Save" .Type}}(ctx context.Context, object *models.{{.Type}}) error { return defaultDatastore().{{funcName "Save" .Type}}(ctx, object) }
func (ds *SqlDatastore) {{funcName "Save" .Type}}(ctx context.Context, object *models.{{.Type}}) error {
	return ds.db.Save(object).Error
}

{{- $type := .Type}}
{{ range $idx, $key := .Keys -}}
{{ $fn := (print "Delete" $type $key.FuncName) -}}
// {{$fn}} soft-deletes the record by its key ({{$key.CallParams}}).
func {{$fn}}(ctx context.Context, {{ $key.Args }}) error { return defaultDatastore().{{$fn}}(ctx, {{$key.CallParams}}) }
func (ds *SqlDatastore) {{$fn}}(ctx context.Context, {{ $key.Args }}) error {
	return ds.db.{{ $key.WhereClause }}.Delete(&models.{{$type}}{}).Error
}

{{ end }}

// Delete{{.Type}} soft-deletes the record.
func {{funcName "Delete" .Type}}(ctx context.Context, record *models.{{.Type}}) error { return defaultDatastore().{{funcName "Delete" .Type}}(ctx, record) }
func (ds *SqlDatastore) {{funcName "Delete" .Type}}(ctx context.Context, record *models.{{.Type}}) error {
	return ds.db.Delete(record).Error
}

{{- $type := .Type}}
{{ range $idx, $key := .Keys -}}

{{ $fn := (print "Get" $type $key.FuncName) -}}
// {{$fn}} gets an instance of {{$type}} by its key ({{$key.CallParams}}).
func {{$fn}}(ctx context.Context, {{ $key.Args }}) (*models.{{$type}}, error) { return defaultDatastore().{{$fn}}(ctx, {{$key.CallParams}}) }
func (ds *SqlDatastore) {{$fn}}(ctx context.Context, {{ $key.Args }}) (*models.{{$type}}, error) {
	record := models.{{$type}}{}
	if err := ds.db.{{ $key.WhereClause }}.First(&record).Error; err != nil {
		return nil, err
	}

	return &record, nil
}

{{ $fn := (print "CheckDeleted" $type $key.FuncName) -}}
// {{$fn}} checks to see if an instance of {{$type}} was soft deleted by its key ({{$key.CallParams}}).
func {{$fn}}(ctx context.Context, {{ $key.Args }}) (bool, error) { return defaultDatastore().{{$fn}}(ctx, {{$key.CallParams}}) }
func (ds *SqlDatastore) {{$fn}}(ctx context.Context, {{ $key.Args }}) (bool, error) {
	record := models.{{$type}}{}
	if err := ds.db.Unscoped().{{ $key.WhereClause }}.First(&record).Error; err != nil {
		return false, err
	}

	return record.DeletedAt != nil, nil
}

{{ end }}

{{- end }}
`))

var daoTestTemplate = template.Must(template.New("").Funcs(
	template.FuncMap{
		"funcName": functionNameGen,
	},
).Parse(`// Copyright {{ .Timestamp.Year }} the Service Broker Project Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Code generated by go generate; DO NOT EDIT.

package db_service

import (
	"context"
	"testing"
	"time"

	"github.com/GoogleCloudPlatform/gcp-service-broker/brokerapi/brokers/models"
	"github.com/jinzhu/gorm"
)

func newInMemoryDatastore(t *testing.T) *SqlDatastore {
	testDb, err := gorm.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Error opening test database %s", err)
	}

	{{range .Models}}testDb.CreateTable(models.{{.Type}}{})
	{{end}}
	return &SqlDatastore{db: testDb}
}


{{- range .Models}}

func create{{.Type}}Instance() ({{.PrimaryKeyType}}, models.{{.Type}}) {
	testPk := {{.PrimaryKeyType}}(42)

	instance := models.{{.Type}}{}
	instance.ID = testPk
{{range $k, $v := .ExampleFields}}	instance.{{$k}} = {{ printf "%#v" $v}}
{{end}}

	return testPk, instance
}

func ensure{{.Type}}FieldsMatch(t *testing.T, expected, actual *models.{{.Type}}) {
{{range $k, $v := .ExampleFields}}
	if expected.{{$k}} != actual.{{$k}} {
		t.Errorf("Expected field {{$k}} to be %#v, got %#v", expected.{{$k}}, actual.{{$k}})
	}
{{end}}
}

func TestSqlDatastore_{{.Type}}DAO(t *testing.T) {
	ds := newInMemoryDatastore(t)
	testPk, instance := create{{.Type}}Instance()
	testCtx := context.Background()

	// on startup, there should be no objects to find or delete
	if count, err := ds.{{funcName "Count" .Type .PrimaryKeyField}}(testCtx, testPk); count != 0 || err != nil {
		t.Fatalf("Expected count to be 0 and error to be nil got count: %d, err: %v", count, err)
	}

	if _, err := ds.{{funcName "Get" .Type .PrimaryKeyField}}(testCtx, testPk); err != gorm.ErrRecordNotFound {
		t.Errorf("Expected an ErrRecordNotFound trying to get non-existing PK got %v", err)
	}

	if _, err := ds.{{funcName "CheckDeleted" .Type .PrimaryKeyField}}(testCtx, testPk); err != gorm.ErrRecordNotFound {
		t.Errorf("Expected an ErrRecordNotFound trying to check deletion status of a non-existing PK got %v", err)
	}

	// Should be able to create the item
	beforeCreation := time.Now()
	if err := ds.{{funcName "Create" .Type}}(testCtx, &instance); err != nil {
		t.Errorf("Expected to be able to create the item %#v, got error: %s", instance, err)
	}
	afterCreation := time.Now()

	// after creation we should be able to get the item
	ret, err := ds.{{funcName "Get" .Type .PrimaryKeyField}}(testCtx, testPk)
	if err != nil {
		t.Errorf("Expected no error trying to get saved item, got: %v", err)
	}

	if ret.CreatedAt.Before(beforeCreation) || ret.CreatedAt.After(afterCreation) {
		t.Errorf("Expected creation time to be between  %v and %v got %v", beforeCreation, afterCreation, ret.CreatedAt)
	}

	if !ret.UpdatedAt.Equal(ret.CreatedAt) {
		t.Errorf("Expected initial update time to equal creation time, but got update: %v, create: %v", ret.UpdatedAt, ret.CreatedAt)
	}

	// Ensure non-gorm fields were deserialized correctly
	ensure{{.Type}}FieldsMatch(t, &instance, ret)

	// we should be able to update the item and it will have a new updated time
	if err := ds.{{funcName "Save" .Type}}(testCtx, ret); err != nil {
		t.Errorf("Expected no error trying to get update %#v , got: %v", ret, err)
	}

	if !ret.UpdatedAt.After(ret.CreatedAt) {
		t.Errorf("Expected update time to be after create time after update, got update: %#v create: %#v", ret.UpdatedAt, ret.CreatedAt)
	}

	// after deleting the item we should not be able to get it
	deleted, err := ds.{{funcName "CheckDeleted" .Type .PrimaryKeyField}}(testCtx, testPk)
	if err != nil {
		t.Errorf("Expected no error when checking if a non-deleted thing was deleted")
	}
	if deleted {
		t.Errorf("Expected a non-deleted instance to not be marked as deleted but it was.")
	}

	if err := ds.{{funcName "Delete" .Type .PrimaryKeyField}}(testCtx, testPk); err != nil {
		t.Errorf("Expected no error when deleting by pk got: %v", err)
	}

	// we should be able to see that it was soft-deleted
	deleted, err = ds.{{funcName "CheckDeleted" .Type .PrimaryKeyField}}(testCtx, testPk)
	if err != nil {
		t.Errorf("Expected no error when checking if a non-deleted thing was deleted")
	}
	if !deleted {
		t.Errorf("Expected a deleted instance to marked as deleted but it was not.")
	}

	// after deleting the item we should not be able to get it
	if _, err := ds.{{funcName "Get" .Type .PrimaryKeyField}}(testCtx, testPk); err != gorm.ErrRecordNotFound {
		t.Errorf("Expected ErrRecordNotFound after delete but got %v", err)
	}
}

{{- $type := .Type}}{{ $pk := .PrimaryKeyField }}
{{ range $idx, $key := .Keys -}}
{{ $fn := (print "Get" $type $key.FuncName) -}}
func TestSqlDatastore_{{$fn}}(t *testing.T) {
	ds := newInMemoryDatastore(t)
	_, instance := create{{$type}}Instance()
	testCtx := context.Background()

	if _, err := ds.{{$fn}}(testCtx, {{$key.ExampleArgs "instance"}}); err != gorm.ErrRecordNotFound {
		t.Errorf("Expected an ErrRecordNotFound trying to get non-existing record got %v", err)
	}

	beforeCreation := time.Now()
	if err := ds.{{funcName "Create" $type}}(testCtx, &instance); err != nil {
		t.Errorf("Expected to be able to create the item %#v, got error: %s", instance, err)
	}
	afterCreation := time.Now()

	// after creation we should be able to get the item
	ret, err := ds.{{$fn}}(testCtx, {{$key.ExampleArgs "instance"}})
	if err != nil {
		t.Errorf("Expected no error trying to get saved item, got: %v", err)
	}

	if ret.CreatedAt.Before(beforeCreation) || ret.CreatedAt.After(afterCreation) {
		t.Errorf("Expected creation time to be between  %v and %v got %v", beforeCreation, afterCreation, ret.CreatedAt)
	}

	if !ret.UpdatedAt.Equal(ret.CreatedAt) {
		t.Errorf("Expected initial update time to equal creation time, but got update: %v, create: %v", ret.UpdatedAt, ret.CreatedAt)
	}

	// Ensure non-gorm fields were deserialized correctly
	ensure{{$type}}FieldsMatch(t, &instance, ret)
}

{{ $fn := (print "CheckDeleted" $type $key.FuncName) -}}
func TestSqlDatastore_{{$fn}}(t *testing.T) {
	ds := newInMemoryDatastore(t)
	_, instance := create{{$type}}Instance()
	testCtx := context.Background()

	if _, err := ds.{{$fn}}(testCtx, {{$key.ExampleArgs "instance"}}); err != gorm.ErrRecordNotFound {
		t.Errorf("Expected an ErrRecordNotFound trying to get non-existing record got %v", err)
	}

	if err := ds.{{funcName "Create" $type}}(testCtx, &instance); err != nil {
		t.Errorf("Expected to be able to create the item %#v, got error: %s", instance, err)
	}

	deleted, err := ds.{{$fn}}(testCtx, {{$key.ExampleArgs "instance"}})
	if err != nil {
		t.Errorf("Expected no error when checking if a non-deleted thing was deleted")
	}
	if deleted {
		t.Errorf("Expected a non-deleted instance to not be marked as deleted but it was.")
	}

	if err := ds.{{funcName "Delete" $type}}(testCtx, &instance); err != nil {
		t.Errorf("Expected no error when deleting by pk got: %v", err)
	}

	// we should be able to see that it was soft-deleted
	deleted, err = ds.{{$fn}}(testCtx, {{$key.ExampleArgs "instance"}})
	if err != nil {
		t.Errorf("Expected no error when checking if a non-deleted thing was deleted")
	}
	if !deleted {
		t.Errorf("Expected a deleted instance to marked as deleted but it was not.")
	}
}

{{ $fn := (print "Count" $type $key.FuncName) -}}
func TestSqlDatastore_{{$fn}}(t *testing.T) {
	ds := newInMemoryDatastore(t)
	_, instance := create{{$type}}Instance()
	testCtx := context.Background()

	// on startup, there should be no objects to find or delete
	if count, err := ds.{{$fn}}(testCtx, {{$key.ExampleArgs "instance"}}); count != 0 || err != nil {
		t.Fatalf("Expected count to be 0 and error to be nil got count: %d, err: %v", count, err)
	}

	if err := ds.{{funcName "Create" $type}}(testCtx, &instance); err != nil {
		t.Errorf("Expected to be able to create the item %#v, got error: %s", instance, err)
	}

	// on startup, there should be no objects to find or delete
	if count, err := ds.{{$fn}}(testCtx, {{$key.ExampleArgs "instance"}}); count != 1 || err != nil {
		t.Fatalf("Expected count to be 1 and error to be nil got count: %d, err: %v", count, err)
	}
}
{{ end }}


{{- end }}
`))