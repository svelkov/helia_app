# Robno: the new document screens and their differences from the legacy program

This file summarizes the robno documents ported from the legacy WinDev program (the windows that "Snimi nalog" opens):
what each new screen does and where it deliberately differs from the legacy. The companion file for the komercijalno
module is `docs/komercijalno-dokumenti.md`.

The code is in `internal/domain/model_robno*.go`, `internal/repository/robno_*_repository.go`,
`internal/service/robno/robnadokumenta_*_service.go`, `internal/handler/robno/rob*_handler.go` and
`frontend/templates/robno/rob*.templ`.

## How the screens open

In "Unos dokumenta" the user saves a robni nalog (magacin, vrsta naloga, vrsta dokumenta). As in the legacy SWITCH of
"Snimi nalog", the group of the vrsta dokumenta (`dokvrsta.grpdok`) chooses the screen, and the screen opens as a dialog
over the tab. The registry is `robnoDokumentaUnosEkrani` in `internal/handler/robno/robnadokumentaunos_handler.go`. A
group without a screen shows "Program za ovu opciju nije instaliran (legacy window)".

| Group(s) | Legacy window | New screen | Routes |
|----------|---------------|------------|--------|
| FAK, FRP, FZR | rob_Fakture | Fakture veleprodaje | `/api/robno-dokumenta/fakture/...` |
| PRE, FUR, FCO, RCO, KPF | rob_Fakture_Usluga | Fakture usluga | `/api/robno-dokumenta/usluge/...` |
| ARA, ARU | rob_AvansniRacuni | Avansni računi (variant of Fakture usluga) | `/api/robno-dokumenta/usluge/...` |
| IRT, IRP | rob_Interni_racuni | Interni prenosi – dostavnice | `/api/robno-dokumenta/interni/...` |
| PRO | rob_ProFakture | Profakture | `/api/robno-dokumenta/profakture/...` |
| POR | rob_Porudzbenica | Porudžbenice | `/api/robno-dokumenta/porudzbenice/...` |
| KNO, KNZ | rob_KnjiznoPismoFIN | Knjižna pisma | `/api/robno-dokumenta/knjizna-pisma/...` |
| OPD, POP, POT | rob_opsti_dokumenti_opd | Opšti dokumenti | `/api/robno-dokumenta/opsti/...` |
| FIN | rob_opsti_dokumenti_FIN | Opšti dokumenti (FIN variant) | same |
| KOL | rob_opsti_dokumenti_KOL | Opšti dokumenti (KOL variant) | same |
| MNV | rob_opsti_dokumenti_MNV | Opšti dokumenti (MNV variant) | same |
| NIV | rob_Nivelacije_Veleprodaja | Nivelacije veleprodaje (opšti variant) | same |
| PRI, PRU | rob_Prenosnice | Prenosnice (opšti variant) | same |
| GRD | rob_ZadauzGRADILISTA | Zaduženja gradilišta (opšti variant) | same |
| OTR | rob_Trebovanja | Trebovanja (opšti variant) | same |
| PTG | rob_opsti_dokPOPIS | Popis u toku godine (opšti variant) | same |
| SIV | rob_ZaduzSitanInventar | Sitan inventar (opšti variant) | same |
| MNI | rob_Nivelacija_maloprodaje | Nivelacije maloprodaje | `/api/robno-dokumenta/nivelacije-mp/...` |
| PLT | rob_Kalkulacije_Veleprodaja | Kalkulacije veleprodaje | `/api/robno-dokumenta/kalkulacije-vp/...` |
| KAL | rob_Kalkulacije_Maloprodaja | Kalkulacije maloprodaje | `/api/robno-dokumenta/kalkulacije-mp/...` |

**Not built yet** (registered without a screen): KPK (knjižna odobrenja and zaduženja of rob_Fakture), DIR
(rob_Fakture_maloprodaje), IRS (interni računi of the proizvodnja). Two checks of the legacy SWITCH are not in the
registry yet: `fvr.etunosprod = 2` (FAK in stages, pr_FaktureOtp) and IRS only from a magacin of tip P or L.

Unlike komercijalno, the robno documents are stored in `rdok` / `rpro` and post to the nalog (`rnal`), the stanje
(`rsta`) and the zalihe by lot (`drsta`) as each stavka is saved.

---

## Common to all screens

**Two layouts:**

