-- Menu and SubMenu Tables Creation and Data Insert

-- Create menuitems table
create table if not exists menuitems (
    id serial PRIMARY KEY,
    menuname varchar(100) unique not null,
    displayname varchar(200) not null,
    icon varchar(100),
    sortorder int DEFAULT 0,
    xdatunosa timestamp without time zone,
    xdatizmene timestamp without time zone,
    xopunos character varying(30) COLLATE pg_catalog."default",
    xopizmene character varying(30) COLLATE pg_catalog."default"
);

create table if not exists submenuitems (
    id serial PRIMARY KEY,
    menuid int not null REFERENCES menuitems(id) ON DELETE CASCADE,
    submenuname varchar(200) not null,
    urlmenu varchar(255),
    icon varchar(100),
    sortorder int DEFAULT 0,
    xdatunosa timestamp without time zone,
    xdatizmene timestamp without time zone,
    xopunos character varying(30) COLLATE pg_catalog."default",
    xopizmene character varying(30) COLLATE pg_catalog."default"
);

-- Clear existing data (optional, comment out if you want to preserve existing data)
-- DELETE FROM submenuitems;
-- DELETE FROM menuitems;

-- Insert Menu Items
INSERT INTO menuitems (menuname, displayname, icon, sortorder) VALUES
('opsti_podaci', 'Opšti podaci/Sifarnici', '', 1),
('finansijsko', 'Finansijsko', '', 2),
('komercijalno_poslovanje', 'Komercijalno poslovanje', '', 3),
('robno', 'Robno', '', 4),
('blagajna', 'Blagajna', '', 5),
('proizvodnja', 'Proizvodnja', '', 6),
('osnovna_sredstva', 'Osnovna sredstva', '', 7)
ON CONFLICT (menuname) DO NOTHING;

-- Insert Sub-Menu Items for "Opšti podaci"
INSERT INTO submenuitems (menuid, submenuname, urlmenu, icon, sortorder) 
SELECT id, 'Kupci/Dobavljači', 'api/partneri/all', 'partneri', 1 FROM menuitems WHERE menuname = 'opsti_podaci'
UNION ALL
SELECT id, 'Vrste naloga', 'api/tipdok/all', 'vrstenaloga', 2 FROM menuitems WHERE menuname = 'opsti_podaci'
UNION ALL
SELECT id, 'Vrste dokumenata', 'api/dokvrsta/all', 'vrstedokumenta', 3 FROM menuitems WHERE menuname = 'opsti_podaci'
UNION ALL
SELECT id, 'Vrste poreskih knjiga (KPR i KIR)', 'api/fvknjrac/all', 'vrsteporknjige', 4 FROM menuitems WHERE menuname = 'opsti_podaci'
UNION ALL
SELECT id, 'Vrste evidencija PDV (EV PDV)', 'api/fvepdv/all', 'vrsteevpdv', 5 FROM menuitems WHERE menuname = 'opsti_podaci'
UNION ALL
SELECT id, 'Organizacione jedinice', 'api/orgjed/all', 'oj', 6 FROM menuitems WHERE menuname = 'opsti_podaci'
UNION ALL
SELECT id, 'Mesta troškova', 'api/mestotroska/all', 'mtroska', 7 FROM menuitems WHERE menuname = 'opsti_podaci'
UNION ALL
SELECT id, 'Države', 'api/drzava/all', 'drzave', 8 FROM menuitems WHERE menuname = 'opsti_podaci'
UNION ALL
SELECT id, 'Opštine', 'api/sifop/all', 'opstine', 9 FROM menuitems WHERE menuname = 'opsti_podaci'
UNION ALL
SELECT id, 'Mesta', 'api/sifmesto/all', 'mesta', 10 FROM menuitems WHERE menuname = 'opsti_podaci'
UNION ALL
SELECT id, 'Banke', 'api/banke/all', 'banke', 11 FROM menuitems WHERE menuname = 'opsti_podaci'
UNION ALL
SELECT id, 'Banke za izvozne fakture', 'api/bnkizv/all', 'bankeizv', 12 FROM menuitems WHERE menuname = 'opsti_podaci'
UNION ALL
SELECT id, 'Šifre plaćanja za domaći promet', 'api/sifplizvodi/all', 'sifplizv', 13 FROM menuitems WHERE menuname = 'opsti_podaci'
ON CONFLICT DO NOTHING;

