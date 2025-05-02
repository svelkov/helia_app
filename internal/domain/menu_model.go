package domain

var MenuData = map[string][]SubMenuItem{
	"Opšti podaci": {
		SubMenuItem{URL: "api/partneri/all", Name: "Kupci/Dobavljači", Icon: "M15 9h3.75M15 12h3.75M15 15h3.75M4.5 19.5h15a2.25 2.25 0 0 0 2.25-2.25V6.75A2.25 2.25 0 0 0 19.5 4.5h-15a2.25 2.25 0 0 0-2.25 2.25v10.5A2.25 2.25 0 0 0 4.5 19.5Zm6-10.125a1.875 1.875 0 1 1-3.75 0 1.875 1.875 0 0 1 3.75 0Zm1.294 6.336a6.721 6.721 0 0 1-3.17.789 6.721 6.721 0 0 1-3.168-.789 3.376 3.376 0 0 1 6.338 0Z"},
		SubMenuItem{URL: "api/tipdok/all", Name: "Vrste naloga", Icon: "M12 7.5h1.5m-1.5 3h1.5m-7.5 3h7.5m-7.5 3h7.5m3-9h3.375c.621 0 1.125.504 1.125 1.125V18a2.25 2.25 0 0 1-2.25 2.25M16.5 7.5V18a2.25 2.25 0 0 0 2.25 2.25M16.5 7.5V4.875c0-.621-.504-1.125-1.125-1.125H4.125C3.504 3.75 3 4.254 3 4.875V18a2.25 2.25 0 0 0 2.25 2.25h13.5M6 7.5h3v3H6v-3Z"},
		SubMenuItem{URL: "api/dokvrsta/all", Name: "Vrste dokumenata", Icon: "M3.75 12h16.5m-16.5 3.75h16.5M3.75 19.5h16.5M5.625 4.5h12.75a1.875 1.875 0 0 1 0 3.75H5.625a1.875 1.875 0 0 1 0-3.75Z"},
		SubMenuItem{URL: "api/fvknjrac/all", Name: "Vrste poreskih knjiga (KPR i KIR)", Icon: "M6.429 9.75 2.25 12l4.179 2.25m0-4.5 5.571 3 5.571-3m-11.142 0L2.25 7.5 12 2.25l9.75 5.25-4.179 2.25m0 0L21.75 12l-4.179 2.25m0 0 4.179 2.25L12 21.75 2.25 16.5l4.179-2.25m11.142 0-5.571 3-5.571-3"},
		SubMenuItem{URL: "api/fvepdv/all", Name: "Vrste evidencija PDV (EV PDV)", Icon: "m9 14.25 6-6m4.5-3.493V21.75l-3.75-1.5-3.75 1.5-3.75-1.5-3.75 1.5V4.757c0-1.108.806-2.057 1.907-2.185a48.507 48.507 0 0 1 11.186 0c1.1.128 1.907 1.077 1.907 2.185ZM9.75 9h.008v.008H9.75V9Zm.375 0a.375.375 0 1 1-.75 0 .375.375 0 0 1 .75 0Zm4.125 4.5h.008v.008h-.008V13.5Zm.375 0a.375.375 0 1 1-.75 0 .375.375 0 0 1 .75 0Z"},
		SubMenuItem{URL: "api/orgjed/all", Name: "Organizacione jedinice", Icon: "M3.375 19.5h17.25m-17.25 0a1.125 1.125 0 0 1-1.125-1.125M3.375 19.5h7.5c.621 0 1.125-.504 1.125-1.125m-9.75 0V5.625m0 12.75v-1.5c0-.621.504-1.125 1.125-1.125m18.375 2.625V5.625m0 12.75c0 .621-.504 1.125-1.125 1.125m1.125-1.125v-1.5c0-.621-.504-1.125-1.125-1.125m0 3.75h-7.5A1.125 1.125 0 0 1 12 18.375m9.75-12.75c0-.621-.504-1.125-1.125-1.125H3.375c-.621 0-1.125.504-1.125 1.125m19.5 0v1.5c0 .621-.504 1.125-1.125 1.125M2.25 5.625v1.5c0 .621.504 1.125 1.125 1.125m0 0h17.25m-17.25 0h7.5c.621 0 1.125.504 1.125 1.125M3.375 8.25c-.621 0-1.125.504-1.125 1.125v1.5c0 .621.504 1.125 1.125 1.125m17.25-3.75h-7.5c-.621 0-1.125.504-1.125 1.125m8.625-1.125c.621 0 1.125.504 1.125 1.125v1.5c0 .621-.504 1.125-1.125 1.125m-17.25 0h7.5m-7.5 0c-.621 0-1.125.504-1.125 1.125v1.5c0 .621.504 1.125 1.125 1.125M12 10.875v-1.5m0 1.5c0 .621-.504 1.125-1.125 1.125M12 10.875c0 .621.504 1.125 1.125 1.125m-2.25 0c.621 0 1.125.504 1.125 1.125M13.125 12h7.5m-7.5 0c-.621 0-1.125.504-1.125 1.125M20.625 12c.621 0 1.125.504 1.125 1.125v1.5c0 .621-.504 1.125-1.125 1.125m-17.25 0h7.5M12 14.625v-1.5m0 1.5c0 .621-.504 1.125-1.125 1.125M12 14.625c0 .621.504 1.125 1.125 1.125m-2.25 0c.621 0 1.125.504 1.125 1.125m0 1.5v-1.5m0 0c0-.621.504-1.125 1.125-1.125m0 0h7.5"},
		SubMenuItem{URL: "api/mestotroska/all", Name: "Mesta troškova", Icon: "M19.5 14.25v-2.625a3.375 3.375 0 0 0-3.375-3.375h-1.5A1.125 1.125 0 0 1 13.5 7.125v-1.5a3.375 3.375 0 0 0-3.375-3.375H8.25m0 11.625h4.5m-4.5 2.25h4.5m2.121 1.527c-1.171 1.464-3.07 1.464-4.242 0-1.172-1.465-1.172-3.84 0-5.304 1.171-1.464 3.07-1.464 4.242 0M10.5 2.25H5.625c-.621 0-1.125.504-1.125 1.125v17.25c0 .621.504 1.125 1.125 1.125h12.75c.621 0 1.125-.504 1.125-1.125V11.25a9 9 0 0 0-9-9Z"},
		SubMenuItem{URL: "api/drzava/all", Name: "Države", Icon: "M9 6.75V15m6-6v8.25m.503 3.498 4.875-2.437c.381-.19.622-.58.622-1.006V4.82c0-.836-.88-1.38-1.628-1.006l-3.869 1.934c-.317.159-.69.159-1.006 0L9.503 3.252a1.125 1.125 0 0 0-1.006 0L3.622 5.689C3.24 5.88 3 6.27 3 6.695V19.18c0 .836.88 1.38 1.628 1.006l3.869-1.934c.317-.159.69-.159 1.006 0l4.994 2.497c.317.158.69.158 1.006 0Z"},
		SubMenuItem{URL: "api/sifop/all", Name: "Opštine", Icon: "M12 21v-8.25M15.75 21v-8.25M8.25 21v-8.25M3 9l9-6 9 6m-1.5 12V10.332A48.36 48.36 0 0 0 12 9.75c-2.551 0-5.056.2-7.5.582V21M3 21h18M12 6.75h.008v.008H12V6.75Z"},
		SubMenuItem{URL: "api/sifmesto/all", Name: "Mesta", Icon: "M13.5 21v-7.5a.75.75 0 0 1 .75-.75h3a.75.75 0 0 1 .75.75V21m-4.5 0H2.36m11.14 0H18m0 0h3.64m-1.39 0V9.349M3.75 21V9.349m0 0a3.001 3.001 0 0 0 3.75-.615A2.993 2.993 0 0 0 9.75 9.75c.896 0 1.7-.393 2.25-1.016a2.993 2.993 0 0 0 2.25 1.016c.896 0 1.7-.393 2.25-1.015a3.001 3.001 0 0 0 3.75.614m-16.5 0a3.004 3.004 0 0 1-.621-4.72l1.189-1.19A1.5 1.5 0 0 1 5.378 3h13.243a1.5 1.5 0 0 1 1.06.44l1.19 1.189a3 3 0 0 1-.621 4.72M6.75 18h3.75a.75.75 0 0 0 .75-.75V13.5a.75.75 0 0 0-.75-.75H6.75a.75.75 0 0 0-.75.75v3.75c0 .414.336.75.75.75Z"},
		SubMenuItem{URL: "api/banke/all", Name: "Banke", Icon: "M3.75 21h16.5M4.5 3h15M5.25 3v18m13.5-18v18M9 6.75h1.5m-1.5 3h1.5m-1.5 3h1.5m3-6H15m-1.5 3H15m-1.5 3H15M9 21v-3.375c0-.621.504-1.125 1.125-1.125h3.75c.621 0 1.125.504 1.125 1.125V21"},
		SubMenuItem{URL: "api/bnkizv/all", Name: "Banke za izvozne fakture", Icon: "M2.25 21h19.5m-18-18v18m10.5-18v18m6-13.5V21M6.75 6.75h.75m-.75 3h.75m-.75 3h.75m3-6h.75m-.75 3h.75m-.75 3h.75M6.75 21v-3.375c0-.621.504-1.125 1.125-1.125h2.25c.621 0 1.125.504 1.125 1.125V21M3 3h12m-.75 4.5H21m-3.75 3.75h.008v.008h-.008v-.008Zm0 3h.008v.008h-.008v-.008Zm0 3h.008v.008h-.008v-.008Z"},
		SubMenuItem{URL: "api/sifplizvodi/all", Name: "Šifre plaćanja za domaći promet", Icon: "M8.242 5.992h12m-12 6.003H20.24m-12 5.999h12M4.117 7.495v-3.75H2.99m1.125 3.75H2.99m1.125 0H5.24m-1.92 2.577a1.125 1.125 0 1 1 1.591 1.59l-1.83 1.83h2.16M2.99 15.745h1.125a1.125 1.125 0 0 1 0 2.25H3.74m0-.002h.375a1.125 1.125 0 0 1 0 2.25H2.99"},
	},
	"Finansijsko": {
		SubMenuItem{URL: "api/fkpl/all", Name: "Kontni plan", Icon: ""},
		SubMenuItem{URL: "api/nalozi/all/tipdok", Name: "Nalozi", Icon: ""},
		SubMenuItem{URL: "", Name: "Promet", Icon: ""},
		SubMenuItem{URL: "", Name: "Salda konta", Icon: ""},
		SubMenuItem{URL: "", Name: "Kompenzacije", Icon: ""},
		SubMenuItem{URL: "", Name: "Otvorene stavke", Icon: ""},
		SubMenuItem{URL: "", Name: "Obračun kamate", Icon: ""},
		SubMenuItem{URL: "", Name: "Bilansi", Icon: ""},
		SubMenuItem{URL: "", Name: "Poreske knjige", Icon: ""},
		SubMenuItem{URL: "", Name: "Učitavanje i knjiženje izvoda", Icon: ""},
	},
	"Komercijalno poslovanje": {
		SubMenuItem{URL: "", Name: "Zahtev za nabavku", Icon: ""},
		SubMenuItem{URL: "", Name: "Zahtev za ponudu", Icon: ""},
		SubMenuItem{URL: "", Name: "Porudžbenice dobavljačima", Icon: ""},
		SubMenuItem{URL: "", Name: "Prijemnica materijala/sirovine", Icon: ""},
		SubMenuItem{URL: "", Name: "Profakture", Icon: ""},
		SubMenuItem{URL: "", Name: "Nalozi za isporuku", Icon: ""},
		SubMenuItem{URL: "", Name: "Otpremanje robe", Icon: ""},
	},
	"Robno": {
		SubMenuItem{URL: "", Name: "Robna Dokumenta", Icon: ""},
		SubMenuItem{URL: "", Name: "Promet po artiklima i kontima", Icon: ""},
		SubMenuItem{URL: "", Name: "Stanje po artiklima i kontima", Icon: ""},
		SubMenuItem{URL: "", Name: "Izveštaji pregled ulaza/izlaza", Icon: ""},
		SubMenuItem{URL: "", Name: "Artikli", Icon: ""},
		SubMenuItem{URL: "", Name: "Jedinice mere", Icon: ""},
		SubMenuItem{URL: "", Name: "Robne grupe", Icon: ""},
		SubMenuItem{URL: "", Name: "Robne podgrupe", Icon: ""},
		SubMenuItem{URL: "", Name: "Mesta isporuke", Icon: ""},
		SubMenuItem{URL: "", Name: "Komercijalisti", Icon: ""},
		SubMenuItem{URL: "", Name: "Poreske stope", Icon: ""},
		SubMenuItem{URL: "", Name: "Vrste dokumenata", Icon: ""},
		SubMenuItem{URL: "", Name: "Cenovnik", Icon: ""},
		SubMenuItem{URL: "", Name: "Tipovi knjižnih pisama", Icon: ""},
		SubMenuItem{URL: "", Name: "Magacini", Icon: ""},
		SubMenuItem{URL: "", Name: "Komercijalni podaci", Icon: ""},
		SubMenuItem{URL: "", Name: "Kraj poslovne godine", Icon: ""},
	},
	"Blagajna": {
		SubMenuItem{URL: "", Name: "Dnevnik blagajne", Icon: ""},
		SubMenuItem{URL: "", Name: "Kontiranje blagajne", Icon: ""},
		SubMenuItem{URL: "", Name: "Knjiženje blagajne", Icon: ""},
		SubMenuItem{URL: "", Name: "Specifikacija apoena", Icon: ""},
		SubMenuItem{URL: "", Name: "Valute", Icon: ""},
		SubMenuItem{URL: "", Name: "Kursna lista", Icon: ""},
		SubMenuItem{URL: "", Name: "Apoeni", Icon: ""},
		SubMenuItem{URL: "", Name: "Tipovi blagajne", Icon: ""},
		SubMenuItem{URL: "", Name: "Vrste stavki blagajne", Icon: ""},
		SubMenuItem{URL: "", Name: "Dodatna knjiženja", Icon: ""},
		SubMenuItem{URL: "", Name: "Putni nalozi", Icon: ""},
		SubMenuItem{URL: "", Name: "Dnevnice", Icon: ""},
		SubMenuItem{URL: "", Name: "Tipovi troškova", Icon: ""},
		SubMenuItem{URL: "", Name: "Prevozna sredstva", Icon: ""},
	},
	"Proizvodnja": {
		SubMenuItem{URL: "", Name: "Matični podaci normativa", Icon: ""},
		SubMenuItem{URL: "", Name: "Konta troškova", Icon: ""},
		SubMenuItem{URL: "", Name: "Radni nalog za proizvodnju", Icon: ""},
		SubMenuItem{URL: "", Name: "Unos dokumenta", Icon: ""},
		SubMenuItem{URL: "", Name: "Pregled dokumenta", Icon: ""},
		SubMenuItem{URL: "", Name: "Specifikacija dokumenta", Icon: ""},
		SubMenuItem{URL: "", Name: "Knjiženje dokumenta", Icon: ""},
		SubMenuItem{URL: "", Name: "Promet po artiklima i kontima", Icon: ""},
		SubMenuItem{URL: "", Name: "Stanje po artiklima i kontima", Icon: ""},
		SubMenuItem{URL: "", Name: "Trenutno stanje zaliha", Icon: ""},
		SubMenuItem{URL: "", Name: "Pregled stanja i cena artikala", Icon: ""},
		SubMenuItem{URL: "", Name: "Obračunska kalkulacija", Icon: ""},
		SubMenuItem{URL: "", Name: "Kalkulacija proizvoda", Icon: ""},
		SubMenuItem{URL: "", Name: "Sitan inventar - zaduženje", Icon: ""},
		SubMenuItem{URL: "", Name: "Izveštaji", Icon: ""},
		SubMenuItem{URL: "", Name: "Kraj poslovne godine", Icon: ""},
	},
	"Osnovna sredstva": {
		SubMenuItem{URL: "", Name: "Maticni podaci osnovnih sredstava", Icon: ""},
		SubMenuItem{URL: "", Name: "Uknjižavanje, promene", Icon: ""},
		SubMenuItem{URL: "", Name: "Obračun amortizacije", Icon: ""},
		SubMenuItem{URL: "", Name: "Obračun poreske amortizacije", Icon: ""},
		SubMenuItem{URL: "", Name: "Amortizacione grupe", Icon: ""},
		SubMenuItem{URL: "", Name: "Amortizacione podgrupe", Icon: ""},
		SubMenuItem{URL: "", Name: "Organizacione jedinice", Icon: ""},
		SubMenuItem{URL: "", Name: "Rukovaoci", Icon: ""},
		SubMenuItem{URL: "", Name: "Objekti", Icon: ""},
		SubMenuItem{URL: "", Name: "Konta nabavke i ispravke", Icon: ""},
		SubMenuItem{URL: "", Name: "Koeficijenti revalorizacije", Icon: ""},
		SubMenuItem{URL: "", Name: "Stanje osnovnih sredstava", Icon: ""},
		SubMenuItem{URL: "", Name: "Kartica osnovnog sredstva", Icon: ""},
		SubMenuItem{URL: "", Name: "Izveštaji", Icon: ""},
		SubMenuItem{URL: "", Name: "Kraj poslovne godine", Icon: ""},
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