- **Collapsible panels** (the faktura pattern: fakture, usluge, avansni, interni, profakture, porudžbenice, knjižna
  pisma, nivelacije MP, both kalkulacije). The header panel is open until "Sačuvaj"; then it locks and the stavke panel
  opens. "Izmeni" in the header title opens the header again. A second tab ("Prikaz …" / "Pregled …") shows the
  documents of the vrsta and the stavke of the selected one.
- **One window with two parts** (the opšti dokumenti and their variants). The header is enabled until "Sačuvaj" /
  "Snimi dok."; then the stavke are enabled. "Izmeni" switches back. Same "Pregled" tab.

**Shared procedures** (ported once, used by every screen):

| Legacy | Go |
|--------|----|
| SnimiRDOK | `SaveRobniDokument` (insert or update of `rdok`, nalog count, stavke datum, avansi, extra writes in the same transaction) |
| Brisanje of a document | `robBrisiDokument` / `robBrisiDokumentSa`: refuses a fiscalized (`knjige[4] = "S"`) or posted (`knjige[1] = "D"`) document |
| AzurirajRNAL, AzurirajRSTA, AzurirajDRSTA | `AzurirajRNAL`, `AzurirajRSTA`, `AzurirajDRSTA`, `robAzurirajStanjeStavke` |
| AZurirajStanja, BrisiStanja | `AzurirajStanja`, `BrisiStanja` |
| AzurRSTA_ZN, AzurDRSTA_ZN | `robnadokumenta_stanjazn_service.go` (used by the komercijalno prijemnica) |
| IzracunajNCENA | `IzracunajNCENA` (prosečna nabavna cena) |
| ObracunFAKT, PreracunRATA | `ObracunFakture`, `PreracunRataDokumenta` |
| NoviBrojRDOK, exit of EDT_DOKUM | `GetNoviBrojDokumenta`, `GetRobFaktureDokument` |

**Left out on every screen** (your general rules):

- The SR counter (no `sr` table) and the RSAL / RSIN updates.
- The Da/Ne questions that only warn: "Unos bez OTK broja?" (except on the kalkulacije), "Oslobođeno PDV-a?", "datum
  outside the poslovna godina", "izvorni dokument of another kupac / already used". A question that stops the save
  becomes a field error, e.g. "Nema dovoljno zaliha.".
- The Republika Srpska behavior (JIB, `gnDRZAVA`): only the Serbian one; the kupac shows its PIB.
- `UpdUNOS` / `UpdIZMENE` (the repository fills the audit columns) and `HErrorLock` (the lock middleware locks the
  nalog).
- `BrisiRNAL` / `ObrisiRNAL` on close (deleting an empty nalog or an empty document): the screens only release the lock.
- The "+" / "Nova šifra" buttons that open other šifarnici (partneri, konta, komercijalisti, artikli).
- `profak` (the profaktura link) on the screens where you said "skip profak" (fakture, interni).

**Other common changes:**

- Searches are TableCombos (no search buttons); choosing a kupac / dobavljač fills naziv, adresa, PTT, mesto, PIB.
- OJ / mesto troška: the mesto troška combo lists only the chosen OJ's mesta; a mesto troška must belong to its OJ.
- Number fields are `TextInputField` with `FieldType: "number"` (količina 3 decimals, the rest 2).
- A stavka is changed with "Izmeni" on its row, not by typing in the grid.
- Inserts write the NOT NULL columns the legacy didn't know (`rdok.epp_polje`, `rpro.taxcat`, `rpro.taxexreco`).

---

## 1. Fakture veleprodaje (rob_Fakture, FAK / FRP / FZR)

**What it does:** header with kupac, mesto isporuke, OJ / mesto troška, komercijalista, valuta, uslovi prodaje, broj
rata, poreska kategorija with napomena, nastanak PDV obaveze, POPDV, tekst (`rdok.foot`), finansijsko stanje of the kupac
(the legacy 4×3 grid) and the avansi grid. Save: `ValidateRobFakture` → `SaveRobniDokument` → `PreracunRataDokumenta` →
`ObracunFakture` (iznos, kasa skonto, ugovoreni rabat, closing of the avansi). Stavke as `rob_FAKTURE_DET` (cena, rabat,
neto cena, zadnja cena of the kupac, tax category / exemption). Tab "Prikaz faktura". Print as before.

**Differences from the legacy:**

