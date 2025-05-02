package domain

type TableRow struct {
	ID        string
	Fields    []string
	HasUpdate bool
	HasDelete bool
}

type TableData struct {
	ContentTitle string
	TableID      string
	Headers      []Fields
	Rows         []TableRow
	PageSize     int
	CurrentPage  int
	TotalPages   int
	TotalRecords int
	StartRecord  int
	EndRecord    int
	URLPrefix    string
	HXInclude    string
	ShowActions  bool
	ShowAdd      bool
	ShowUpdate   bool
}

// Data structures for rendering templates
type MenuItem struct {
	Name string
	Path string
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