-- Insert Sub-Menu Items for "Finansijsko"
INSERT INTO submenuitems (menuid, submenuname, urlmenu, icon, sortorder) 
SELECT id, 'Tipovi analitike', 'api/tipanalitike/all', 'tipanalitike', 1 FROM menuitems WHERE menuname = 'finansijsko'
UNION ALL
SELECT id, 'Kontni plan', 'api/fkpl/all', 'kontniplan', 2 FROM menuitems WHERE menuname = 'finansijsko'
UNION ALL
SELECT id, 'Nalozi', 'api/nalozi/all/tipdok', 'fin_nalozi', 3 FROM menuitems WHERE menuname = 'finansijsko'
UNION ALL
SELECT id, 'Dnevnik knjiženja', 'api/dnevnik', '', 4 FROM menuitems WHERE menuname = 'finansijsko'
UNION ALL
SELECT id, 'Promet', 'api/promet', 'fin_promet', 5 FROM menuitems WHERE menuname = 'finansijsko'
UNION ALL
SELECT id, 'Salda konta', 'api/salda', 'fin_saldakonta', 6 FROM menuitems WHERE menuname = 'finansijsko'
UNION ALL
SELECT id, 'Kompenzacije', 'api/kompenzacije', 'fin_kompenzacije', 7 FROM menuitems WHERE menuname = 'finansijsko'
UNION ALL
SELECT id, 'Otvorene stavke', '/api/otvorenestavke', 'fin_otvorenestavke', 8 FROM menuitems WHERE menuname = 'finansijsko'
UNION ALL
SELECT id, 'Bilansi', '/api/bilansi', 'fin_bilansi', 9 FROM menuitems WHERE menuname = 'finansijsko'
UNION ALL
SELECT id, 'Poreske knjige i prijava', '/api/poreskeknjige', '', 10 FROM menuitems WHERE menuname = 'finansijsko'
UNION ALL
SELECT id, 'POPDV, PPPDV i PPV prijave', '/api/poreskeprijave', '', 11 FROM menuitems WHERE menuname = 'finansijsko'
UNION ALL
SELECT id, 'Evidencija predhodnog poreza EPP', '/api/fsepp', '', 12 FROM menuitems WHERE menuname = 'finansijsko'
UNION ALL
SELECT id, 'Učitavanje i knjiženje izvoda', '/api/izvodi', '', 13 FROM menuitems WHERE menuname = 'finansijsko'
UNION ALL
SELECT id, 'Obračun kamate', '/api/kamate', 'fin_obracunkamate', 14 FROM menuitems WHERE menuname = 'finansijsko'
ON CONFLICT DO NOTHING;

-- Insert Sub-Menu Items for "Komercijalno poslovanje"
INSERT INTO submenuitems (menuid, submenuname, urlmenu, icon, sortorder) 
SELECT id, 'Zahtev za nabavku', '', '', 1 FROM menuitems WHERE menuname = 'komercijalno_poslovanje'
UNION ALL
SELECT id, 'Zahtev za ponudu', '', '', 2 FROM menuitems WHERE menuname = 'komercijalno_poslovanje'
UNION ALL
SELECT id, 'Porudžbenice dobavljačima', '', '', 3 FROM menuitems WHERE menuname = 'komercijalno_poslovanje'
UNION ALL
SELECT id, 'Prijemnica materijala/sirovine', '', '', 4 FROM menuitems WHERE menuname = 'komercijalno_poslovanje'
UNION ALL
SELECT id, 'Profakture', '', '', 5 FROM menuitems WHERE menuname = 'komercijalno_poslovanje'
UNION ALL
SELECT id, 'Nalozi za isporuku', '', '', 6 FROM menuitems WHERE menuname = 'komercijalno_poslovanje'
UNION ALL
SELECT id, 'Otpremanje robe', '', '', 7 FROM menuitems WHERE menuname = 'komercijalno_poslovanje'
ON CONFLICT DO NOTHING;

