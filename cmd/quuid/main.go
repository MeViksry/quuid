package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/MeViksry/quuid"
)

type result struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

func main() {
	kind := flag.String("type", "id", "id, strong, sortable, uuid4, uuid6, uuid7, uuid8, or token")
	count := flag.Int("n", 1, "number of identifiers to generate")
	asJSON := flag.Bool("json", false, "emit JSON")
	namespace := flag.String("namespace", "", "namespace for deterministic uuid8")
	data := flag.String("data", "", "data for deterministic uuid8")
	flag.Parse()

	if *count < 1 || *count > 1_000_000 {
		fatalf("-n must be between 1 and 1000000")
	}

	results := make([]result, 0, *count)
	for i := 0; i < *count; i++ {
		value, err := generate(*kind, *namespace, *data)
		if err != nil {
			fatalf("%v", err)
		}
		results = append(results, result{Type: *kind, Value: value})
	}

	if *asJSON {
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		if *count == 1 {
			if err := encoder.Encode(results[0]); err != nil {
				fatalf("encode JSON: %v", err)
			}
			return
		}
		if err := encoder.Encode(results); err != nil {
			fatalf("encode JSON: %v", err)
		}
		return
	}

	for _, item := range results {
		fmt.Println(item.Value)
	}
}

func generate(kind, namespace, data string) (string, error) {
	switch kind {
	case "id":
		id, err := quuid.NewID()
		return id.String(), err
	case "strong":
		id, err := quuid.NewStrongID()
		return id.String(), err
	case "sortable":
		id, err := quuid.NewSortableID()
		return id.String(), err
	case "uuid4":
		id, err := quuid.NewUUIDv4()
		return id.String(), err
	case "uuid6":
		id, err := quuid.NewUUIDv6()
		return id.String(), err
	case "uuid7":
		id, err := quuid.NewUUIDv7()
		return id.String(), err
	case "uuid8":
		if namespace != "" || data != "" {
			return quuid.DeriveUUIDv8(namespace, []byte(data)).String(), nil
		}
		id, err := quuid.NewUUIDv8()
		return id.String(), err
	case "token":
		token, err := quuid.NewToken()
		return token.Secret(), err
	default:
		return "", fmt.Errorf("unsupported type %q", kind)
	}
}

func fatalf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "quuid: "+format+"\n", args...)
	os.Exit(1)
}
