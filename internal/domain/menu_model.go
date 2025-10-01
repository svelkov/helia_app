package domain

type MenuResponse struct {
	Submenu []string `json:"submenu"`
}
type SubMenuItem struct {
	Name string
	URL  string
	Icon string
}

type MenuItem struct {
	Name        string
	DisplayName string
	Icon        string
	SubMenus    []SubMenuItem // Direct child of MenuItem
}

type MenuDataItems struct {
	CurrentMenu    string
	CurrentSubMenu string
	MenuItems      []MenuItem
}

var MenuData = MenuDataItems{
	CurrentMenu: "Opšti podaci",
	MenuItems: []MenuItem{
		{
			Name:        "opsti_podaci",
			DisplayName: "Opšti podaci/Sifarnici",
			SubMenus: []SubMenuItem{
				{URL: "api/partneri/all", Name: "Kupci/Dobavljači", Icon: "partneri"},
				{URL: "api/tipdok/all", Name: "Vrste naloga", Icon: "vrstenaloga"},
				{URL: "api/dokvrsta/all", Name: "Vrste dokumenata", Icon: "vrstedokumenta"},
				{URL: "api/fvknjrac/all", Name: "Vrste poreskih knjiga (KPR i KIR)", Icon: "vrsteporknjige"},
				{URL: "api/fvepdv/all", Name: "Vrste evidencija PDV (EV PDV)", Icon: "vrsteevpdv"},
				{URL: "api/orgjed/all", Name: "Organizacione jedinice", Icon: "oj"},
				{URL: "api/mestotroska/all", Name: "Mesta troškova", Icon: "mtroska"},
				{URL: "api/drzava/all", Name: "Države", Icon: "drzave"},
				{URL: "api/sifop/all", Name: "Opštine", Icon: "opstine"},
				{URL: "api/sifmesto/all", Name: "Mesta", Icon: "mesta"},
				{URL: "api/banke/all", Name: "Banke", Icon: "banke"},
				{URL: "api/bnkizv/all", Name: "Banke za izvozne fakture", Icon: "bankeizv"},
				{URL: "api/sifplizvodi/all", Name: "Šifre plaćanja za domaći promet", Icon: "sifplizv"},
			},
		},
		{
			Name:        "finansijsko",
			DisplayName: "Finansijsko",
			SubMenus: []SubMenuItem{
				{URL: "api/fkpl/all", Name: "Kontni plan", Icon: "kontniplan"},
				{URL: "api/nalozi/all/tipdok", Name: "Nalozi", Icon: "fin_nalozi"},
				{URL: "api/promet", Name: "Promet", Icon: "fin_promet"},
				{URL: "", Name: "Salda konta", Icon: "fin_saldakonta"},
				{URL: "", Name: "Kompenzacije", Icon: "fin_kompenzacije"},
				{URL: "", Name: "Otvorene stavke", Icon: ""},
				{URL: "", Name: "Obračun kamate", Icon: ""},
				{URL: "", Name: "Bilansi", Icon: ""},
				{URL: "", Name: "Poreske knjige", Icon: ""},
				{URL: "", Name: "Učitavanje i knjiženje izvoda", Icon: ""},
			},
		},
		{

			Name:        "komercijalno_poslovanje",
			DisplayName: "Komercijalno poslovanje",
			SubMenus: []SubMenuItem{
				{URL: "", Name: "Zahtev za nabavku", Icon: ""},
				{URL: "", Name: "Zahtev za ponudu", Icon: ""},
				{URL: "", Name: "Porudžbenice dobavljačima", Icon: ""},
				{URL: "", Name: "Prijemnica materijala/sirovine", Icon: ""},
				{URL: "", Name: "Profakture", Icon: ""},
				{URL: "", Name: "Nalozi za isporuku", Icon: ""},
				{URL: "", Name: "Otpremanje robe", Icon: ""},
			},
		},
		{
			Name:        "robno",
			DisplayName: "Robno",
			SubMenus: []SubMenuItem{
				{URL: "", Name: "Robna Dokumenta", Icon: ""},
				{URL: "", Name: "Promet po artiklima i kontima", Icon: ""},
				{URL: "", Name: "Stanje po artiklima i kontima", Icon: ""},
				{URL: "", Name: "Izveštaji pregled ulaza/izlaza", Icon: ""},
				{URL: "", Name: "Artikli", Icon: ""},
				{URL: "", Name: "Jedinice mere", Icon: ""},
				{URL: "", Name: "Robne grupe", Icon: ""},
				{URL: "", Name: "Robne podgrupe", Icon: ""},
				{URL: "", Name: "Mesta isporuke", Icon: ""},
				{URL: "", Name: "Komercijalisti", Icon: ""},
				{URL: "", Name: "Poreske stope", Icon: ""},
				{URL: "", Name: "Vrste dokumenata", Icon: ""},
				{URL: "", Name: "Cenovnik", Icon: ""},
				{URL: "", Name: "Tipovi knjižnih pisama", Icon: ""},
				{URL: "", Name: "Magacini", Icon: ""},
				{URL: "", Name: "Komercijalni podaci", Icon: ""},
				{URL: "", Name: "Kraj poslovne godine", Icon: ""},
			},
		},
		{
			Name:        "blagajna",
			DisplayName: "Blagajna",
			SubMenus: []SubMenuItem{
				{URL: "", Name: "Dnevnik blagajne", Icon: ""},
				{URL: "", Name: "Kontiranje blagajne", Icon: ""},
				{URL: "", Name: "Knjiženje blagajne", Icon: ""},
				{URL: "", Name: "Specifikacija apoena", Icon: ""},
				{URL: "", Name: "Valute", Icon: ""},
				{URL: "", Name: "Kursna lista", Icon: ""},
				{URL: "", Name: "Apoeni", Icon: ""},
				{URL: "", Name: "Tipovi blagajne", Icon: ""},
				{URL: "", Name: "Vrste stavki blagajne", Icon: ""},
				{URL: "", Name: "Dodatna knjiženja", Icon: ""},
				{URL: "", Name: "Putni nalozi", Icon: ""},
				{URL: "", Name: "Dnevnice", Icon: ""},
				{URL: "", Name: "Tipovi troškova", Icon: ""},
				{URL: "", Name: "Prevozna sredstva", Icon: ""},
			},
		},
		{
			Name:        "proizvodnja",
			DisplayName: "Proizvodnja",
			SubMenus: []SubMenuItem{
				{URL: "", Name: "Matični podaci normativa", Icon: ""},
				{URL: "", Name: "Konta troškova", Icon: ""},
				{URL: "", Name: "Radni nalog za proizvodnju", Icon: ""},
				{URL: "", Name: "Unos dokumenta", Icon: ""},
				{URL: "", Name: "Pregled dokumenta", Icon: ""},
				{URL: "", Name: "Specifikacija dokumenta", Icon: ""},
				{URL: "", Name: "Knjiženje dokumenta", Icon: ""},
				{URL: "", Name: "Promet po artiklima i kontima", Icon: ""},
				{URL: "", Name: "Stanje po artiklima i kontima", Icon: ""},
				{URL: "", Name: "Trenutno stanje zaliha", Icon: ""},
				{URL: "", Name: "Pregled stanja i cena artikala", Icon: ""},
				{URL: "", Name: "Obračunska kalkulacija", Icon: ""},
				{URL: "", Name: "Kalkulacija proizvoda", Icon: ""},
				{URL: "", Name: "Sitan inventar - zaduženje", Icon: ""},
				{URL: "", Name: "Izveštaji", Icon: ""},
				{URL: "", Name: "Kraj poslovne godine", Icon: ""},
			},
		},
		{
			Name:        "osnovna_sredstva",
			DisplayName: "Osnovna sredstva",
			SubMenus: []SubMenuItem{
				{URL: "", Name: "Maticni podaci osnovnih sredstava", Icon: ""},
				{URL: "", Name: "Uknjižavanje, promene", Icon: ""},
				{URL: "", Name: "Obračun amortizacije", Icon: ""},
				{URL: "", Name: "Obračun poreske amortizacije", Icon: ""},
				{URL: "api/oamgrp/all", Name: "Amortizacione grupe", Icon: ""},
				{URL: "", Name: "Amortizacione podgrupe", Icon: ""},
				{URL: "", Name: "Organizacione jedinice", Icon: ""},
				{URL: "", Name: "Rukovaoci", Icon: ""},
				{URL: "", Name: "Objekti", Icon: ""},
				{URL: "", Name: "Konta nabavke i ispravke", Icon: ""},
				{URL: "", Name: "Koeficijenti revalorizacije", Icon: ""},
				{URL: "", Name: "Stanje osnovnih sredstava", Icon: ""},
				{URL: "", Name: "Kartica osnovnog sredstva", Icon: ""},
				{URL: "", Name: "Izveštaji", Icon: ""},
				{URL: "", Name: "Kraj poslovne godine", Icon: ""},
			},
		},
	},
}

