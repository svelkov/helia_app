package domain

var MenuData = map[string][]SubMenuItem{
	"Opšti podaci": {
		SubMenuItem{URL: "api/partneri/all", Name: "Kupci/Dobavljači", Icon: "partneri"},
		SubMenuItem{URL: "api/tipdok/all", Name: "Vrste naloga", Icon: "vrstenaloga"},
		SubMenuItem{URL: "api/dokvrsta/all", Name: "Vrste dokumenata", Icon: "vrstedokumenta"},
		SubMenuItem{URL: "api/fvknjrac/all", Name: "Vrste poreskih knjiga (KPR i KIR)", Icon: "vrsteporknjige"},
		SubMenuItem{URL: "api/fvepdv/all", Name: "Vrste evidencija PDV (EV PDV)", Icon: "vrsteevpdv"},
		SubMenuItem{URL: "api/orgjed/all", Name: "Organizacione jedinice", Icon: "oj"},
		SubMenuItem{URL: "api/mestotroska/all", Name: "Mesta troškova", Icon: "mtroska"},
		SubMenuItem{URL: "api/drzava/all", Name: "Države", Icon: "drzave"},
		SubMenuItem{URL: "api/sifop/all", Name: "Opštine", Icon: "opstine"},
		SubMenuItem{URL: "api/sifmesto/all", Name: "Mesta", Icon: "mesta"},
		SubMenuItem{URL: "api/banke/all", Name: "Banke", Icon: "banke"},
		SubMenuItem{URL: "api/bnkizv/all", Name: "Banke za izvozne fakture", Icon: "bankeizv"},
		SubMenuItem{URL: "api/sifplizvodi/all", Name: "Šifre plaćanja za domaći promet", Icon: "sifplizv"},
	},
	"Finansijsko": {
		SubMenuItem{URL: "api/fkpl/all", Name: "Kontni plan", Icon: "kontniplan"},
		SubMenuItem{URL: "api/nalozi/all/tipdok", Name: "Nalozi", Icon: "fin_nalozi"},
		SubMenuItem{URL: "", Name: "Promet", Icon: "fin_nalozi1"},
		SubMenuItem{URL: "", Name: "Salda konta", Icon: "fin_nalozi2"},
		SubMenuItem{URL: "", Name: "Kompenzacije", Icon: "fin_nalozi3"},
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
		SubMenuItem{URL: "api/oamgrp/all", Name: "Amortizacione grupe", Icon: ""},
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
