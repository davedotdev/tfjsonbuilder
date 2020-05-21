#### TFJSONBuilder

This project loads the Terraform JSON Schema and creates a TF JSON configuration blob from the schema. This is a tool for a programmatic tool-chain.

#### Usage

```bash
go get github.com/davedotdev/tfjsonbuilder
cd $GOPATH/src/github.com/davedotdev/tfjsonbuilder
go mod download
go build
```

I've included two schemas for testing purposes and there are two main modes of operation. Print everything and JSON output only.

You should be able to create a file named `thing.tf.json` and use the output of this tool (specifically `schemaexample1.json`) as input for another project under this GitHub account called: `canonicaltfprovider` found (here)[https://github.com/davedotdev/canonicaltfprovider]. This provider does nothing other than exercise various input mechanisms (which I wrote coincidentally to create this project).

__Print Everything__
`./tfjsonbuilder -file schemaexample1.json`

```bash
# Output
resource_schema: canonical
attributes: float1, schema: number
attributes: id, schema: string
attributes: int1, schema: number
attributes: liststring9, schema: list of string
attributes: map1, schema: map of string
attributes: string1, schema: string
attributes: bool1, schema: bool
nested block: listset3, type: list
	nested block attribute: string3, schema: string
	nested block attribute: bool3, schema: bool
	nested block attribute: float3, schema: number
	nested block attribute: int3, schema: number
nested block: set2, type: set
	nested block attribute: bool2, schema: bool
	nested block attribute: float2, schema: number
	nested block attribute: int2, schema: number
	nested block attribute: string2, schema: string
nested block: setnested4, type: set
	nested block attribute: float4, schema: number
	nested block attribute: int4, schema: number
	nested block attribute: string4, schema: string
	nested block attribute: bool4, schema: bool
	nested nested block: listset8, type: list
		nested block attribute: string8, schema: string
		nested block attribute: bool8, schema: bool
		nested block attribute: float8, schema: number
		nested block attribute: int8, schema: number
	nested nested block: set5, type: set
		nested block attribute: bool5, schema: bool
		nested block attribute: float5, schema: number
		nested block attribute: int5, schema: number
		nested block attribute: string5, schema: string
		nested nested block: listset7, type: list
			nested block attribute: bool7, schema: bool
			nested block attribute: float7, schema: number
			nested block attribute: int7, schema: number
			nested block attribute: string7, schema: string
		nested nested block: set6, type: set
			nested block attribute: bool6, schema: bool
			nested block attribute: float6, schema: number
			nested block attribute: int6, schema: number
			nested block attribute: string6, schema: string

Valid JSON: {"resource":{"canonical":{"config-group-name":{"float1": 0, "int1": 2, "liststring9": ["bar3", "barbar3"], "map1": {"foo4": "bar4", "foofoo4": "barbar4"}, "string1": "foo5", "bool1": false, "listset3": [{"string3": "foo7", "bool3": false, "float3": 9, "int3": 10}], "set2": {"bool2": false, "float2": 12, "int2": 13, "string2": "foo14"}, "setnested4": {"float4": 15, "int4": 16, "string4": "foo17", "bool4": false, "listset8": [{"string8": "foo19", "bool8": false, "float8": 21, "int8": 22}], "set5": {"bool5": false, "float5": 24, "int5": 25, "string5": "foo26", "listset7": [{"bool7": false, "float7": 28, "int7": 29, "string7": "foo30"}], "set6": {"bool6": false, "float6": 32, "int6": 33, "string6": "foo34"}}}}}}}
```

__JSON Only__
`./tfjsonbuilder -file schemaexample1.json -getJSON`

```bash
# Output
{"resource":{"canonical":{"config-group-name":{"bool1": false, "float1": 1, "int1": 3, "liststring9": ["bar4", "barbar4"], "map1": {"foo5": "bar5", "foofoo5": "barbar5"}, "string1": "foo6", "set2": {"bool2": false, "float2": 8, "int2": 9, "string2": "foo10"}, "setnested4": {"bool4": false, "float4": 12, "int4": 13, "string4": "foo14", "listset8": [{"bool8": false, "float8": 16, "int8": 17, "string8": "foo18"}], "set5": {"bool5": false, "float5": 20, "int5": 21, "string5": "foo22", "listset7": [{"bool7": false, "float7": 24, "int7": 25, "string7": "foo26"}], "set6": {"bool6": false, "float6": 28, "int6": 29, "string6": "foo30"}}}, "listset3": [{"bool3": false, "float3": 32, "int3": 33, "string3": "foo34"}]}}}}
```


