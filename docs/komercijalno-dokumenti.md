# Komercijalno: the new documents and their differences from the legacy program

This file summarizes the komercijalno screens ported from the legacy WinDev program (seven documents and the
eOtpremnica, which is the otpremnica with the data of the e-otpremnica): what each new screen does and where it
deliberately differs from the legacy. All of them live in `internal/{domain,repository,service,handler}`
(`komercijalno` packages / `komercijalno_*` files), `frontend/templates/komercijalno` and
`frontend/templates/reports/komercijalno`.

| # | Document | Legacy window | Vrsta (group) | Route (menu URL) |
|---|----------|---------------|---------------|------------------|
| 1 | Zahtev za nabavku | pr_NarudzbenicaPROIZV | 300 (INR) | `/api/komercijalno/zahtevi-nabavka` |
| 2 | Zahtev za ponudu | pr_ZahtevPONUDA | 310 (ZPN) | `/api/komercijalno/zahtevi-ponuda` |
| 3 | Porudžbenica dobavljaču | pr_PorudzbenicaDOBAVLJ | 320 (NRD) | `/api/komercijalno/porudzbenice-dobavljaca` |
| 4 | Prijemnica materijala po porudžbenici | pr_PrijemnicaSIROVINE | ULS (198/199) or PLT (110/111) | `/api/komercijalno/prijemnice-sirovina` |
| 5 | Profaktura (predračun) | pr_Profakture | 360 (PRF) | `/api/komercijalno/profakture` |
| 6 | Nalog za isporuku (porudžbenica kupca) | pr_PorudzbenicaProdaja | 350 (NRK) | `/api/komercijalno/nalozi-isporuka` |
| 7 | Otpremnica robe/gotovih proizvoda | pr_Otpremnica | FPR 188/189, FAK 130–137, FRP 132/139 | `/api/komercijalno/otpremnice` |
| 8 | eOtpremnica (the otpremnica with the data of the e-otpremnica) | pr_eOTPREMNICA + IW_Otpremnice | same as 7 | `/api/komercijalno/eotpremnice` |

The documents form one flow:

- **Nabavka:** zahtev za nabavku → zahtev za ponudu → porudžbenica dobavljaču → prijemnica (posts ulaz to the stanje).
- **Prodaja:** profaktura → nalog za isporuku (reserves stock) → otpremnica (creates and posts the robna faktura). The
  eOtpremnica edits the same otpremnice and adds the data of the e-otpremnica.

Every document is stored in `edok` / `epro` (not in `rdok` / `rpro` like the robno documents). Only the otpremnica
writes to `rdok` / `rpro`, when it is confirmed. The eOtpremnica also uses the tables `enot` (faze vožnje) and `eopma`
(opasne materije) and the columns of `migrations/003_eotpremnica.sql`.

The menu rows (`submenuitems.urlmenu`) are added by hand in the database; the code has no menu entries. "Otpremanje
robe" must point to `/api/komercijalno/otpremnice` and "eOtpremnica" to `/api/komercijalno/eotpremnice`.

---

## Common to all documents

**How the screens work (different from the legacy windows):**

- **Main screen first.** The legacy window opens on tab 1 (the entry) with the pregled on tab 2. The new screen opens
  on the pregled ("Pregled ..."): filters (magacin, Nepotvrđeni / Potvrđeni, "Prikaži dokumenta za sve magacine"), a
  search, the list, and "Detalji dokumenta" (the stavke of the clicked row).
- **Entry in a dialog.** "Dodaj" and "Izmeni" of the list open the entry dialog (the legacy tab 1). "Briši" is a row
  action of the list with "Brisanje dokumenta?". The dialog has no "Briši" and no "Novi dok.".
- **Header and stavke as two panels** (the pattern of the robno fakture): "Sačuvaj" locks the header and opens the
  stavke; "Izmeni" in the header title opens the header again.