// Osnovna sredstva
//	TreeAdd(TREE_OS, "Sifarnici"+TAB+","VST01366.png", "VST01366.png", "", tvLast)
// TreeAdd(TREE_OS, "Šifarnici"+TAB+"","VST01366.png", "VST01366.png", "", tvLast)

// TreeAdd(TREE_OS, "Izveštaji","VST01330b_16_1.png", "VST01330b_16_1.png", "", tvLast)
// TreeAdd(TREE_OS, "Izveštaji"+TAB+"Kartica osnovnog sredstva","VST01366.png", "VST01366.png", "", tvLast)
// TreeAdd(TREE_OS, "Izveštaji"+TAB+"Stanje osnovnih sredstava","VST01330b_16_1.png", "VST01330b_16_1.png", "", tvLast)
// TreeAdd(TREE_OS, "Izveštaji"+TAB+"Stanje osnovnih sredstava"+TAB+"Početno stanje","VST01366.png", "VST01366.png", "", tvLast)
// TreeAdd(TREE_OS, "Izveštaji"+TAB+"Stanje osnovnih sredstava"+TAB+"Trenutno stanje","VST01366.png", "VST01366.png", "", tvLast)
// TreeAdd(TREE_OS, "Izveštaji"+TAB+"Stanja po obračunu","VST01366.png", "VST01366.png", "", tvLast)
// TreeAdd(TREE_OS, "Izveštaji"+TAB+"Predračun amortizacije","VST01366.png", "VST01366.png", "", tvLast)
// TreeAdd(TREE_OS, "Izveštaji"+TAB+"Prikaz naloga","VST01366.png", "VST01366.png", "", tvLast)
// TreeAdd(TREE_OS, "Izveštaji"+TAB+"Prikaz prometa konta","VST01366.png", "VST01366.png", "", tvLast)
// TreeAdd(TREE_OS, "Izveštaji"+TAB+"Štampa po obračunu","VST01366.png", "VST01366.png", "", tvLast)

