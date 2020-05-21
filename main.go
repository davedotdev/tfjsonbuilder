/*
	This really terrible code creates a canonical JSON Terraform resource from the Terraform provider schema fed to it in a file.

	It works. Is it pretty? No. Hell no.

	I need more time to figure out the Hashicorp way of building resources and dumping them out to JSON. Lots of the functions and structs they use are
	not exported from the packge >_< so had to do it this way. Le sigh.

	David Gee, 2020, dave.dev
*/

package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"

	tfjson "github.com/hashicorp/terraform-json"
)

var (
	ctr             int
	printEverything bool
)

const (
	attributeStrPattern    = `%q: %q`
	attributeNonStrPattern = `%q: %v`
	cfgGroupName           = "config-group-name"
)

// JSON is a simple func to test that the JSON is valid
func isJSON(str string) bool {
	var jsonStr map[string]interface{}
	err := json.Unmarshal([]byte(str), &jsonStr)
	return err == nil

}

// PrintWrapper is a dirty function that just prints to screen what it receives
func PrintWrapper(a ...interface{}) {

	if printEverything {
		fmt.Print(a...)
	}
}

func main() {
	ctr = 0
	printEverything = true

	path := flag.String("file", "", "File containing the TF JSON Schema")
	getJSON := flag.Bool("getJSON", false, "Return canonical JSON TF config for input TF Schema in JSON")
	diffSchema := flag.Bool("diffschema", false, "diff output instead of writing")
	flag.Parse()

	f, err := os.Open(*path)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	defer f.Close()

	parsed := &tfjson.ProviderSchemas{}

	dec := json.NewDecoder(f)
	dec.DisallowUnknownFields()
	if err = dec.Decode(parsed); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	out, err := json.MarshalIndent(parsed, "", "  ")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	out = append(out, byte('\n'))

	if *diffSchema {
		var diffCmd string
		if _, err := exec.LookPath("colordiff"); err == nil {
			diffCmd = "colordiff"
		} else {
			diffCmd = "diff"
		}

		cmd := exec.Command(diffCmd, "-urN", *path, "-")
		cmd.Stdin = bytes.NewBuffer(out)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			if err.(*exec.ExitError).ProcessState.ExitCode() > 1 {
				os.Exit(1)
			}
		} else {
			fmt.Fprintln(os.Stderr, "[no diff]")
		}
	} else if *getJSON {
		// Don't print out anything other than the JSON
		printEverything = false
	}
	// Print out all the things

	// Build the canonical JSON plan from the schema and spit it out
	// We're going to have some fun doing this, mainly arbitrarily iterating
	// our way through the KV values for resource blocks and nested blocks.

	// This is a dirty hack because the TF marshallers are too rigid to work with.
	// The issue is, we don't know what any given schema looks like ahead of time.
	// This is dynamic in nature and thus, we need to generate canonical JSON from any schema input.
	resourceStrings := []string{}
	depth := 0
	for _, v1 := range parsed.Schemas {
		// First level goes through each provider_schemas

		for k2, v2 := range v1.ResourceSchemas {
			attributeStr := `{"resource":{`
			attributeStr += `"` + k2 + `":`
			attributeStr += `{"config-group-name":{`

			PrintWrapper(fmt.Sprintf("resource_schema: %v\n", k2))

			// this is the rest of our JSON string entry
			attLength := 0
			nestedLength := 0

			// Pre-check for ID
			if _, ok := v2.Block.Attributes["id"]; ok {
				attLength++
			}

			for k3, v3 := range v2.Block.Attributes {

				if k3 == "id" {
					ctr++
					goto RELOOP
				}

				for i := 0; i < depth; i++ {
					PrintWrapper("\t")
				}

				// Attributes for top layer block
				switch v3.AttributeType.FriendlyName() {
				case "string":
					if v3.Computed {
						attributeStr += fmt.Sprintf(attributeStrPattern, k3, "computed")
					} else if k3 == "resource_name" {
						attributeStr += fmt.Sprintf(attributeStrPattern, k3, cfgGroupName)
					} else {
						attributeStr += fmt.Sprintf(attributeStrPattern, k3, fmt.Sprintf("foo%d", ctr))
					}
				case "bool":
					attributeStr += fmt.Sprintf(attributeNonStrPattern, k3, "false")
				case "number":
					attributeStr += fmt.Sprintf(attributeNonStrPattern, k3, ctr)
				case "list of string":
					attributeStr += "\"" + k3 + "\"" + ": ["
					attributeStr += fmt.Sprintf("\"bar%d\", \"barbar%d\"", ctr, ctr)
					attributeStr += "]"
				case "map of string":
					attributeStr += "\"" + k3 + "\"" + ": {"
					attributeStr += fmt.Sprintf("\"foo%d\": \"bar%d\", \"foofoo%d\": \"barbar%d\"", ctr, ctr, ctr, ctr)
					attributeStr += "}"
				case "list of number":
					attributeStr += fmt.Sprintf(attributeNonStrPattern, k3, ctr)
				case "map of number":
					attributeStr += fmt.Sprintf(attributeNonStrPattern, k3, ctr)
				case "list of bool":
					attributeStr += "["
					attributeStr += fmt.Sprintf(attributeNonStrPattern, k3, "true, false, true")
					attributeStr += "]"
				case "map of bool":
					attributeStr += "{"
					attributeStr += fmt.Sprintf(attributeNonStrPattern, k3, "\"thing%d\": true, \"thingy%d\": false")
					attributeStr += "}"
				default:
					panic(fmt.Sprintf("CAUGHT EXCEPTION: %v, %v ", k3, v3))
				}

				attLength++

				if attLength < len(v2.Block.Attributes) {
					attributeStr += ", "
				} else {
					if len(v2.Block.NestedBlocks) != 0 {
						attributeStr += ", "
					}
				}

				ctr++

				//	This is tactically placed. Helps us to debug for issues
			RELOOP:
				PrintWrapper(fmt.Sprintf("attributes: %v, schema: %v\n", k3, v3.AttributeType.FriendlyName()))
			}

			for k4, v4 := range v2.Block.NestedBlocks {
				for i := 0; i < depth; i++ {
					PrintWrapper("\t")
				}

				PrintWrapper(fmt.Sprintf("nested block: %v, type: %v\n", k4, v4.NestingMode))

				// TODO: Depending on the type (set/list), we need to create brackets {} and []

				switch v4.NestingMode {
				case "single", "map", "set":
					attributeStr += fmt.Sprintf("%q: {", k4)
				case "list":
					attributeStr += fmt.Sprintf("%q: [{", k4)
				}

				printNested(v4, depth, &attributeStr)

				switch v4.NestingMode {
				case "single", "map", "set":
					attributeStr += "}"
				case "list":
					attributeStr += "}]"
				}

				nestedLength++
				if nestedLength < len(v2.Block.NestedBlocks) {
					attributeStr += ", "
				}

			}
			attributeStr += "}}}}"
			resourceStrings = append(resourceStrings, attributeStr)
		}
	}

	// Exit point for processing
	PrintWrapper("\n")

	// Let's do some basic JSON validation
	for _, v := range resourceStrings {
		result := isJSON(v)
		if result {
			PrintWrapper("Valid JSON: ", v)
			PrintWrapper("\n")
		} else {
			PrintWrapper("Not valid JSON: ", v)
			PrintWrapper("\n")
		}

		if !printEverything {
			// Return cursor to beginning of screen
			fmt.Print("\r" + v + "\n")
		}
	}
}