- **Changing a stavka** is done with "Izmeni" on its row and a form, not by typing into the grid.
- **Štampa** is always visible in the dialog and disabled until the document is saved. On the main screen it prints the
  clicked row.
- **Titles** are in capitals (`translator.Title`), like finansijsko and robno.
- **Only new documents can change.** Izmeni / Briši are offered only on documents that the legacy also lets you change;
  a confirmed one shows the legacy message why.
- **Not ported in any screen:** the "Nova šifra" / "+ …" buttons that open other šifarnici (partneri, komercijalisti,
  artikli), the Republika Srpska PIB/JIB variant, and per-vrsta custom report names (`dokvrsta.dodoznfak`).

**Technical:**

- All SQL is in `KomercijalnoRepository` (embeds `RobnoRepository`), restricted to the session's god/kar, with golden
  tests; every statement is also checked by the read-only `TestRealDBExplain`.
- Confirmations that post stock reuse the robno service (`AzurirajRSTA`, `AzurirajDRSTA`, `AzurirajRNAL`,
  `ObracunFakture`, `IzracunajNCENA`).
- Translations are in `en.json`, `sr.json` and `ср.json`.

---

## 1. Zahtev za nabavku (pr_NarudzbenicaPROIZV)

**What it does:** a request of a magacin for materials. Header: vrsta, magacin, broj, datum, referent (the user's
komercijalista by default), OJ, mesto troška, napomena. Stavke: artikal (tip S, A, R, D, K) with its nabavna cena and the
tražena količina. Potvrda sets status 40. Print `PR_RPT_ZAHTEV_NABAVKA`.

**Differences from the legacy:**

- The main-screen / dialog layout described above (the first document converted to it).
- A šifra typed directly is accepted for any artikal tip; the search lists only tipovi S, A, R, D, K.

---

## 2. Zahtev za ponudu (pr_ZahtevPONUDA)

**What it does:**

- Built from the free stavke (ststatus "S") of potvrđeni zahtevi za nabavku in the magacin: tick them and "Sačuvaj" or
  "Dodaj izabrane stavke". A taken source stavka is marked "R" (`rproid1` links the copy) and is free again when its copy
  or the ponuda is deleted.
- Stavke can also be entered by hand.
- Header: rok nabavke ↔ očekivani datum isporuke, referent, OJ (required), mesto troška, tekst, valuta with the srednji
  kurs of the datum (`klista`), kurs, važi za, and the list of dobavljači (saved in `edok.vezniid` as "id;id;").
- Potvrda (`pr_PotvrdaPONUDE`): a dialog of the ponuda's dobavljači; Da sets status 40 and writes the chosen
  dobavljač's konto/šifra to `fkto`/`fana`.
- Print `PR_RPT_ZAHTEV_PONUDA` for a dobavljač chosen in a dialog.

**Differences from the legacy:**

| Legacy | New |
|--------|-----|
| A popup grid of all dobavljači (~62,000 rows) | A search box over the analitike of konto 43; picks are listed with × |
| `vezniid` (varchar 80) overflows with a database error | Validation message when the list is too long |
| "Sačuvaj" takes any ticked stavka | Only stavke still free on potvrđeni zahtevi are taken |

New shared piece: the kurs lookup (`robno_valute_repository.go`: `valute.vaziza`, `klista.srednji` of the datum, else 1);
`klista` was added to the tables without a period in `.claude/rules/database.md`.

---

## 3. Porudžbenica dobavljaču (pr_PorudzbenicaDOBAVLJ)

**What it does:**

- An order to one dobavljač (konto default `rvr.fdob` + šifra), optionally linked to a potvrđen zahtev za ponudu; choosing
  one copies its rok, način dopremanja, magacin, OJ, mesto troška and dobavljač.
- Stavke by hand (cena from `rcene.ncena`, entered in the document's valuta and converted to RSD with kurs / važi za;
  rabat), or from the linked ponuda with Dodaj, Dodaj sve, Vrati poslednju, Vrati sve.
