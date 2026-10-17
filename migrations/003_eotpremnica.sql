-- The data of the eOtpremnica (the legacy pr_eOTPREMNICA with its internal window IW_Otpremnice): the columns of the
-- otpremnica (edok) and of its stavke (epro) the e-otpremnica sends, the faze vožnje (enot) and the opasne materije
-- (eopma) of an otpremnica.
--
-- The columns STVDATOTP, ULICAODR, ULBRODR, MESTOODR, POSBRODR, SIFOBJODR, FISNAP and REFERKUP come from the Baza.wda
-- analysis (16.10.2026). There they are NOT NULL without a default; here every new column has a default (empty
-- text, 0, false) and the timestamps are nullable, so the inserts of edok and epro that do not know them (the other
-- komercijalno documents, the legacy windows) keep working. jm1 and jm2 (already NOT NULL without a default) get the
-- default '' for the same reason. The tables of the legacy analysis live in the schema baza with lower case names.

-- The otpremnica (edok).
ALTER TABLE baza.edok
    -- DET1: the tip, the stvarni datum, povrat robe, the documents, the kontakt osobe of the pošiljalac and of the
    -- primalac (the primalac itself is fkto/fana, its kontakt ime is referkup).
    ADD COLUMN IF NOT EXISTS tipeotp VARCHAR(3) NOT NULL DEFAULT 'Ext',
    ADD COLUMN IF NOT EXISTS stvdatotp TIMESTAMP,
    ADD COLUMN IF NOT EXISTS povratrobe BOOLEAN NOT NULL DEFAULT false,
    ADD COLUMN IF NOT EXISTS brnarudz VARCHAR(50) NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS brokvspor VARCHAR(50) NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS kontime VARCHAR(50) NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS konttel VARCHAR(30) NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS kontmail VARCHAR(100) NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS referkup VARCHAR(30) NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS telkup VARCHAR(30) NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS mailkup VARCHAR(100) NOT NULL DEFAULT '',
    -- DET2: the način otpreme, the lično preuzimanje / lična dostava, the mesta, the prevoznik (fkpl) and the vozač
    -- (its ime is vozac, its registarski broj brvozila).
    ADD COLUMN IF NOT EXISTS nacotp SMALLINT NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS licime VARCHAR(50) NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS licprez VARCHAR(50) NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS licbrlk VARCHAR(20) NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS mestoutov VARCHAR(100) NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS mestoisp VARCHAR(100) NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS prevkto VARCHAR(6) NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS prevana VARCHAR(6) NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS vozprez VARCHAR(50) NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS vozmail VARCHAR(100) NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS voztel VARCHAR(30) NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS vozdozv VARCHAR(30) NOT NULL DEFAULT '',
    -- DET3: the adresa otpreme and the adresa odredišta (its mesto isporuke is mi, its napomena fisnap).
    ADD COLUMN IF NOT EXISTS ulicaotp VARCHAR(50) NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS ulbrotp VARCHAR(10) NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS mestootp VARCHAR(20) NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS posbrotp VARCHAR(10) NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS sifobjotp VARCHAR(10) NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS ulicaodr VARCHAR(50) NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS ulbrodr VARCHAR(10) NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS mestoodr VARCHAR(20) NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS posbrodr VARCHAR(10) NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS sifobjodr VARCHAR(10) NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS fisnap VARCHAR(1024) NOT NULL DEFAULT '',
    -- DET5: the text of the roba treće strane (roba3str is its checkbox).
    ADD COLUMN IF NOT EXISTS roba3strtxt VARCHAR(1024) NOT NULL DEFAULT '';

ALTER TABLE baza.edok ALTER COLUMN jm1 SET DEFAULT '';
ALTER TABLE baza.edok ALTER COLUMN jm2 SET DEFAULT '';

COMMENT ON COLUMN baza.edok.tipeotp IS 'eOtpremnica: tip otpremnice (Ext eksterna, Int interna)';
COMMENT ON COLUMN baza.edok.stvdatotp IS 'eOtpremnica: stvarni datum i vreme otpreme';
COMMENT ON COLUMN baza.edok.povratrobe IS 'eOtpremnica: povrat robe';
COMMENT ON COLUMN baza.edok.brnarudz IS 'eOtpremnica: broj narudžbenice/ponude';
COMMENT ON COLUMN baza.edok.brokvspor IS 'eOtpremnica: broj okvirnog sporazuma';
COMMENT ON COLUMN baza.edok.kontime IS 'eOtpremnica: kontakt osoba of the pošiljalac, ime';
COMMENT ON COLUMN baza.edok.referkup IS 'eOtpremnica: kontakt osoba of the primalac, ime';
COMMENT ON COLUMN baza.edok.nacotp IS 'eOtpremnica: način otpreme (1 sopstveni prevoz, 2 prevoznik, 3 prevoz primaoca, 4 lično preuzimanje, 5 lična dostava)';
COMMENT ON COLUMN baza.edok.prevkto IS 'eOtpremnica: konto of the prevoznik (način otpreme 2)';
COMMENT ON COLUMN baza.edok.prevana IS 'eOtpremnica: analitička šifra of the prevoznik (način otpreme 2)';
COMMENT ON COLUMN baza.edok.fisnap IS 'eOtpremnica: napomena of the adresa odredišta';
COMMENT ON COLUMN baza.edok.roba3strtxt IS 'eOtpremnica: roba treće strane';