// Recursive function to unfold the nested-blocks
func printNested(t *tfjson.SchemaBlockType, depth int, attributeStr *string) {
	attLength := 0
	nestedLength := 0

	// Pre-check for ID
	if _, ok := t.Block.Attributes["id"]; ok {
		attLength++
	}

	depth++
	for k1, v1 := range t.Block.Attributes {

		for i := 0; i < depth; i++ {
			PrintWrapper("\t")
		}

		if k1 == "id" {
			ctr++
			goto RELOOP
		}

		// Attributes for top layer block
		switch v1.AttributeType.FriendlyName() {
		case "string":
			if v1.Computed {
				*attributeStr += fmt.Sprintf(attributeStrPattern, k1, "computed")
			} else if k1 == "resource_name" {
				*attributeStr += fmt.Sprintf(attributeStrPattern, k1, cfgGroupName)
			} else {
				*attributeStr += fmt.Sprintf(attributeStrPattern, k1, fmt.Sprintf("foo%d", ctr))
			}
		case "bool":
			*attributeStr += fmt.Sprintf(attributeNonStrPattern, k1, "false")
		case "number":
			*attributeStr += fmt.Sprintf(attributeNonStrPattern, k1, ctr)
		case "list of string":
			*attributeStr += "\"" + k1 + "\"" + ": ["
			*attributeStr += fmt.Sprintf("\"bar%d\", \"barbar%d\"", ctr, ctr)
			*attributeStr += "]"
		case "map of string":
			*attributeStr += "\"" + k1 + "\"" + ": {"
			*attributeStr += fmt.Sprintf("\"foo%d\": \"bar%d\", \"foofoo%d\": \"barbar%d\"", ctr, ctr, ctr, ctr)
			*attributeStr += "}"
		case "list of number":
			*attributeStr += fmt.Sprintf(attributeNonStrPattern, k1, ctr)
		case "map of number":
			*attributeStr += fmt.Sprintf(attributeNonStrPattern, k1, ctr)
		case "list of bool":
			*attributeStr += "["
			*attributeStr += fmt.Sprintf(attributeNonStrPattern, k1, "true, false, true")
			*attributeStr += "]"
		case "map of bool":
			*attributeStr += "{"
			*attributeStr += fmt.Sprintf(attributeNonStrPattern, k1, "\"thing%d\": true, \"thingy%d\": false")
			*attributeStr += "}"
		default:
			panic(fmt.Sprintf("CAUGHT EXCEPTION: %v, %v", k1, v1))
		}

		attLength++

		if (attLength < len(t.Block.Attributes)) && attLength != 0 {
			*attributeStr += ", "
		} else {
			if len(t.Block.NestedBlocks) != 0 {
				*attributeStr += ", "
			}
		}

		ctr++

	RELOOP:
		PrintWrapper(fmt.Sprintf("nested block attribute: %v, schema: %v\n", k1, v1.AttributeType.FriendlyName()))
	}
	for k2, v2 := range t.Block.NestedBlocks {
		for i := 0; i < depth; i++ {
			fmt.Print("\t")
		}
		PrintWrapper(fmt.Sprintf("nested nested block: %v, type: %v\n", k2, v2.NestingMode))

		if nestedLength == 0 {
			switch v2.NestingMode {
			case "single", "map", "set":
				*attributeStr += fmt.Sprintf("%q: {", k2)
			case "list":
				*attributeStr += fmt.Sprintf("%q: [{", k2)
			}
		} else {
			switch v2.NestingMode {
			case "single", "map", "set":
				*attributeStr += fmt.Sprintf(", %q: {", k2)
			case "list":
				*attributeStr += fmt.Sprintf(", %q: [{", k2)
			}
		}

		// TODO: Depending on the type (set/list), we need to create brackets {} and []

		printNested(v2, depth, attributeStr)

		switch v2.NestingMode {
		case "single", "map", "set":
			*attributeStr += "}"
		case "list":
			*attributeStr += "}]"
		}

		nestedLength++
	}
}