- Header groups as in the legacy: uslovi nabavke, referent / valuta / kurs, OJ / mesto troška / napomena, posebni uslovi
  nabavke and prateća dokumentacija; "UVOZNI DOKUMENT – INO PARTNER" for a foreign valuta.
- Potvrda: stavke exist and every cena > 0, status 40.
- Print `PR_RPT_PORUDZB_DOB` (Serbian) or `…DOBINO` (English, foreign valuta), with "Naziv artikla u štampi" and
  "Komercijalni opis u štampi".

**Differences from the legacy:**

| Legacy | New |
|--------|-----|
| OJ and mesto troška are required | Optional ("-"); a chosen mesto troška must belong to the chosen OJ |
| Ponuda stavke are copied without a link and can be copied twice | A taken ponuda stavka is marked "R" and freed again when its copy is deleted |
| The pregled filters on a date range that is never set | The whole period (god/kar) is shown |
| Changing the valuta with stavke asks Da/Ne in a window | Browser confirm; prices are not recalculated (as in the legacy) |

---

## 4. Prijemnica materijala po porudžbenici (pr_PrijemnicaSIROVINE)

**What it does:**

- A receipt of a potvrđena porudžbenica dobavljaču (`vrdokid`). Vrste: group ULS (moduli PRO + SVI) in magacini of tip
  P, S, L, D; group PLT (ROB + SVI) in magacini of tip V. Header copied from the porudžbenica (dobavljač, rok, uslovi,
  valuta, komercijalista, OJ, mesto troška).
- Stavke from the porudžbenica (Dodaj / Dodaj sve: cena less rabat, the next OTK "001", "002", …; the same artikal twice
  is refused) or by hand (konto of the magacin, količina na dokumentu, primljena, serija required); Vrati poslednju /
  Vrati sve.
- **Pakovanje:** količina po pakovanju × broj pakovanja per stavka, saved in `kolpak`.
- **Karton identifikacije** print for the selected stavka (`RPT_KI`).
- **Potvrda** posts each stavka with the port of `AzurRSTA_ZN` / `AzurDRSTA_ZN`
  (`internal/service/robno/robnadokumenta_stanjazn_service.go`), always with the lot (OTK, serija, rok trajanja). A
  magacin with `analiza = 1` gets status 30 and the stock goes to `kolicner` (unavailable); otherwise status 40.
- Print `PR_RPT_PRIJEMNICA`.

**Differences from the legacy:**

| Legacy | New |
|--------|-----|
| "Vrati poslednju" removes the lowest rbr (reverse-sorted grid) | Removes the highest rbr |
| Storno sign only applied when a row is edited | Also applied on Dodaj |
| `kolpak` rows stay behind | Deleted with their stavka, with "Vrati sve" and with the document |
| Saving a saved prijemnica changes only `dadok` | Sets both `dadok` and `datotp` |
| A potvrđena prijemnica can still be deleted / its stavke changed | Blocked |
| Robno DRSTA drops the lot for tipzal 1 | The lot is always posted (as `AzurDRSTA_ZN` does) |

Not ported (agreed): editing the tražena količina in the porudžbenica grid, and "Dozvoli izmenu broja".

---

## 5. Profaktura / predračun (pr_Profakture)

**What it does:**

- Header: vrsta naloga za knjiženje (tipdok), magacin (tip G or V), vrsta, broj, datum; kupac (konto default `rvr.fkup`;
  choosing it fills its ugovoreni rabat and rok); rok ↔ valuta date; % ugovoreni rabat and % kase with their amounts;
  način isporuke, uslovi plaćanja; tip cenovnika (only tipcene 2); paritet, izvozna izjava; mesto isporuke, referent
  prodaje, OJ, mesto troška, tekst, žiro račun, poreska kategorija with its "član … stav …" text, valuta with kurs and
  važi za, uslovi kamate, po ugovoru.
