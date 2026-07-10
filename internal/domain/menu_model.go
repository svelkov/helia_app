package domain

type MenuResponse struct {
	Submenu []string `json:"submenu"`
}

type MenuItem struct {
	ID          int    `db:"id"`
	MenuName    string `db:"menuname"`
	DisplayName string `db:"displayname"`
	Icon        string `db:"icon"`
	SortOrder   int    `db:"sortorder"`
	SubMenus    []SubMenuItem
}

type SubMenuItem struct {
	ID          int    `db:"id"`
	MenuID      int    `db:"menuid"`
	SubMenuName string `db:"submenuname"`
	Url         string `db:"urlmenu"`
	Icon        string `db:"icon"`
	SortOrder   int    `db:"sortorder"`
}

type MenuDataItems struct {
	CurrentMenu    string
	CurrentSubMenu string
	MenuItems      []MenuItem
}

// var MenuData = MenuDataItems{
// 	CurrentMenu: "Opšti podaci",
// 	MenuItems: []MenuItem{
// 		{
// 			Name:        "opsti_podaci",
// 			DisplayName: "Opšti podaci/Sifarnici",
// 			SubMenus: []SubMenuItem{
// 				{URL: "api/partneri/all", Name: "Kupci/Dobavljači", Icon: "partneri"},
// 				{URL: "api/tipdok/all", Name: "Vrste naloga", Icon: "vrstenaloga"},
// 				{URL: "api/dokvrsta/all", Name: "Vrste dokumenata", Icon: "vrstedokumenta"},
// 				{URL: "api/fvknjrac/all", Name: "Vrste poreskih knjiga (KPR i KIR)", Icon: "vrsteporknjige"},
// 				{URL: "api/fvepdv/all", Name: "Vrste evidencija PDV (EV PDV)", Icon: "vrsteevpdv"},
// 				{URL: "api/orgjed/all", Name: "Organizacione jedinice", Icon: "oj"},
// 				{URL: "api/mestotroska/all", Name: "Mesta troškova", Icon: "mtroska"},
// 				{URL: "api/drzava/all", Name: "Države", Icon: "drzave"},
// 				{URL: "api/sifop/all", Name: "Opštine", Icon: "opstine"},
// 				{URL: "api/sifmesto/all", Name: "Mesta", Icon: "mesta"},
// 				{URL: "api/banke/all", Name: "Banke", Icon: "banke"},
// 				{URL: "api/bnkizv/all", Name: "Banke za izvozne fakture", Icon: "bankeizv"},
// 				{URL: "api/sifplizvodi/all", Name: "Šifre plaćanja za domaći promet", Icon: "sifplizv"},
// 			},
// 		},
// 		{
// 			Name:        "finansijsko",
// 			DisplayName: "Finansijsko",
// 			SubMenus: []SubMenuItem{
// 				{URL: "api/tipanalitike/all", Name: "Tipovi analitike", Icon: "tipanalitike"},
// 				{URL: "api/fkpl/all", Name: "Kontni plan", Icon: "kontniplan"},
// 				{URL: "api/nalozi/all/tipdok", Name: "Nalozi", Icon: "fin_nalozi"},
// 				{URL: "api/dnevnik", Name: "Dnevnik knjiženja", Icon: "fin_dnevnik"},
// 				{URL: "api/promet", Name: "Promet", Icon: "fin_promet"},
// 				{URL: "api/salda", Name: "Salda konta", Icon: "fin_saldakonta"},
// 				{URL: "api/kompenzacije", Name: "Kompenzacije", Icon: "fin_kompenzacije"},
// 				{URL: "/api/otvorenestavke", Name: "Otvorene stavke", Icon: "fin_otvorenestavke"},
// 				{URL: "/api/bilansi", Name: "Bilansi", Icon: "fin_bilansi"},
// 				{URL: "/api/poreskeknjige", Name: "Poreske knjige i prijava", Icon: ""},
// 				{URL: "/api/poreskeprijave", Name: "POPDV, PPPDV i PPV prijave", Icon: ""},
// 				{URL: "/api/fsepp", Name: "Evidencija predhodnog poreza EPP", Icon: ""},
// 				{URL: "/api/izvodi", Name: "Učitavanje i knjiženje izvoda", Icon: ""},
// 				{URL: "/api/kamate", Name: "Obračun kamate", Icon: "fin_obracunkamate"},
// 			},
// 		},
// 		{

