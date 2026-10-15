-- Create the appo table: the plaćen avans of a kalkulacija veleprodaje (the legacy APPO of the Baza.wda
-- analysis, the grid "Plaćen avans" of rob_Kalkulacije_Veleprodaja): per row the osnovica and the porez of an
-- avansni račun of the dobavljač, its poreska oznaka (po) with its stopa (pdv) and rpor row, the broj (rac) and
-- the datum (dadok) of the avansni račun, the document it belongs to (rdokid) and its vrsta and broj.
-- The tables of the legacy analysis live in the schema baza with lower case names (like rppo, its pair for the
-- plaćen porez).

CREATE TABLE IF NOT EXISTS baza.appo (
    appoid BIGSERIAL PRIMARY KEY,
    god INTEGER DEFAULT 0,
    kar INTEGER DEFAULT 0,
    vrd INTEGER DEFAULT 0,
    dokum INTEGER DEFAULT 0,
    po SMALLINT DEFAULT 0,
    ppor FLOAT8 DEFAULT 0,
    konto VARCHAR(6),
    sifra VARCHAR(6),
    dadok DATE,
    rac VARCHAR(16),
    osn FLOAT8 DEFAULT 0,
    pdv FLOAT8 DEFAULT 0,
    rdokid INTEGER DEFAULT 0,
    rporid INTEGER DEFAULT 0,
    xopunos VARCHAR(30),
    xopizmene VARCHAR(30),
    xdatunosa TIMESTAMP,
    xdatizmene TIMESTAMP
);

-- The indexes of the legacy analysis.
CREATE INDEX IF NOT EXISTS idx_appo_god ON baza.appo (god);
CREATE INDEX IF NOT EXISTS idx_appo_kar ON baza.appo (kar);
CREATE INDEX IF NOT EXISTS idx_appo_vrd ON baza.appo (vrd);
CREATE INDEX IF NOT EXISTS idx_appo_dokum ON baza.appo (dokum);
CREATE INDEX IF NOT EXISTS idx_appo_rdokid ON baza.appo (rdokid);
CREATE INDEX IF NOT EXISTS idx_appo_rporid ON baza.appo (rporid);
CREATE INDEX IF NOT EXISTS idx_appo_godkarvrd ON baza.appo (god, kar, vrd);
CREATE INDEX IF NOT EXISTS idx_appo_godkarvrddokum ON baza.appo (god, kar, vrd, dokum);

COMMENT ON TABLE baza.appo IS 'Plaćen avans of a kalkulacija veleprodaje (legacy APPO)';
COMMENT ON COLUMN baza.appo.osn IS 'Osnovica of the avans';
COMMENT ON COLUMN baza.appo.ppor IS 'Porez of the avans';
COMMENT ON COLUMN baza.appo.po IS 'Poreska oznaka (tarifa)';
COMMENT ON COLUMN baza.appo.pdv IS 'Stopa of the poreska oznaka on the datum';
COMMENT ON COLUMN baza.appo.rac IS 'Broj of the avansni račun';
COMMENT ON COLUMN baza.appo.dadok IS 'Datum of the avansni račun';
COMMENT ON COLUMN baza.appo.rdokid IS 'The kalkulacija (rdok) the avans belongs to';

-- Rollback:
-- DROP TABLE IF EXISTS baza.appo;
