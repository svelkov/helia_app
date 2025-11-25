package domain

type TableRow struct {
	ID        string
	Fields    []string
	HasUpdate bool
	HasDelete bool
}

type TableData struct {
	ContentTitle   string
	TableID        string
	Headers        []Fields
	Rows           []TableRow
	Pagination     PaginationData
	URLPrefix      string
	URLGetAll      string
	HxInclude      string
	HxTarget       string
	BtnAdd         Button
	BtnUpdate      Button
	BtnDelete      Button
	BtnPrint       Button
	BtnExportExcel Button
	BtnExportPDF   Button
	SearchEnabled  bool
	ShowActions    bool
	ShowPagination bool
	DetailTarget   string
	DetailURL      string
	ExportFilename string
	HasExportExcel bool
	HasExportPdf   bool
	FuncClick      string
	FuncDblClick   string
	DestField      string //use for popup dialog to return the value in the control
}

type PaginationData struct {
	PageSize     int
	CurrentPage  int
	TotalPages   int
	TotalRecords int
	StartRecord  int
	EndRecord    int
	PageSizes    []int
	HxInclude    string
	HxVals       string
}

type PageData struct {
	Title    string
	MainMenu []MenuItem
	SideMenu []MenuItem
	Content  string
}

type ContentData struct {
	ContentTitle string
	Content      interface{}
}

type InputControl struct {
	ID           string
	Label        string
	Type         string
	Placeholder  string
	Value        string
	HxActionURL  string
	HxTarget     string
	HxSwap       string
	HxTrigger    string
	HxInclude    string
	Autocomplete string
	OnKeyUp      string
	Class        string
}