-- Insert Sub-Menu Items for "Robno"
INSERT INTO submenuitems (menuid, submenuname, urlmenu, icon, sortorder) 
SELECT id, 'Robna Dokumenta', '', '', 1 FROM menuitems WHERE menuname = 'robno'
UNION ALL
SELECT id, 'Promet po artiklima i kontima', '', '', 2 FROM menuitems WHERE menuname = 'robno'
UNION ALL
SELECT id, 'Stanje po artiklima i kontima', '', '', 3 FROM menuitems WHERE menuname = 'robno'
UNION ALL
SELECT id, 'Izveštaji pregled ulaza/izlaza', '', '', 4 FROM menuitems WHERE menuname = 'robno'
UNION ALL
SELECT id, 'Artikli', '', '', 5 FROM menuitems WHERE menuname = 'robno'
UNION ALL
SELECT id, 'Jedinice mere', 'api/jedmere/all', '', 6 FROM menuitems WHERE menuname = 'robno'
UNION ALL
SELECT id, 'Robne grupe', '', '', 7 FROM menuitems WHERE menuname = 'robno'
UNION ALL
SELECT id, 'Robne podgrupe', '', '', 8 FROM menuitems WHERE menuname = 'robno'
UNION ALL
SELECT id, 'Mesta isporuke', '', '', 9 FROM menuitems WHERE menuname = 'robno'
UNION ALL
SELECT id, 'Komercijalisti', '', '', 10 FROM menuitems WHERE menuname = 'robno'
UNION ALL
SELECT id, 'Poreske stope', '', '', 11 FROM menuitems WHERE menuname = 'robno'
UNION ALL
SELECT id, 'Vrste dokumenata', '', '', 12 FROM menuitems WHERE menuname = 'robno'
UNION ALL
SELECT id, 'Cenovnik', '', '', 13 FROM menuitems WHERE menuname = 'robno'
UNION ALL
SELECT id, 'Tipovi knjižnih pisama', '', '', 14 FROM menuitems WHERE menuname = 'robno'
UNION ALL
SELECT id, 'Magacini', '', '', 15 FROM menuitems WHERE menuname = 'robno'
UNION ALL
SELECT id, 'Komercijalni podaci', '', '', 16 FROM menuitems WHERE menuname = 'robno'
UNION ALL
SELECT id, 'Kraj poslovne godine', '', '', 17 FROM menuitems WHERE menuname = 'robno'
ON CONFLICT DO NOTHING;

-- Insert Sub-Menu Items for "Blagajna"
INSERT INTO submenuitems (menuid, submenuname, urlmenu, icon, sortorder) 
SELECT id, 'Dnevnik blagajne', '', '', 1 FROM menuitems WHERE menuname = 'blagajna'
UNION ALL
SELECT id, 'Kontiranje blagajne', '', '', 2 FROM menuitems WHERE menuname = 'blagajna'
UNION ALL
SELECT id, 'Knjiženje blagajne', '', '', 3 FROM menuitems WHERE menuname = 'blagajna'
UNION ALL
SELECT id, 'Specifikacija apoena', '', '', 4 FROM menuitems WHERE menuname = 'blagajna'
UNION ALL
SELECT id, 'Valute', '', '', 5 FROM menuitems WHERE menuname = 'blagajna'
UNION ALL
SELECT id, 'Kursna lista', '', '', 6 FROM menuitems WHERE menuname = 'blagajna'
UNION ALL
SELECT id, 'Apoeni', '', '', 7 FROM menuitems WHERE menuname = 'blagajna'
UNION ALL
SELECT id, 'Tipovi blagajne', '', '', 8 FROM menuitems WHERE menuname = 'blagajna'
UNION ALL
SELECT id, 'Vrste stavki blagajne', '', '', 9 FROM menuitems WHERE menuname = 'blagajna'
UNION ALL
SELECT id, 'Dodatna knjiženja', '', '', 10 FROM menuitems WHERE menuname = 'blagajna'
UNION ALL
SELECT id, 'Putni nalozi', '', '', 11 FROM menuitems WHERE menuname = 'blagajna'
UNION ALL
SELECT id, 'Dnevnice', '', '', 12 FROM menuitems WHERE menuname = 'blagajna'
UNION ALL
SELECT id, 'Tipovi troškova', '', '', 13 FROM menuitems WHERE menuname = 'blagajna'
UNION ALL
SELECT id, 'Prevozna sredstva', '', '', 14 FROM menuitems WHERE menuname = 'blagajna'
ON CONFLICT DO NOTHING;