-- The stavke (epro): the akcizni podaci.
ALTER TABLE baza.epro
    ADD COLUMN IF NOT EXISTS gtin VARCHAR(14) NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS akckat SMALLINT NOT NULL DEFAULT 1,
    ADD COLUMN IF NOT EXISTS tippak SMALLINT NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS grligu NUMERIC(16,3) NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS sifrom VARCHAR(30) NOT NULL DEFAULT '';

COMMENT ON COLUMN baza.epro.gtin IS 'eOtpremnica: GTIN of the artikal';
COMMENT ON COLUMN baza.epro.akckat IS 'eOtpremnica: akcizna kategorija (1 -, 2 duvan, 3 kafa, 4 alkohol, 5 naftni derivati, 6 nikotinski proizvodi)';
COMMENT ON COLUMN baza.epro.tippak IS 'eOtpremnica: tip pakovanja of the duvan (1 paleta, 2 boks, 3 masterkejs, 4 paklica)';
COMMENT ON COLUMN baza.epro.grligu IS 'eOtpremnica: gramaža / litraža / gustina';
COMMENT ON COLUMN baza.epro.sifrom IS 'eOtpremnica: šifra robne marke (duvan)';

-- The faze vožnje of an otpremnica (the legacy grid TABLE_PR_QRY_ENOT): the redni broj, the način otpreme, the
-- prevoznik (konto and šifra, when there is one) and its naziv or the ime i prezime.
CREATE TABLE IF NOT EXISTS baza.enot (
    enotid BIGSERIAL PRIMARY KEY,
    god INTEGER DEFAULT 0,
    kar INTEGER DEFAULT 0,
    rdokid INTEGER NOT NULL,
    rbr INTEGER NOT NULL DEFAULT 0,
    nacotp SMALLINT NOT NULL DEFAULT 0,
    fkto VARCHAR(6) NOT NULL DEFAULT '',
    fana VARCHAR(6) NOT NULL DEFAULT '',
    naziv VARCHAR(100) NOT NULL DEFAULT '',
    xopunos VARCHAR(30),
    xopizmene VARCHAR(30),
    xdatunosa TIMESTAMP,
    xdatizmene TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_enot_godkar ON baza.enot (god, kar);
CREATE INDEX IF NOT EXISTS idx_enot_rdokid ON baza.enot (rdokid);

COMMENT ON TABLE baza.enot IS 'eOtpremnica: faze vožnje of an otpremnica (edok)';

-- The opasne materije of an otpremnica (the legacy grid TABLE_PR_QRY_EOPMA): the naziv, the opis / vrednost and
-- the komentar.
CREATE TABLE IF NOT EXISTS baza.eopma (
    eopmaid BIGSERIAL PRIMARY KEY,
    god INTEGER DEFAULT 0,
    kar INTEGER DEFAULT 0,
    rdokid INTEGER NOT NULL,
    rbr INTEGER NOT NULL DEFAULT 0,
    naziv VARCHAR(100) NOT NULL DEFAULT '',
    opis VARCHAR(250) NOT NULL DEFAULT '',
    komentar VARCHAR(250) NOT NULL DEFAULT '',
    xopunos VARCHAR(30),
    xopizmene VARCHAR(30),
    xdatunosa TIMESTAMP,
    xdatizmene TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_eopma_godkar ON baza.eopma (god, kar);
CREATE INDEX IF NOT EXISTS idx_eopma_rdokid ON baza.eopma (rdokid);

COMMENT ON TABLE baza.eopma IS 'eOtpremnica: opasne materije of an otpremnica (edok)';

-- Rollback:
-- DROP TABLE IF EXISTS baza.eopma;
-- DROP TABLE IF EXISTS baza.enot;
-- ALTER TABLE baza.epro DROP COLUMN IF EXISTS gtin, DROP COLUMN IF EXISTS akckat, DROP COLUMN IF EXISTS tippak,
--     DROP COLUMN IF EXISTS grligu, DROP COLUMN IF EXISTS sifrom;
-- ALTER TABLE baza.edok DROP COLUMN IF EXISTS tipeotp, DROP COLUMN IF EXISTS stvdatotp, ... (every column added above);
-- ALTER TABLE baza.edok ALTER COLUMN jm1 DROP DEFAULT, ALTER COLUMN jm2 DROP DEFAULT;