| Legacy | New |
|--------|-----|
| Search buttons and popup grids | TableCombos (broj dokumenta `robni-dokumenti`, konto, analitika, artikli) |
| Avansi grid edited inline, with "Del – Brisanje stavke" | "+ Dodaj avans" and "Obriši" per row; the closed amount is computed by the obračun (read-only) |
| Komercijalista checked by šifra (`sifkom`) | Checked by `komid` (the combo sends the key) |
| Nastanak PDV saved by the window | Saved in `rdok.descriptioncode` (the e-Faktura tax point codes 35, 3, 432, 0; 999 = "-") |
| Broj dokumenta not checked against other nalozi | Must not be used by another nalog for the same magacin and vrsta |
| Choosing an existing broj marks the header saved | Only fills the header; "Sačuvaj" updates by `rdokid` |
| Za period of KPF: Od always disabled | Od follows the checkbox |

---

## 2. Fakture usluga (rob_Fakture_Usluga + DET2: PRE, FUR, FCO, RCO, KPF)

**What it does:** the faktura header without paritet, izvozna izjava, banka – izvoz, težine, kutije and palete; "Za
period" only for KPF, "Datum smanjenja" only for a KPF storno. Stavke: šifra 0 = free text (opis usluge, JM, tarifa typed;
KPF has šifra 0 fixed); cene from `rsta` (vpcena, else `rcene.cena`); for vrsta 280 the last cena to the kupac; "Dod. opis
usluge" saved per document in `dokusldet` (old RTF texts shown as plain text). An obračun variant computes the PDV an
avans closes from the avans's own stopa.

**Differences from the legacy:** the stavke query left-joins `rsif`, so šifra-0 stavke appear in every grid (before
they were dropped). Kept as in the legacy although surprising: deleting a stavka also deletes the document's Dod. opis
usluge and doesn't lower `rdok.brst`; Datum smanjenja is validated but not saved (`datpro` is skipped).

**Not ported:** the "Fiskalni" button (it writes files to `C:\MIKROELEKTRONIKA`), the rate popup.

---

## 3. Avansni računi (rob_AvansniRacuni, ARA / ARU)

**What it does:** opens on the Fakture usluga screen. Header without valuta, rate, rok, narudžbenica, mesto isporuke,
vozač, javna nabavka and avansi grid; profaktura shown; uslovi plaćanja "AVANS"; nastanak PDV "-", "Datum avansa", "Ne
nastaje". Extra checks: broj not 0, the partner has the analitička konta of `fvknjrac` I/3. A new document gets its
`avansfakt` row. Stavke under artikal šifra 0: opis, količina 1, tarifa with stopa, poreska kategorija with osnov
oslobođenja, cena, iznos; the osnovice and PDV are added to the avans by stopa (20 → 1, 10 → 2, 0 → 3).

**Differences from the legacy:**

| Legacy | New |
|--------|-----|
| `avansfakt.rdokid` = 0 for "not closed" | NULL (the column now has a foreign key) |
| Changing a stavka applies the difference at the new tarifa | Takes the old amount back by its old tarifa first |
| Deleting a stavka keeps `rdok.brst` | Lowers it |
| Deleting a document leaves its `avansfakt` row | Removes it; refused when the avans is closed by a faktura or was sent to SEF |
| Faktura-only checks (kredit limit, budžetski ugovor) | Skipped |

**Data:** the šifra-0 artikal must exist in `rsif` of the period; the legacy `KreirajRSIF` can't be copied because
`rsif` needs five NOT NULL foreign keys (boje, debljine, oblici, metali, modmetala) whose tables are empty.

---

## 4. Interni prenosi – dostavnice (rob_Interni_racuni + rob_InterniRacuni_DET, IRT / IRP)

**What it does:** header with prodavnica (konto default `rvr.fprod`, šifra, naziv, adresa, mesto), kupac (IRT), OJ /
mesto troška, uslovi prodaje. Stavke: artikal by šifra or barkod (> 6 characters); for IRT the cene from `rsta` (nabavna
cena magacin: prosnc, rabat 0; else cena, rabat from vma, nabavna, marža, MP cena, MP sa porezom); model "D" takes mcenap;
zadnja cena of the prodavnica; recalculation of the vrednosti, nabavna, marža and MP cena as in the legacy. Save as
CreateRPRO / UpdateRPRO (rcene added when missing); delete renumbers. The konto entered stays for the next stavka.