-- Insert Sub-Menu Items for "Proizvodnja"
INSERT INTO submenuitems (menuid, submenuname, urlmenu, icon, sortorder) 
SELECT id, 'Matični podaci normativa', '', '', 1 FROM menuitems WHERE menuname = 'proizvodnja'
UNION ALL
SELECT id, 'Konta troškova', '', '', 2 FROM menuitems WHERE menuname = 'proizvodnja'
UNION ALL
SELECT id, 'Radni nalog za proizvodnju', '', '', 3 FROM menuitems WHERE menuname = 'proizvodnja'
UNION ALL
SELECT id, 'Unos dokumenta', '', '', 4 FROM menuitems WHERE menuname = 'proizvodnja'
UNION ALL
SELECT id, 'Pregled dokumenta', '', '', 5 FROM menuitems WHERE menuname = 'proizvodnja'
UNION ALL
SELECT id, 'Specifikacija dokumenta', '', '', 6 FROM menuitems WHERE menuname = 'proizvodnja'
UNION ALL
SELECT id, 'Knjiženje dokumenta', '', '', 7 FROM menuitems WHERE menuname = 'proizvodnja'
UNION ALL
SELECT id, 'Promet po artiklima i kontima', '', '', 8 FROM menuitems WHERE menuname = 'proizvodnja'
UNION ALL
SELECT id, 'Stanje po artiklima i kontima', '', '', 9 FROM menuitems WHERE menuname = 'proizvodnja'
UNION ALL
SELECT id, 'Trenutno stanje zaliha', '', '', 10 FROM menuitems WHERE menuname = 'proizvodnja'
UNION ALL
SELECT id, 'Pregled stanja i cena artikala', '', '', 11 FROM menuitems WHERE menuname = 'proizvodnja'
UNION ALL
SELECT id, 'Obračunska kalkulacija', '', '', 12 FROM menuitems WHERE menuname = 'proizvodnja'
UNION ALL
SELECT id, 'Kalkulacija proizvoda', '', '', 13 FROM menuitems WHERE menuname = 'proizvodnja'
UNION ALL
SELECT id, 'Sitan inventar - zaduženje', '', '', 14 FROM menuitems WHERE menuname = 'proizvodnja'
UNION ALL
SELECT id, 'Izveštaji', '', '', 15 FROM menuitems WHERE menuname = 'proizvodnja'
UNION ALL
SELECT id, 'Kraj poslovne godine', '', '', 16 FROM menuitems WHERE menuname = 'proizvodnja'
ON CONFLICT DO NOTHING;

-- Insert Sub-Menu Items for "Osnovna sredstva"
INSERT INTO submenuitems (menuid, submenuname, urlmenu, icon, sortorder) 
SELECT id, 'Maticni podaci osnovnih sredstava', '', '', 1 FROM menuitems WHERE menuname = 'osnovna_sredstva'
UNION ALL
SELECT id, 'Uknjižavanje, promene', '', '', 2 FROM menuitems WHERE menuname = 'osnovna_sredstva'
UNION ALL
SELECT id, 'Obračun amortizacije', '', '', 3 FROM menuitems WHERE menuname = 'osnovna_sredstva'
UNION ALL
SELECT id, 'Obračun poreske amortizacije', '', '', 4 FROM menuitems WHERE menuname = 'osnovna_sredstva'
UNION ALL
SELECT id, 'Amortizacione grupe', 'api/oamgrp/all', '', 5 FROM menuitems WHERE menuname = 'osnovna_sredstva'
UNION ALL
SELECT id, 'Amortizacione podgrupe', '', '', 6 FROM menuitems WHERE menuname = 'osnovna_sredstva'
UNION ALL
SELECT id, 'Organizacione jedinice', '', '', 7 FROM menuitems WHERE menuname = 'osnovna_sredstva'
UNION ALL
SELECT id, 'Rukovaoci', '', '', 8 FROM menuitems WHERE menuname = 'osnovna_sredstva'
UNION ALL
SELECT id, 'Objekti', '', '', 9 FROM menuitems WHERE menuname = 'osnovna_sredstva'
UNION ALL
SELECT id, 'Konta nabavke i ispravke', '', '', 10 FROM menuitems WHERE menuname = 'osnovna_sredstva'
UNION ALL
SELECT id, 'Koeficijenti revalorizacije', '', '', 11 FROM menuitems WHERE menuname = 'osnovna_sredstva'
UNION ALL
SELECT id, 'Stanje osnovnih sredstava', '', '', 12 FROM menuitems WHERE menuname = 'osnovna_sredstva'
UNION ALL
SELECT id, 'Kartica osnovnog sredstva', '', '', 13 FROM menuitems WHERE menuname = 'osnovna_sredstva'
UNION ALL
SELECT id, 'Izveštaji', '', '', 14 FROM menuitems WHERE menuname = 'osnovna_sredstva'
UNION ALL
SELECT id, 'Kraj poslovne godine', '', '', 15 FROM menuitems WHERE menuname = 'osnovna_sredstva'
ON CONFLICT DO NOTHING;

-- Query to verify data
-- SELECT m.id, m.name, m.displayname, COUNT(s.id) as submenu_count 
-- FROM menuitems m 
-- LEFT JOIN submenuitems s ON m.id = s.menuid 
-- GROUP BY m.id, m.name, m.displayname 
-- ORDER BY m.sortorder;