// TreeAdd(TREE_OS, "Kraj poslovne godine","VST01330b_16_1.png", "VST01330b_16_1.png", "", tvLast)
// TreeAdd(TREE_OS, "Kraj poslovne godine"+TAB+"Štampa popisnih listi","VST01366.png", "VST01366.png", "", tvLast)
// TreeAdd(TREE_OS, "Kraj poslovne godine"+TAB+"","VST01366.png", "VST01366.png", "", tvLast)
// TreeAdd(TREE_OS, "Kraj poslovne godine"+TAB+"Viškovi i manjkovi","VST01366.png", "VST01366.png", "", tvLast)
// TreeAdd(TREE_OS, "Kraj poslovne godine"+TAB+"Prepis u novu","VST01366.png", "VST01366.png", "", tvLast)

//  TreeAdd(TREE_Proizvodnja, "Promet po artiklima i kontima"+TAB+"Promet kartice artikala","VST01366.png", "VST01366.png", "", tvLast)
//  TreeAdd(TREE_Proizvodnja, "Promet po artiklima i kontima"+TAB+"Promet kartice","VST01366.png", "VST01366.png", "", tvLast)

//  TreeAdd(TREE_Proizvodnja, "","VST01330b_16_1.png", "VST01330b_16_1.png", "", tvLast)
//  TreeAdd(TREE_Proizvodnja, "Stanje po artiklima i kontima"+TAB+"Prikaz stanja","VST01366.png", "VST01366.png", "", tvLast)
// TreeAdd(TREE_Proizvodnja, "Stanje po artiklima i kontima"+TAB+"Prikaz stanja","VST01366.png", "VST01366.png", "", tvLast)
// TreeAdd(TREE_Proizvodnja, "Stanje po artiklima i kontima"+TAB+"Prikaz salda","VST01366.png", "VST01366.png", "", tvLast)
// TreeAdd(TREE_Proizvodnja, "Stanje po artiklima i kontima"+TAB+"Svodjenje zaliha","VST01366.png", "VST01366.png", "", tvLast)

// TreeAdd(TREE_Proizvodnja, "","VST01330b_16_1.png", "VST01330b_16_1.png", "", tvLast)
// TreeAdd(TREE_Proizvodnja, "Izveštaji"+TAB+"Izveštaj o ostvarenoj","VST01366.png", "VST01366.png", "", tvLast)
// TreeAdd(TREE_Proizvodnja, "Izveštaji"+TAB+"Ostvarena proizvodnja","VST01366.png", "VST01366.png", "", tvLast)
// TreeAdd(TREE_Proizvodnja, "Izveštaji"+TAB+"Pregled proizvodnje","VST01366.png", "VST01366.png", "", tvLast)
// TreeAdd(TREE_Proizvodnja, "Izveštaji"+TAB+"Izveštaj o utrošenim","VST01366.png", "VST01366.png", "", tvLast)