**Differences from the legacy:** the broj dokumenta is checked against other nalozi (the legacy didn't); a changed
document also updates the prodavnica of its stavke (`rpro.pkto` / `pana`); the stanje check gives a field error
"Nema dovoljno zaliha." (a changed stavka counts its own old količina as stanje).

---

## 5. Profakture (rob_ProFakture + rob_PROFAKTURE_DET, PRO)

**What it does:** header like the faktura (dokument, napomena / broj rata, valuta, uslovi prodaje, kupac, mesto
isporuke, OJ / mesto troška, komercijalista, izvoz group only for a foreign valuta). Saved through `SaveRobniDokument`
without avansi, rate or obračun (as the legacy). Stavke: cena from `rsta.cena`, zadnja cena of the kupac and "Koristi
poslednji rabat", foreign cena converted by the kurs, šifra 0 with typed naziv / JM. Posting to the nalog and the stanje.
Print with the Profaktura / Ponuda radio.

**Decisions taken:** `rpro.ncena` saved as 0 (as the legacy, whose `xNabCena` is never set); no "Oslobođeno PDV-a?";
izvoz group hidden for the domestic valuta; datum of a new document today (rejected if after the datum naloga); deleting
a stavka renumbers only `rbr`.

**Not ported:** dospelo / saldo of the kupac (not on the screenshot). Kurs and važi za are typed by hand.

---

## 6. Porudžbenice (rob_Porudzbenica + rob_PORUDZBENICE_DET, POR)

**What it does:** a robni dokument of a dobavljač (`rdok.fkto` / `fana`) with stavke at the nabavna cena. New document:
today's datum, uslovi plaćanja "VIRMANOM", opis "Porudzbenica br: N". Validation: datum, dobavljač required and exists,
a given OJ and komercijalista exist, broj and rok whole numbers. Stavke: artikal by šifra or barkod, naziv, JM, tarifa
with stopa, stanje and nabavna cena (`rsta.ncena`, else `rcene.ncena`); iznos = količina × nabavna cena (negative for a
storno); posting with AzurirajRNAL and AzurirajRSTA. Delete renumbers rbr and brst.

**As in the legacy:** the zalihe by lot (`drsta`) are not updated.

---

## 7. Knjižna pisma (rob_KnjiznoPismoFIN, KNO / KNZ)

**What it does:**

- **Without a vezni dokument:** only the amounts (osnovica, % poreza, porez, iznos); no stavke.
- **With a vezni dokument** (a faktura of groups FAK / FRP / FZR of the magacin and year, filtered by kupac): the kupac,
  OJ, mesto troška, rabat, kasa and valuta are copied and the kupac locked. The stavke panel has the faktura's stavke
  ("Dodaj", "Dodaj sve") and the korigovane stavke (`rproid1` → the original; "Izmeni", "Obriši", "Vrati sve"); after
  every change the document is recalculated (`Preracun`) and the change posted to the nalog.
- "Pregl. Vr. Fakt.": new values, faktura values and razlika. Tip knjižnog pisma (`tipknpisma`). Print as before.

**Differences from the legacy:**

| Legacy | New |
|--------|-----|
| Changing the rabat in the faktura grid writes into the issued faktura's stavka | The issued faktura is never changed |
| "Dodaj" refuses a stavka corrected in any pismo, "Dodaj sve" checks nothing | Both skip stavke already in this pismo |
| "Vrati jedan" deletes the last row with an empty confirmation | "Obriši" on the chosen row with a confirmation and renumbering; "Vrati sve" with a confirmation |
| Iznos = osnovica × (1 + porez / 100) (amount mixed with percentage) | Osnovica + porez, porez = osnovica × % / 100 |
| Delete takes back the sum of `rpro.iznos` | Takes back the document's iznos (what `Preracun` posted) |
| Copied stavke have no kupac | Get the pismo's `fkto` / `fana` |
| "Pregled vrednosti fakture" matches rows by position | By stopa |
| Delete updates the stanje (commented out in the legacy) | Only the nalog, as the legacy effectively does |

New: the next-document insert adds a default column only when the screen didn't send it (fixed the duplicate `bi`
column of the profaktura insert).

---

## 8. Opšti dokumenti (rob_opsti_dokumenti_opd, OPD / POP / POT)