- Stavke: šifra search over the magacin's zalihe ("Prikaz artikala sa stanjem > 0"), lot list for tipzal 2, price and
  rabat from `rcene` or the cenovnik (tipcene 2); a foreign-valuta price is entered in the valuta and stored in RSD; the
  planska cena comes from `pinorm` in tip G magacini. Header totals are recalculated after every stavka change.
- Statuses: 10 new, 20 potvrđena (Potvrda dok.), 40 nalog formed.
- **Form. nalog:** creates the porudžbenica kupca (vrsta 350) with a copy of the header and stavke, reserves the stock
  of each stavka (`drsta.rez`, `rsta.rez`), sets the profaktura to 40.
- Prints `PR_RPT_PROFAKTURA` (domestic, with PDV per stopa and slovima) and `PR_RPT_PROFAKTURAIZV` (English, valuta, no
  PDV).

**Differences from the legacy:**

| Legacy | New |
|--------|-----|
| Tipdok list: only tipdok whose `grpvrd` names the vrsta (none for 360 in the data) | Those, else all tipdok of PRO/ROB/SVI except "00" |
| Domestic print computes the kasa from `VKASE / 100` (the amount) | Uses the kasa percentage, like the izvoz print and the recalculation |
| Form. nalog reserves only `drsta.rez` of an existing lot | Reserves `drsta.rez` (a missing lot gets a row) and `rsta.rez` |

Kept as in the legacy although they look like legacy bugs: the formed nalog doesn't copy the kurs; in a foreign valuta
the default price is the RSD price from `rcene` unchanged; OJ and mesto troška are required.

---

## 6. Nalog za isporuku / porudžbenica kupca (pr_PorudzbenicaProdaja)

**What it does:**

- Two tabs on the main screen: **"Pregled naloga za isporuku"** (the start) and **"Pregled delimično realizovanih
  naloga"** (status 40 with a stavka delivered partly; such rows are red).
- The entry dialog is the profaktura's dialog with the nalog fields:
  - **Broj profakture:** a search over the potvrđene profakture of the magacin; choosing one fills the header. Saving
    copies its stavke, reserves them and sets the profaktura to 40.
  - **Banka za uplatu – izvoz** (`sifbank`), **Status dokumenta**.
  - **Finansijsko stanje kupca** (`SaldoKUPCA`): proknjiženo from `fpro`, neproknjiženo from the robne fakture not yet
    posted (with PDV), and the totals.
- Stavke reserve their lot in `drsta.rez` and `rsta.rez`; editing adjusts the reservation, deleting releases it.
  **"Skidaj zalihe po starosti"** (tipzal 2) splits the količina over the lots with free stock, oldest rok first.
- Potvrda: stavke exist and both prices > 0; status 20.
- Print `PR_RPT_NALOGPRODAJA` ("NALOG ZA OTPREMU BR.:").

**Differences from the legacy:**

| Legacy | New |
|--------|-----|
| `rsta.rez` only incremented for the split stavke, never decremented | `drsta.rez` and `rsta.rez` both kept right on every reservation and release (a release never takes `rsta.rez` below 0) |
| The remainder stavka without a lot gets its reservation only when its `drsta` row is new | Always |
| Splitting by age with no free stock adds nothing | The whole količina goes on a stavka without a lot |
| Editing a stavka overwrites the planska cena with the prodajna cena | The planska cena is kept |
| A potvrđen nalog (20) can still be changed | Only status 10 can be changed or deleted |
| Deleting a nalog sets its profaktura back to 10 | Back to 20, so it can be chosen again |
| Print shows the broj/datum of the otpremnica (empty before it exists) | Prints the nalog's broj and datum |
| Choosing a profaktura puts its kasa amount into the kasa % field | Takes the kasa percentage |
| Choosing a profaktura sets the unsaved nalog to 40 | Not done |
| The kurs stays when a profaktura is chosen | The kurs of the profaktura's valuta on the nalog's datum |
| New nalog stavke copy the tražena količina into `kolic` (the port did so) | `kolic` starts at 0; the otpremnica writes the delivered količina |
| "Delimično realizovani" lists every status-40 nalog with `kolic ≠ trazkolic` | Only after its otpremnica is confirmed |