// TreeAdd(TREE_Robno, "Šifarnici"+TAB+"","VST01366.png", "VST01366.png", "", tvLast)
// TreeAdd(TREE_Robno, "Šifarnici"+TAB+"","VST01366.png", "VST01366.png", "", tvLast)
// TreeAdd(TREE_Robno, "","VST01366.png", "VST01366.png", "", tvLast)
// TreeAdd(TREE_Robno, "Magacin-Konto","VST01366.png", "VST01366.png", "", tvLast)
// // TreeAdd(TREE_Robno, "Podesavanje konta","VST01366.png", "VST01366.png", "", tvLast)

// TreeAdd(TREE_Robno, "","VST01330b_16_1.png", "VST01330b_16_1.png", "", tvLast)
// TreeAdd(TREE_Robno, "Robna Dokumenta"+TAB+"Unos dokumenata","VST01366.png", "VST01366.png", "", tvLast)
// TreeAdd(TREE_Robno, "Robna Dokumenta"+TAB+"Pregled dokumenata","VST01366.png", "VST01366.png", "", tvLast)
// TreeAdd(TREE_Robno, "Robna Dokumenta"+TAB+"Specifikacija dokumenata","VST01366.png", "VST01366.png", "", tvLast)
// TreeAdd(TREE_Robno, "Robna Dokumenta"+TAB+"Kontiranje dokumenata","VST01366.png", "VST01366.png", "", tvLast)
// TreeAdd(TREE_Robno, "Robna Dokumenta"+TAB+"Prepis dokumenata","VST01366.png", "VST01366.png", "", tvLast)
// TreeAdd(TREE_Robno, "Robna Dokumenta"+TAB+"Prikaz ukupne obrade","VST01366.png", "VST01366.png", "", tvLast)
// TreeAdd(TREE_Robno, "Robna Dokumenta"+TAB+"Prikaz naloga","VST01366.png", "VST01366.png", "", tvLast)
// TreeAdd(TREE_Robno, "Robna Dokumenta"+TAB+"Prikaz dokumenata","VST01366.png", "VST01366.png", "", tvLast)
// TreeAdd(TREE_Robno, "Robna Dokumenta"+TAB+"Prikaz dokumenata","VST01366.png", "VST01366.png", "", tvLast)

// TreeAdd(TREE_Robno, "","VST01330b_16_1.png", "VST01330b_16_1.png", "", tvLast)
// TreeAdd(TREE_Robno, "Promet po artiklima i kontima"+TAB+"Promet kartice artikala","VST01366.png", "VST01366.png", "", tvLast)
// TreeAdd(TREE_Robno, "Promet po artiklima i kontima"+TAB+"Promet kartice","VST01366.png", "VST01366.png", "", tvLast)
// //	   TreeAdd(TREE_Robno, "Promet po artiklima i kontima"+TAB+"Stampanje odjava","VST01366.png", "VST01366.png", "", tvLast)

// TreeAdd(TREE_Robno, "", "VST01330b_16_1.png", "VST01330b_16_1.png", "", tvLast)
// TreeAdd(TREE_Robno, "Stanje po artiklima i kontima"+TAB+"Prikaz stanja","VST01366.png", "VST01366.png", "", tvLast)
// TreeAdd(TREE_Robno, "Stanje po artiklima i kontima"+TAB+"Prikaz stanja","VST01366.png", "VST01366.png", "", tvLast)
// //	   TreeAdd(TREE_Robno, "Stanje po artiklima i kontima"+TAB+"Prikaz stanja vise artikala po pocetku naziva","VST01366.png", "VST01366.png", "", tvLast)
// TreeAdd(TREE_Robno, "Stanje po artiklima i kontima"+TAB+"Prikaz salda","VST01366.png", "VST01366.png", "", tvLast)
// TreeAdd(TREE_Robno, "Stanje po artiklima i kontima"+TAB+"Svodjenje stanja zaliha","VST01366.png", "VST01366.png", "", tvLast)
// //	   TreeAdd(TREE_Robno, "Stanje po artiklima i kontima"+TAB+"Svodjenje zaliha na prosecnu cenu","VST01366.png", "VST01366.png", "", tvLast)