**What it does:** header with vrsta, broj, datum, izvorni dokument, opis, protivstav konto and šifra, OJ / mesto troška
and OJ / mesto troška gde ide roba (a new document gets the first OJ with its first mesto troška). Stavke: artikal with
the prosečna cena for vrste whose opis contains "PROS", nabavna-cena magacini, lot stanje and tip artikla check for tip
zaliha 2; cena read-only in a nabavna-cena magacin unless the nalog opis contains "servis". Save (CreateRpro): rpro with
konprosnc, rcene if missing, RNAL, RSTA, DRSTA; a duguje document also updates the prosečna nabavna cena. Delete
(BrisiRPRO): the reverse, an empty DRSTA lot removed, `IzracunajNCENA` for a duguje document. No change of a stavka (the
legacy Izmene is commented out).

**Differences from the legacy:**

- OJ / mesto troška optional (empty or "-" saved as NULL); a given value must exist and the mesto troška belong to the OJ.
- A new document (also "Novi dok.") starts with the datum naloga; the legacy uses today for "Novi dok.".
- The nalog stavka count after a delete is recounted after the row is deleted (it was one too high).
- `IzracunajNCENA` after a delete ran after the commit, not in the same transaction (at the time of the port).
- No lot popup for tip zaliha 2: OTK, serija and rok are typed.
- Empty document not deleted on close.

### FIN (rob_opsti_dokumenti_FIN)
Header: vrsta, broj, datum, izvorni dokument, opis, OJ, mesto troška (no protivstav, no second OJ / MT). Stavke: konto,
šifra, stanje, iznos, JM; saved with količina and cena 0, iznos negative for a storno; RNAL and RSTA only (no DRSTA or
rcene, as the legacy); delete without the DRSTA cleanup or the nabavna cena recalculation.

### KOL (rob_opsti_dokumenti_KOL)
Header as FIN. Stavke: konto, šifra with stanje, količina, JM, and OTK / serija / rok for tip zaliha 2; saved with cena
and iznos 0; RSTA and DRSTA by lot for tip zaliha 2. No stanje check (commented out in the legacy); delete without
`IzracunajNCENA`.

### MNV (rob_opsti_dokumenti_MNV)
Header: vrsta, broj, datum, izvorni dokument, opis. Stavke: naziv (free text), količina, stara cena, nova cena, efekat
nivelacije; saved under šifra 0 without konto; posted with `AzurirajStanja` / `BrisiStanja`.

- **Efekat:** količina × nova − količina × stara on every change (the legacy set iznos = količina × cena on the exit of
  the količina).
- **Redni broj:** highest rbr + 1 (the legacy used `RDOK.RBR + 1`, which never grows).
- **Print:** designed here (RPT_MNVSTAMPA layout was not available).
- No MNV vrsta exists in the data, so the screen was never opened from a nalog.

### NIV – Nivelacije veleprodaje (rob_Nivelacije_Veleprodaja)
Stavke: artikal with količina and stara cena from the stanje of the konto (tip zaliha 2: the lot total, only when it
equals the stanje, else "Proverite ukupne i pojedinačne zalihe…"); nova cena / procenat / razlika. Save: one stavka per
lot for tip zaliha 2; iznos = količina × (nova − stara); posted to RNAL, RSTA, DRSTA; every lot gets the nova cena.
Delete with the legacy warning, cene stay.

- **Header save blocked** for a magacin not kept at the prodajna cena ("Zalihe se ne vode po prodajnoj ceni!"); the
  legacy showed the message and continued.
- The save uses the konto and cene of the form (the legacy read variables filled on exit; same result).
- The Pregled tab uses the shared stavke columns (not TABLE_RPRO1).

### PRI / PRU – Prenosnice (rob_Prenosnice)
Header: magacin gde ide roba (default: the first one other than the source), konto zaliha gde ide roba, OJ / mesto
troška of both sides; "Naziv artikla u štampi" and "Komercijalni opis u štampi". Save creates or updates the kontra ulaz
document (vrsta 113) in the nalog of the destination magacin (created if missing), linked both ways by `vrdokid`; each
stavka gets its kontra stavka (`rproid1`), the destination's prosečna nabavna cena is updated; a sirovine magacin swaps
cena and nabavna cena. Vrsta 113 opens on the Pregled tab only.