"Br. otpremnice" and "Br. fakture" in the list come through the otpremnica (`edok.vrdokid` = the nalog,
`rdok.edokid` = the otpremnica).

Not ported (no code yet): "Pov. Avansa" (`pr_PovezivanjeAVANSA`), "Fakt. Pregled" (`PR_RPT_FAKT_OTP_TEST`). "Teren
prodaje" on the print is empty (no source in the legacy).

---

## 7. Otpremnica robe/gotovih proizvoda (pr_Otpremnica)

**What it does:**

- Made from a **potvrđen nalog za isporuku** (status 20) of the magacin; saving it sets the nalog to 40 and copies its
  stavke (isporučena = tražena, negative for a storno vrsta).
- The vrsta is the vrsta of the robna faktura that the potvrda creates: FPR from a G/S magacin, FAK/FRP from a V/M
  magacin.
- Until the potvrda the broj otpremnice and broj fakture are temporary (990001+). "Dozvoli izmenu br. OTP i br. Fakture"
  lets the user type the real ones (checked to be unique).
- Header: OJ, mesto troška, vrsta naloga za knjiženje (from the nalog), kupac with its **napomena** (saved in
  `edok.napomena`), poreska kategorija, žiro račun, vozač, način otpreme, broj ugovora, broj vozila, težine, kutije,
  palete (the transport rows aligned in two label columns), POPDV polje (PDV obveznik), valuta / kurs / važi za (only
  for a foreign kupac, `partneri.ter` = 3).
- Stavke ("Količine po nalogu za otpremu"): "Izmeni" changes the isporučena količina and, with "Dozvoli izmenu OTK,
  Seriju, Rok", the lot.
- A changed kurs recalculates the prodajne cene (cena u valuti × kurs / važi za); the prodajna vrednost is
  `PreracunFAKTURAPROIZVGLOBEDOK`.