// 			Name:        "komercijalno_poslovanje",
// 			DisplayName: "Komercijalno poslovanje",
// 			SubMenus: []SubMenuItem{
// 				{URL: "", Name: "Zahtev za nabavku", Icon: ""},
// 				{URL: "", Name: "Zahtev za ponudu", Icon: ""},
// 				{URL: "", Name: "Porudžbenice dobavljačima", Icon: ""},
// 				{URL: "", Name: "Prijemnica materijala/sirovine", Icon: ""},
// 				{URL: "", Name: "Profakture", Icon: ""},
// 				{URL: "", Name: "Nalozi za isporuku", Icon: ""},
// 				{URL: "", Name: "Otpremanje robe", Icon: ""},
// 			},
// 		},
// 		{
// 			Name:        "robno",
// 			DisplayName: "Robno",
// 			SubMenus: []SubMenuItem{
// 				{URL: "", Name: "Robna Dokumenta", Icon: ""},
// 				{URL: "", Name: "Promet po artiklima i kontima", Icon: ""},
// 				{URL: "", Name: "Stanje po artiklima i kontima", Icon: ""},
// 				{URL: "", Name: "Izveštaji pregled ulaza/izlaza", Icon: ""},
// 				{URL: "", Name: "Artikli", Icon: ""},
// 				{URL: "api/jedmere/all", Name: "Jedinice mere", Icon: ""},
// 				{URL: "", Name: "Robne grupe", Icon: ""},
// 				{URL: "", Name: "Robne podgrupe", Icon: ""},
// 				{URL: "", Name: "Mesta isporuke", Icon: ""},
// 				{URL: "", Name: "Komercijalisti", Icon: ""},
// 				{URL: "", Name: "Poreske stope", Icon: ""},
// 				{URL: "", Name: "Vrste dokumenata", Icon: ""},
// 				{URL: "", Name: "Cenovnik", Icon: ""},
// 				{URL: "", Name: "Tipovi knjižnih pisama", Icon: ""},
// 				{URL: "", Name: "Magacini", Icon: ""},
// 				{URL: "", Name: "Komercijalni podaci", Icon: ""},
// 				{URL: "", Name: "Kraj poslovne godine", Icon: ""},
// 			},
// 		},
// 		{
// 			Name:        "blagajna",
// 			DisplayName: "Blagajna",
// 			SubMenus: []SubMenuItem{
// 				{URL: "", Name: "Dnevnik blagajne", Icon: ""},
// 				{URL: "", Name: "Kontiranje blagajne", Icon: ""},
// 				{URL: "", Name: "Knjiženje blagajne", Icon: ""},
// 				{URL: "", Name: "Specifikacija apoena", Icon: ""},
// 				{URL: "", Name: "Valute", Icon: ""},
// 				{URL: "", Name: "Kursna lista", Icon: ""},
// 				{URL: "", Name: "Apoeni", Icon: ""},
// 				{URL: "", Name: "Tipovi blagajne", Icon: ""},
// 				{URL: "", Name: "Vrste stavki blagajne", Icon: ""},
// 				{URL: "", Name: "Dodatna knjiženja", Icon: ""},
// 				{URL: "", Name: "Putni nalozi", Icon: ""},
// 				{URL: "", Name: "Dnevnice", Icon: ""},
// 				{URL: "", Name: "Tipovi troškova", Icon: ""},
// 				{URL: "", Name: "Prevozna sredstva", Icon: ""},
// 			},
// 		},
// 		{
// 			Name:        "proizvodnja",
// 			DisplayName: "Proizvodnja",
// 			SubMenus: []SubMenuItem{
// 				{URL: "", Name: "Matični podaci normativa", Icon: ""},
// 				{URL: "", Name: "Konta troškova", Icon: ""},
// 				{URL: "", Name: "Radni nalog za proizvodnju", Icon: ""},
// 				{URL: "", Name: "Unos dokumenta", Icon: ""},
// 				{URL: "", Name: "Pregled dokumenta", Icon: ""},
// 				{URL: "", Name: "Specifikacija dokumenta", Icon: ""},
// 				{URL: "", Name: "Knjiženje dokumenta", Icon: ""},
// 				{URL: "", Name: "Promet po artiklima i kontima", Icon: ""},
// 				{URL: "", Name: "Stanje po artiklima i kontima", Icon: ""},
// 				{URL: "", Name: "Trenutno stanje zaliha", Icon: ""},
// 				{URL: "", Name: "Pregled stanja i cena artikala", Icon: ""},
// 				{URL: "", Name: "Obračunska kalkulacija", Icon: ""},
// 				{URL: "", Name: "Kalkulacija proizvoda", Icon: ""},
// 				{URL: "", Name: "Sitan inventar - zaduženje", Icon: ""},
// 				{URL: "", Name: "Izveštaji", Icon: ""},
// 				{URL: "", Name: "Kraj poslovne godine", Icon: ""},
// 			},
// 		},
// 		{
// 			Name:        "osnovna_sredstva",
// 			DisplayName: "Osnovna sredstva",
// 			SubMenus: []SubMenuItem{
// 				{URL: "", Name: "Maticni podaci osnovnih sredstava", Icon: ""},
// 				{URL: "", Name: "Uknjižavanje, promene", Icon: ""},
// 				{URL: "", Name: "Obračun amortizacije", Icon: ""},
// 				{URL: "", Name: "Obračun poreske amortizacije", Icon: ""},
// 				{URL: "api/oamgrp/all", Name: "Amortizacione grupe", Icon: ""},
// 				{URL: "", Name: "Amortizacione podgrupe", Icon: ""},
// 				{URL: "", Name: "Organizacione jedinice", Icon: ""},
// 				{URL: "", Name: "Rukovaoci", Icon: ""},
// 				{URL: "", Name: "Objekti", Icon: ""},
// 				{URL: "", Name: "Konta nabavke i ispravke", Icon: ""},
// 				{URL: "", Name: "Koeficijenti revalorizacije", Icon: ""},
// 				{URL: "", Name: "Stanje osnovnih sredstava", Icon: ""},
// 				{URL: "", Name: "Kartica osnovnog sredstva", Icon: ""},
// 				{URL: "", Name: "Izveštaji", Icon: ""},
// 				{URL: "", Name: "Kraj poslovne godine", Icon: ""},
// 			},
// 		},
// 	},
// }