| Legacy | New |
|--------|-----|
| Default destination: the first magacin, maybe the source | The first other than the source |
| A changed destination konto is used only until the window closes | Stored on the kontra document |
| Kontra stavke not renumbered after a delete (uninitialised `nPOMID`) | Renumbered |
| "Also print the kontra document?" | Not asked; print it from vrsta 113 |
| Prenosnica print joined the destination by `mag` | Joined by `magaciniid` (`rdok.magid1`) |

OJ / mesto troška are required here (as the legacy), unlike the other opšti documents.

### GRD – Zaduženja gradilišta (rob_ZadauzGRADILISTA)
Header: OJ / mesto troška, konto and šifra gradilišta (must exist in the kontni plan). A saved document changes only
the izvorni dokument, opis and gradilište (as SnimiRDOK). Stavke: konto, šifra, stanje, količina, cena (from the last
prenosnica ulaz, `NadjiZadnjuVPC`; 0 without one), stock check. Differences: the save adds 1 to the document's broj
stavki (the legacy only subtracted on delete); the delete also reverses `drsta`; OJ / mesto troška optional.

### OTR – Trebovanja (rob_Trebovanja)
GRD variant with konto / šifra troška; all header fields change on a saved document. No cena and no stock check
(commented out in the legacy). Differences: OTK / serija / rok are stored on the stavka (the legacy showed but didn't save
them), without posting lot stock; broj stavki counted on save. **No print yet** (RPT_ROB_TREBOVANJA not received). No OTR
vrsta exists in the data.

### PTG – Popis u toku godine (rob_opsti_dokPOPIS)
Header with konto and šifra radnika (`pkto` / `pana`). Stavke: artikal from rsif / rcene, stanje from `drsta` (tip zaliha 2)
or `rsta`; saving updates the nalog and the document's prodvred, nabvred, vporez, iznos and brst, not the stanje.
Differences: the delete also reverses the nalog and document totals (the legacy left them, so they drifted); OJ / mesto
troška optional; tip zaliha 0 treated as 1.

### SIV – Sitan inventar (rob_ZaduzSitanInventar)
Header with both OJ / mesto troška pairs (all four required, as the legacy) and konto / šifra radnika; a saved document
keeps its radnik (`ModifyRDOK`). Stavke: cene from the lot (tip zaliha 2, with stanje, not storno) or the stanje; a
nabavna cena of 0 from `rcene`; iznos = količina × cena, or × nabavna cena in a sirovine magacin; posting to the lot only
for tip zaliha 2. Pregled tab with the legacy columns. Decisions: konto radnika starts empty (the legacy filled `1033`);
"Datum profakture" column left out; datum naloga for a new document.

---

## 9. Nivelacije maloprodaje (rob_Nivelacija_maloprodaje + _DET, MNI)

**What it does:** own screen with collapsible panels. Header: dokument, prodavnica (konto default `rvr.fprod`) saved as
both `pkto/pana` and `fkto/fana`, OJ / mesto troška. Stavke under šifra 0 and konto `999999`: naziv, JM, stara / nova MP
cena sa PDV, "Unesi procenat", stara / nova tarifa with "Promeni tarifu"; iznos = količina × (nova − stara).

**Differences from the legacy** (your 8 decisions):

| Legacy | New |
|--------|-----|
| Add posts to konto `""`, delete to `999999` | `999999` both ways |
| Delete takes back količina 0 | Takes back the količina |
| DRSTA: never on add, always on delete | Never (šifra 0 has no lots) |
| Renumbering after delete (query with id 0, does nothing) | Renumbers |
| Razlika / procenat mix total and per unit | Per unit; iznos rounded to 2 decimals |
| `EditRpro` exists but no button calls it | "Izmeni" on the row |
| Tarifa only ≠ 0 | Must exist in `rpor` |
| Pregled stavke: the generic grid | The entry grid's columns |
| New document dated today (then rejected after the datum naloga) | Datum naloga |

---

## 10. Kalkulacije veleprodaje (rob_Kalkulacije_Veleprodaja, PLT)

**What it does:** three collapsible panels: header, zavisni troškovi, stavke. After the header save, "Da li postoje
eksterni zavisni troškovi za ovaj dokument?" opens the zavisni troškovi (Da) or the stavke (Ne).

