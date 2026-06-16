package clitable_test

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/DavidGamba/dgtools/clitable"
)

func ExampleSimpleTable() {
	fmt.Println("String Slices")
	simpleData := [][]string{{"Hello", "1"}, {"World", "2"}}
	clitable.NewTablePrinter().Print(clitable.SimpleTable{Data: simpleData})

	// Output: String Slices
	// ┌───────┬───┐
	// │ Hello │ 1 │
	// ╞═══════╪═══╡
	// │ World │ 2 │
	// └───────┴───┘
}

func ExampleCSVTable() {
	fmt.Println("CSV")
	r := strings.NewReader("Hello,1\nWorld,2\n")
	clitable.NewTablePrinter().FprintCSVReader(os.Stdout, r)

	// Output: CSV
	// ┌───────┬───┐
	// │ Hello │ 1 │
	// ╞═══════╪═══╡
	// │ World │ 2 │
	// └───────┴───┘
}

func ExampleCSVTable_tsv() {
	fmt.Println("TSV")
	r := strings.NewReader("Hello	1\nWorld	2\n")
	clitable.NewTablePrinter().Separator('\t').FprintCSVReader(os.Stdout, r)

	// Output: TSV
	// ┌───────┬───┐
	// │ Hello │ 1 │
	// ╞═══════╪═══╡
	// │ World │ 2 │
	// └───────┴───┘
}

func ExampleMapTable() {
	fmt.Println("Map Slices Simple")
	mapData := []map[string]any{
		{"a": "c", "b": "d"},
		{"a": "e", "b": "f"},
	}
	clitable.NewTablePrinter().Print(clitable.MapTable{MapList: mapData})
	fmt.Println("Map Slices Nested")
	mapData = []map[string]any{
		{"a": map[string]any{"c": "g", "h": "i"}, "b": "d"},
		{"a": "e", "b": map[string]any{"f": 123, "j": 456}},
	}
	clitable.NewTablePrinter().Print(clitable.MapTable{MapList: mapData})

	// Output: Map Slices Simple
	// ┌───┬───┐
	// │ a │ b │
	// ╞═══╪═══╡
	// │ c │ d │
	// ├───┼───┤
	// │ e │ f │
	// └───┴───┘
	// Map Slices Nested
	// ┌─────────────┬─────────────┐
	// │ a           │ b           │
	// ╞═════════════╪═════════════╡
	// │ {           │ d           │
	// │   "c": "g", │             │
	// │   "h": "i"  │             │
	// │ }           │             │
	// ├─────────────┼─────────────┤
	// │ e           │ {           │
	// │             │   "f": 123, │
	// │             │   "j": 456  │
	// │             │ }           │
	// └─────────────┴─────────────┘
}

type Data struct {
	Info []struct {
		Name string
		ID   int
	}
}

func (d *Data) RowIterator() <-chan clitable.Row {
	c := make(chan clitable.Row)
	go func() {
		c <- clitable.Row{Fields: []string{"", "Name", "ID"}}
		for i, row := range d.Info {
			c <- clitable.Row{Fields: []string{strconv.Itoa(i + 1), row.Name, strconv.Itoa(row.ID)}}
		}
		close(c)
	}()
	return c
}

func ExampleTable() {
	fmt.Println("Custom Structs")
	data := &Data{[]struct {
		Name string
		ID   int
	}{{"Hello", 1}, {"World", 2}}}
	clitable.NewTablePrinter().Fprint(os.Stdout, data)

	// Output: Custom Structs
	// ┌───┬───────┬────┐
	// │   │ Name  │ ID │
	// ╞═══╪═══════╪════╡
	// │ 1 │ Hello │ 1  │
	// ├───┼───────┼────┤
	// │ 2 │ World │ 2  │
	// └───┴───────┴────┘
}