// TreeAdd(TREE_Robno, "","VST01330b_16_1.png", "VST01330b_16_1.png", "", tvLast)
// //	   TreeAdd(TREE_Robno, "Izvestaji pregled ulaza/izlaza"+TAB+"Prikaz stanja artikala po grupama od-do","VST01366.png", "VST01366.png", "", tvLast)
// TreeAdd(TREE_Robno, "Izveštaji pregled ulaza/izlaza"+TAB+"Promet po grupi","VST01366.png", "VST01366.png", "", tvLast)
// TreeAdd(TREE_Robno, "Izveštaji pregled ulaza/izlaza"+TAB+"Promet po grupi","VST01366.png", "VST01366.png", "", tvLast)
// TreeAdd(TREE_Robno, "Izveštaji pregled ulaza/izlaza"+TAB+"Prodaja po kupcima","VST01366.png", "VST01366.png", "", tvLast)
// TreeAdd(TREE_Robno, "Izveštaji pregled ulaza/izlaza"+TAB+"Nabavka po dobavljačima","VST01366.png", "VST01366.png", "", tvLast)
// TreeAdd(TREE_Robno, "Izveštaji pregled ulaza/izlaza"+TAB+"Lager Lista, Ulaz/Izlaz,","VST01366.png", "VST01366.png", "", tvLast)
// TreeAdd(TREE_Robno, "Izveštaji pregled ulaza/izlaza"+TAB+"Izveštaj zaduženja","VST01366.png", "VST01366.png", "", tvLast)
// TreeAdd(TREE_Robno, "Izveštaji pregled ulaza/izlaza"+TAB+"Izveštaj zaduženja","VST01366.png", "VST01366.png", "", tvLast)

// TreeAdd(TREE_Robno, "","VST01330b_16_1.png", "VST01330b_16_1.png", "", tvLast)
// TreeAdd(TREE_Robno, "Komercijalni podaci"+TAB+"Prikaz kartice","VST01366.png", "VST01366.png", "", tvLast)
// TreeAdd(TREE_Robno, "Komercijalni podaci"+TAB+"Prikaz salda","VST01366.png", "VST01366.png", "", tvLast)
// TreeAdd(TREE_Robno, "Komercijalni podaci"+TAB+"Prikaz prodaje","VST01366.png", "VST01366.png", "", tvLast)
// TreeAdd(TREE_Robno, "Komercijalni podaci"+TAB+"Realizacija","VST01366.png", "VST01366.png", "", tvLast)
// TreeAdd(TREE_Robno, "Komercijalni podaci"+TAB+"Realizacija","VST01366.png", "VST01366.png", "", tvLast)
// TreeAdd(TREE_Robno, "Komercijalni podaci"+TAB+"Pregled realizacije","VST01366.png", "VST01366.png", "", tvLast)
// TreeAdd(TREE_Robno, "Komercijalni podaci"+TAB+"Pregled učešća artikla","VST01366.png", "VST01366.png", "", tvLast)
// TreeAdd(TREE_Robno, "Komercijalni podaci"+TAB+"Pregled učešća grupe","VST01366.png", "VST01366.png", "", tvLast)

// TreeAdd(TREE_Robno, "","VST01330b_16_1.png", "VST01330b_16_1.png", "", tvLast)
// TreeAdd(TREE_Robno, "Kraj poslovne godine"+TAB+"Popisne liste","VST01366.png", "VST01366.png", "", tvLast)
// TreeAdd(TREE_Robno, "Kraj poslovne godine"+TAB+"Obrada viškova-manjkova","VST01366.png", "VST01366.png", "", tvLast)
// TreeAdd(TREE_Robno, "Kraj poslovne godine"+TAB+"Prepis stanja artikala","VST01366.png", "VST01366.png", "", tvLast)

//)
//	TreeExpandAll(TREE_Proizvodnja)

// //**********************************************************************************************
// TreeAdd(TREE_Servis, "Izbor poslovne godine","VST01366.png", "VST01366.png", "", tvLast)
// TreeAdd(TREE_Servis, "Održavanje opštih podataka","VST01366.png", "VST01366.png", "", tvLast)
// TreeAdd(TREE_Servis,"Parametri konekcije","VST01366.png", "VST01366.png", "", tvLast)
// TreeAdd(TREE_Servis,"Servis","VST01366.png", "VST01366.png", "", tvLast)
// TreeAdd(TREE_Servis,"Šema knjiženja","VST01366.png", "VST01366.png", "", tvLast)