- **Header:** dobavljač (default `rvr.fdob`), mesto nabavke, valuta / kurs, uslovi nabavke, OJ / mesto troška (required),
  EPP with its polje, POPDV (type "U"), "PDV se može odbiti", veza porudžbenice (`rdok.edokid`), the plaćen porez
  (`rppo`) and plaćen avans (`appo`) grids (inline rows, "›" copies porez to avans, saved with the header). Duplicate
  izvorni dokument asks "Nastavak unosa?".
- **Zavisni troškovi** (`rztr` with their `rppo` rows): own Sačuvaj / Odustani, Izmeni / Obriši per row.
- **Stavke:** nabavna cena from fakturna cena, rabat, zavisni / interni trošak (-/%/d) and the document's external %
  (recomputed on the server); VP and MP marža / cena both ways; devizna fakturna cena by the kurs; "Unos bez OTK broja?"
  for tip zaliha 2. Posting to the nalog, stanje with all cene, lots, `rcene`, VP cena and prosečna nabavna cena.
- **Delete document:** warns about artikli that would go negative, then deletes stavke, `rppo`, `appo`, `rztr`, `kepu`
  and the document, and recomputes the prosečna cena.

**Differences from the legacy:**

1. A new document starts with the datum naloga, not today.
2. Editing a stavka recomputes the prosečna nabavna cena with `IzracunajNCENA` (the legacy removal formula used
   `rpro.cena`).
3. Deleting a stavka lowers `rdok.brst`.
4. `rcene`: `mcena` = MP cena without porez, `mcenap` = new MP cena, `tz` / `tzi` as "-", "%", "d" (the legacy
   `AzurirajRCENE` was not available).
5. Deleted porez / avans rows take effect with the header (or trošak) save; rows with an osnovica need an existing poreska
   oznaka.
6. The `rpor` row in force on the date (the legacy view picked the oldest).
7. The narudžbenica search lists the `edok` documents of the magacin (the legacy parameter "40" unknown).
8. The negative-zalihe warning and the delete confirmation are one dialog.

**Not ported:** the automatic nivelacija ("Stekli su se uslovi…", `KreirajNIVELACIJU`) and its print, "Nalepnice",
changing the valuta in the stavke.

**Database:** needs `migrations/002_create_appo_table.sql` (the `appo` table was missing).

---

## 11. Kalkulacije maloprodaje (rob_Kalkulacije_Maloprodaja + _Stavke, KAL)

**What it does:** header with dokument, način / rok plaćanja, tekst, napomena, dobavljač (no default konto, as the
legacy), prodavnica (`pkto` / `pana`, šifra of at least 2 characters), OJ / mesto troška, plaćen porez grid, EPP, POPDV
(only when the vrsta asks for it, default its polje), "PDV se može odbiti"; "Nastavak unosa?" for a duplicate izvorni
dokument. Stavke: šifra ≠ 0 only in a maloprodaja magacin (tipmag "M"); šifra 0 with typed naziv, JM and tarifa; posl.
cena (last MP cena from `rcene`) prefills the MP cena; nabavna = fakturna − rabat + zavisni trošak; MP cena = (nabavna +
marža) + porez and back. Delete: negative-stanje warning, then stavke, `rppo`, `appo` and the document (no `rztr` /
`kepu`, no prosečna cena recompute).

**Differences from the legacy:**

- "Izmeni" on the stavka row (the legacy window had only Sačuvaj / Briši).
- The stavka konto is the document's prodavnica.
- Zavisni trošak oznaka "-", "%", "d" (the legacy wrote "D").
- EPP polje list as on PLT (the legacy filtered by tipdok with a query not available).
- Način plaćanja saved for existing documents too (the legacy only for a new one).
- Deleting a stavka lowers `rdok.brst`; iznos rounded to 2 decimals.

---

## Open items

- Screens not built: KPK, DIR, IRS; the `fvr.etunosprod = 2` path (pr_FaktureOtp).
- Trebovanje print (RPT_ROB_TREBOVANJA).
- PLT automatic nivelacija and "Nalepnice".
- The interni računi of the proizvodnja print (IRS / PPR, `ROB_RPT_INTRAC_PROIZV`): its body code was not received.
- Šifra-0 artikal in `rsif` for every godina / kartica (avansni računi, KPF stavke, MNV / MNI stanje rows).

## Notes on testing

Most screens were tested read-only or inside rolled-back transactions on the local database, and with render and route
tests; none was clicked through end to end in a browser during the port. MNV and OTR vrste don't exist in the data, and
there were no KOL, NIV or avansni documents to open.