- **Potvrda dok.** (checks: year not closed, stavke, status 10, nastanak PDV chosen, JM of the lot stavke, OTK question,
  prices > 0, total ≠ 0, enough stock per lot, storno lot had an ulaz; then "Da li želite da potvrdite dokument?"), in
  one transaction:
  - the robni nalog of the tipdok, datum and magacin, or a new "Aut. kreiran nalog";
  - the robna faktura (`rdok`, `edokid` = the otpremnica) with its stavke (`rpro`), the cena by `magacini.nacvodzal`
    (3 nabavna / prosečna nabavna, 1 veleprodajna, else the stavka's cena), tax category S20/S10 by tarifa;
  - a price row (`rcene`) for an artikal without one;
  - izlaz posted to `rsta` and `drsta`, the nalog totals, the avansi of the nalog;
  - the reservations of all stavke released; the delivered količine written back to the nalog's stavke;
  - the otpremnica set to 40 with its real numbers; then the faktura's obračun (`ObracunFakture`).
- Print `PR_RPT_OTPREMNICA` (Serbian, or the English header of the same report for a foreign valuta, with the exchange
  rate and Total RSD), with "Komercijalni opis u štampi".

**Differences from the legacy:**

| Legacy | New |
|--------|-----|
| Print options: otpremnica, račun, otpremnica za mesto isporuke, packing list | One print: the otpremnica (English header for a foreign valuta) |
| Napomena saved into `partneri.napomena` | Saved into `edok.napomena` (a new otpremnica takes it from the nalog) |
| Lot change: subtracts the tražena from the old lot, adds the isporučena to the new one | Moves the tražena reservation from the old lot to the new one |
| Potvrda releases only the delivered stavke | Releases the reservation of every stavka |
| A storno release adds to the reservation | Releases it |
| Stock check compares the tražena količina | Compares the isporučena količina |
| Prodajna vrednost recalculated only after a kurs change | On every Sačuvaj |
| Obračun of the faktura only with avansi | Always (iznos, ugovoreni rabat, kasa, avansi) |
| A confirmed otpremnica can be deleted (its faktura stays) | Only status 10 can be deleted; the nalog goes back to 20 (legacy 10) |
| A negative isporučena količina is not checked | Refused |
| Print title "Storno" decided by the knjiženje of the vrsta | By its predznak |
| Foreign prices on the print: prodajna cena / kurs / važi za | Cena u valuti |
| The OTK question is asked per stavka | Asked once, listing the stavke |
| Nastanak PDV and Opis dok. in the header | Next to "Potvrda dok." (the locked header can't send them); nastanak checked at the potvrda |
| The faktura has only the old šifre (mi, sifval, sifbank) | Also `fispid`, `idvalute`, `bnkizvid`, so the robno screens show them |

Not ported: the zbirni dokument (`BTN_TraziZbirnidok`, `SnimiRDOK_IRZbirno`), the interni-račun path (`SnimiRDOK_IR`,
vrste IRT/PPR/IRS, the firm with PIB 105690651 and its "Nema dovoljno zaliha, da li ste sigurni?"), the prints "Štampaj
račun", "za mesto isporuke" and the packing list. PRE and FZR vrste don't exist in the data.

---

## 8. eOtpremnica (pr_eOTPREMNICA with the internal window IW_Otpremnice)

**What it does:**

- The same otpremnice as section 7 (same `edok` / `epro`, vrste, nalog, numbers, Potvrda dok., print and pregled):
  the otpremnica handler and service run with the eOtpremnica's own routes (`/api/komercijalno/eotpremnice/...`). Both
  menu items list the same documents.
- The entry dialog has the header and one collapsible control per legacy screenshot, in this order:
  - **DET1 – Osnovni podaci, pošiljalac i primalac:**
    - Opasne materije / Roba treće strane (they enable their groups in DET5; both groups are always
      shown, greyed out until their checkbox is ticked);
    - tip otpremnice (Eksterna / Interna), broj otpremnice (`brotp`, read-only);
    - težina isporuke (bruto) and zapremina isporuke (bruto) with their jedinice mere (-, g, kg, t / -, m3, l);
    - stvarni datum i vreme otpreme, planirani početak and završetak otpreme;
    - broj paketa (`totboxs`) with povrat robe; broj ugovora, narudžbenice/ponude, okvirnog sporazuma;
    - the pošiljalac (the firm from `fvr`) and the primalac (the kupac, `fkto` / `fana`): naziv, adresa, mesto, PIB,
      JBKJS, matični broj, each with a kontakt osoba (ime, telefon, imejl).
  - **DET2 – Način otpreme, prevoznik i vozač:**
    - način otpreme: Sopstveni prevoz, Prevoznik, Prevoz primaoca, Lično preuzimanje, Lična dostava;
    - for Lično preuzimanje / Lična dostava: ime, prezime, broj lične karte;
    - for a prevoz (1–3): mesto utovara, mesto isporuke, the prevoznik (1: the firm; 2: konto + šifra of the kontni
      plan with their searches; 3: the primalac) and the vozač (ime, prezime, imejl, telefon, registarski broj vozila,
      broj vozačke dozvole);
    - faze vožnje: način otpreme and naziv / ime i prezime, "Dodaj novu fazu vožnje", "Briši" per row.
  - **DET3 – Adresa otpreme i adresa odredišta:** ulica, broj, mesto, poštanski broj, šifra objekta of each; the adresa
    odredišta comes from the mesto isporuke of the primalac (its search), with a napomena.
  - **DET4 – Stavke:** the stavke control of the otpremnica, unchanged.
  - **DET5 – Opasne materije, roba treće strane i napomena:** opasne materije (naziv, opis / vrednost, komentar; "Dodaj",
    "Briši"), roba treće strane, the napomena of the otpremnica (`edok.napomena`), the notes "*" / "**".
- A new eOtpremnica: tip Eksterna, adresa otpreme from the magacin; choosing the nalog fills the primalac and the adresa
  odredišta of the nalog's mesto isporuke.
- The controls are saved with the header ("Sačuvaj" sends them; one transaction with the otpremnica), lock with it and
  open again with "Izmeni". The faze vožnje and opasne materije are written again on every save; the opasne materije
  only with "Opasne materije", the roba treće strane only with its checkbox, the lično podaci and the prevoznik only for
  their način otpreme.
- Validation (besides the otpremnica's): tip, jedinice mere and način otpreme from their lists, valid datumi i vremena,
  an entered prevoznik must exist in the kontni plan, a faza needs način and naziv, an opasna materija its naziv.
- Deleting an eOtpremnica also deletes its faze vožnje and opasne materije. Deleting the same document from "Otpremanje
  robe" leaves them in `enot` / `eopma` (they belong to no document any more; the otpremnica screen does not depend on
  migration 003).

**Where the data is stored** (`migrations/003_eotpremnica.sql`; to be mirrored in the WinDev analysis):

| Control | Field | Column |
|---------|-------|--------|
| DET1 | Opasne materije, Roba treće strane | `edok.opasnemat`, `edok.roba3str` |
| DET1 | Tip otpremnice | `edok.tipeotp` (Ext / Int) |
| DET1 | Jedinica mere + težina, jedinica mere + zapremina | `edok.jm1` + `bttowght`, `edok.jm2` + `bttovol` |
| DET1 | Stvarni / planirani početak / planirani završetak | `edok.stvdatotp`, `datpoc`, `datzav` |
| DET1 | Broj paketa, Povrat robe | `edok.totboxs`, `edok.povratrobe` |
| DET1 | Broj ugovora, narudžbenice/ponude, okvirnog sporazuma | `edok.brugov`, `brnarudz`, `brokvspor` |
| DET1 | Kontakt pošiljaoca: ime, telefon, imejl | `edok.kontime`, `konttel`, `kontmail` |
| DET1 | Kontakt primaoca: ime, telefon, imejl | `edok.referkup`, `telkup`, `mailkup` |
| DET2 | Način otpreme | `edok.nacotp` (1–5) |
| DET2 | Lično: ime, prezime, broj lične karte | `edok.licime`, `licprez`, `licbrlk` |
| DET2 | Mesto utovara, mesto isporuke | `edok.mestoutov`, `mestoisp` |
| DET2 | Prevoznik konto, šifra | `edok.prevkto`, `prevana` |
| DET2 | Vozač: ime, prezime, imejl, telefon, reg. broj, vozačka | `edok.vozac`, `vozprez`, `vozmail`, `voztel`, `brvozila`, `vozdozv` |
| DET2 | Faze vožnje | table `enot` (rdokid, rbr, nacotp, naziv; fkto / fana reserved) |
| DET3 | Adresa otpreme | `edok.ulicaotp`, `ulbrotp`, `mestootp`, `posbrotp`, `sifobjotp` |
| DET3 | Mesto isporuke, adresa odredišta, napomena | `edok.mi`, `ulicaodr`, `ulbrodr`, `mestoodr`, `posbrodr`, `sifobjodr`, `fisnap` |
| DET4 | Akcizni podaci of a stavka (not used yet) | `epro.gtin`, `akckat`, `tippak`, `grligu`, `sifrom` |
| DET5 | Opasne materije | table `eopma` (rdokid, rbr, naziv, opis, komentar) |
| DET5 | Roba treće strane, napomena | `edok.roba3strtxt`, `edok.napomena` |

The columns of the analysis DDL (16.10.2026) are `NOT NULL` without a default; 003 gives every new column a default and
leaves the timestamps nullable, and sets the default `''` of `jm1` / `jm2`, so the inserts of the other documents (and
of the legacy windows) that don't write these columns keep working.

**Differences from the legacy:**

| Legacy | New |
|--------|-----|
| The internal window's fields are never saved (no code) | Saved in `edok`, `enot`, `eopma` (migration 003) |
| Vozač, broj vozila, broj ugovora, težina, broj kutija, napomena in the header | In DET1 / DET2 / DET5 (same columns); the način otpreme text of the nalog is kept hidden |
| Primalac konto / šifra searchable (and for "Interna" the konto `rvr.fprod`) | The kupac of the otpremnica, read-only; the tip is only saved |
| Grids edited inline | A line with "Dodaj" (checks the required fields) and "Briši" per row |
| "*" fields "obavezna pri slanju" | Marked only; no check (sending is not ported) |
| The window grows with the layout (`UpdateLayout`) | The dialog scrolls; the controls are collapsible |

**Not ported:** the stavke entry of DET4 (Dodaj with GTIN, akcizna kategorija, tip pakovanja, gramaža, šifra robne
marke: the procedures `BTN_DODAJDET`, `Validacija`, `CreateRpro`, `EditRpro` will be sent later; the `epro` columns
already exist); the "eOtpremnica" button (sending to SEF via `rob_ElektronskeFakture`); the zbirni dokument; the "+"
buttons of the šifarnici.

**Files:** `model_komercijalno_eotpremnica.go`, `komercijalno_eotpremnica_repository.go` (+ golden test),
`eotpremnica_service.go` (+ test), `eotpremnica_handler.go` (+ test), `eotpremnica.templ`, `eotpremnica_script.templ`.

---

## Shared code changed along the way

- `KomPregled` (`pregled.templ`): optional tabs and an optional status filter; the vrsta filter hidden when a screen has
  no vrste.
- The profaktura dialog takes a variant (`ProVarijanta`) so the nalog za isporuku reuses it; the profaktura repository
  and service are parameterized by the group (`komProVarijanta`).
- Shared table (`components/table.templ`): a normal row can carry CSS classes (`common.ClassRowRazlika` for the red rows).
- `DialogOk`: keeps line breaks in its message.
- Stanje allow-list: `kolicner` and `rez`; `Izmena.NeNegativno` keeps a sum at 0 or above.
- Reservation helpers `komRezervisi` / `komRezervisiLot` shared by the profaktura, the nalog and the otpremnica.
- The otpremnica handler is a variant: its urls come from `otpPutanje` (`otpPutanjeOtp`, `eotpPutanje`), and
  `EotpremnicaHandler` registers the same actions under its own routes. The service has hooks: `saveOtpremnica` (what
  the screen saves in the header's transaction) and `deleteOtpremnica` (what it deletes first); `KomOtpDokument.Eotp`
  switches the dialog to the eOtpremnica controls; `komOtpNalog` calls `komEotpNalog` when it exists.
- Searches: `zahtevi-nabavka`, `artikli-proizvodnja`, `zahtevi-ponuda`, `dobavljaci-ponude`, `porudzbenice-potvrdjene`,
  `artikli-sirovine`, `artikli-zalihe`, `cenovnici`, `profakture-potvrdjene`, `nalozi-potvrdjeni`.
- `.claude/rules/database.md`: `klista`, `rcenovnikhdr` and `klasacen` added to the tables without a period.

## Notes on the data (at the time of the port)

- `pinorm` is empty: in a tip G magacin the profaktura and the nalog refuse stavke without a planska cena.
- `rcenovnikhdr` is empty: the cenovnik prices (tipcene 2) can't be tried.
- No otpremnice, ULS prijemnice or avansi linked to a nalog exist yet: those paths are checked by tests only.
- Vrsta 199 ("STORNO KALKULACIJA ULAZA SIROVINE") has predznak "+": confirming it books a normal ulaz.
- Migration 003 is not applied yet: until it runs, the eOtpremnica screen fails and its 8 statements fail the read-only
  `TestRealDBExplain` (all other statements pass).
- None of the screens has been clicked through in a browser by the porting work; the HTML is covered by render tests.
