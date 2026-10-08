package domain

// TableComboConfig configures components.TableCombo: a text input that searches while the user types
// and shows the matching rows in a table under the input (partneri, konta, artikli, ...). A click on a
// row (or Enter on the highlighted row) selects it: the key of the row goes to the hidden input Name
// (what the form submits) and its text to the visible input.
//
// The ids of the parts are derived from ID: the visible input is ID+"-input", the hidden input
// ID+"-value" (or ValueID), the details line ID+"-details" and the dropdown with the results
// ID+"-results".
type TableComboConfig struct {
	ID string
	// Name is the name of the hidden input with the key of the selected row (submitted with the form).
	Name string
	// ValueID is the id of the hidden input (default ID+"-value"): a screen whose scripts and
	// hx-include read the value of an existing field by its id (e.g. "konto") keeps that id here.
	ValueID string
	// Value is the key of the selected row and DisplayValue its text in the input (both empty when
	// nothing is selected yet, e.g. a new document); Details is the optional line under the input
	// (e.g. the adresa and the PIB of the selected partner).
	Value        string
	DisplayValue string
	Details      string
	// DefaultValue is the value of the hidden input while nothing is selected (e.g. "00" for the od
	// šifre, "999999" for the do šifre of a range): it is the value at the start (shown in the input
	// when DisplayValue is empty), after the text is cleared and when the user leaves the field without
	// selecting a row.
	DefaultValue string

	HasLabel    bool
	LabelText   string // translated with translator.Label
	Placeholder string // translated with translator.Label
	ClassLabel  string
	ClassInput  string
	// ClassWrapper is the class of the element around the input and the dropdown (default
	// "relative w-full").
	ClassWrapper string

	// SearchURL is called (GET) while the user types; it answers with components.TableComboResults.
	// The text of the input is sent as the parameter SearchParam (default "q").
	SearchURL   string
	SearchParam string
	// HxTrigger fires the search (default "input changed delay:300ms, focus[tableComboEmpty(this)]":
	// while typing, and on the focus only when nothing is selected yet). On the focus the text of the
	// input is selected, so typing replaces it.
	HxTrigger string
	// HxVals and HxInclude add parameters to the search (e.g. the konto of the partneri).
	HxVals    string
	HxInclude string
	// MinLength is the number of characters the search needs ("" = search from the first one).
	MinLength string

	TabIndex string
	Disabled bool
	Required bool
	// OnSelect is the name of a JavaScript function called after a row is selected, with the id of
	// the combo, the key and the cells of the selected row: fn(comboID, key, cells).
	OnSelect string
}

// TableComboResults is the answer of the search of a components.TableCombo: the matching rows in a
// TableData (Headers: the columns with their Label, Width and TextAlign; Rows: ID is the key of the
// row, Fields its cells).
type TableComboResults struct {
	// ComboID is the ID of the TableComboConfig the results belong to.
	ComboID string
	Table   TableData
	// TextColumns are the indexes of the cells shown in the input after the selection, joined with
	// " - " (default the first two, e.g. "šifra - naziv"); DetailColumns are the cells shown in the
	// details line under the input (default none).
	TextColumns   []int
	DetailColumns []int
	// EmptyText is shown when no row matches (translated with translator.Label; default "Nema
	// rezultata").
	EmptyText string
}

// SearchRowDto is one row of a search of the SearchService: the key of the row and its cells as
// text (every search selects its columns as key, c1, c2, ...).
type SearchRowDto struct {
	Key string `db:"key"`
	C1  string `db:"c1"`
	C2  string `db:"c2"`
	C3  string `db:"c3"`
	C4  string `db:"c4"`
	C5  string `db:"c5"`
	C6  string `db:"c6"`
	C7  string `db:"c7"`
	C8  string `db:"c8"`
}

// Cells returns the first n cells of the row.
func (r SearchRowDto) Cells(n int) []string {
	cells := []string{r.C1, r.C2, r.C3, r.C4, r.C5, r.C6, r.C7, r.C8}
	if n < len(cells) {
		cells = cells[:n]
	}
	return cells
}
